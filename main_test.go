package main

import (
	"fmt"
	"testing"
)

func runTrace(addr string) (result TracerouteResult)   {
	options := TracerouteOptions{ TimeoutMs:100,Retries:1}
	result = trace(addr, &options)
	return
}

func TestTraceWithUrl(t *testing.T)  {
	result := runTrace("www.google.com")
	if len(result.Hops) == 0 {
		t.Errorf("Test Failed. Expected at Least One Hop")
	}
}

func TestTraceWithIP(t *testing.T) {
	result := runTrace("8.8.8.8")
	if len(result.Hops) == 0 {
		t.Errorf("Test Failed. Expected at Least One Hop")
	}
}

func TestCalcRank(t *testing.T) {
	var tr TracerouteResult
	tr.Hops = []TracerouteHop{}

	var dummyAddr [4]byte
	dummyAddr,_ = getLocalAddress()

	tr.Hops = append(tr.Hops, TracerouteHop{Time:1000, Address:dummyAddr})
	tr.Hops = append(tr.Hops, TracerouteHop{Time:2000, Address:dummyAddr})
	tr.Hops = append(tr.Hops, TracerouteHop{Time:4000, Address:dummyAddr})
	tr.Hops = append(tr.Hops, TracerouteHop{Time:5000000, Address:dummyAddr})

	output := calculateRank(tr)
	fmt.Println(output.Hops)
	if (len(tr.Hops)-1 != len(output.Hops)) {
		t.Errorf("Test Failed. Expected Three Item")

	}
}


