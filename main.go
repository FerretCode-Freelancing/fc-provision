package main

import (
	"fmt"

	"github.com/ferretcode-freelancing/fc-provision/api"
)

func main() {
	fmt.Println("Starting builder...")

	api.NewApi()

	fmt.Println("API started.")
}
