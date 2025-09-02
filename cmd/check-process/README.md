# check-process

A Sensu check plugin for verifying processes are running.

## Features

- **Process Monitoring**: Verifies that specific processes are running
- **Regular Expression Support**: Use regex patterns to match process names or command lines
- **Command Line Matching**: Matches against full command line arguments, not just process names
- **Process Listing**: Shows matched processes with their PIDs
- **Self-Exclusion**: Automatically excludes the check process itself from results
- **Cross-Platform Support**: Works on Linux, macOS, Windows, and other platforms

## Usage

```bash
check-process [OPTIONS]
```

### Options

- `-p, --regexp_pattern` - Regular expression pattern to match processes (default: "a_process_name")

## Examples

```bash
# Check for nginx process
check-process -p nginx

# Check for Java processes
check-process -p "java"

# Check for specific Java application
check-process -p "java.*MyApplication"

# Check for processes with specific arguments
check-process -p "python.*manage.py"

# Check for multiple processes using regex OR
check-process -p "(nginx|apache2)"

# Check for process with specific port
check-process -p "node.*:3000"
```

## Exit Codes

- **0 (OK)**: One or more processes matching the pattern were found
- **2 (CRITICAL)**: No processes matching the pattern were found
- **3 (ERROR)**: Error occurred while checking processes (invalid regex, permission issues, etc.)

## Output Examples

**Process Found:**
```
 - (1234) /usr/sbin/nginx -g daemon off;
 - (1235) nginx: worker process
 - (1236) nginx: worker process
CheckProcess OK: Process [nginx]: 3 occurence(s)
```

**Process Not Found:**
```
CheckProcess CRITICAL: Unable to find process [apache2]
```

**Multiple Matches:**
```
 - (5678) java -jar application.jar
 - (5679) java -Xmx2g -jar worker.jar
CheckProcess OK: Process [java]: 2 occurence(s)
```

## Regular Expression Examples

| Pattern | Matches |
|---------|---------|
| `nginx` | Any process with "nginx" in its command line |
| `^nginx` | Processes starting with "nginx" |
| `\.jar$` | Processes ending with ".jar" |
| `python.*\.py` | Python scripts |
| `node.*app\.js` | Node.js applications |
| `(mysql\|maria)` | MySQL or MariaDB processes |
| `java.*-Xmx[0-9]+g` | Java processes with memory settings |

## Use Cases

- **Service Monitoring**: Ensure critical services are running
- **Application Health**: Verify application processes are active
- **Cluster Monitoring**: Check that all required cluster components are running
- **Dependency Checking**: Verify dependent services before starting applications
- **Process Count Validation**: Ensure expected number of worker processes

## Notes

- Matches against the full command line, not just the process name
- The check process itself is automatically excluded from results
- Lists all matching processes with their PIDs before the status message
- Regular expressions use Go's regexp syntax (RE2)
- Requires appropriate permissions to read process information