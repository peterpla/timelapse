package main

import (
	"log"
	"os"

	"github.com/GoogleCloudPlatform/functions-framework-go/funcframework"
	"github.com/peterpla/timelapse/capture"
)

// main uses the Functions Framework for Go to run the Cloud Function locally.
// See https://github.com/GoogleCloudPlatform/functions-framework-go/
func main() {
	funcframework.RegisterHTTPFunction("/", capture.WebcamImage)

	// Use PORT environment variable, or default to 8080.
	port := "8080"
	if envPort := os.Getenv("PORT"); envPort != "" {
		port = envPort
	}

	if err := funcframework.Start(port); err != nil {
		log.Fatalf("funcframework.Start: %v\n", err)
	}
}
