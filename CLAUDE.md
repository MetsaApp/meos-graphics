# Claude Context for MeOS Graphics API

## Project Overview

This is a Go-based REST API server that connects to MeOS (orienteering event software) and provides competition data for graphics displays. The application can run in two modes:
- **Normal mode**: Connects to a real MeOS server
- **Simulation mode**: Generates test data for development

## Key Technical Details

### Architecture
- **Language**: Go 1.23+
- **Web Framework**: gin-gonic/gin
- **Templating**: a-h/templ for type-safe HTML templates
- **CSS Framework**: Tailwind CSS (compiled locally)
- **JavaScript**: HTMX for dynamic updates (served locally)
- **Architecture Pattern**: Clean architecture with separation of concerns
- **Concurrency**: Thread-safe state management using sync.RWMutex
- **Logging**: Dual logging to console and file

### Project Structure
```
meos-graphics/
├── cmd/meos-graphics/        # Application entry point
├── internal/                 # Private packages (not importable externally)
│   ├── handlers/            # HTTP request handlers
│   ├── logger/              # Logging functionality  
│   ├── meos/                # MeOS server integration
│   ├── middleware/          # HTTP middleware
│   ├── models/              # Domain models
│   ├── simulation/          # Simulation mode implementation
│   ├── state/               # Thread-safe state management
│   └── web/                 # Web UI handlers and templates
│       └── templates/       # Templ template files
├── web/                     # Web UI assets
│   ├── src/                 # Source CSS
│   └── static/              # Static files served by the app
│       ├── css/             # Compiled CSS
│       └── js/              # JavaScript files (HTMX)
├── logs/                    # Log files (auto-created)
└── Makefile                 # Build automation
```

### Running in WSL
When running in WSL, the MeOS server connection uses the Windows host IP (configured in `internal/meos/config.go`). The default is `192.168.112.1:2009`.

## Common Tasks

### Building and Running

#### Prerequisites
```bash
# Install dependencies
make deps
```

#### Development
```bash
# Build CSS and templates, then run
make run

# Run in simulation mode
make run-sim

# Development mode with hot reload (requires air)
make dev

# Watch CSS changes (in separate terminal)
make css-watch
```

#### Production Build
```bash
# Full build (templates, CSS, swagger docs, and binary)
make full-build

# Run the compiled binary
./bin/meos-graphics
```

### Testing Endpoints
```bash
# Check health
curl http://localhost:8090/health

# Get classes
curl http://localhost:8090/classes

# Get start list for class 1
curl http://localhost:8090/classes/1/startlist

# Get results for class 1
curl http://localhost:8090/classes/1/results

# Get split standings for class 1
curl http://localhost:8090/classes/1/splits
```

### Checking Logs
```bash
# View latest logs
tail -f logs/meos-graphics-$(date +%Y-%m-%d).log

# Check for errors
grep ERROR logs/meos-graphics-$(date +%Y-%m-%d).log
```

## Important Implementation Details

### MeOS Data Flow
1. MeOS adapter polls the MeOS server every second
2. XML data is parsed into internal models
3. State is updated atomically with proper locking
4. HTTP handlers read from state to serve API requests

### Simulation Mode
- Runs a 15-minute cycle that auto-restarts
- Phase 1 (0-3 min): Start lists only
- Phase 2 (3-10 min): Competitors progressively finish
- Phase 3 (10-15 min): Stable results
- All competitors have the same start time
- Times include deciseconds for realistic formatting

### Thread Safety
- All state access goes through the `state.State` struct
- Read operations use RLock()
- Write operations use Lock()
- Handler methods make defensive copies of data

### Time Handling
- MeOS uses deciseconds (1/10 second) internally
- All times are converted to Go's time.Duration/time.Time
- Display formatting preserves deciseconds (e.g., "1:23.4")

## Code Style Guidelines

When modifying this codebase:

1. **Package Organization**: Keep packages focused on a single responsibility
2. **Error Handling**: Always wrap errors with context using fmt.Errorf
3. **Logging**: Use appropriate log levels (Info, Error, Debug)
4. **Testing**: Run with simulation mode to verify changes
5. **Concurrency**: Always protect shared state with appropriate locks
6. **Commits**: Use conventional commits for automatic versioning

## Versioning and Release Process

### Conventional Commits
This project uses conventional commits for automatic semantic versioning:

- `feat:` - New features → minor version bump (0.x.0)
- `fix:` - Bug fixes → patch version bump (0.0.x)
- `feat!:` or `BREAKING CHANGE:` → major version bump (x.0.0)
- `docs:`, `style:`, `refactor:`, `test:`, `chore:` → no version bump

### Making Changes
1. Create feature branch from `main`
2. Make changes with conventional commits
3. Run `scripts/check.sh` and make it pass (see Local validation)
4. Push branch and create PR
5. After merge, release-please will handle versioning

### Local validation
GitHub Actions CI is off to save paid minutes, so nothing checks a pull request on GitHub. Run `scripts/check.sh` and make it pass before pushing.

- `scripts/check.sh` is `make check`, the steps `.github/workflows/test.yml` ran: templ generate, go build, go test, golangci-lint, and the CLI and Swagger docs are current. It runs the Go toolchain go.mod names. There is no slow suite, so there is no `--full`.
- The pre-push hook (`.githooks/pre-push`) runs `scripts/check.sh`. Claude Code sessions turn it on at start; humans turn it on once with `git config core.hooksPath .githooks`. `git push --no-verify` skips it.

### Release Automation
Releases are fully automated via GitHub Actions:

1. **release-please** monitors `main` branch
2. Creates/updates a release PR when changes detected
3. The release PR will automatically update version strings in:
   - `internal/version/version.go` - Go constant
   - `cmd/meos-graphics/main.go` - Swagger @version annotation
4. When release PR is merged:
   - Creates GitHub release with changelog
   - Builds binaries for all platforms
   - Publishes Docker images to ghcr.io
   - Tags release with semantic version

### Docker Images
- `ghcr.io/metsaapp/meos-graphics:latest`
- `ghcr.io/metsaapp/meos-graphics:vX.Y.Z`
- Multi-platform: linux/amd64, linux/arm64

## Common Issues and Solutions

### Port Already in Use
```bash
# Find and kill process on port 8090
lsof -i :8090 | grep LISTEN | awk '{print $2}' | xargs kill
```

### MeOS Connection Failed
- Check if MeOS server is running on Windows host
- Verify WSL can reach Windows host IP
- Check firewall settings on Windows

### Simulation Not Working
- Ensure no real MeOS connection interferes
- Check logs for simulation phase messages
- Verify time calculations in generator.go

## Testing Strategies

1. **Unit Testing**: Test individual packages in isolation
2. **Integration Testing**: Use simulation mode for end-to-end tests
3. **Manual Testing**: Use curl or a REST client to verify endpoints
4. **Load Testing**: Simulation mode can handle rapid requests

## Swagger Documentation

### Single Source of Truth
The Swagger/OpenAPI documentation is generated from annotations in `cmd/meos-graphics/main.go`. To ensure consistency:

1. **All API metadata must be defined in main.go annotations**
2. **Never manually edit generated files** (docs.go, swagger.json, swagger.yaml)
3. **Validation script** at `scripts/validate-swagger.sh` ensures docs match source annotations
4. **Pre-commit hooks** and `scripts/check.sh` check for drift between source and generated docs

### Updating API Documentation
```bash
# Update annotations in cmd/meos-graphics/main.go, then:
swag init -g cmd/meos-graphics/main.go --parseDependency --parseInternal

# Validate consistency
./scripts/validate-swagger.sh
```

## Important Instructions

- NEVER add the Claude Code byline to pull requests or commits
- This includes removing "🤖 Generated with [Claude Code](https://claude.ai/code)" or "Co-Authored-By: Claude <noreply@anthropic.com>"
- Keep commit messages clean and professional without AI attribution