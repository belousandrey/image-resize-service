package main

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/ReneKroon/ttlcache"
	"github.com/pkg/errors"
	"github.com/spf13/pflag"
)

var allowedContentTypes = map[string]bool{"image/jpeg": true}

// ResizeHandler covers all routine with file download, image conversion and resize, client responses
func ResizeHandler(c *ttlcache.Cache, ttl int, reg Registry, i Imager, d Downloader) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		fx := NewImageFixture()

		if r.Method != "GET" {
			fx.respondWithError(w, http.StatusMethodNotAllowed, errors.New("only GET method allowed"))
			return
		}

		err := fx.getParamsFromRequest(w, r)
		if err != nil {
			fx.respondWithError(w, http.StatusBadRequest, err)
			return
		}

		if fx.upToDate(r, c) {
			fx.respondWithRedirect(w)
			return
		}

		exists := fx.GetFromCache(c)
		if !exists {
			fx.File.Path, err = d.StoreFileToTemp(fx.Params.URL)
			if err != nil {
				fx.respondWithError(w, http.StatusInternalServerError, err)
				return
			}

			fx.SetToCache(c, reg)
		} else {
			resized, existsResized := fx.FindInCache(c)
			if existsResized {
				b, err := ioutil.ReadFile(resized)
				if err == nil {
					buffer := new(bytes.Buffer)
					buffer.Write(b)
					fx.respondWithImage(w, buffer, fx.Params.URL, ttl)
					return
				}
			}
		}

		fx.File.Handler, err = i.Open(fx.File.Path)
		if err != nil {
			fx.respondWithError(w, http.StatusInternalServerError, err)
			return
		}
		defer fx.File.Handler.Close()

		err = fx.checkFileContentType(allowedContentTypes)
		if err != nil {
			fx.respondWithError(w, http.StatusBadRequest, err)
			return
		}

		err = i.Decode(fx.File.Handler)
		if err != nil {
			fx.respondWithError(w, http.StatusInternalServerError, err)
			return
		}

		i.Resize(uint(fx.Params.Width), uint(fx.Params.Height))
		resized, err := i.StoreResizedToTempFile()
		if err != nil {
			fx.respondWithError(w, http.StatusInternalServerError, err)
		}
		fx.UpdateValueInCache(c, resized, reg)

		buffer, err := i.Encode()
		if err != nil {
			fx.respondWithError(w, http.StatusInternalServerError, err)
			return
		}

		fx.respondWithImage(w, buffer, fx.Params.URL, ttl)
	}
}

// FormHandler loads page with simple form for image resize
func FormHandler(p int) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		t, err := template.ParseFiles("tmpl/upload.html")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		t.Execute(w, p)
	}
}

func main() {
	port, ttl := readFlags()

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
		err := reg.Cleanup()
		if err != nil {
			fmt.Println("remove temp files: ", err.Error())
		}

		os.Exit(1)
	}(signals, registry)

	fmt.Println("Listening on http://localhost:" + strconv.Itoa(port))
	http.HandleFunc("/", FormHandler(port))
	http.HandleFunc("/upload", ResizeHandler(cache, ttl, registry, NewImager(), NewDownloader()))
	http.ListenAndServe(":"+strconv.Itoa(port), nil)
}

func readFlags() (port, ttl int) {
	pflag.IntVarP(&port, "port", "p", 8080, "system port number")
	pflag.IntVarP(&ttl, "ttl", "t", 3600, "image cache in seconds")
	pflag.Parse()

	return
}
