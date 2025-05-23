# Claude Context for MeOS Graphics API

## Project Overview

This is a Go-based REST API server that connects to MeOS (orienteering event software) and provides competition data for graphics displays. The application can run in two modes:
- **Normal mode**: Connects to a real MeOS server
- **Simulation mode**: Generates test data for development

## Key Technical Details

### Architecture
- **Language**: Go 1.21+
- **Web Framework**: gin-gonic/gin
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
│   └── state/               # Thread-safe state management
└── logs/                    # Log files (auto-created)
```

### Running in WSL
When running in WSL, the MeOS server connection uses the Windows host IP (configured in `internal/meos/config.go`). The default is `192.168.112.1:2009`.

## Common Tasks

### Running the Server
```bash
# Normal mode (connects to MeOS)
go run ./cmd/meos-graphics

# Simulation mode (test data)
go run ./cmd/meos-graphics --simulation
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

## Future Enhancements to Consider

- WebSocket support for real-time updates
- Configuration file support (YAML/JSON)
- Docker containerization
- Prometheus metrics endpoint
- OpenAPI/Swagger documentation
- Database persistence option
- Multi-event support