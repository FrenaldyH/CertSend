package logger

import (
	"io"
	"log/slog"
	"os"
)

var Log *slog.Logger

// InitLog initializes the package-level Log variable with a structured
// JSON logger (see log/slog) that writes simultaneously to stdout and
// to today's log file inside the given path (see makeFileLogs). It must be
// called once, typically at application startup, before Log is used
// elsewhere; calling it again replaces the previous logger and opens
// a new file handle without closing the old one.
func InitLog(path string) error {
	file, err := makeFileLogs(path)
	if err != nil {
		return err
	}

	multi := io.MultiWriter(os.Stdout, file)

	Log = slog.New(slog.NewJSONHandler(
		multi,
		&slog.HandlerOptions{
			Level:     slog.LevelDebug,
			AddSource: true,
		},
	))

	return nil
}
