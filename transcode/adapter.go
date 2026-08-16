package transcode

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

type Adapter struct {
	mu      sync.Mutex
	library NativeLibrary
	session NativeSession
	state   State
}

func OpenLibrary(loader Loader, path string) (*Adapter, error) {
	if loader == nil {
		return nil, fmt.Errorf("load local transcoder library: loader is required")
	}
	if path == "" {
		return nil, fmt.Errorf("load local transcoder library: path is required")
	}

	library, err := loader.Load(path)
	if err != nil {
		return nil, fmt.Errorf("load local transcoder library: %w", err)
	}
	if library == nil {
		return nil, fmt.Errorf("load local transcoder library: loader returned no library")
	}

	return &Adapter{library: library, state: StateReady}, nil
}

func (a *Adapter) OpenFile(inputPath, outputPath string, bitrateKbps int) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.state != StateReady {
		return &StateError{Operation: "open file", State: a.state}
	}
	if inputPath == "" || outputPath == "" || bitrateKbps <= 0 {
		return &NativeError{Operation: "open file", Code: CodeInvalidArgument}
	}

	session, code := a.library.OpenFile(inputPath, outputPath, bitrateKbps)
	if code != CodeOK {
		return &NativeError{Operation: "open file", Code: code}
	}
	if session == nil {
		return &NativeError{Operation: "open file", Code: CodeInternal}
	}

	a.session = session
	a.state = StateFileOpen
	return nil
}

func (a *Adapter) SetResolution(width, height int) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.state != StateFileOpen {
		return &StateError{Operation: "set target resolution", State: a.state}
	}
	if width <= 0 || height <= 0 {
		return &NativeError{Operation: "set target resolution", Code: CodeInvalidArgument}
	}

	if code := a.session.SetResolution(width, height); code != CodeOK {
		return &NativeError{Operation: "set target resolution", Code: code}
	}
	return nil
}

func (a *Adapter) Start(ctx context.Context, progress ProgressFunc) error {
	if ctx == nil {
		ctx = context.Background()
	}

	a.mu.Lock()
	if a.state != StateFileOpen {
		state := a.state
		a.mu.Unlock()
		return &StateError{Operation: "start transcoding", State: state}
	}
	if err := ctx.Err(); err != nil {
		a.state = StateCancelled
		a.mu.Unlock()
		return fmt.Errorf("start transcoding: %w", err)
	}
	session := a.session
	a.state = StateRunning
	a.mu.Unlock()

	callback := func(percent int) {
		select {
		case <-ctx.Done():
			_ = session.Cancel()
		default:
			if progress != nil {
				progress(percent)
			}
		}
	}

	code := session.Start(callback)

	a.mu.Lock()
	defer a.mu.Unlock()
	if code == CodeOK {
		a.state = StateSucceeded
		return nil
	}
	if code == CodeCancelled {
		a.state = StateCancelled
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("start transcoding: %w", err)
		}
		return &NativeError{Operation: "start transcoding", Code: code}
	}
	a.state = StateFailed
	return &NativeError{Operation: "start transcoding", Code: code}
}

func (a *Adapter) Cancel() error {
	a.mu.Lock()
	if a.state != StateRunning {
		state := a.state
		a.mu.Unlock()
		return &StateError{Operation: "cancel transcoding", State: state}
	}
	session := a.session
	a.mu.Unlock()

	code := session.Cancel()
	if code != CodeOK {
		return &NativeError{Operation: "cancel transcoding", Code: code}
	}

	a.mu.Lock()
	a.state = StateCancelled
	a.mu.Unlock()
	return nil
}

func (a *Adapter) Status() State {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.state
}

func (a *Adapter) Release() error {
	a.mu.Lock()
	if a.state == StateReleased {
		a.mu.Unlock()
		return nil
	}
	session := a.session
	library := a.library
	a.session = nil
	a.library = nil
	a.state = StateReleased
	a.mu.Unlock()

	var releaseErrors []error
	if session != nil {
		if code := session.Close(); code != CodeOK {
			releaseErrors = append(releaseErrors, &NativeError{Operation: "release transcoding session", Code: code})
		}
	}
	if library != nil {
		if code := library.Close(); code != CodeOK {
			releaseErrors = append(releaseErrors, &NativeError{Operation: "release transcoding library", Code: code})
		}
	}
	return errors.Join(releaseErrors...)
}
