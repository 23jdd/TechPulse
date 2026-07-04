package main

import "techpulse/internal/scheduler"

func main() {
	scheduler.RunStandalone("search", 8085)
}
