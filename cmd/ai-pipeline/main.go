package main

import "techpulse/internal/scheduler"

func main() {
	scheduler.RunStandalone("ai-pipeline", 8084)
}
