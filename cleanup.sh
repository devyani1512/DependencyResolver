#!/bin/bash

echo " Cleaning up DependencyResolver test environments..."

# Python app cleanup
echo "  → Cleaning Python app..."
rm -rf examples/python-app/venv
rm -rf examples/python-app/__pycache__
rm -f examples/python-app/requirements.txt
rm -f examples/python-app/docker-compose.yml
rm -f examples/python-app/Dockerfile
rm -f examples/python-app/.dockerignore

# Node app cleanup
echo "  → Cleaning Node app..."
rm -rf examples/node-app/node_modules
rm -f examples/node-app/package.json
rm -f examples/node-app/package-lock.json
rm -f examples/node-app/docker-compose.yml
rm -f examples/node-app/Dockerfile
rm -f examples/node-app/.dockerignore

# Docker cleanup
echo "  → Stopping Docker containers..."
cd examples/python-app 2>/dev/null && docker-compose down 2>/dev/null
cd ../..
cd examples/node-app 2>/dev/null && docker-compose down 2>/dev/null
cd ../..

echo " Cleanup complete!"
echo ""
echo " To test again:"
echo "   ./depsol run examples/python-app"
echo "   ./depsol dockerize examples/python-app"