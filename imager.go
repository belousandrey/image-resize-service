package main

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"io/ioutil"
	"os"

	"github.com/nfnt/resize"
)

// Imager is an interface that works with images
type Imager interface {
	Open(path string) (*os.File, error)
	Decode(reader io.Reader) error
	Encode() (*bytes.Buffer, error)
	EncodeToWriter(writer io.Writer) error
	Resize(width, height uint)
	StoreResizedToTempFile() (string, error)
}

// NewImager returns new Images object
func NewImager() Imager {
	return &Images{}
}

// Imager contains original and resized image objects for request
// Imager can encode/decode and resize JPEG pictures
type Images struct {
	original image.Image
	resized  image.Image
}

// Open returns file handler and error
func (i *Images) Open(path string) (*os.File, error) {
	return os.Open(path)
}

// Decode JPEG image to io.Reader
func (i *Images) Decode(reader io.Reader) error {
	var err error
	i.original, err = jpeg.Decode(reader)
	return err
}

// Encode JPEG image into new bytes buffer
func (i *Images) Encode() (*bytes.Buffer, error) {
	buffer := new(bytes.Buffer)
	err := jpeg.Encode(buffer, i.resized, &jpeg.Options{Quality: jpeg.DefaultQuality})
	return buffer, err
}

// EncodeToWriter encodes JPEG image to io.Writer
func (i *Images) EncodeToWriter(writer io.Writer) error {
	return jpeg.Encode(writer, i.resized, &jpeg.Options{Quality: jpeg.DefaultQuality})
}

// Resize image with provided width and height
func (i *Images) Resize(width, height uint) {
	i.resized = resize.Resize(width, height, i.original, resize.Lanczos3)
}

// StoreResizedToTempFile stores resized image into temporary file and returns path
func (i *Images) StoreResizedToTempFile() (string, error) {
	if i.resized == nil {
		return "", fmt.Errorf("no resized image yet")
	}

	file, err := ioutil.TempFile("", "")
	if err != nil {
		return "", err
	}

	if err = i.EncodeToWriter(file); err != nil {
		return "", err
	}

	return file.Name(), nil
}
