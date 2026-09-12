package service

import (
	"log"
	"time"

	"42tui/api"
	deployment "42tui/deployment"
)

func Startup() (bool, error) {

	// Build container with the server
	log.Println("[Startup 0/3] building container...")
	if err := deployment.Build(); err != nil {
		log.Println("[Startup 0/3] build error:", err)
		return false, err
	}

	// Run the server container once is ready
	log.Println("[Startup 1/3] running container...")
	if _, err := deployment.Run(); err != nil {
		log.Println("[Startup 1/3] run error:", err)
		return false, err
	}

	// Wait until the API is ready (has a timeout)
	log.Println("[Startup 2/3] waiting for API endpoint: /status")
	err := api.WaitForRequestTo(
		"/status",
		func(resp map[string]interface{}) bool {
			log.Println("[Startup 2/3] /status response:", resp)
			return resp["status"] == "ok"
		},
		10*time.Second,
		30*time.Millisecond,
	)
	if err != nil {
		log.Println("[Startup 2/3] API error:", err)
		return false, err
	}

	log.Println("[Startup 3/3] READY")
	return true, nil
}