package main

import "techpulse/internal/scheduler"

func main() {
	scheduler.RunStandalone("scheduler", 8081)
}
