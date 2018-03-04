package main

import (
	"github.com/ReneKroon/ttlcache"
)

// MetaData contains local file paths to original image and to all resized images
// original is a string with path to original image
// resized is a map with all resized images:
// { width1: { height1: path11, height2: path12 }, width2: {height1: path21}}
type MetaData struct {
	original string
	resized  map[uint64]map[uint64]string
}

// NewMetaData returns new MetaData object
// accepts file path to original image
func NewMetaData(ofp string) *MetaData {
	return &MetaData{
		original: ofp,
		resized:  make(map[uint64]map[uint64]string),
	}
}

// SetToCache puts into cache image metadata struct by image Etag
func (fx *ImageFixture) SetToCache(c *ttlcache.Cache, reg Registry) {
	md := NewMetaData(fx.File.Path)
	//pretty.Printf("put to cache | key = %s | value = %s | metadata = %+v\n", fx.File.Etag, fx.File.Path, md)
	c.Set(fx.File.Etag, md)

	reg.AddFileToRegistry(fx.File.Path)
}

func (fx *ImageFixture) getImageMetaDataFromCache(c *ttlcache.Cache) (*MetaData, bool) {
	value, exists := c.Get(fx.File.Etag)
	//pretty.Printf("get from cache | key = %s | value = %+v | exists = %t\n", fx.File.Etag, value, exists)
	if !exists {
		return nil, false
	}

	md, ok := value.(*MetaData)
	if !ok {
		fx.RemoveFromCache(c)
	}
	return md, true
}

// GetFromCache extracts from cache image metadata struct by image Etag
func (fx *ImageFixture) GetFromCache(c *ttlcache.Cache) bool {
	md, exists := fx.getImageMetaDataFromCache(c)
	if !exists {
		return false
	}

	fx.File.Path = md.original
	return true
}

// UpdateValueInCache updates image metadata struct by image Etag
// i.e. if new resized image added
func (fx *ImageFixture) UpdateValueInCache(c *ttlcache.Cache, resized string, reg Registry) {
	md, exists := fx.getImageMetaDataFromCache(c)
	if !exists {
		// wtf?
		return
	}

	_, ok := md.resized[fx.Params.Width]
	if !ok {
		md.resized[fx.Params.Width] = make(map[uint64]string)
	}
	md.resized[fx.Params.Width][fx.Params.Height] = resized

	reg.AddFileToRegistry(resized)
}

// RemoveFromCache deletes image metadata struct by image Etag
func (fx *ImageFixture) RemoveFromCache(c *ttlcache.Cache) {
	//pretty.Printf("remove from cache | key = %s\n", fx.File.Etag)
	c.Remove(fx.File.Etag)
}

// FindInCache searches in cache resized image by width and height
func (fx *ImageFixture) FindInCache(c *ttlcache.Cache) bool {
	md, exists := fx.getImageMetaDataFromCache(c)
	if !exists {
		return false
	}

	_, ok := md.resized[fx.Params.Width]
	if !ok {
		return false
	}

	_, ok = md.resized[fx.Params.Width][fx.Params.Height]
	if !ok {
		return false
	}

	return true
}
