# MeOS Graphics API

A Go-based REST API server that connects to MeOS (orienteering event software) and provides competition data for graphics displays.

## Project Structure

```
meos-graphics/
├── cmd/
│   └── meos-graphics/
│       └── main.go           # Application entry point
├── internal/                  # Private application code
│   ├── handlers/             # HTTP request handlers
│   │   ├── handlers.go       # Main handler implementations
│   │   └── types.go          # Response type definitions
│   ├── logger/               # Logging functionality
│   │   └── logger.go         # Logger initialization and configuration
│   ├── meos/                 # MeOS integration
│   │   ├── adapter.go        # MeOS adapter implementation
│   │   ├── config.go         # MeOS configuration
│   │   └── types.go          # MeOS XML type definitions
│   ├── middleware/           # HTTP middleware
│   │   └── logger.go         # Request logging middleware
│   ├── models/               # Domain models
│   │   └── models.go         # Core data structures
│   └── state/                # Application state management
│       └── state.go          # Global state with thread-safe access
└── logs/                     # Log files directory
```

## Features

- Connects to MeOS information server via XML API
- Polls for updates every second
- Thread-safe in-memory state management
- REST API endpoints for competition graphics
- Logging to both console and file

## API Endpoints

- `GET /health` - Health check endpoint
- `GET /classes` - List all competition classes
- `GET /classes/:classId/startlist` - Get start list for a class
- `GET /classes/:classId/results` - Get results with positions and radio times
- `GET /classes/:classId/splits` - Get split time standings at each control

## Configuration

### Command-Line Flags

All configuration is done through command-line flags. For a complete reference of all available flags, see [docs/CLI_FLAGS.md](docs/CLI_FLAGS.md).

Common flags:
- `--simulation` - Run in simulation mode (no MeOS server required)
- `--meos-host <hostname>` - MeOS server hostname or IP (default: localhost)
- `--meos-port <port>` - MeOS server port (default: 2009, use 'none' to omit port)
- `--poll-interval <duration>` - How often to fetch updates from MeOS (default: 1s)
- `--version` - Show version information
- `--help` - Show help for all available flags

Examples:
```bash
# Connect to local MeOS server
go run ./cmd/meos-graphics

# Connect to remote MeOS server
go run ./cmd/meos-graphics --meos-host 192.168.1.100 --meos-port 8080

# Connect without port (for reverse proxy setups)
go run ./cmd/meos-graphics --meos-host meos.example.com --meos-port none

# Use faster polling for more responsive updates
go run ./cmd/meos-graphics --poll-interval 200ms
```

### Poll Interval Details

The `--poll-interval` flag accepts Go duration strings:
- Minimum: 100ms
- Maximum: 1 hour
- Default: 1s

Lower intervals provide more responsive updates but increase network traffic and server load.

## Running

### Using Docker (Recommended)
```bash
# Pull from GitHub Container Registry
docker pull ghcr.io/metsaapp/meos-graphics:latest

# Run in normal mode
docker run -p 8090:8090 ghcr.io/metsaapp/meos-graphics:latest

# Run in simulation mode
docker run -p 8090:8090 ghcr.io/metsaapp/meos-graphics:latest --simulation

# With custom MeOS server
docker run -p 8090:8090 ghcr.io/metsaapp/meos-graphics:latest --meos-host 192.168.1.100 --meos-port 8080

# With custom poll interval (default: 1s)
docker run -p 8090:8090 ghcr.io/metsaapp/meos-graphics:latest --poll-interval 500ms

# With persistent logs
docker run -p 8090:8090 -v $(pwd)/logs:/app/logs ghcr.io/metsaapp/meos-graphics:latest
```

### From Source
```bash
# Normal mode
go run ./cmd/meos-graphics

# Simulation mode
go run ./cmd/meos-graphics --simulation

# Custom MeOS server
go run ./cmd/meos-graphics --meos-host 192.168.1.100 --meos-port 8080

# Custom poll interval
go run ./cmd/meos-graphics --poll-interval 200ms

# Show version
go run ./cmd/meos-graphics --version
```

The server will start on port 8090.

## Simulation Mode

The simulation mode generates test data for development and testing without requiring a MeOS server. It runs a 15-minute cycle:

- **Minutes 0-3**: Start list phase - all competitors registered but not started
- **Minutes 3-10**: Running phase - competitors progressively receive split times and finish
- **Minutes 10-15**: Results phase - all competitors finished, stable results
- **After 15 minutes**: Automatic restart with fresh data

Features:
- All competitors start at the same time
- Realistic time variations with deciseconds
- 3 classes with different numbers of radio controls
- Random clubs and names from predefined lists

## Dependencies

- [gin-gonic/gin](https://github.com/gin-gonic/gin) - HTTP web framework

## Versioning and Releases

This project uses [Semantic Versioning](https://semver.org/) and [Conventional Commits](https://www.conventionalcommits.org/) for automatic version management.

### Conventional Commits

All commits should follow the conventional commit format:

```
<type>(<scope>): <subject>

<body>

<footer>
```

Types:
- `feat:` - New features (triggers minor version bump)
- `fix:` - Bug fixes (triggers patch version bump)
- `docs:` - Documentation changes
- `style:` - Code style changes (formatting, etc.)
- `refactor:` - Code refactoring
- `test:` - Test additions or corrections
- `chore:` - Maintenance tasks
- `ci:` - CI/CD changes
- `build:` - Build system changes

Breaking changes:
- Add `!` after type: `feat!: breaking change`
- Or add `BREAKING CHANGE:` in the commit body
- These trigger major version bumps

### Release Process

Releases are fully automated using [release-please](https://github.com/googleapis/release-please):

1. Merge PRs with conventional commits to `main`
2. Release-please creates/updates a release PR
3. When the release PR is merged:
   - A new GitHub release is created
   - Version numbers are bumped
   - CHANGELOG.md is generated
   - Binary artifacts are built for multiple platforms
   - Docker images are published to ghcr.io

### Docker Images

Docker images are automatically published to GitHub Container Registry:

- `ghcr.io/metsaapp/meos-graphics:latest` - Latest release
- `ghcr.io/metsaapp/meos-graphics:vX.Y.Z` - Specific version

Images are multi-platform (linux/amd64, linux/arm64).

## Architecture

The application follows Go best practices with clear separation of concerns:

- **Models** define the core domain objects
- **MeOS adapter** handles communication with the MeOS server
- **State** provides thread-safe storage and access to competition data
- **Handlers** implement the REST API endpoints
- **Logger** provides structured logging to file and console
- **Middleware** handles cross-cutting concerns like request logging

## Development

### Updating CLI Documentation

The CLI documentation in `docs/CLI_FLAGS.md` is automatically generated from the code. To update it:

```bash
# Generate documentation
go run ./cmd/generate-docs
```

The CI pipeline validates that the documentation is up-to-date. If you add or modify command-line flags, you must regenerate the documentation and commit the changes.