package main

import (
	"bytes"
	"fmt"
	"image/jpeg"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"

	"syscall"

	"github.com/ReneKroon/ttlcache"
	"github.com/kr/pretty"
	"github.com/nfnt/resize"
	"github.com/pkg/errors"
	"github.com/spf13/pflag"
)

var allowedContentTypes = map[string]bool{"image/jpeg": true}

func withCache(c *ttlcache.Cache, ttl int) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		fx := NewImageFixture()

		if r.Method != "GET" {
			fx.respondWithError(w, http.StatusMethodNotAllowed, errors.New("only GET method allowed"))
			return
		}

		pretty.Println("If-Modified-Since:", r.Header.Get("If-Modified-Since"), "| If-None-Match:", r.Header.Get("If-None-Match"))

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
			fx.File.Path, err = downloadFileToTemp(w, fx.Params.URL)
			if err != nil {
				fx.respondWithError(w, http.StatusInternalServerError, err)
			}

			fx.SetToCache(c)
		}

		fx.File.Handler, err = os.Open(fx.File.Path)
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

		img, err := jpeg.Decode(fx.File.Handler)
		if err != nil {
			fx.respondWithError(w, http.StatusInternalServerError, err)
			return
		}

		resized := resize.Resize(uint(fx.Params.Width), uint(fx.Params.Height), img, resize.Lanczos3)
		resizedFile, err := ioutil.TempFile("", "")
		jpeg.Encode(resizedFile, resized, &jpeg.Options{jpeg.DefaultQuality})
		fx.UpdateValueInCache(c, resizedFile.Name())

		buffer := new(bytes.Buffer)
		err = jpeg.Encode(buffer, resized, nil)
		if err != nil {
			fx.respondWithError(w, http.StatusInternalServerError, err)
			return
		}

		fx.respondWithImage(w, buffer, fx.Params.URL, ttl)
	}
}

//func formHandler(p int) func(http.ResponseWriter, *http.Request) {
//	return func(w http.ResponseWriter, r *http.Request) {
//		t, err := template.ParseFiles("tmpl/upload.html")
//		if err != nil {
//			respondWithError(w, http.StatusInternalServerError, err)
//		}
//
//		t.Execute(w, p)
//	}
//}

func main() {
	port, ttl := readFlags()

	cache := ttlcache.NewCache()
	cache.SetTTL(time.Second * time.Duration(ttl))
	cache.SetWithTTL(registry, make([]string, 0), ttlcache.ItemNotExpire)

	// chan to capture SIGTERM
	signals := make(chan os.Signal, 1)

	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	go func(signals <-chan os.Signal) {
		s := <-signals
		pretty.Println("CAUGHT SIGNAL:", s)
		err := cleanTempFiles(cache)
		if err != nil {
			log.Println("remove temp files: ", err.Error())
		}

		os.Exit(1)
	}(signals)

	fmt.Println("Listening on http://localhost:" + strconv.Itoa(port))
	//http.HandleFunc("/", formHandler(port))
	http.HandleFunc("/upload", withCache(cache, ttl))
	http.ListenAndServe(":"+strconv.Itoa(port), nil)
}

func readFlags() (port, ttl int) {
	pflag.IntVarP(&port, "port", "p", 8080, "system port number")
	pflag.IntVarP(&ttl, "ttl", "t", 3600, "image cache in seconds")
	pflag.Parse()

	return
}
