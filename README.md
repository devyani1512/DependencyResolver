# DependencyResolver

Automatic development environment bootstrap tool.

## Features

- Auto-detects Python/Node.js projects
- Identifies required services (PostgreSQL, Redis)
- Generates docker-compose.yml automatically
- Starts development environment with one command

## Installation
```bash
go build -o depsol cmd/depsol/main.go
```

## Usage
```bash
./DepedencyResolver ./examples/python-app
```

## Architecture

See `docs/architecture.md` for details.