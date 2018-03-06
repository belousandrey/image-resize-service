package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/ReneKroon/ttlcache"
	"github.com/belousandrey/image-resize-service/mock"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const URL = "/upload"

func TestResizeHandler(t *testing.T) {
	cache := ttlcache.NewCache()
	reg := NewRegistry()
	ttl := 60

	logger := log.New(ioutil.Discard, "", 0)

	t.Run("wrong method", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		rec := httptest.NewRecorder()
		req := httptest.NewRequest("POST", URL, nil)
		handler := &resizeHandler{cache, ttl, reg, mock.NewMockImager(ctrl), mock.NewMockDownloader(ctrl), logger}
		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
	})
	t.Run("negative width", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", URL, nil)
		req.Form = url.Values{"url": {"http://example.com/image.jpg"}, "width": {"-100"}, "height": {"100"}}

		handler := &resizeHandler{cache, ttl, reg, mock.NewMockImager(ctrl), mock.NewMockDownloader(ctrl), logger}
		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
	t.Run("negative height", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", URL, nil)
		req.Form = url.Values{"url": {"http://example.com/image.jpg"}, "width": {"100"}, "height": {"-100"}}

		handler := &resizeHandler{cache, ttl, reg, mock.NewMockImager(ctrl), mock.NewMockDownloader(ctrl), logger}
		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
	t.Run("bad image URL", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", URL, nil)
		req.Form = url.Values{"url": {"wrong URL"}, "width": {"100"}, "height": {"100"}}

		handler := &resizeHandler{cache, ttl, reg, mock.NewMockImager(ctrl), mock.NewMockDownloader(ctrl), logger}
		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
	t.Run("wrong content type", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		imageLocation := "https://golang.org/gopher.png"
		original := "testdata/wrong_content_type.png"
		width, height := 100, 100

		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", URL, nil)
		req.Form = url.Values{"url": {imageLocation}, "width": {strconv.Itoa(width)}, "height": {strconv.Itoa(height)}}

		downloader := mock.NewMockDownloader(ctrl)
		downloader.EXPECT().StoreFileToTemp(imageLocation).Return(original, nil).Times(1)

		fh, err := os.Open(original)
		require.NoError(t, err)

		imager := mock.NewMockImager(ctrl)
		imager.EXPECT().Open(original).Return(fh, nil).Times(1)

		handler := &resizeHandler{cache, ttl, reg, imager, downloader, logger}
		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		imageLocation := "https://golang.org/gopher.jpg"
		original, resized := "testdata/gopher.original.jpg", "testdata/gopher.100.100.jpg"
		width, height := 100, 100

		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", URL, nil)
		req.Form = url.Values{"url": {imageLocation}, "width": {strconv.Itoa(width)}, "height": {strconv.Itoa(height)}}

		downloader := mock.NewMockDownloader(ctrl)
		downloader.EXPECT().StoreFileToTemp(imageLocation).Return(original, nil).Times(1)

		fh, err := os.Open(original)
		require.NoError(t, err)

		imager := mock.NewMockImager(ctrl)
		imager.EXPECT().Open(original).Return(fh, nil).Times(1)
		imager.EXPECT().Decode(fh).Return(nil).Times(1)
		imager.EXPECT().Resize(uint(width), uint(height)).Times(1)
		imager.EXPECT().StoreResizedToTempFile().Return(resized, nil).Times(1)

		b, err := ioutil.ReadFile(resized)
		require.NoError(t, err)

		buffer := new(bytes.Buffer)
		buffer.Write(b)
		imager.EXPECT().Encode().Return(buffer, nil).Times(1)

		handler := &resizeHandler{cache, ttl, reg, imager, downloader, logger}
		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "image/jpeg", rec.Header().Get("Content-Type"))
		assert.Equal(t, "3609", rec.Header().Get("Content-Length"))
		assert.NotEmpty(t, rec.Header().Get("Etag"))
		assert.Equal(t, fmt.Sprintf("max-age:%d, public", ttl), rec.Header().Get("Cache-Control"))

		lm, err := time.Parse(time.RFC1123, rec.Header().Get("Last-Modified"))
		require.NoError(t, err)
		assert.WithinDuration(t, time.Now(), lm, time.Duration(24*time.Hour)) // because of timezones
		exp, err := time.Parse(time.RFC1123, rec.Header().Get("Expires"))
		require.NoError(t, err)
		assert.WithinDuration(t, time.Now(), exp, time.Duration(24*time.Hour)) // because of timezones
	})
	t.Run("not modified", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		imageLocation := "https://golang.org/gopher.jpg"
		width, height := 100, 100

		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", URL, nil)
		req.Form = url.Values{"url": {imageLocation}, "width": {strconv.Itoa(width)}, "height": {strconv.Itoa(height)}}
		req.Header.Set("If-None-Match", "70c8cb786769432edd9f1cd55cf1b135")

		handler := &resizeHandler{cache, ttl, reg, mock.NewMockImager(ctrl), mock.NewMockDownloader(ctrl), logger}
		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNotModified, rec.Code)
	})
}
