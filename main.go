package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
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

const (
	masterPath = "/Users/peterplamondon/Downloads/timelapse/"
	masterFile = "timelapse.json"
	timeLayout = "3:04:05 PM" // see https://godoc.org/time#Time.Format and https://ednsquare.com/story/date-and-time-manipulation-golang-with-examples------cU1FjK
)

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

	var wg sync.WaitGroup
	for i, tld := range *srv.mtld {
		log.Printf("%s, launching goroutine #%d, %s, poll interval: %d", sn, i, tld.Name, srv.config.pollSecs)
		wg.Add(1)
		go func(ctx context.Context, tld TLDef, pollInterval int) {
			log.Printf("goroutine handling TLDef %s (%p)\n", tld.Name, &tld)

			tld.SetCaptureTimes(time.Now()) // calculate all capture times for today
			tld.UpdateIndexOfNext()         // determine which capture time comes next

			for {
				select {
				case <-ctx.Done():
					log.Printf("goroutine handling TLDef %s exiting after ctx.Done\n", tld.Name)
					wg.Done()
					return
				default:
					if tld.IsTimeForCapture() {
						// capture and store the image
						log.Printf("time for TLD %s capture\n", tld.Name)
						tld.UpdateIndexOfNext()
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

// handleHome is the handler for "/"
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

// handlenew is the handler for webform submissions with new TLDef specifications
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

		tld := newTLDef()

		decoder := formam.NewDecoder(nil)
		if err := decoder.Decode(r.Form, tld); err != nil {
			log.Printf("%s, decoder.Decode: %v\n", sn, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		// validate the TLDef we just decoded
		if err := srv.validate.Struct(tld); err != nil {
			log.Printf("%s, handleNew: %v\n", sn, err)
			http.Error(w, err.Error(), http.StatusBadRequest)
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

		srv.mtld.Append(tld)
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
	viper.BindEnv("path") // treats as upper-cased SetEnvPrefix value + "_" + upper-cased "path"
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
	Latitude     float64     `json:"latitude" formam:"latitude" validate:"latitude,required"`         // Latitude of webcam
	Longitude    float64     `json:"longitude" formam:"longitude" validate:"longitude,required"`      // Longitude of webcam
	FirstTime    bool        `json:"firstTime" formam:"firstTime"`                                    // First capture at specific time
	FirstSunrise bool        `json:"firstSunrise" formam:"firstSunrise"`                              // First capture at "Sunrise + offset"
	LastTime     bool        `json:"lastTime" formam:"lastTime"`                                      // Last capture at specific time
	LastSunset   bool        `json:"lastSunset" formam:"lastSunset"`                                  // Last capture at "Sunset - offset"
	Additional   int         `json:"additional" formam:"additional" validate:"min=0,max=16,required"` // Additional captures per day (in addition to First and Last)
	FolderPath   string      `json:"folder" formam:"folder" validate:"dir,required"`                  // Folder path to store captures
	TZWebcam     string      `json:"-"`                                                               // timezone of the webcam (e.g., "America/Los_Angeles")
	SunriseUTC   time.Time   `json:"-"`                                                               // sunrise at webcam lat/long (UTC)
	SolarNoonUTC time.Time   `json:"-"`                                                               // solar noon at webcam lat/long (UTC)
	SunsetUTC    time.Time   `json:"-"`                                                               // sunset at webcam lat/long (UTC)
	CaptureTimes []time.Time `json:"-"`                                                               // Times (in time zone where the code is running) to capture images
	IndexOfNext  int         `json:"-"`                                                               // index in CaptureTimes[] of next (future) capture time
}

// newTLDef initializes a TLDef structure
func newTLDef() *TLDef {
	tld := TLDef{}
	tld.CaptureTimes = []time.Time{} // prefer an empty slice so json.Marshal() will marshall to produces "[]"

	return &tld
}

// SetCaptureTimes calculate all capture times for the specified date
func (tld *TLDef) SetCaptureTimes(date time.Time) error {
	sn := "main.TLDef.SetCaptureTimes"
	var err error

	log.Printf("%s, %s, date: %v\n", sn, tld.Name, date)

	if err = tld.SetSolar(date); err != nil { // set sunrise, solar noon, and sunset for specified date
		log.Printf("%s, %s, tld.SetSolar: %v\n", sn, tld.Name, err)
		return err
	}

	tld.CaptureTimes = append(tld.CaptureTimes, TimeToSecond(date)) // HACK

	// TODO: get timezone of lat/long
	// TODO: if selected, get sunrise and sunset times at lat/long [for specified date]
	// TODO: calculate and append to CaptureTimes the local time [for specified date] corresponding to first capture time at webcam
	// TODO: if additional > 0, calculate the span between additional captures (i.e., split up span from first to last in "additional" segments)
	// TODO: for ... calculate and append local capture time [for specified date] for next span
	// TODO: calculate and append to CaptureTimes the local time [for specified date] corresponding to last capture time at webcam

	// TODO: sort slice's capture times into increasing order; superfluous if logic above is correct

	log.Printf("%s, %s, tld.CaptureTimes: %+v\n", sn, tld.Name, tld.CaptureTimes)
	return nil
}

// UpdateIndexOfNext increments IndexOfNext to reference the next
// CaptureTime element, or after today's captures have been performed,
// updates CaptureTimes with tomorrow's capture times
func (tld TLDef) UpdateIndexOfNext() {
	sn := "main.tld.UpdateIndexOfNext"

	log.Printf("%s, %s, enter IndexOfNext: %d\n", sn, tld.Name, tld.IndexOfNext)

	if (tld.IndexOfNext + 1) > len(tld.CaptureTimes)-1 { // fell off the end of the slice = processed all of today's captures
		tomorrow := time.Now().AddDate(0, 0, 1)
		tld.SetCaptureTimes(tomorrow) // setup tomorrow's capture times
		tld.IndexOfNext = 0           // next capture time is tomorrow's first time
		log.Printf("%s, %s, reset CaptureTimes to tomorrow; tld.IndexOfNext: %d\n", sn, tld.Name, tld.IndexOfNext)
		return
	}

	current := tld.CaptureTimes[tld.IndexOfNext]
	next := tld.CaptureTimes[tld.IndexOfNext+1]
	if !next.After(current) { // next element expected to be after (later than) current element
		msg := fmt.Sprintf("%s, next entry NOT after current entry, current: %s, next %s\n", sn, current, next)
		panic(msg)
	}

	tld.IndexOfNext++
	log.Printf("%s, %s, tld.IndexOfNext: %d\n", sn, tld.Name, tld.IndexOfNext)
}

// NextCaptureTime returns the time of the next capture
func (tld TLDef) NextCaptureTime() time.Time {
	next := tld.CaptureTimes[tld.IndexOfNext]
	return next
}

// const Margin = 100 * time.Millisecond

// IsTimeForCapture determines if it's time to capture an image
func (tld TLDef) IsTimeForCapture() bool {
	b := time.Now().After(tld.NextCaptureTime())
	return b
}

// TimeToSecond returns the provided time rounded down to an even second
func TimeToSecond(t time.Time) time.Time {
	// sn := "timelapse.TimeToSecond"

	toSecond := t.Truncate(time.Second)
	// layout := "Jan 2 2006 15:04:05 -0700 MST"
	// log.Printf("%s, parameter: %s, return value: %s\n", sn, t.Format(time.RFC3339Nano), toSecond.Format(layout))
	return toSecond
}

// ********** ********** ********** ********** ********** **********

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
	// log.Printf("mtld.Read, contents of %s: %q\n", masterFile, data)

	err = json.Unmarshal(data, srv.mtld)
	if err != nil {
		log.Printf("%s, json.Unmarshal: %v\n", sn, err)
		return err
	}

	// validate the TLDefs we just read
	mtldSlice := *srv.mtld
	for i, tld := range mtldSlice { // validate the TLDef structs within mtld
		if err := srv.validate.Struct(tld); err != nil {
			log.Printf("%s, validate.Struct, element %d (%s): %v\n", sn, i, tld.Name, err)
			return err
		}
	}

	return nil
}

// Write writes the masterTLDefs to the master timelapse definitions file
func (mtld masterTLDefs) Write() error {
	sn := "mtld.Write"

	var buf []byte
	var err error

	// validate the TLDefs in mtld before writing them
	mtldSlice := mtld
	for i, tld := range mtldSlice { // validate the TLDef structs within mtld
		if err := srv.validate.Struct(tld); err != nil {
			log.Printf("%s, validate.Struct, element %d (%s): %v\n", sn, i, tld.Name, err)
			return err
		}
	}

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

// ********** ********** ********** ********** ********** **********

// SSDayInfo holds times of sunrise, sunset and twilight based on date, latitude, and longitude
type SSDayInfo struct { // all times are UTC
	linkTLDef                 *TLDef // link to associated TLDef
	Date                      time.Time
	Latitude                  float64 `json:"-" validate:"latitude,required"`  // Latitude of webcam
	Longitude                 float64 `json:"-" validate:"longitude,required"` // Longitude of webcam
	TZName                    string
	TZLoc                     time.Location
	SSDISunrise               string `json:"sunrise"`
	SSDISunset                string `json:"sunset"`
	SSDISolarNoon             string `json:"solar_noon"`
	DayLength                 string `json:"day_length"`
	CivilTwilightBegin        string `json:"civil_twilight_begin"`
	CivilTwilightEnd          string `json:"civil_twilight_end"`
	NauticalTwilightBegin     string `json:"nautical_twilight_begin"`
	NauticalTwilightEnd       string `json:"nautical_twilight_end"`
	AstronomicalTwilightBegin string `json:"astronomical_twilight_begin"`
	AstronomicalTwilightEnd   string `json:"astronomical_twilight_end"`
}

// Solar holds sun-related times
type Solar struct {
	SunriseUTC   time.Time
	SolarNoonUTC time.Time
	SunsetUTC    time.Time
}

// SetSolar updates the TLDef with sunrise, solar noon, and sunset
// times (UTC) on the specified date and latitude/longitude
func (tld *TLDef) SetSolar(t time.Time) error {
	sn := "main.tld.SetSolar"

	log.Printf("%s, %s, enter, t: %v\n", sn, tld.Name, t)

	ssdi := NewSSDayInfo(tld)
	ssdi.Date = t

	query := ssdi.buildQuery()
	req, err := http.NewRequest("GET", query, nil)
	if err != nil {
		log.Printf("%s, %s, http.NewRequest: %v\n", sn, tld.Name, err)
		return err
	}

	client := &http.Client{Timeout: time.Second * 2}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("%s, %s, http.Client.Do: %v", sn, tld.Name, err)
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("%s, %s, ioutil.ReadAll: %v", sn, tld.Name, err)
		return err
	}
	log.Printf("%s, %s, body: %s", sn, tld.Name, body)

	// strip off outer structure
	wrapperStart := []byte(`{"results":`)
	wrapperEnd := []byte(`,"status":"OK"}`)
	if bytes.Index(body, wrapperStart) >= 0 {
		tmpBody := bytes.TrimPrefix(body, wrapperStart)
		body = bytes.TrimSuffix(tmpBody, wrapperEnd)
	}
	// log.Printf("%s, trimmed body: %s\n", sn, body)

	if err := json.Unmarshal(body, &ssdi); err != nil { // unmarshall all provided fields
		log.Printf("%s, %s, json.Unmarshal: %v", sn, tld.Name, err)
		return err
	}

	// TODO: incorporate passed-in date into ssdi.* times before parsing

	if tld.SunriseUTC, err = time.Parse(timeLayout, ssdi.SSDISunrise); err != nil {
		log.Printf("%s, %s, time.Parse(%s): %v", sn, tld.Name, ssdi.SSDISunrise, err)
		return err
	}
	log.Printf("%s, %s, SSDISunrise: %s, s.SunriseUTC: %v\n", sn, tld.Name, ssdi.SSDISunrise, tld.SunriseUTC)

	if tld.SolarNoonUTC, err = time.Parse(timeLayout, ssdi.SSDISolarNoon); err != nil {
		log.Printf("%s, %s, time.Parse(%s): %v", sn, tld.Name, ssdi.SSDISolarNoon, err)
		return err
	}
	if tld.SunsetUTC, err = time.Parse(timeLayout, ssdi.SSDISunset); err != nil {
		log.Printf("%s, %s. time.Parse(%s): %v", sn, tld.Name, ssdi.SSDISunset, err)
		return err
	}

	log.Printf("%s, %s (%p), SunriseUTC: %v, SolarNoonUTC: %v, SunsetUTC: %v\n", sn, tld.Name, tld, tld.SunriseUTC, tld.SolarNoonUTC, tld.SunsetUTC)
	return nil
}

// NewSSDayInfo returns an initialized SSDayInfo struct
func NewSSDayInfo(tld *TLDef) *SSDayInfo {
	ssdi := &SSDayInfo{}
	ssdi.linkTLDef = tld
	ssdi.Latitude = tld.Latitude
	ssdi.Longitude = tld.Longitude

	return ssdi
}

// Sunrise returns the local time of sunrise retrieves data to populate SSDayInfo
func (ssdi *SSDayInfo) Sunrise() (time.Time, error) {
	sn := "main.SSDayInfo.Sunrise"

	localLoc, err := time.LoadLocation("Local") // timezone where this code is running
	if err != nil {
		msg := fmt.Sprintf("%s, time.LoadLocation(\"Local\"): %v", sn, err)
		panic(msg)
	}

	utcSunrise, err := time.Parse(timeLayout, ssdi.SSDISunrise) // time in UTC
	if err != nil {
		log.Printf("%s, time.Parse(%s): %v", sn, ssdi.SSDISunrise, err)
		return time.Time{}, err
	}

	localSunrise := utcSunrise.In(localLoc) // local time

	log.Printf("%s, sunrise %s local, %s UTC\n", sn, localSunrise, utcSunrise)
	return localSunrise, nil
}

// buildQuery returns the query string to get data for SSDayInfo
func (ssdi SSDayInfo) buildQuery() string {
	sn := "main.SSDayInfo.buildQuery"

	queryParams := url.Values{}

	queryParams.Add("lat", fmt.Sprintf("%.7f", ssdi.linkTLDef.Latitude))
	queryParams.Add("lng", fmt.Sprintf("%.7f", ssdi.linkTLDef.Longitude))

	year, month, day := ssdi.Date.Date()
	date := fmt.Sprintf("%4d-%02d-%02d", year, month, day)
	queryParams.Add("date", date)

	query := "https://api.sunrise-sunset.org/json?"
	query += queryParams.Encode()

	log.Printf("%s, query: %q\n", sn, query)
	return query
}
