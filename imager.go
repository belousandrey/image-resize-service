package main

import (
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"io/ioutil"

	"github.com/nfnt/resize"
)

// Imager is an interface that works with images
type Imager interface {
	Decode(reader io.Reader) error
	Encode(writer io.Writer) error
	Resize(width, height uint)
	StoreToTempFile() (string, error)
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

// Decode JPEG image to io.Reader
func (i *Images) Decode(reader io.Reader) error {
	var err error
	i.original, err = jpeg.Decode(reader)
	return err
}

// Encode JPEG image to io.Writer
func (i *Images) Encode(writer io.Writer) error {
	return jpeg.Encode(writer, i.resized, &jpeg.Options{jpeg.DefaultQuality})
}

// Resize image with provided width and height
func (i *Images) Resize(width, height uint) {
	i.resized = resize.Resize(width, height, i.original, resize.Lanczos3)
}

// StoreToTempFile stores resized image into temporary file and returns path
func (i *Images) StoreToTempFile() (string, error) {
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
