package main

import (
	"fmt"
	"testing"
	"time"
)

func runTrace(addr string) (result TracerouteResult)   {
	options := TracerouteOptions{ TimeoutMs:100,Retries:1}
	result = trace(addr, &options)
	return
}

func createDummyHops() (tr TracerouteResult) {
	var dummyAddr [4]byte
	dummyAddr,_ = getLocalAddress()
	tr.Hops = append(tr.Hops, TracerouteHop{Time:1000, Address:dummyAddr})
	tr.Hops = append(tr.Hops, TracerouteHop{Time:2000, Address:dummyAddr})
	tr.Hops = append(tr.Hops, TracerouteHop{Time:1000000, Address:dummyAddr})
	tr.Hops = append(tr.Hops, TracerouteHop{Time:5000000, Address:dummyAddr})
	return tr
}

//func TestTraceWithUrl(t *testing.T)  {
//	result := runTrace("www.google.com")
//	if len(result.Hops) == 0 {
//		t.Errorf("Test Failed. Expected at Least One Hop")
//	}
//}
//
//func TestTraceWithIP(t *testing.T) {
//	result := runTrace("8.8.8.8")
//	if len(result.Hops) == 0 {
//		t.Errorf("Test Failed. Expected at Least One Hop")
//	}
//}


func TestCalcRankNumberOfItems(t *testing.T) {
	var tr = createDummyHops()
	output := calculateRank(tr)
	fmt.Println(output.Hops)
	if (len(tr.Hops)-1 != len(output.Hops)) {
		t.Errorf("Test Failed. Expected Three Item")

	}
}

func TestCalcRankHighestDelayPlacedAtFirstofList(t *testing.T) {
	var tr = createDummyHops()
	output := calculateRank(tr)
	if output.Hops[0].Time != time.Duration(4 * time.Millisecond) {
		t.Errorf("Test Failed. List is not Sorted")
	}
}


