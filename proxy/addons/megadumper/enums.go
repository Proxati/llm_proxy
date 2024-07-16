package megadumper

// LogDestination is an enum for the destination for where the logs are stored
type LogDestination int

const (
	// WriteToFile logs to a single file
	WriteToFile LogDestination = iota

	// WriteToDir logs to a directory
	WriteToDir

	// WriteToStdOut logs to standard out
	WriteToStdOut
)
