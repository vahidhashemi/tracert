# Tracert

Tracert is a golang based traceroute tool.

  - Trace to destination using ip / url
  - Show a sorted list of delays between hops from highes delay to lowest
  - only works with IPv4

### Technical Spec

This tool uses standard go libraries and developed using go 1.12.5 based on osX.

### How to Compie
```sh
$ go build -o trace main.go models.go
```
### How to use
```sh
$ trace -address www.example.com
```
Since This application is based on system calls you need to have root privlleges to run it. 

### Usage
There are optional parameters you can pass:
```
Usage : trace -address=address/ip [-port=portnumber][-maxhops=hopnumber][-timeout=milliseconds][-retry=retriesnumber]
  -address string
    	Enter an URL or an IP e.g. www.google.com or 1.1.1.
  -maxhops int
    	Enter Maximum Hop for tracing (default 30)
  -port int
    	Enter a Port Number (default 33443)
  -retry int
    	Enter Number of Retries (default 3)
  -timeout int
    	Enter Timeout in Milliseconds (1000 = 1s) (default 1000)
```

### Run Test
You need to have root privilleges for running tests

```sh
$ sudo go test
```



