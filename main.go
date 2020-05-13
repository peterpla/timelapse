package main

import (
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
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

func main() {
	defer catch() // implements recover so panics reported
	sn := "timelapse"

	srv = newServer()
	srv.router.ServeFiles("/static/*filepath", http.Dir("static"))
	srv.router.GET("/new", srv.handleNew())
	srv.router.GET("/", srv.handleHome())

	// TODO: #4 read Timelapse Definitions (TLDef) master list
	// TODO: #5 create Go routine to handle each TLDef
	// TODO: #6 Go routine cleanup on SIGTERM, etc.

	initTemplates(&srv.tmpl, "./templates", ".html")

	// TODO: #7 handler for webform to enter new TLDef
	// TODO: #8 handler on webform submit adds new TLDef to master list

	hs := http.Server{
		Addr:         ":" + srv.config.port,
		Handler:      middleware.LogReqResp(srv.router), // mux
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Printf("Starting service %s listening on port %s", sn, srv.config.port)

	go startListening(&hs, "main") // call ListenAndServe from a separate go routine so main can listen for signals

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	s := <-signals
	log.Printf("\n%s received signal %s, terminating", sn, s.String())

}

// ********** ********** ********** ********** ********** **********

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

func initTemplates(t **template.Template, dir string, ext string) {
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

	*t = template.Must(template.ParseFiles(allFiles...)) // parses all .tmpl files in the 'templates' folder
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

		tld := &TLDef{}

		decoder := formam.NewDecoder(nil)
		if err := decoder.Decode(r.Form, tld); err != nil {
			log.Printf("decoder.Decode: %v\n", err)
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

		// log.Printf("handleNew, TLDef: %+v", tld)

		srv.mtld.Add(tld)

		http.Redirect(w, r, "/", http.StatusSeeOther)

		// log.Printf("%s.%s, duration %v\n", sn, mn, time.Now().Sub(startTime))
		return
	}
}

// ********** ********** ********** ********** ********** **********

// Config holds application-wide configuration info
type Config struct {
	path string
	port string
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

// ********** ********** ********** ********** ********** **********

// TLDef represents a Timelapse capture definition
type TLDef struct {
	Name         string `json:"name" formam:"name" validate:"required"`                 // Friendly name of this timelapse definition
	URL          string `json:"webcamUrl" formam:"webcamUrl" validate:"required"`       // URL of webcam image
	FirstTime    bool   `json:"firstTime" formam:"firstTime" validate:"required"`       // First capture at specific time
	FirstSunrise bool   `json:"firstSunrise" formam:"firstSunrise" validate:"required"` // First capture at "Sunrise + offset"
	LastTime     bool   `json:"lastTime" formam:"lastTime" validate:"required"`         // Last capture at specific time
	LastSunset   bool   `json:"lastSunset" formam:"lastSunset" validate:"required"`     // Last capture at "Sunset - offset"
	Additional   int    `json:"additional" formam:"additional" validate:"required"`     // Additional captures per day (in addition to First and Last)
	FolderPath   string `json:"folder" formam:"folder" validate:"required"`             // Folder path to store captures
}

type masterTLDefs []*TLDef

// NewMasterTLDefs returns a new (empty) master timelapse definition object
func newMasterTLDefs() *masterTLDefs {
	var mtld = new(masterTLDefs)
	log.Printf("newMasterTLDefs, mtld: %p, %+v\n", mtld, *mtld)
	return mtld
}

func (mtld *masterTLDefs) Add(newTLD *TLDef) error {
	*mtld = append(*mtld, newTLD)

	// msg := fmt.Sprintf("mtld.Add after append, mtld: %p\n", mtld)
	// for i, tld := range *mtld {
	// 	msg = msg + fmt.Sprintf("[%d] %+v\n", i, tld)
	// }
	// log.Print(msg)

	return nil
}
