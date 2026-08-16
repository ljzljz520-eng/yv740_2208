package transcode

import "fmt"

type NativeCode int

const (
	CodeOK              NativeCode = 0
	CodeInvalidArgument NativeCode = 10
	CodeInvalidState    NativeCode = 20
	CodeInputOpen       NativeCode = 30
	CodeOutputWrite     NativeCode = 40
	CodeResolution      NativeCode = 50
	CodeCancelled       NativeCode = 60
	CodeInternal        NativeCode = 70
)

type State string

const (
	StateReady     State = "ready"
	StateFileOpen  State = "file_open"
	StateRunning   State = "running"
	StateSucceeded State = "succeeded"
	StateCancelled State = "cancelled"
	StateFailed    State = "failed"
	StateReleased  State = "released"
)

type ProgressFunc func(percent int)

type NativeSession interface {
	SetResolution(width, height int) NativeCode
	Start(ProgressFunc) NativeCode
	Cancel() NativeCode
	Close() NativeCode
}

type NativeLibrary interface {
	OpenFile(inputPath, outputPath string, bitrateKbps int) (NativeSession, NativeCode)
	Close() NativeCode
}

type Loader interface {
	Load(path string) (NativeLibrary, error)
}

type NativeError struct {
	Operation string
	Code      NativeCode
}

func (e *NativeError) Error() string {
	return fmt.Sprintf("%s failed: native code %d (%s)", e.Operation, e.Code, codeText(e.Code))
}

type StateError struct {
	Operation string
	State     State
}

func (e *StateError) Error() string {
	return fmt.Sprintf("cannot %s while transcoder is %s", e.Operation, e.State)
}

func codeText(code NativeCode) string {
	switch code {
	case CodeOK:
		return "ok"
	case CodeInvalidArgument:
		return "invalid argument"
	case CodeInvalidState:
		return "invalid state"
	case CodeInputOpen:
		return "input open error"
	case CodeOutputWrite:
		return "output write error"
	case CodeResolution:
		return "resolution rejected"
	case CodeCancelled:
		return "cancelled"
	case CodeInternal:
		return "internal error"
	default:
		return "unknown error"
	}
}
