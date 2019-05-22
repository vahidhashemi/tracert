package main

import "time"

type TracerouteOptions struct {
	Port      int
	MaxHops   int
	TimeoutMs int64
	Retries   int
}

type TracerouteHop struct {
	Address [4]byte
	Time time.Duration
}

type TracerouteResult struct {
	Hops []TracerouteHop
}

type Distance struct {
	Title string
	Time time.Duration
}
type RankedHop struct {
	Hops []Distance

}
