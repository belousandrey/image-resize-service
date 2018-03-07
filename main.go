package main //package main

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/ReneKroon/ttlcache"
	"github.com/pkg/errors"
	"github.com/spf13/pflag"
)

var allowedContentTypes = map[string]bool{"image/jpeg": true}

// resizeHandler is a struct to serve resize handler
type resizeHandler struct {
	cache      *ttlcache.Cache
	ttl        int
	reg        Registry
	imager     Imager
	downloader Downloader
	logger     *log.Logger
}

// ServeHTTP passes request to ResizeHandler and logs results
func (fh *resizeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	status, err := ResizeHandler(w, r, fh.cache, fh.ttl, fh.reg, fh.imager, fh.downloader)
	if err != nil {
		fh.logger.SetPrefix("ERROR: ")
		fh.logger.Println("status:", status, "| ", err.Error())
	} else {
		fh.logger.SetPrefix("INFO: ")
		fh.logger.Println("status: ", status, "| resized image from "+strings.ToLower(r.Form.Get("url")))
	}
}

// ResizeHandler covers all routine with file download, image conversion and resize, client responses
func ResizeHandler(w http.ResponseWriter, r *http.Request, c *ttlcache.Cache, ttl int, reg Registry, i Imager, d Downloader) (int, error) {
	fx := NewImageFixture()

	if r.Method != http.MethodGet {
		return fx.respondWithError(w, http.StatusMethodNotAllowed, errors.New("only GET method allowed"))
	}

	err := fx.getParamsFromRequest(w, r)
	if err != nil {
		return fx.respondWithError(w, http.StatusBadRequest, err)
	}

	if fx.upToDate(r, c) {
		return fx.respondWithRedirect(w)
	}

	exists := fx.GetFromCache(c)
	if !exists {
		fx.File.Path, err = d.StoreFileToTemp(fx.Params.URL)
		if err != nil {
			return fx.respondWithError(w, http.StatusInternalServerError, err)
		}

		fx.SetToCache(c, reg)
	} else {
		resized, existsResized := fx.FindInCache(c)
		if existsResized {
			b, err := ioutil.ReadFile(resized)
			if err == nil {
				buffer := new(bytes.Buffer)
				buffer.Write(b)
				return fx.respondWithImage(w, buffer, fx.Params.URL, ttl)
			}
		}
	}

	fx.File.Handler, err = i.Open(fx.File.Path)
	if err != nil {
		return fx.respondWithError(w, http.StatusInternalServerError, err)
	}
	defer fx.File.Handler.Close()

	err = fx.checkFileContentType(allowedContentTypes)
	if err != nil {
		return fx.respondWithError(w, http.StatusBadRequest, err)
	}

	err = i.Decode(fx.File.Handler)
	if err != nil {
		return fx.respondWithError(w, http.StatusInternalServerError, err)
	}

	i.Resize(uint(fx.Params.Width), uint(fx.Params.Height))
	resized, err := i.StoreResizedToTempFile()
	if err != nil {
		fx.respondWithError(w, http.StatusInternalServerError, err)
	}
	fx.UpdateValueInCache(c, resized, reg)

	buffer, err := i.Encode()
	if err != nil {
		return fx.respondWithError(w, http.StatusInternalServerError, err)
	}

	return fx.respondWithImage(w, buffer, fx.Params.URL, ttl)
}

// formHandler is simple struct to serve form for image resize
type formHandler struct {
	port int
}

// ServeHTTP loads page with simple form for image resize
func (fh *formHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("tmpl/upload.html")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	t.Execute(w, fh.port)
}

func main() {
	port, ttl := readFlags()

	logger := log.New(os.Stdout, "", log.LstdFlags)

	// key-value storage with expiring keys
	cache := ttlcache.NewCache()
	cache.SetTTL(time.Second * time.Duration(ttl))

	// storage for all generated temp files
	registry := NewRegistry()

	// chan to capture SIGTERM
	signals := make(chan os.Signal, 1)

	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	go func(signals <-chan os.Signal, reg Registry) {
		<-signals
		exitCode := 0
		err := reg.Cleanup()
		if err != nil {
			fmt.Println("remove temp files: ", err.Error())
			exitCode = 1
		}

		os.Exit(exitCode)
	}(signals, registry)

	mux := http.NewServeMux()
	mux.Handle("/", &formHandler{port: port})
	mux.Handle("/upload", &resizeHandler{cache: cache, ttl: ttl, reg: registry, imager: NewImager(), downloader: NewDownloader(), logger: logger})

	fmt.Println("Listening on http://localhost:" + strconv.Itoa(port))
	http.ListenAndServe(":"+strconv.Itoa(port), mux)
}

func readFlags() (port, ttl int) {
	pflag.IntVarP(&port, "port", "p", 8080, "system port number")
	pflag.IntVarP(&ttl, "ttl", "t", 3600, "image cache in seconds")
	pflag.Parse()

	return
}
