# netscan

This is a simple network scanning utility, which pings every ip in given network.


# Usage
Run with:

```go run main.go 192.168.0.0/24```

or build and run:

```go build main.go```

```./main.exe 192.168.0.0/24```

IP address could be network address like ```192.168.0.0/24``` or host address like ```192.168.0.42/24``` but it is necessary to provide a mask. 
