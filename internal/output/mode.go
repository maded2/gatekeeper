package output

// Mode indicates the execution mode.
type Mode int

const (
	Local Mode = iota
	CI
)

func (m Mode) String() string {
	switch m {
	case Local:
		return "local"
	case CI:
		return "ci"
	default:
		return "unknown"
	}
}
