package main

import "techpulse/internal/scheduler"

func main() {
	scheduler.RunStandalone("fetcher", 8082)
}
