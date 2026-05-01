# depsolver — PubGrub-Style Python Dependency Resolver

A production-grade dependency conflict resolution engine in Go, implementing a
**PubGrub-inspired algorithm** — the same class of algorithm used in Dart's `pub`
package manager and modern resolvers like Poetry's.

---

## Architecture

```
resolver/
  main.go        — CLI entry point
  types.go       — Core types: Version, Constraint, VersionSet, Decision, Incompatibility
  version.go     — PEP 440 version parsing
  normalize.go   — Import alias normalization (cv2 → opencv-python, etc.)
  registry.go    — PyPI REST client with caching
  graph.go       — Dependency graph construction and tree printing
  pubgrub.go     — PubGrub conflict solver (incompatibilities, backjumping)
  solver.go      — Orchestration pipeline
  lockfile.go    — Deterministic lockfile generation
```

---

## Resolution Pipeline

```
Input requirements.txt
        ↓
Package name normalization  (cv2 → opencv-python)
        ↓
PyPI metadata fetch         (GET /pypi/<pkg>/json)
        ↓
Transitive dependency fetch (latest version's requires_dist)
        ↓
Constraint accumulation     (merge all constraints per package)
        ↓
PubGrub solver loop
  ├── chooseVersion()        choose best compatible version
  ├── propagate constraints  push new constraints to dependents
  ├── conflict detection     check all prior decisions
  ├── backjumping            skip unrelated decisions
  └── learnIncompat()        record learned incompatibilities
        ↓
Dependency graph construction
        ↓
Lockfile generation          (depsolver.lock)
```

---

## PubGrub Algorithm

PubGrub was designed by Natalie Weizenbaum (2018) to solve the limitations of
greedy dependency solvers (like pip's original resolver):

### Key concepts

**Incompatibilities**
A clause stating: "these package+version combinations cannot all be selected
simultaneously." Example:
```
flask>=3.0 AND werkzeug<2.0  →  incompatible (flask 3.x needs werkzeug 2.x+)
```

**Decision levels**
Each package assignment is tagged with a level. When a conflict is detected,
the solver can backjump directly to the level where the conflict originates,
skipping unrelated choices.

**Unit propagation**
If all terms of an incompatibility are satisfied except one, that one is
*forced*. This is the same propagation rule used in SAT solvers.

**Learned incompatibilities**
When a conflict can't be resolved at the current level, the solver derives a
new incompatibility by resolving (combining) existing ones and adds it to the
global set. This prevents the solver from making the same mistake again.

### Why better than greedy resolution?

| Property              | Greedy (pip ≤21)     | PubGrub                    |
|-----------------------|----------------------|----------------------------|
| Conflict explanation  | ✗ cryptic or none    | ✓ human-readable           |
| Backtracking          | ✗ none               | ✓ backjumps over irrelevant|
| Conflict learning     | ✗ none               | ✓ learns incompatibilities |
| Completeness          | ✗ may miss solutions | ✓ finds solution if exists |

---

## Version Constraints

Supports PEP 440 constraint operators:

| Operator | Meaning              | Example       |
|----------|----------------------|---------------|
| `==`     | Exact version        | `flask==3.0`  |
| `>=`     | Minimum version      | `werkzeug>=2` |
| `<=`     | Maximum version      | `jinja2<=3.2` |
| `>`      | Strictly greater     | `click>7`     |
| `<`      | Strictly less        | `urllib3<2`   |
| `!=`     | Exclude version      | `numpy!=1.24` |
| `~=`     | Compatible release   | `requests~=2` |

Multiple constraints are combined with AND:
```
urllib3>=1.21.1,<3
```

---

## Import Normalization

Common Python imports are automatically mapped to their PyPI names:

| Import name    | PyPI package      |
|----------------|-------------------|
| `cv2`          | `opencv-python`   |
| `PIL`          | `Pillow`          |
| `sklearn`      | `scikit-learn`    |
| `bs4`          | `beautifulsoup4`  |
| `yaml`         | `PyYAML`          |
| `psycopg2`     | `psycopg2-binary` |
| `dotenv`       | `python-dotenv`   |

---

## Running the Tool

### Prerequisites
- Go 1.21+
- Internet access (fetches from pypi.org)

### Build and run

```bash
cd resolver
go run . resolve requirements.txt
go run . resolve-verbose requirements.txt   # shows solver trace
```

### Run demos

```bash
# All three demos
go run . demo

# Individual demos
go run . demo1    # Simple dependencies: flask + numpy
go run . demo2    # Transitive: flask + psycopg2
go run . demo3    # Conflict detection: requests==2.25 vs urllib3>=2
```

### Example output

```
╔══════════════════════════════════════════════╗
║        depsolver — PubGrub Resolver          ║
╚══════════════════════════════════════════════╝

→ Fetching package metadata from PyPI...
  → Fetching flask from PyPI
  → Fetching numpy from PyPI
→ Running PubGrub solver...
  [solver] Decision [L1]: flask == 3.0.3
  [solver] Decision [L2]: werkzeug == 3.0.2
  [solver] Decision [L3]: jinja2 == 3.1.4
  [solver] Decision [L4]: click == 8.1.7
  [solver] Decision [L5]: numpy == 2.4.2

✓ Resolution complete — 5 packages resolved

Resolved packages:
──────────────────────────────────────
  click                          8.1.7
  flask                          3.0.3
  jinja2                         3.1.4
  numpy                          2.4.2
  werkzeug                       3.0.2

╔══════════════════════════════════════════════╗
║              depsolver.lock                  ║
╚══════════════════════════════════════════════╝
click==8.1.7
flask==3.0.3
jinja2==3.1.4
numpy==2.4.2
werkzeug==3.0.2
```

---

## Conflict Example (Demo 3)

Input:
```
requests==2.25.0
urllib3>=2
```

`requests 2.25` depends on `urllib3>=1.21.1,<1.27`, but we also require
`urllib3>=2`. These constraints are incompatible.

Output:
```
╔══════════════════════════════════════════════╗
║           CONFLICT EXPLANATION               ║
╚══════════════════════════════════════════════╝

Conflict #1 — Package: urllib3
  Constraints: >=1.21.1,<1.27, >=2
  Reason: no version of urllib3 satisfies constraints [>=1.21.1,<1.27, >=2]
  Incompatibility: learned from conflict on urllib3

Suggested resolution:
  → Try relaxing constraints on urllib3
    Latest available: 2.2.3
  → Use requests>=2.28 which supports urllib3>=2
```

---

## Lockfile Format

`depsolver.lock` is deterministic (alphabetically sorted):

```
# depsolver.lock
# Generated: 2025-01-15T10:30:00Z
# This file is automatically generated. Do not edit manually.

click==8.1.7
flask==3.0.3
jinja2==3.1.4
numpy==2.4.2
werkzeug==3.0.2
```

---

## References

- [PubGrub paper by Natalie Weizenbaum (2018)](https://nex3.medium.com/pubgrub-2fb6470504f)
- [PEP 440 — Python version specifiers](https://peps.python.org/pep-0440/)
- [PEP 508 — Dependency specification](https://peps.python.org/pep-0508/)
- [PyPI JSON API](https://warehouse.pypa.io/api-reference/json.html)
