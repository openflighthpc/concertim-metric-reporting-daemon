package processing

// Recorder is an interface for recording the processed metrics.
type Recorder interface {
	Record(*Result) error
}
