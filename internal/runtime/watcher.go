package runtime

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"github.com/SSOHEB/codrao/internal/schema"
	"github.com/fsnotify/fsnotify"
)

// Watcher monitors a configuration file for changes, re-applies defaults and environment
// variable overrides, re-validates the output, and notifies registered listeners of updates.
type Watcher[T any] struct {
	ast          *schema.AST
	filePath     string
	current      *T
	currentMu    sync.RWMutex
	listeners    []func(newConfig *T)
	errListeners []func(err error)
	listenersMu  sync.Mutex
	watcher      *fsnotify.Watcher
	stopChan     chan struct{}
	stopOnce     sync.Once
	opts         RuntimeOptions
	reloadMu     sync.Mutex
}

// NewWatcher loads the initial config, validates it, and returns a new Watcher.
// If the initial config is invalid, it fails fast.
func NewWatcher[T any](ast *schema.AST, path string, opts RuntimeOptions) (*Watcher[T], error) {
	cfg, _, err := LoadAndPrepareFile[T](ast, path, opts)
	if err != nil {
		return nil, fmt.Errorf("initial config load failed: %w", err)
	}

	fw, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create fsnotify watcher: %w", err)
	}

	dir := filepath.Dir(path)
	if err := fw.Add(dir); err != nil {
		fw.Close()
		return nil, fmt.Errorf("failed to watch directory %s: %w", dir, err)
	}

	return &Watcher[T]{
		ast:      ast,
		filePath: path,
		current:  cfg,
		watcher:  fw,
		stopChan: make(chan struct{}),
		opts:     opts,
	}, nil
}

// Current returns the most recently loaded valid configuration in a thread-safe manner.
func (w *Watcher[T]) Current() *T {
	w.currentMu.RLock()
	defer w.currentMu.RUnlock()
	return w.current
}

// OnReload registers a callback to be executed when the configuration is successfully reloaded.
func (w *Watcher[T]) OnReload(fn func(newConfig *T)) {
	w.listenersMu.Lock()
	defer w.listenersMu.Unlock()
	w.listeners = append(w.listeners, fn)
}

// OnError registers a callback to be executed when a reload attempt fails validation or parsing.
func (w *Watcher[T]) OnError(fn func(err error)) {
	w.listenersMu.Lock()
	defer w.listenersMu.Unlock()
	w.errListeners = append(w.errListeners, fn)
}

// Start begins monitoring the configuration file's parent directory.
// We watch the parent directory instead of the file itself because:
// 1. Editors saving via temp-file-rename (like Vim/IntelliJ) replace the inode, which breaks file-level watchers.
// 2. If the file is deleted then recreated, directory watching survives and naturally picks up the recreation.
func (w *Watcher[T]) Start(ctx context.Context) error {
	var timer *time.Timer
	var timerMu sync.Mutex

	// triggerReload implements a 100ms debounce window to prevent rapid successive filesystem
	// writes (e.g. IDE autosave every keystroke) from triggering multiple successive reloads.
	triggerReload := func() {
		timerMu.Lock()
		defer timerMu.Unlock()
		if timer != nil {
			timer.Stop()
		}
		timer = time.AfterFunc(100*time.Millisecond, func() {
			w.reload()
		})
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-w.stopChan:
			return nil
		case event, ok := <-w.watcher.Events:
			if !ok {
				return nil
			}

			// Filter only events matching our target filename
			if filepath.Clean(event.Name) != filepath.Clean(w.filePath) {
				continue
			}

			// Watch for Write (contents modified), Create (temp file replaced/created), or Rename (temp file moved over)
			if (event.Op & (fsnotify.Write | fsnotify.Create | fsnotify.Rename)) != 0 {
				triggerReload()
			}

		case err, ok := <-w.watcher.Errors:
			if !ok {
				return nil
			}
			w.reportError(err)
		}
	}
}

// Stop terminates the directory file monitoring cleanly. It is safe to call multiple times.
func (w *Watcher[T]) Stop() error {
	var err error
	w.stopOnce.Do(func() {
		close(w.stopChan)
		err = w.watcher.Close()
	})
	return err
}

func (w *Watcher[T]) reload() {
	w.reloadMu.Lock()
	defer w.reloadMu.Unlock()

	cfg, _, err := LoadAndPrepareFile[T](w.ast, w.filePath, w.opts)
	if err != nil {
		w.reportError(err)
		return
	}

	w.currentMu.Lock()
	w.current = cfg
	w.currentMu.Unlock()

	w.listenersMu.Lock()
	listeners := make([]func(newConfig *T), len(w.listeners))
	copy(listeners, w.listeners)
	w.listenersMu.Unlock()

	for _, fn := range listeners {
		fn(cfg)
	}
}

func (w *Watcher[T]) reportError(err error) {
	w.listenersMu.Lock()
	listeners := make([]func(err error), len(w.errListeners))
	copy(listeners, w.errListeners)
	w.listenersMu.Unlock()

	for _, fn := range listeners {
		fn(err)
	}
}
