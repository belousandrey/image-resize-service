package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/pkg/errors"
)

// Downloader is an interface that works with
type Downloader interface {
	DownloadFile(URL string) (io.ReadCloser, error)
	StoreFileToTemp(URL string) (string, error)
}

// Downloader is a type to process downloads
type Downloads struct{}

// NewDownloader returns new object Downloader
func NewDownloader() Downloader {
	return &Downloads{}
}

// StoreFileToTemp saves file content to temporary file and returns path
func (d *Downloads) StoreFileToTemp(URL string) (string, error) {
	content, err := d.DownloadFile(URL)
	if err != nil {
		return "", err
	}

	tempFile, err := ioutil.TempFile("", "")
	if err != nil {
		return "", err
	}

	_, err = io.Copy(tempFile, content)
	if err != nil {
		return "", err
	}
	content.Close()

	return tempFile.Name(), nil
}

// DownloadFile downloads file by URL and returns content
func (d *Downloads) DownloadFile(URL string) (io.ReadCloser, error) {
	req, err := http.NewRequest("GET", URL, nil)
	if err != nil {
		return nil, errors.Wrap(err, "create request object")
	}

	req.Close = true

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "download file by URL")
	}

	// Check server response
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status: %s", resp.Status)
	}

	return resp.Body, nil
}
