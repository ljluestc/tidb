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

## AI Agent Workflow

### 1. Requirements Discovery
- **Primary Source**: `PRD.md` (Always prioritize this if present).
- **Secondary**: `requirements.txt`, `README.md`, or specific task files.
- **Goal**: Understand the full scope before writing code.

### 2. Implementation Protocol
- **Branching**: Work on a dedicated feature branch (e.g., `feat/implementation-details`).
- **Development**:
  - Analyze code structure.
  - Implement changes in `src/` or relevant directories.
  - Adhere to existing code style.
- **Verification**:
  - Run build commands (see above).
  - Run test suite (see above).
  - Ensure no regressions.

### 3. Delivery
- **Commit**: Use conventional commits (e.g., `feat: ...`, `fix: ...`).
- **PR Creation**:
  - Push branch: `git push -u origin <branch-name>`
  - Create a Pull Request against the main branch.
  - Summary: Link to `PRD.md` requirements solved.

## Task Implementation
1. **Analyze Requirements**: Refer to `README.md` for detailed feature specifications and system design.
2. **Implementation**: Modify source code in the respective directories (e.g., `src/`, `internal/`).
3. **Verification**: Run provided build and test commands (see above) to ensure correctness.
4. **Push Changes**:
   - Commit changes: `git commit -m "feat: implement <feature>"`
   - Push to remote: `git push origin <branch-name>`
