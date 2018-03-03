package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"image"
	"net/http"
	"os"
)

type ImageFixture struct {
	Params struct {
		URL    string
		Width  uint64
		Height uint64
	}
	File struct {
		ContentType string
		Path        string
		Etag        string
		Handler     *os.File
	}
	Image image.Image
}

func NewImageFixture() *ImageFixture {
	return &ImageFixture{}
}

func (fx *ImageFixture) SetParams(u string, w, h uint64) {
	fx.Params.URL, fx.Params.Width, fx.Params.Height = u, w, h

	hasher := md5.New()
	hasher.Write([]byte(fx.Params.URL))
	fx.File.Etag = hex.EncodeToString(hasher.Sum(nil))
}

func (fx *ImageFixture) checkFileContentType(allowed map[string]bool) error {
	// Only the first 512 bytes are used to sniff the content type.
	buffer := make([]byte, 512)

	_, err := fx.File.Handler.Read(buffer)
	if err != nil {
		return err
	}

	fx.File.Handler.Seek(0, 0)

	// Use the net/http package's handy DectectContentType function. Always returns a valid
	// content-type by returning "application/octet-stream" if no others seemed to match.
	fx.File.ContentType = http.DetectContentType(buffer)

	if _, ok := allowed[fx.File.ContentType]; !ok {
		return fmt.Errorf("%s image format is not allowed", fx.File.ContentType)
	}

	return nil
}
