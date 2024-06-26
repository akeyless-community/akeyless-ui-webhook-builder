// main.go

package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/akeyless-community/akeyless-ui-webhook-builder/cmd"
)

func main() {
	// Set up channel to catch SIGINT
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	// Run the command in a goroutine
	errChan := make(chan error)
	go func() {
		errChan <- cmd.Execute()
	}()

	// Wait for either the command to finish or an interrupt
	select {
	case err := <-errChan:
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	case <-c:
		fmt.Println("\nOperation cancelled by user.")
		os.Exit(0)
	}
}
