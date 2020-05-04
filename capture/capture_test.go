// Package capture retrieves webcam images
package capture

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/GoogleCloudPlatform/functions-framework-go/funcframework"
)

func TestMain(m *testing.M) {
	funcframework.RegisterHTTPFunction("/", WebcamImage)

	// Use PORT environment variable, or default to 8080.
	port := "8080"
	if envPort := os.Getenv("PORT"); envPort != "" {
		port = envPort
	}

	go startFramework(port) // call ListenAndServe from a separate go routine so main can listen for signals

	exitcode := m.Run()
	os.Exit(exitcode)
}

// startFramework starts funcframework which calls ListenAndServe
func startFramework(port string) {
	if err := funcframework.Start(port); err != nil {
		log.Fatalf("funcframework.Start: %v\n", err)
	}
}

func TestWebcamImage(t *testing.T) {
	var emptyString string

	tests := []struct {
		name   string
		args   WebcamRequest
		status int
		body   *string
	}{
		{"empty body",
			WebcamRequest{},
			http.StatusBadRequest,
			nil,
		},
		{"valid",
			WebcamRequest{
				URL:         "https://www.nps.gov/webcams-lavo/kyvc_webcam1.jpg",
				TimeoutSec:  10,
				FileRoot:    "kyvc_webcam1",
				FilePattern: "yyyyMMdd_hhmmss",
			},
			http.StatusOK,
			&emptyString,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			// WebcamImage(tt.args.w, tt.args.r)

			rr := httptest.NewRecorder()

			body, err := json.Marshal(tt.args)
			if err != nil {
				t.Fatalf("json.Marshall: %v", err)
			}
			log.Printf("body: %q\n", string(body))
			req := httptest.NewRequest("POST", "/", bytes.NewReader(body))

			WebcamImage(rr, req)

			if rr.Result().StatusCode != tt.status {
				t.Errorf("%s: got status %v, want %v", tt.name, rr.Result().StatusCode, tt.status)
			}

			if tt.body != nil {
				var b []byte
				if b, err = ioutil.ReadAll(rr.Body); err != nil {
					t.Fatalf("%s: ReadAll: %v", tt.name, err)
				}
				t.Errorf("%s: got body %q, want %q", tt.name, string(b), *tt.body)
			}
		})
	}
}
