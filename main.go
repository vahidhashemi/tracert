package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"sort"
	"syscall"
	"time"
)
var defaultPortNumber int = 33443
var maxHops int = 30
var timeoutMs int64 = 1000
var retries int = 3
func main()  {


	urlParam := flag.String("address", "", "Enter an URL or an IP e.g. www.google.com or 1.1.1.")
	portParam := flag.Int("port",defaultPortNumber, "Enter a Port Number")
	maxhopsParam := flag.Int("maxhops", maxHops, "Enter Maximum Hop for tracing")
	timeoutMsParam := flag.Int64("timeout",timeoutMs, "Enter Timeout in Milliseconds (1000 = 1s)")
	retriesParam := flag.Int("retry", retries, "Enter Number of Retries")


	flag.Parse()
	if *urlParam == "" {
		fmt.Println("Usage : trace -address=address/ip [-port=portnumber][-maxhops=hopnumber][-timeout=milliseconds][-retry=retriesnumber]")
		fmt.Println("")
		flag.CommandLine.SetOutput(os.Stdout)
		flag.PrintDefaults()
		os.Exit(2)
	}
	options := TracerouteOptions{Port:*portParam, MaxHops:*maxhopsParam,TimeoutMs:*timeoutMsParam,Retries:*retriesParam }
	result := trace(*urlParam, &options)
	endresult := calculateRank(result)
	printArray(endresult)
}

func getLocalAddress() (addr [4]byte, err error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return
	}

	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if len(ipnet.IP.To4()) == net.IPv4len {
				log.Println("Found IP address: ", ipnet.IP.String())
				copy(addr[:], ipnet.IP.To4())
				return
			}
		}
	}
	err = errors.New("Check Your Internet Connection")
	return
}

func getDestinationAddress(dest string) (destAddr [4]byte, err error) {
	addrs, err := net.LookupHost(dest)
	if err != nil {
		return
	}
	addr := addrs[0]
	log.Println("Destination address: ", addr)

	ipAddr, err := net.ResolveIPAddr("ip", addr)
	if err != nil {
		return
	}
	copy(destAddr[:], ipAddr.IP.To4())
	return
}



func defaultOptions(options *TracerouteOptions) {
	if options.Port == 0 {
		options.Port = defaultPortNumber
	}
	if options.MaxHops == 0 {
		options.MaxHops = maxHops
	}
	if options.TimeoutMs == 0 {
		options.TimeoutMs = timeoutMs
	}
	if options.Retries == 0 {
		options.Retries = retries
	}
}

func exitWithError(err error) {
	fmt.Printf("Error %v\n", err)
	os.Exit(1)
}

func calculateRank(input TracerouteResult) (ranks RankedHop)  {
	hopsLen := len(input.Hops)
	hops := input.Hops
	ranks.Hops = []Distance{}

	for i :=0; i<hopsLen; i++ {
		if i+1 < hopsLen {
			delta := timeAbs(hops[i+1].Time - hops[i].Time)
			info := fmt.Sprintf("time between hops %d and %d ",i,i+1)
			ranks.Hops = append(ranks.Hops, Distance{Title: info , Time: delta})
		}
	}
	sort.Slice(ranks.Hops, func(i, j int) bool {
		return ranks.Hops[i].Time > ranks.Hops[j].Time
	})
	return ranks
}

func timeAbs(time time.Duration) time.Duration {
	if (time < 0) {
		time = time * -1
	}
	return time
}

func printArray(data RankedHop)  {
	for _,element := range data.Hops{
		fmt.Println(element)
	}
}

func setSocket(domain int,typ int, protocol int ) (fileDescriptor int) {
	var err error
	fileDescriptor, err = syscall.Socket(domain, typ, protocol)
	if (err != nil){
		exitWithError(err)
	}
	return fileDescriptor
}

func trace(dest string, options *TracerouteOptions) (result TracerouteResult) {
	result.Hops = []TracerouteHop{}
	defaultOptions(options)

	socketAddr, err := getLocalAddress()
	if (err != nil) {
		exitWithError(err)
	}

	destAddr, err := getDestinationAddress(dest)
	if (err != nil) {
		exitWithError(err)
	}

	tv := syscall.NsecToTimeval(1000 * 1000 * options.TimeoutMs)
	if err != nil {
		exitWithError(err)
	}

	ttl := 1
	retry := 0
	traceCount := 0
	for {
		start := time.Now()

		recvSocket := setSocket(syscall.AF_INET, syscall.SOCK_RAW, syscall.IPPROTO_ICMP)
		sendSocket := setSocket(syscall.AF_INET, syscall.SOCK_DGRAM, syscall.IPPROTO_UDP)

		if err := syscall.SetsockoptInt(sendSocket, 0x0, syscall.IP_TTL, ttl); err != nil {
			exitWithError(err)
		}
		if err := syscall.SetsockoptTimeval(recvSocket, syscall.SOL_SOCKET, syscall.SO_RCVTIMEO, &tv); err != nil {
			exitWithError(err)
		}

		defer syscall.Close(recvSocket)
		defer syscall.Close(sendSocket)

		if err := syscall.Bind(recvSocket, &syscall.SockaddrInet4{Port: options.Port, Addr: socketAddr}); err != nil {
			exitWithError(err)
		}
		if err := syscall.Sendto(sendSocket, []byte{0x0}, 0, &syscall.SockaddrInet4{Port: options.Port, Addr: destAddr}); err != nil {
			exitWithError(err)
		}

		var buffer = make([]byte, 512)
		n, from, err := syscall.Recvfrom(recvSocket, buffer, 0)
		elapsed := time.Since(start)
		if err == nil {
			currAddr := from.(*syscall.SockaddrInet4).Addr

			result.Hops = append(result.Hops, TracerouteHop{currAddr, elapsed})

			log.Println(traceCount, "- ", "Received n=", n, ", from=", currAddr, ", t=", elapsed)
			traceCount++;

			ttl += 1
			retry = 0

			if ttl > options.MaxHops || currAddr == destAddr {
				return result
			}
		} else {
			retry += 1
			if retry > options.Retries {
				ttl += 1
				retry = 0
			}
		}

	}
}