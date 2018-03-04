package main

import (
	"fmt"
	"os"

	"github.com/pkg/errors"
)

// Registry for all temp files
type Registry map[string]struct{}

// NewRegistry returns new Registry object
func NewRegistry() Registry {
	return make(Registry, 0)
}

// AddFileToRegistry adds new record with file path
func (r Registry) AddFileToRegistry(file string) {
	r[file] = struct{}{}
}

// Cleanup removes all files recorded in registry and cleans registry keys
func (r Registry) Cleanup() error {
	for k := range r {
		err := os.Remove(k)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("could not remove temp file %s", k))
		}

		delete(r, k)
	}

	return nil
}
