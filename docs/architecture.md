# DependencyResolver Architecture

## Overview

DependencyResolver follows a pipeline architecture:
```
Scan → Graph → Conflict → Services → Docker → Ready
```

## Modules

1. **Scanner**: Detects project type and dependencies
2. **Graph**: Models dependency relationships
3. **Conflict**: Detects unsatisfiable constraints
4. **Services**: Maps dependencies to Docker services
5. **Docker**: Generates and runs docker-compose
6. **Bootstrap**: Orchestrates the entire flow

## Design Decisions

- Go for performance and simplicity
- docker-compose for service orchestration
- No complex SAT solver (future work)