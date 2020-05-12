package main

import "github.com/go-playground/validator"

var validate *validator.Validate

func main() {
	// TODO: env var: TIMELAPSE_DEFINITIONS (path to timelapse.json file)
	// TODO: env var: TIMELAPSE_PORT (HTTP port to listen on)

	// TODO: command line arguments

	validate = validator.New()

	// TODO: read Timelapse Definitions (TLDef) master list
	// TODO: create Go routine to handle each TLDef
	// TODO: Go routine cleanup on SIGTERM, etc.

	// TODO: handler for webform to enter new TLDef
	// TODO: handler on webform submit adds new TLDef to master list
}

// TLDef represents a Timelapse capture definition
type TLDef struct {
	Name           string `json:"name,omitempty" validate:"required"`             // Friendly name of this timelapse definition
	URL            string `json:"url,omitempty" validate:"required"`              // URL of webcam image
	First          string `json:"first,omitempty" validate:"required"`            // Time or special value, e.g., "Sunrise + 1 hour"
	Last           string `json:"last,omitempty" validate:"required"`             // Time or special value, e.g., "Sunset - 1 hour", "none"
	CapturesPerDay int    `json:"captures_per_day,omitempty" validate:"required"` // Total captures per day, including First and Last
	FolderPath     string `json:"folder_path,omitempty" validate:"required"`      // Fully-qualified path of folder to store captures
}
