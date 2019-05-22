package main

import (
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"sort"
	"syscall"
	"time")

func main()  {
	options := TracerouteOptions{}
	result := trace("8.8.8.8", &options)
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
		options.Port = 33434
	}
	if options.MaxHops == 0 {
		options.MaxHops = 30
	}
	if options.TimeoutMs == 0 {
		options.TimeoutMs = 1000
	}
	if options.Retries == 0 {
		options.Retries = 3
	}
}

func exitWithError(err error) {
	fmt.Printf("%v\n", err)
	os.Exit(1)
}

func calculateRank(input TracerouteResult) (ranks RankedHop)  {
	hopsLen := len(input.Hops)
	hops := input.Hops
	ranks.Hops = []Distance{}

	for i :=0; i<hopsLen; i++ {
		if i+1 < hopsLen {
			delta := timeAbs(hops[i].Time - hops[i+1].Time)
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
		exitWithError(nil)
	}

	destAddr, err := getDestinationAddress(dest)
	if (err != nil) {
		exitWithError(nil)
	}

	tv := syscall.NsecToTimeval(1000 * 1000 * options.TimeoutMs)
	if err != nil {
		exitWithError(err)
	}

	ttl := 1
	retry := 0
	traceCount := 0
	for {
		traceCount++;
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

			ttl += 1
			retry = 0

			if ttl > options.MaxHops || currAddr == destAddr {
				return result
			}
		} else {
			retry += 1
			log.Print(traceCount, "- ", "* t=", elapsed)
			if retry > options.Retries {
				ttl += 1
				retry = 0
			}
		}

	}
}