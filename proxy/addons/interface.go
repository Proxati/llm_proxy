package addons

// ClosableAddon is an interface that defines an addon that can be closed
type ClosableAddon interface {
	String() string
	Close() error
}
