package main

import "testing"

func runTrace(addr string) (result TracerouteResult)   {
	options := TracerouteOptions{ }
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
