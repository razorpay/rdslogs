package tracker

// Tracker is an interface to store the marker and other logFile related information
type Tracker interface {
	// Read and Write latest marker
	ReadLatestMarker(dbname string) string
	WriteLatestMarker(dbname string, marker string)
}
