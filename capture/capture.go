// Package capture retrieves webcam images
package capture

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/c2h5oh/datasize"
	"github.com/colinplamondon/thebeam/beam-server-go/pkg/jsonBody"
	"github.com/go-playground/validator"
)

var validate *validator.Validate // use a single instance of Validate, it caches struct info

// TODO: if Cloud Function, add Error Reporting Client to publish errors to StackDriver Error Reporting https://cloud.google.com/error-reporting/docs/reference/libraries#client-libraries-usage-go
// TODO: if Cloud Function, init global variables which *may* survive if execution environment is recycled

// WebcamImage is an Google Cloud Function that retrieves a webcam image
// from the specified URL.
func WebcamImage(w http.ResponseWriter, r *http.Request) {

	// endpoint := r.URL.Path
	// if endpoint != "/" { // only the root endpoint is supported
	// 	// log.Printf("unsupported path: %q\n", endpoint)
	// 	w.WriteHeader(http.StatusNotFound)
	// 	return
	// }

	var wcrIn = WebcamRequest{}

	if err := jsonBody.Decode(w, r, &wcrIn); err != nil { // decode incoming webcam request into WebcamRequest struct
		log.Printf("jsonBody.Decode: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	validate = validator.New()
	if err := validate.Struct(wcrIn); err != nil { // validate the struct contents
		log.Printf("validation error: %v\n", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	query := wcrIn.URL
	webcamReq, err := http.NewRequest("GET", query, nil)
	if err != nil {
		msg := fmt.Sprintf("http.NewRequest: %v", err)
		log.Fatal(msg)
	}

	client := &http.Client{Timeout: time.Second * 10}

	resp, err := client.Do(webcamReq)
	if err != nil {
		msg := fmt.Sprintf("client.Do: %v", err)
		log.Fatal(msg)
	}
	defer resp.Body.Close()

	// establish file path and name
	// fileDir = os.TempDir()
	// fileName = wcrIn.FileRoot + "_"

	tmpFile, err := ioutil.TempFile(os.TempDir(), wcrIn.FileRoot+"-")
	if err != nil {
		msg := fmt.Sprintf("ioutil.TempFile: %v", err)
		log.Fatal(msg)
	}
	defer tmpFile.Close()

	written, err := io.Copy(tmpFile, resp.Body) // Use io.Copy to just dump the response body to the file. This supports huge files
	if err != nil {
		msg := fmt.Sprintf("io.Copy: %v", err)
		log.Fatal(msg)
	}
	log.Printf("%s created, size %s", tmpFile.Name(), datasize.ByteSize(written).HumanReadable())

	// respBody, err := json.Marshal(something)
	// if err != nil {
	// 	http.Error(w, err.Error(), http.StatusInternalServerError)
	// 	return
	// }
	// w.Header().Set("Content-Type", "application/json")
	// w.Write(respBody)
}

// WebcamRequest holds incoming request parameters
type WebcamRequest struct {
	URL         string `json:"url,omitempty" validate:"required"`
	TimeoutSec  int    `json:"timeout,omitempty" validate:"required"`
	FilePath    string `json:"file_path,omitempty" validate:"required"`    // path to folder for saved images, e.g., gs://elated-practice-224603-kohm-yah-mah-nee
	FileRoot    string `json:"file_root,omitempty" validate:"required"`    // filename root for saved images, e.g., "webcam7" results in "webcam7_yyyyMMdd_hhmmss"
	FilePattern string `json:"file_pattern,omitempty" validate:"required"` // date/time pattern for saved images, e.g., "yyyyMMdd_hhmmss" results in "webcam7_yyyyMMdd_hhmmss"
}
