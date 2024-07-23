package writers

// MegaDumpWriter abstracts the different types of log storage targets
type MegaDumpWriter interface {
	Write(identifier string, bytes []byte) (int, error)
	String() string
}
