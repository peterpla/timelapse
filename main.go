package main

import (
	"log"
	"os"

	"github.com/go-playground/validator"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// Config holds application-wide configuration info
type Config struct {
	path string
	port string
}

var config Config
var validate *validator.Validate

func main() {
	c := Config{}
	c.Load()
	log.Printf("main: path=%q, port=%q\n", c.path, c.port)

	// validate = validator.New()

	// TODO: read Timelapse Definitions (TLDef) master list
	// TODO: create Go routine to handle each TLDef
	// TODO: Go routine cleanup on SIGTERM, etc.

	// TODO: handler for webform to enter new TLDef
	// TODO: handler on webform submit adds new TLDef to master list
}

// Load populates Config with application configuration info
func (c *Config) Load() {

	pflag.StringVar(&c.path, "path", "./", "path to folder containing timelapse.json")
	pflag.StringVar(&c.port, "port", "8099", "HTTP port to listen on")
	var help bool
	pflag.BoolVarP(&help, "help", "h", false, "show usage information")
	pflag.Parse()

	if help {
		pflag.PrintDefaults()
		os.Exit(0)
	}

	viper.BindPFlag("path", pflag.Lookup("path"))
	viper.BindPFlag("port", pflag.Lookup("port"))

	viper.SetEnvPrefix("timelapse")
	viper.AutomaticEnv()
	viper.BindEnv("path") // treats as upper-cased SerEnvPrefix value + "_" + upper-cased "path" (BindEnv argument)
	viper.BindEnv("port")

	c.path = viper.GetString("path")
	c.port = viper.GetString("port")
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
