package main

import (
	"bytes"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

func (fx *ImageFixture) respondWithError(w http.ResponseWriter, status int, err error) (int, error) {
	w.WriteHeader(status)
	return status, err
}

func (fx *ImageFixture) respondWithImage(w http.ResponseWriter, buffer *bytes.Buffer, URL string, ttl int) (int, error) {
	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("Content-Length", strconv.Itoa(len(buffer.Bytes())))

	w.Header().Set("Etag", fx.File.Etag)
	w.Header().Set("Cache-Control", fmt.Sprintf("max-age:%d, public", ttl))
	w.Header().Set("Last-Modified", time.Now().Add(time.Second*600*-1).Format(http.TimeFormat))
	w.Header().Set("Expires", time.Now().Add(time.Second*time.Duration(ttl)).Format(http.TimeFormat))

	_, err := w.Write(buffer.Bytes())
	if err != nil {
		return fx.respondWithError(w, http.StatusInternalServerError, err)
	}

	return http.StatusOK, nil
}

func (fx *ImageFixture) respondWithRedirect(w http.ResponseWriter) (int, error) {
	w.WriteHeader(http.StatusNotModified)

	return http.StatusNotModified, nil
}
