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

// Recorder records the processed metrics.
type Recorder interface {
	Record(Result) error
}

type ScriptRecorder struct {
	Path   string
	Logger zerolog.Logger
}

func NewScriptRecorder(logger zerolog.Logger, config config.Recorder) *ScriptRecorder {
	return &ScriptRecorder{
		Path:   config.Path,
		Logger: logger.With().Str("component", "recorder").Logger(),
	}
}

func (sr *ScriptRecorder) Record(result *Result) error {
	buf := &bytes.Buffer{}
	enc := json.NewEncoder(buf)
	if err := enc.Encode(result); err != nil {
		return errors.Wrap(err, "encoding results")
	}

	cmd := exec.Command(sr.Path)
	cmd.Stdin = bytes.NewReader(buf.Bytes())
	out, err := cmd.Output()
	if err != nil {
		msg := "executing script"
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			if strings.Contains(exitErr.Error(), sr.Path) || strings.Contains(string(exitErr.Stderr), sr.Path) {
				return errors.Wrapf(exitErr, "%s: %s:", msg, exitErr.Stderr)
			} else {
				return errors.Wrapf(exitErr, "%s: %s: %s:", msg, sr.Path, exitErr.Stderr)
			}
		}
		return errors.Wrap(err, msg)
	}
	sr.Logger.Info().
		Str("output", string(out)).
		Msg("recorded results")
	return nil
}
