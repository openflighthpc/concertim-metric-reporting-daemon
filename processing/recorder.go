package main

import (
	"bytes"
	"encoding/json"
	"os/exec"
	"strings"

	"github.com/alces-flight/concertim-metric-reporting-daemon/processing/config"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

// Recorder is an interface for recording the processed metrics.
type Recorder interface {
	Record(Result) error
}

// ScriptRecorder implements the Recorder interface and records the metrics by
// calling the configured script.
//
// The processed metrics are converted to JSON and sent to the script over
// standard input.
type ScriptRecorder struct {
	Args   []string
	Path   string
	Logger zerolog.Logger
}

// NewScriptRecorder returns a new ScriptRecorder.
func NewScriptRecorder(logger zerolog.Logger, config config.Recorder) *ScriptRecorder {
	return &ScriptRecorder{
		Args:   config.Args,
		Path:   config.Path,
		Logger: logger.With().Str("component", "recorder").Logger(),
	}
}

// Record calls the configured script providing the given results encoded as
// JSON over standard input.
//
// The output of the script is logged.  It is assumed that very little output
// will be generated.
func (sr *ScriptRecorder) Record(result *Result) error {
	buf := &bytes.Buffer{}
	enc := json.NewEncoder(buf)
	if err := enc.Encode(result); err != nil {
		return errors.Wrap(err, "encoding results")
	}
	args := sr.Args
	if args == nil {
		args = []string{}
	}
	cmd := exec.Command(sr.Path, args...)
	sr.Logger.Debug().Str("cmd", cmd.String()).Msg("recording")
	cmd.Stdin = bytes.NewReader(buf.Bytes())
	out, err := cmd.Output()
	if err != nil {
		msg := "executing script"
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			if strings.Contains(exitErr.Error(), sr.Path) || strings.Contains(string(exitErr.Stderr), sr.Path) {
				return errors.Wrapf(exitErr, "%s: %s", msg, exitErr.Stderr)
			}
			return errors.Wrapf(exitErr, "%s: %s: %s", msg, sr.Path, exitErr.Stderr)
		}
		return errors.Wrap(err, msg)
	}
	sr.Logger.Info().
		Str("output", string(out)).
		Msg("completed")
	return nil
}
