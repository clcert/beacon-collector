package main

import "time"

const secondMarkInit = 0

func main() {
	oneSecondInNs := 1000000000
	for {
		now := time.Now()
		timeToWait := time.Duration(oneSecondInNs*(60-now.Second()) + (1000000000 - now.Nanosecond()) - oneSecondInNs)
		println("Waiting for the current minute to end...")
		time.Sleep(timeToWait)
		collectEvents()
	}
	//collectEvents()
}
