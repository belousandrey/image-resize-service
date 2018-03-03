package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/ReneKroon/ttlcache"
	"github.com/kr/pretty"
	"github.com/pkg/errors"
)

func downloadFileToTemp(w http.ResponseWriter, URL string) (string, error) {
	content, err := downloadFile(URL)
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

	pretty.Println("temporary file path:", tempFile.Name())
	return tempFile.Name(), nil
}

func downloadFile(url string) (io.ReadCloser, error) {
	pretty.Println("download file from:", url)
	req, err := http.NewRequest("GET", url, nil)
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

func cleanTempFiles(c *ttlcache.Cache) error {
	pretty.Println("prepare to clean temp files")
	value, exists := c.Get(registry)
	if !exists {
		return fmt.Errorf("temp files registry not found in cache")
	}

	tempFiles, ok := value.([]string)
	if !ok {
		return fmt.Errorf("temp files registry contains %T not []string", value)
	}

	for _, e := range tempFiles {
		pretty.Println("remove temp file:", e)
		err := os.Remove(e)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("could not remove temp file %s", e))
		}
	}

	return nil
}
