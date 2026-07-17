package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"nats-router-explorer/internal/nats"
	"nats-router-explorer/internal/websocket"
)

func main() {
	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		natsURL = "nats://localhost:4222"
	}

	var hub *websocket.Hub
	var natsService *nats.Service
	var err error

	// 1. define the callback that the NATS service calls upon receiving a message
	onMessage := func(subject string, payload []byte) {
		if hub != nil {
			hub.Broadcast(subject, payload)
		}
	}

	// 2. initialize the NATS service by injecting the onMessage callback
	natsService, err = nats.NewService(natsURL, onMessage)
	if err != nil {
		log.Fatalf("failed to connect to NATS: %v", err)
	}
	defer natsService.Close()

	// 3. define subscription control functions based on dynamic client interest
	subscribeFn := func(subject string) error {
		if natsService != nil {
			return natsService.Subscribe(subject)
		}
		return fmt.Errorf("NATS service is not initialized")
	}

	unsubscribeFn := func(subject string) {
		if natsService != nil {
			natsService.Unsubscribe(subject)
		}
	}

	// 4. initialize the WebSocket Hub by injecting the subscription control functions
	hub = websocket.NewHub(subscribeFn, unsubscribeFn)

	// 5. configure the HTTP route for the WebSocket handler
	http.HandleFunc("/ws", hub.ServeWS)

	// 6. channel to intercept interruptions and perform a graceful shutdown
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		fmt.Println("Server running on :3000")
		if err := http.ListenAndServe(":3000", nil); err != nil && err != http.ErrServerClosed {
			log.Fatalf("failed to run HTTP server: %v", err)
		}
	}()

	<-stopChan
	fmt.Println("Shutting down server...")
}
