A little experiment to write a Go application that can relay incoming TCP connections to a dynamically
requested endpoint.

## Example

In terminal 1:

```bash
nc -k -l 8091
```

In terminal 2:

```bash
go run main.go --in 8090
```

In terminal 3:

```bash
nc 8090
```

Enter the following command into terminal 3:

```text
CONNECT :8091
```

That command line and any other lines you type will get relayed and displayed in terminal 1. Press Control-D to
stop the connection.