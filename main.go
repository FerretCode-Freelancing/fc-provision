package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/ferretcode-freelancing/fc-provision/api"
)

func main() {
	fmt.Println("Starting builder...")

	done := api.StartBuilder()

	fmt.Println("API started.")

	shutdown := make(chan os.Signal, 1)

	signal.Notify(shutdown, syscall.SIGTERM)
	signal.Notify(shutdown, syscall.SIGINT)
	signal.Notify(shutdown, syscall.SIGQUIT)

	select {
	case <-shutdown:
		done <- struct{}{}
	}
}
