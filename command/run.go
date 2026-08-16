package command

import (
	"context"
	"errors"
	"fmt"
	"io"

	"example.com/golabel/transcodewrap/transcode"
)

type Options struct {
	LibraryPath string
	InputPath   string
	OutputPath  string
	Width       int
	Height      int
	BitrateKbps int
}

func Run(ctx context.Context, options Options, output io.Writer) (err error) {
	adapter, err := transcode.OpenLibrary(transcode.LocalLoader{}, options.LibraryPath)
	if err != nil {
		return err
	}
	defer func() {
		if releaseErr := adapter.Release(); releaseErr != nil {
			err = errors.Join(err, releaseErr)
		}
	}()

	if err := adapter.OpenFile(options.InputPath, options.OutputPath, options.BitrateKbps); err != nil {
		return err
	}
	if err := adapter.SetResolution(options.Width, options.Height); err != nil {
		return err
	}
	if err := adapter.Start(ctx, func(percent int) {
		fmt.Fprintf(output, "progress: %d%%\n", percent)
	}); err != nil {
		return err
	}

	fmt.Fprintf(output, "task succeeded: %s\n", options.OutputPath)
	return nil
}
