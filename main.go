package main

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/go-playground/validator"
	"github.com/julienschmidt/httprouter"
	"github.com/monoculum/formam"
	"github.com/peterpla/lead-expert/pkg/middleware"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var srv *server

const masterPath = "/Users/peterplamondon/Downloads/timelapse/"
const masterFile = "timelapse.json"

func main() {
	defer catch() // implements recover so panics reported
	sn := "timelapse"

	runtime.GOMAXPROCS(2)

	// use context and cancel with goroutines to handle Ctrl+C
	ctx, cancel := context.WithCancel(context.Background())

	srv = newServer()
	if err := srv.mtld.Read(filepath.Join(masterPath, masterFile)); err != nil {
		msg := fmt.Sprintf("main, srv.mtld.Read: %v", err)
		panic(msg)
	}

	// TODO: #5 create Go routine to handle each TLDef
	var wg sync.WaitGroup
	for i, tld := range *srv.mtld {
		log.Printf("%s, launching goroutine #%d, %s, poll interval: %d", sn, i, tld.Name, srv.config.pollSecs)
		wg.Add(1)
		go func(ctx context.Context, tld TLDef, pollInterval int) {
			log.Printf("goroutine handling TLDef %s (%p)\n", tld.Name, &tld)

			tld.SetCaptureTimes() // calculate all capture times

			for {
				select {
				case <-ctx.Done():
					log.Printf("goroutine handling TLDef %s exiting after ctx.Done\n", tld.Name)
					wg.Done()
					return
				default:
					if tld.IsTimeForCapture() {
						// capture and store the image
					}
				}
				// log.Printf("TLDef %s sleeping for %d seconds...\n", tld.Name, pollInterval)
				time.Sleep(time.Duration(pollInterval) * time.Second)
			}
		}(ctx, tld, srv.config.pollSecs)
		time.Sleep(100 * time.Millisecond)
	}

	srv.initTemplates("./templates", ".html")
	srv.router.ServeFiles("/static/*filepath", http.Dir("static"))
	srv.router.GET("/new", srv.handleNew())
	srv.router.GET("/", srv.handleHome())

	hs := http.Server{
		Addr:         ":" + srv.config.port,
		Handler:      middleware.LogReqResp(srv.router), // mux
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	log.Printf("Starting service %s listening on port %s", sn, hs.Addr)
	go startListening(&hs, "main") // call ListenAndServe from a separate go routine so main can listen for signals

	// on handled signals, cancel goroutines and exit
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	defer func() {
		signal.Stop(c)
		cancel()
		wg.Wait()
	}()

	s := <-c
	log.Printf("\n%s received signal %s, terminating", sn, s.String())
}

// startListening invokes ListenAndServe
func startListening(hs *http.Server, sn string) {
	if err := hs.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("%s startListening, ListenAndServe returned err: %+v\n", sn, err)
	}
}

// catch uses recover() and logs it
func catch() {
	sn := "timelapse"
	defer func() {
		if r := recover(); r != nil {
			log.Fatalf("=====> RECOVER in %s.catch, recover() returned: %v\n", sn, r)
		}
	}()
}

// ********** ********** ********** ********** ********** **********

type server struct {
	router   *httprouter.Router
	validate *validator.Validate // use a single instance of Validate, it caches struct info
	config   *Config
	tmpl     *template.Template
	mtld     *masterTLDefs
}

// newServer returns a server struct with router and validation initialized,
// and application configuration loaded
func newServer() *server {
	s := &server{}
	s.router = httprouter.New()
	s.validate = validator.New()
	s.mtld = newMasterTLDefs()

	s.config = &Config{}
	s.config.Load()

	return s
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

func (s *server) handleHome() httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		// startTime := time.Now()

		data := struct {
			Company string
		}{
			Company: "Timelapse",
		}

		srv.tmpl.ExecuteTemplate(w, "layout", data)

		// log.Printf("%s.%s, duration %v\n", sn, mn, time.Now().Sub(startTime))
		return
	}
}

func (s *server) handleNew() httprouter.Handle {

	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		sn := "handleNew"
		// startTime := time.Now()

		if err := r.ParseForm(); err != nil {
			log.Printf("FromForm: r.ParseForm: %v *****ERROR*****\n", err)
			return
		}

		// dump r.Form contents
		// log.Println("After r.ParseForm(), r.Form values:")
		// for key, value := range r.Form {
		// 	log.Printf("%q: %q\n", key, value)
		// }

		tld := TLDef{}

		decoder := formam.NewDecoder(nil)
		if err := decoder.Decode(r.Form, &tld); err != nil {
			log.Printf("%s, decoder.Decode: %v\n", sn, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		// handle checkbox values
		if _, ok := r.Form["firstTime"]; ok {
			tld.FirstTime = true
		}
		if _, ok := r.Form["firstSunrise"]; ok {
			tld.FirstSunrise = true
		}
		if _, ok := r.Form["lastTime"]; ok {
			tld.LastTime = true
		}
		if _, ok := r.Form["lastSunset"]; ok {
			tld.LastSunset = true
		}

		// if the FolderPath directory doesn't exist, create it
		if err := os.MkdirAll(tld.FolderPath, 0664); err != nil { // octal for -rw-rw-r--: owner read/write, group/other read-only
			log.Printf("%s, os.MkdirAll: %v\n", sn, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		// log.Printf("handleNew, TLDef: %+v", tld)

		srv.mtld.Append(&tld)
		if err := srv.mtld.Write(); err != nil {
			log.Printf("%s, srv.mtld.Write: %v\n", sn, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)

		// log.Printf("%s.%s, duration %v\n", sn, mn, time.Now().Sub(startTime))
		return
	}
}

// initTemplates reads and parses template files, and saves the template
// in the server receiver
func (s *server) initTemplates(dir string, ext string) {
	sn := "initTemplates"

	var allFiles []string
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		log.Fatalf("%s: ioutil.ReadDir(%q): %v\n", sn, dir, err)
	}

	// all files in specified directory with specified extension treated as templates
	for _, file := range files {
		filename := file.Name()
		if strings.HasSuffix(filename, ext) {
			allFiles = append(allFiles, filepath.Join(dir, filename))
		}
	}

	s.tmpl = template.Must(template.ParseFiles(allFiles...)) // parses all .tmpl files in the 'templates' folder
}

// ********** ********** ********** ********** ********** **********

// Config holds application-wide configuration info
type Config struct {
	path     string
	pollSecs int
	port     string
}

// Load populates Config with flag and environment variable values
func (c *Config) Load() {

	pflag.StringVar(&c.path, "path", "./", "path to folder containing timelapse.json")
	pflag.IntVar(&c.pollSecs, "poll", 60, "seconds between time checks")
	pflag.StringVar(&c.port, "port", "8099", "HTTP port to listen on")
	var help bool
	pflag.BoolVarP(&help, "help", "h", false, "show usage information")
	pflag.Parse()

	if help {
		pflag.PrintDefaults()
		os.Exit(0)
	}

	viper.BindPFlag("path", pflag.Lookup("path"))
	viper.BindPFlag("poll", pflag.Lookup("poll"))
	viper.BindPFlag("port", pflag.Lookup("port"))

	viper.SetEnvPrefix("timelapse")
	viper.AutomaticEnv()
	viper.BindEnv("path") // treats as upper-cased SerEnvPrefix value + "_" + upper-cased "path" (BindEnv argument)
	viper.BindEnv("poll")
	viper.BindEnv("port")

	c.path = viper.GetString("path")
	c.pollSecs = viper.GetInt("poll")
	c.port = viper.GetString("port")

	log.Printf("Config: %+v\n", c)
}

// ********** ********** ********** ********** ********** **********

// TLDef represents a Timelapse capture definition
type TLDef struct {
	Name         string      `json:"name" formam:"name"`                                              // Friendly name of this timelapse definition
	URL          string      `json:"webcamUrl" formam:"webcamUrl" validate:"url,required"`            // URL of webcam image
	FirstTime    bool        `json:"firstTime" formam:"firstTime"`                                    // First capture at specific time
	FirstSunrise bool        `json:"firstSunrise" formam:"firstSunrise"`                              // First capture at "Sunrise + offset"
	LastTime     bool        `json:"lastTime" formam:"lastTime"`                                      // Last capture at specific time
	LastSunset   bool        `json:"lastSunset" formam:"lastSunset"`                                  // Last capture at "Sunset - offset"
	Additional   int         `json:"additional" formam:"additional" validate:"min=0,max=16,required"` // Additional captures per day (in addition to First and Last)
	FolderPath   string      `json:"folder" formam:"folder" validate:"dir,required"`                  // Folder path to store captures
	Latitude     float64     `json:"latitude" formam:"latitude" validate:"latitude,required"`         // Latitude of webcam
	Longitude    float64     `json:"longitude" formam:"longitude" validate:"longitude,required"`      // Longitude of webcam
	CaptureTimes []time.Time `json:"captureTimes"`                                                    // Times (in time zone where the code is running) to capture images
}

// SetCaptureTimes calculate all capture times
func (tld *TLDef) SetCaptureTimes() error {
	return nil
}

// IsTimeForCapture determines if it's time to capture an image
func (tld TLDef) IsTimeForCapture() bool {
	// if within some delta of capture time, return true
	// else
	return false
}

type masterTLDefs []TLDef

// newMasterTLDefs returns a new (empty) master timelapse definition object
func newMasterTLDefs() *masterTLDefs {
	var mtld = new(masterTLDefs)

	log.Printf("newMasterTLDefs, %p, %+v", mtld, mtld)
	return mtld
}

// Read reads the master timelapse definitions file into the masterTLDefs
// slice of its receiver
func (mtld masterTLDefs) Read(path string) error {
	sn := "mtld.Read"

	data, err := ioutil.ReadFile(path)
	if err != nil {
		log.Printf("%s, ioutil.ReadFile: %v\n", sn, err)
		return err
	}
	if len(data) == 0 {
		return fmt.Errorf("master file empty")
	}
	log.Printf("mtld.Read, contents of %s: %q\n", masterFile, data)

	err = json.Unmarshal(data, srv.mtld)
	if err != nil {
		log.Printf("%s, json.Unmarshal: %v\n", sn, err)
		return err
	}

	mtldSlice := *srv.mtld
	for i, tld := range mtldSlice { // validate the TLDef structs within mtld
		if err := srv.validate.Struct(tld); err != nil {
			log.Printf("%s, validate.Struct, element %d: %v\n", sn, i, err)
			return err
		}
	}

	log.Printf("mtld.Read, %+v", mtld)
	return nil
}

// Write writes the masterTLDefs to the master timelapse definitions file
func (mtld masterTLDefs) Write() error {
	sn := "mtld.Write"

	var buf []byte
	var err error

	if buf, err = json.Marshal(mtld); err != nil {
		log.Printf("%s, json.Marshal: %v\n", sn, err)
		return err
	}

	filename := filepath.Join(masterPath, masterFile)
	if err = ioutil.WriteFile(filename, buf, 0644); err != nil { // -rw-r--r--
		log.Printf("%s, ioutil.WriteFile: %v\n", sn, err)
		return err
	}

	log.Printf("mtld.Write, %s", filename)
	return nil
}

// Append appends a timelapse definition to the masterTLDefs slice
func (mtld *masterTLDefs) Append(newTLD *TLDef) error {
	*mtld = append(*mtld, *newTLD)

	// log.Printf("mtld.Add, %+v", mtld)
	return nil
}
