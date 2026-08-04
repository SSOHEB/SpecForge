package runtime

import (
	"errors"
	"io"
	"os"

	"gopkg.in/yaml.v3"
)

// LoadReader reads YAML from r and unmarshals it into a new instance of T.
// Using generics here ensures that this package has no compile-time dependency
// on any specific generated Config struct layout, avoiding import cycles.
func LoadReader[T any](r io.Reader) (*T, error) {
	var cfg T
	dec := yaml.NewDecoder(r)
	if err := dec.Decode(&cfg); err != nil {
		if errors.Is(err, io.EOF) {
			return nil, &ConfigEmptyError{}
		}
		return nil, &ConfigSyntaxError{Err: err}
	}
	return &cfg, nil
}

// LoadFile reads YAML from the file at path and unmarshals it into a new instance of T.
func LoadFile[T any](path string) (*T, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, &ConfigFileNotFoundError{Path: path, Err: err}
		}
		return nil, err
	}
	defer f.Close()
	return LoadReader[T](f)
}

// Load looks up SPECFORGE_CONFIG_PATH env var (defaulting to "./config.yaml") and loads it into a new instance of T.
func Load[T any]() (*T, error) {
	path := os.Getenv("SPECFORGE_CONFIG_PATH")
	if path == "" {
		path = "./config.yaml"
	}
	return LoadFile[T](path)
}
