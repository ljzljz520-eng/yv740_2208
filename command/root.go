package command

import (
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

func NewRootCommand(stdout, stderr io.Writer) *cobra.Command {
	options := Options{}
	command := &cobra.Command{
		Use:           "transcode",
		Short:         "Transcode an MP4 through a local transcoder library",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return Run(cmd.Context(), options, stdout)
		},
	}
	command.SetOut(stdout)
	command.SetErr(stderr)
	command.Flags().StringVar(&options.LibraryPath, "library", fixturePath("libtranscoder.vlib"), "local transcoder library path")
	command.Flags().StringVar(&options.InputPath, "input", fixturePath("sample.mp4"), "input MP4 path")
	command.Flags().StringVar(&options.OutputPath, "output", "transcoded-low.mp4", "output MP4 path")
	command.Flags().IntVar(&options.Width, "width", 640, "target width")
	command.Flags().IntVar(&options.Height, "height", 360, "target height")
	command.Flags().IntVar(&options.BitrateKbps, "bitrate", 450, "target bitrate in kbps")
	return command
}

func fixturePath(name string) string {
	directory, err := os.Getwd()
	if err != nil {
		return filepath.Join("fixtures", name)
	}
	for {
		if _, err := os.Stat(filepath.Join(directory, "go.mod")); err == nil {
			candidate := filepath.Join(directory, "fixtures", name)
			if _, err := os.Stat(candidate); err == nil {
				return candidate
			}
		}
		parent := filepath.Dir(directory)
		if parent == directory {
			return filepath.Join("fixtures", name)
		}
		directory = parent
	}
}
