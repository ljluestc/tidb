# TiDB

## Project Overview
TiDB is an open-source, distributed SQL database that supports Hybrid Transactional/Analytical Processing (HTAP). It is compatible with MySQL and designed for horizontal scalability, strong consistency, and high availability.

## Technology Stack
- **Language**: Go (Golang)
- **Consensus**: Raft
- **Storage Engines**: TiKV (Row-based), TiFlash (Columnar)
- **Protocol**: MySQL Wire Protocol

## Build and Test Commands
### Prerequisite
Ensure you have Go installed and a working `make` environment.

### Build
```bash
make
```

### Test
```bash
make test
```
Or run specific Go tests:
```bash
go test ./...
```

## Development Conventions
- **Code Style**: Follow the standard Go conventions. The project enforces checks via `make lint`.
- **Contribution**: See `CONTRIBUTING.md` (if available) or the GitHub Developer Guide.
- **Architecture**: The project separates the SQL layer (TiDB) from the Storage logic (TiKV/PD). This repo contains the SQL layer handling parsing, optimization, and execution.

## Testing Strategy
- Unit tests are extensive and cover SQL logic, planner, and executor.
- Integration tests often require a running PD and TiKV cluster (handled by the test harness or Docker).

## Security
- Authentication matches MySQL's authentication system.
- Network communication between nodes should be secured via TLS/SSL.
