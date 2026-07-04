package main

import "techpulse/internal/scheduler"

func main() {
	scheduler.RunStandalone("rag", 8086)
}
