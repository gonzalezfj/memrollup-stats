# memrollup-stats

A memory usage monitoring and statistics tool written in Go.

## Features

- Monitor process memory usage at configurable frequencies
- Collect various memory metrics (RSS, PSS, USS, Shared Clean/Dirty, Private Clean/Dirty)
- Output statistics in CSV or JSON format
- Support for Linux kernel 4.14 or later
- Verbose mode for detailed logging
- Graceful process termination

## Installation

```bash
go install github.com/gonzalezfj/memrollup-stats@latest
```

## Usage

```bash
memrollup-stats -F <frequency_hz> [-o output_file] [-j] [-v] command [args...]
```

### Options

- `-F <frequency_hz>`: Sampling frequency in Hz (default: 1)
- `-o <output_file>`: Output file for statistics (default: stdout)
- `-j`: Output in JSON format
- `-v`: Verbose mode
- `-h`: Show help message
- `-V`: Show version

### Examples

Monitor a process with 10Hz frequency:

```bash
memrollup-stats -F 10 ./my_program
```

Save statistics to CSV file:

```bash
memrollup-stats -F 2 -o stats.csv ./my_program
```

Output JSON for data analysis:

```bash
memrollup-stats -F 1 -j -o stats.json ./my_program
```

## Exit Codes

- `0`: Success
- `1`: General error
- `2`: Invalid arguments
- `3`: Process start failure
- `4`: Permission denied
- `5`: System resource unavailable

## Requirements

- Linux kernel 4.14 or later (for smaps_rollup support)
- Go 1.21 or later

## License

MIT License

## Author

Facundo Gonzalez <facujgg@gmail.com>
