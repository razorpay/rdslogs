package publisher

// We fetch up to 10k lines at a time - buffering several fetches at once
// allows us to hand them off and fetch more while the line processor is doing work.
const lineChanSize = 100000

// Publisher is an interface to write rdslogs entries to a target.
// Current implementations are STDOUT and file
type Publisher interface {
	// Write accepts a long blob of text and writes it to the target
	Write(blob string)

	Close()
}
