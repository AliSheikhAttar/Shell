# Shell

A simple shell implementation in Go.

## Commands

### exit
Exits the shell with an optional status code.

Usage:
- `exit`: Exits with status code 0
- `exit <code>`: Exits with the specified status code
  
Examples:
```bash
$ exit      # exits with status 0
$ exit 2    # exits with status 2
$ exit 1 2  # error: too many arguments
```