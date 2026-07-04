package main

import "techpulse/internal/scheduler"

func main() {
	scheduler.RunStandalone("parser", 8083)
}
