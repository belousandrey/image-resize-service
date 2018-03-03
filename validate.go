package main

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/ReneKroon/ttlcache"
	"github.com/pkg/errors"
)

func (fx *ImageFixture) getParamsFromRequest(w http.ResponseWriter, r *http.Request) error {
	err := fx.getUploadDataFromRequest(r)
	if err != nil {
		fx.respondWithError(w, http.StatusBadRequest, err)
		return err
	}

	err = fx.validateUploadData()
	if err != nil {
		return err
	}

	return nil
}

func (fx *ImageFixture) getUploadDataFromRequest(r *http.Request) error {
	r.ParseForm()

	url := strings.ToLower(r.Form.Get("url"))
	width, err := strconv.ParseUint(r.Form.Get("width"), 10, 32)
	if err != nil {
		return err
	}

	height, err := strconv.ParseUint(r.Form.Get("height"), 10, 32)
	if err != nil {
		return err
	}

	fx.SetParams(url, width, height)

	return nil
}

func (fx *ImageFixture) validateUploadData() error {
	_, err := url.ParseRequestURI(fx.Params.URL)
	if err != nil {
		return errors.Wrap(err, "validate URL from incoming data")
	}

	return nil
}

func (fx *ImageFixture) upToDate(r *http.Request, c *ttlcache.Cache) bool {
	ifNoneMatch := r.Header.Get("If-None-Match")
	if len(ifNoneMatch) > 0 {
		exists := fx.FindInCache(c)
		if exists {
			return true
		}
	}

	return false
}
