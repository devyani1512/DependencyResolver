#!/bin/bash

# ─────────────────────────────────────────────────────────────
#  depsolver cleanup.sh
#  Removes all generated files and stops Docker services
#  Usage: ./cleanup.sh [project_path]
#         If no path given, cleans current directory
# ─────────────────────────────────────────────────────────────

PROJECT="${1:-.}"

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  depsolver — cleanup"
echo "  Project: $(realpath "$PROJECT")"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# ── 1. Stop Docker services ───────────────────────────────────
if [ -f "$PROJECT/docker-compose.yml" ]; then
  echo ""
  echo "→ Stopping Docker services..."
  if command -v docker &>/dev/null; then
    docker compose -f "$PROJECT/docker-compose.yml" down --remove-orphans 2>/dev/null \
      || docker-compose -f "$PROJECT/docker-compose.yml" down --remove-orphans 2>/dev/null \
      || echo "  ⚠  Could not stop services (Docker may not be running)"
    echo "  ✓ Docker services stopped"
  else
    echo "  ⚠  Docker not found — skipping"
  fi
else
  echo ""
  echo "→ No docker-compose.yml found — skipping Docker stop"
fi

# ── 2. Remove generated files ─────────────────────────────────
echo ""
echo "→ Removing generated files..."

FILES=(
  "docker-compose.yml"
  "depsolver.lock"
  "requirements.txt"
  "package.json"
  "package-lock.json"
  "yarn.lock"
)

for f in "${FILES[@]}"; do
  target="$PROJECT/$f"
  if [ -f "$target" ]; then
    rm "$target"
    echo "  ✓ Deleted $f"
  fi
done

# ── 3. Remove node_modules if present ────────────────────────
if [ -d "$PROJECT/node_modules" ]; then
  echo ""
  read -p "→ Found node_modules/ — delete it? [y/N] " confirm
  if [[ "$confirm" =~ ^[Yy]$ ]]; then
    rm -rf "$PROJECT/node_modules"
    echo "  ✓ Deleted node_modules/"
  else
    echo "  Skipped node_modules/"
  fi
fi

# ── 4. Remove __pycache__ and .pyc files ─────────────────────
PYCACHE_COUNT=$(find "$PROJECT" -name "__pycache__" -type d 2>/dev/null | wc -l | tr -d ' ')
PYC_COUNT=$(find "$PROJECT" -name "*.pyc" 2>/dev/null | wc -l | tr -d ' ')

if [ "$PYCACHE_COUNT" -gt 0 ] || [ "$PYC_COUNT" -gt 0 ]; then
  echo ""
  read -p "→ Found $PYCACHE_COUNT __pycache__ dirs and $PYC_COUNT .pyc files — delete? [y/N] " confirm
  if [[ "$confirm" =~ ^[Yy]$ ]]; then
    find "$PROJECT" -name "__pycache__" -type d -exec rm -rf {} + 2>/dev/null
    find "$PROJECT" -name "*.pyc" -delete 2>/dev/null
    echo "  ✓ Deleted Python cache files"
  else
    echo "  Skipped Python cache files"
  fi
fi

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  ✔  Cleanup complete. Run depsol again to regenerate."
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
