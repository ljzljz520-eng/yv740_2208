package transcode_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"example.com/golabel/transcodewrap/transcode"
)

func TestTranscodingReportsProgressAndFinalStatus(t *testing.T) {
	directory := t.TempDir()
	adapter := openAdapter(t, directory)
	defer adapter.Release()
	inputPath := filepath.Join(directory, "input.mp4")
	outputPath := filepath.Join(directory, "output.mp4")
	if err := os.WriteFile(inputPath, []byte("video"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := adapter.OpenFile(inputPath, outputPath, 300); err != nil {
		t.Fatal(err)
	}
	if err := adapter.SetResolution(320, 180); err != nil {
		t.Fatal(err)
	}

	var progress []int
	if err := adapter.Start(context.Background(), func(percent int) {
		progress = append(progress, percent)
	}); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(progress, []int{0, 50, 100}) {
		t.Fatalf("progress = %v", progress)
	}
	if adapter.Status() != transcode.StateSucceeded {
		t.Fatalf("Status() = %q", adapter.Status())
	}
}

func TestCancellationStopsAnActiveTask(t *testing.T) {
	directory := t.TempDir()
	adapter := openAdapter(t, directory)
	defer adapter.Release()
	inputPath := filepath.Join(directory, "input.mp4")
	outputPath := filepath.Join(directory, "output.mp4")
	if err := os.WriteFile(inputPath, []byte("video"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := adapter.OpenFile(inputPath, outputPath, 300); err != nil {
		t.Fatal(err)
	}
	if err := adapter.SetResolution(320, 180); err != nil {
		t.Fatal(err)
	}

	err := adapter.Start(context.Background(), func(percent int) {
		if percent == 50 {
			if cancelErr := adapter.Cancel(); cancelErr != nil {
				t.Errorf("Cancel() error = %v", cancelErr)
			}
		}
	})
	if err == nil {
		t.Fatal("Start() error = nil")
	}
	if adapter.Status() != transcode.StateCancelled {
		t.Fatalf("Status() = %q", adapter.Status())
	}
	if _, statErr := os.Stat(outputPath); !os.IsNotExist(statErr) {
		t.Fatalf("output file state error = %v", statErr)
	}
}

func TestReleasedTranscoderRejectsNewWork(t *testing.T) {
	directory := t.TempDir()
	adapter := openAdapter(t, directory)
	if err := adapter.Release(); err != nil {
		t.Fatal(err)
	}
	if adapter.Status() != transcode.StateReleased {
		t.Fatalf("Status() = %q", adapter.Status())
	}
	if err := adapter.OpenFile("input.mp4", "output.mp4", 300); err == nil {
		t.Fatal("OpenFile() error = nil")
	}
	if err := adapter.Release(); err != nil {
		t.Fatalf("second Release() error = %v", err)
	}
}

func openAdapter(t *testing.T, directory string) *transcode.Adapter {
	t.Helper()
	libraryPath := filepath.Join(directory, "libtranscoder.vlib")
	manifest := fmt.Sprintf(`{"name":"test-library","progress":[0,50,100],"close_session_code":0,"close_library_code":0}`)
	if err := os.WriteFile(libraryPath, []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}
	adapter, err := transcode.OpenLibrary(transcode.LocalLoader{}, libraryPath)
	if err != nil {
		t.Fatal(err)
	}
	return adapter
}
