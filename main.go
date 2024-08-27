package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

func main() {
	server := NewServer()

	// Load configuration from config file
	config, err := loadConfig("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	for _, mapping := range config.Mappings {
		err := server.AddForwarder(mapping.From, mapping.To)
		if err != nil {
			log.Printf("Failed to add forwarder %s -> %s: %v", mapping.From, mapping.To, err)
		}
	}

	// Set up HTTP server
	http.HandleFunc("/stats", func(w http.ResponseWriter, r *http.Request) {
		stats := server.GetStats()
		json.NewEncoder(w).Encode(stats)
	})

	// Start HTTP server
	httpAddr := fmt.Sprintf(":%d", config.Port)
	go func() {
		log.Printf("Starting HTTP server on %s", httpAddr)
		log.Fatal(http.ListenAndServe(httpAddr, nil))
	}()

	fmt.Println("Port forwarding service and HTTP server started")
	fmt.Printf("Visit http://localhost:%d/stats to view statistics\n", config.Port)

	// Keep the main program running
	select {}
}
