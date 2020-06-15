package publisher

// Publisher is an interface to write rdslogs entries to a target.
// Current implementations are STDOUT and file
type Publisher interface {
	// Write accepts a long blob of text and writes it to the target
	Write(blob string)
}
