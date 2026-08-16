package command_test

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"example.com/golabel/transcodewrap/command"
)

func TestCommandConvertsFileAndReportsCompletion(t *testing.T) {
	directory := t.TempDir()
	libraryPath := writeLibrary(t, directory, 0)
	inputPath := filepath.Join(directory, "input.mp4")
	outputPath := filepath.Join(directory, "output.mp4")
	if err := os.WriteFile(inputPath, []byte("video-payload"), 0o644); err != nil {
		t.Fatal(err)
	}

	var output bytes.Buffer
	err := command.Run(context.Background(), command.Options{
		LibraryPath: libraryPath,
		InputPath:   inputPath,
		OutputPath:  outputPath,
		Width:       640,
		Height:      360,
		BitrateKbps: 450,
	}, &output)
	if err != nil {
		t.Fatalf("command error = %v", err)
	}

	converted, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(converted), "width=640 height=360 bitrate_kbps=450") {
		t.Fatalf("converted output = %q", converted)
	}
	if !strings.Contains(output.String(), "progress: 100%") || !strings.Contains(output.String(), "task succeeded") {
		t.Fatalf("command output = %q", output.String())
	}
}

func TestCommandStopsWhenResolutionCannotBeApplied(t *testing.T) {
	directory := t.TempDir()
	libraryPath := writeLibrary(t, directory, 50)
	inputPath := filepath.Join(directory, "input.mp4")
	outputPath := filepath.Join(directory, "output.mp4")
	if err := os.WriteFile(inputPath, []byte("video-payload"), 0o644); err != nil {
		t.Fatal(err)
	}

	var output bytes.Buffer
	err := command.Run(context.Background(), command.Options{
		LibraryPath: libraryPath,
		InputPath:   inputPath,
		OutputPath:  outputPath,
		Width:       1920,
		Height:      1080,
		BitrateKbps: 450,
	}, &output)
	if err == nil || !strings.Contains(err.Error(), "set target resolution") {
		t.Errorf("command error = %v, output = %q", err, output.String())
	}
	if _, statErr := os.Stat(outputPath); !os.IsNotExist(statErr) {
		t.Errorf("output file state error = %v", statErr)
	}
	if strings.Contains(output.String(), "task succeeded") {
		t.Errorf("command output = %q", output.String())
	}
}

func writeLibrary(t *testing.T, directory string, resolutionCode int) string {
	t.Helper()
	path := filepath.Join(directory, "libtranscoder.vlib")
	data := fmt.Sprintf(`{"name":"test-library","progress":[0,50,100],"set_resolution_code":%d}`, resolutionCode)
	if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}
