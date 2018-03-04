package main

import (
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"io/ioutil"

	"github.com/nfnt/resize"
)

// Imager contains original and resized image objects for request
// Imager can encode/decode and resize JPEG pictures
type Imager struct {
	original image.Image
	resized  image.Image
}

// NewImager returns new Imager object
func NewImager() *Imager {
	return &Imager{}
}

// Decode JPEG image to io.Reader
func (i *Imager) Decode(reader io.Reader) error {
	var err error
	i.original, err = jpeg.Decode(reader)
	return err
}

// Encode JPEG image to io.Writer
func (i *Imager) Encode(writer io.Writer) error {
	return jpeg.Encode(writer, i.resized, &jpeg.Options{jpeg.DefaultQuality})
}

// Resize image with provided width and height
func (i *Imager) Resize(width, height uint) {
	i.resized = resize.Resize(width, height, i.original, resize.Lanczos3)
}

// StoreToTempFile stores resized image into temporary file and returns path
func (i *Imager) StoreToTempFile() (string, error) {
	if i.resized == nil {
		return "", fmt.Errorf("no resized image yet")
	}

	file, err := ioutil.TempFile("", "")
	if err != nil {
		return "", err
	}

	if err = i.Encode(file); err != nil {
		return "", err
	}

	return file.Name(), nil
}
