package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"log"
	"context"

	"github.com/google/uuid"
	"github.com/kubemq-io/kubemq-go"
	"github.com/ferretcode-freelancing/fc-provision/api"
	"github.com/ferretcode-freelancing/fc-bus"
)

func main() {
	fmt.Println("Starting builder...")

	ctx := context.Background()

	bus := events.Bus{
		Channel: "build-pipeline",
		ClientId: uuid.NewString(),
		Context: ctx,
		TransportType: kubemq.TransportTypeGRPC,
	}

	client, err := bus.Connect()

	if err != nil {
		log.Fatal(fmt.Sprintf("There was an error starting the builder: %s", err))
	}

	fmt.Println("bus is connected")

	done, err := bus.Subscribe(client, func(msgs *kubemq.ReceiveQueueMessagesResponse, subscribeErr error) {
		fmt.Println("message received")

		err := api.Build(msgs)

		if err != nil {
			log.Printf("There was an error building the image: %s", err)
		}
	})

	if err != nil {
		log.Fatal(fmt.Sprintf("There was an error subscribing to the bus: %s", err))
	}

	fmt.Println("API started.")

	shutdown(done)

	fmt.Println("hi")
}

func shutdown(done chan struct{}) {
	shutdown := make(chan os.Signal, 1)

	signal.Notify(shutdown, syscall.SIGTERM)
	signal.Notify(shutdown, syscall.SIGINT)
	signal.Notify(shutdown, syscall.SIGQUIT)

	select {
	case <-shutdown:
		done <- struct{}{}
	}
}
