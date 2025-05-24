# MeOS Graphics CLI Flags

MeOS Graphics API Server connects to MeOS (orienteering event software) 
and provides competition data for graphics displays.

The server can run in two modes:
- Normal mode: Connects to a real MeOS server
- Simulation mode: Generates test data for development

## Usage

```bash
meos-graphics [flags]
```

## Available Flags

### --meos-host

- **Type**: string
- **Default**: "localhost"
- **Description**: MeOS server hostname or IP address

### --meos-port

- **Type**: string
- **Default**: "2009"
- **Description**: MeOS server port (use 'none' to omit port from URL)

### --poll-interval

- **Type**: duration
- **Default**: 1s
- **Description**: Poll interval for MeOS data updates (e.g., 200ms, 9s, 2m)

### --simulation

- **Description**: Run in simulation mode

### --simulation-duration

- **Type**: duration
- **Default**: 15m0s
- **Description**: Total simulation cycle duration (only with --simulation)

### --simulation-phase-results

- **Type**: duration
- **Default**: 5m0s
- **Description**: Duration of results phase (only with --simulation)

### --simulation-phase-running

- **Type**: duration
- **Default**: 7m0s
- **Description**: Duration of running phase (only with --simulation)

### --simulation-phase-start

- **Type**: duration
- **Default**: 3m0s
- **Description**: Duration of start list phase (only with --simulation)

## Examples

### Run in simulation mode

```bash
meos-graphics --simulation
```

### Connect to custom MeOS server

```bash
meos-graphics --meos-host=10.0.0.5 --meos-port=3000
```

### Use faster poll interval

```bash
meos-graphics --poll-interval=200ms
```

### Connect to MeOS server without specifying port

```bash
meos-graphics --meos-host=meos.example.com --meos-port=none
```

### Show version information

```bash
meos-graphics --version
```

## Environment Variables

Currently, no environment variables are supported. All configuration is done through command-line flags.

## Notes

- The `--poll-interval` flag accepts Go duration strings (e.g., "200ms", "1s", "2m", "1h")
- When using `--meos-port=none`, the port is omitted from the MeOS server URL
- In simulation mode, the application generates test data without connecting to a real MeOS server
