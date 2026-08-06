package config

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/SSOHEB/codrao/internal/parser"
	"github.com/SSOHEB/codrao/internal/runtime"
	"github.com/SSOHEB/codrao/internal/schema"
	"github.com/SSOHEB/codrao/internal/validator"
)

type options struct {
	envPrefix    string
	disableEnv   bool
	metadataPath string
}

type Option func(*options)

func WithEnvPrefix(prefix string) Option {
	return func(o *options) {
		o.envPrefix = prefix
	}
}

func WithoutEnvOverrides() Option {
	return func(o *options) {
		o.disableEnv = true
	}
}

func WithMetadataPath(path string) Option {
	return func(o *options) {
		o.metadataPath = path
	}
}

// ValidationError mirrors internal/validator's error type but is defined
// here so callers never need to import internal/.
type ValidationError struct {
	Path string
	Msg  string
	Rule string
}

type ValidationErrors []ValidationError

func (e ValidationErrors) Error() string {
	var sb strings.Builder
	for i, err := range e {
		if i > 0 {
			sb.WriteString("\n")
		}
		sb.WriteString(fmt.Sprintf("%s: %s (rule: %s)", err.Path, err.Msg, err.Rule))
	}
	return sb.String()
}

func convertValidationErrors(internal []validator.ValidationError) ValidationErrors {
	var publicErrs ValidationErrors
	for _, e := range internal {
		publicErrs = append(publicErrs, ValidationError{
			Path: e.Path,
			Msg:  e.Msg,
			Rule: e.Rule,
		})
	}
	return publicErrs
}

func convertError(err error) error {
	if err == nil {
		return nil
	}
	if valErrs, ok := err.(validator.ValidationErrors); ok {
		return convertValidationErrors(valErrs)
	}
	return fmt.Errorf("config: %w", err)
}

func resolveMetadataPath(configPath string, opts *options) (string, error) {
	if opts.metadataPath != "" {
		return opts.metadataPath, nil
	}
	dir := filepath.Dir(configPath)
	metaPath := filepath.Join(dir, "metadata.yaml")
	if _, err := os.Stat(metaPath); err != nil {
		return "", fmt.Errorf("config: no metadata path provided and no metadata.yaml found alongside %q", configPath)
	}
	return metaPath, nil
}

func Load[T any](configPath string, opts ...Option) (*T, error) {
	o := &options{}
	for _, opt := range opts {
		opt(o)
	}

	metaPath, err := resolveMetadataPath(configPath, o)
	if err != nil {
		return nil, err // resolveMetadataPath already wraps correctly for missing file, but others might need convertError. Wait, instructions say: fmt.Errorf("config: no metadata path provided and no metadata.yaml found alongside %q", configPath)
	}

	rawMeta, err := parser.ParseFile(metaPath)
	if err != nil {
		return nil, convertError(err)
	}

	ast, err := schema.Build(rawMeta)
	if err != nil {
		return nil, convertError(err)
	}

	rOpts := runtime.RuntimeOptions{
		EnvPrefix:  o.envPrefix,
		DisableEnv: o.disableEnv,
	}

	cfg, rawConfig, err := runtime.LoadAndPrepareFile[T](ast, configPath, rOpts)
	if err != nil {
		return nil, convertError(err)
	}

	valErrs := validator.Validate(ast, rawConfig)
	if len(valErrs) > 0 {
		return nil, convertValidationErrors(valErrs)
	}

	return cfg, nil
}

func Watch[T any](configPath string, onReload func(*T, error), opts ...Option) (stop func(), err error) {
	o := &options{}
	for _, opt := range opts {
		opt(o)
	}

	metaPath, err := resolveMetadataPath(configPath, o)
	if err != nil {
		return nil, err
	}

	rawMeta, err := parser.ParseFile(metaPath)
	if err != nil {
		return nil, convertError(err)
	}

	ast, err := schema.Build(rawMeta)
	if err != nil {
		return nil, convertError(err)
	}

	rOpts := runtime.RuntimeOptions{
		EnvPrefix:  o.envPrefix,
		DisableEnv: o.disableEnv,
	}

	w, err := runtime.NewWatcher[T](ast, configPath, rOpts)
	if err != nil {
		return nil, convertError(err)
	}

	w.OnReload(func(newConfig *T) {
		onReload(newConfig, nil)
	})

	w.OnError(func(watchErr error) {
		onReload(nil, convertError(watchErr))
	})

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		_ = w.Start(ctx)
	}()

	stopFunc := func() {
		cancel()
		w.Stop()
	}

	return stopFunc, nil
}

func Validate(configPath, metadataPath string) error {
	rawMeta, err := parser.ParseFile(metadataPath)
	if err != nil {
		return convertError(err)
	}

	ast, err := schema.Build(rawMeta)
	if err != nil {
		return convertError(err)
	}

	_, rawConfig, err := runtime.LoadAndPrepareFile[map[string]any](ast, configPath, runtime.RuntimeOptions{})
	if err != nil {
		return convertError(err)
	}

	valErrs := validator.Validate(ast, rawConfig)
	if len(valErrs) > 0 {
		return convertValidationErrors(valErrs)
	}

	return nil
}
