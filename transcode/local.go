package transcode

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
)

type LocalLoader struct{}

type localManifest struct {
	Name              string     `json:"name"`
	Progress          []int      `json:"progress"`
	OpenCode          NativeCode `json:"open_code"`
	SetResolutionCode NativeCode `json:"set_resolution_code"`
	StartCode         NativeCode `json:"start_code"`
	CancelCode        NativeCode `json:"cancel_code"`
	CloseSessionCode  NativeCode `json:"close_session_code"`
	CloseLibraryCode  NativeCode `json:"close_library_code"`
}

func (LocalLoader) Load(path string) (NativeLibrary, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	decoder := json.NewDecoder(io.LimitReader(file, 1<<20))
	decoder.DisallowUnknownFields()
	var manifest localManifest
	if err := decoder.Decode(&manifest); err != nil {
		return nil, fmt.Errorf("decode library manifest: %w", err)
	}
	if manifest.Name == "" {
		return nil, fmt.Errorf("decode library manifest: name is required")
	}
	if len(manifest.Progress) == 0 {
		manifest.Progress = []int{0, 25, 50, 75, 100}
	}
	previous := -1
	for _, value := range manifest.Progress {
		if value < 0 || value > 100 || value < previous {
			return nil, fmt.Errorf("decode library manifest: progress must be ordered from 0 to 100")
		}
		previous = value
	}
	if manifest.Progress[len(manifest.Progress)-1] != 100 {
		return nil, fmt.Errorf("decode library manifest: progress must finish at 100")
	}

	return &localLibrary{manifest: manifest}, nil
}

type localLibrary struct {
	mu       sync.Mutex
	manifest localManifest
	closed   bool
}

func (l *localLibrary) OpenFile(inputPath, outputPath string, bitrateKbps int) (NativeSession, NativeCode) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.closed {
		return nil, CodeInvalidState
	}
	if l.manifest.OpenCode != CodeOK {
		return nil, l.manifest.OpenCode
	}
	if inputPath == "" || outputPath == "" || bitrateKbps <= 0 {
		return nil, CodeInvalidArgument
	}
	input, err := os.ReadFile(inputPath)
	if err != nil {
		return nil, CodeInputOpen
	}

	return &localSession{
		manifest:    l.manifest,
		input:       input,
		outputPath:  outputPath,
		bitrateKbps: bitrateKbps,
	}, CodeOK
}

func (l *localLibrary) Close() NativeCode {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.closed {
		return CodeOK
	}
	l.closed = true
	return l.manifest.CloseLibraryCode
}

type localSession struct {
	mu          sync.Mutex
	manifest    localManifest
	input       []byte
	outputPath  string
	bitrateKbps int
	width       int
	height      int
	running     bool
	cancelled   bool
	closed      bool
}

func (s *localSession) SetResolution(width, height int) NativeCode {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed || s.running {
		return CodeInvalidState
	}
	if width <= 0 || height <= 0 {
		return CodeInvalidArgument
	}
	if s.manifest.SetResolutionCode != CodeOK {
		return s.manifest.SetResolutionCode
	}
	s.width = width
	s.height = height
	return CodeOK
}

func (s *localSession) Start(progress ProgressFunc) NativeCode {
	s.mu.Lock()
	if s.closed || s.running {
		s.mu.Unlock()
		return CodeInvalidState
	}
	if s.manifest.StartCode != CodeOK {
		s.mu.Unlock()
		return s.manifest.StartCode
	}
	s.running = true
	values := append([]int(nil), s.manifest.Progress...)
	s.mu.Unlock()

	for _, value := range values {
		s.mu.Lock()
		cancelled := s.cancelled
		s.mu.Unlock()
		if cancelled {
			return CodeCancelled
		}
		if progress != nil {
			progress(value)
		}
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if s.cancelled {
		return CodeCancelled
	}
	header := fmt.Sprintf("fixture-transcode width=%d height=%d bitrate_kbps=%d\n", s.width, s.height, s.bitrateKbps)
	output := append([]byte(header), s.input...)
	if err := os.WriteFile(s.outputPath, output, 0o644); err != nil {
		return CodeOutputWrite
	}
	return CodeOK
}

func (s *localSession) Cancel() NativeCode {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed || !s.running {
		return CodeInvalidState
	}
	if s.manifest.CancelCode != CodeOK {
		return s.manifest.CancelCode
	}
	s.cancelled = true
	return CodeOK
}

func (s *localSession) Close() NativeCode {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return CodeOK
	}
	s.closed = true
	s.input = nil
	return s.manifest.CloseSessionCode
}
