package ink

// Error reasons are enumerated here to be used in the Err struct,
// the error type shared across all Ink APIs.
const (
	ErrUnknown = 0
	ErrSyntax  = 1
	ErrRuntime = 2
	ErrSystem  = 40
	ErrAssert  = 100
)

// Err constants represent possible errors that Ink interpreter
// binding functions may return.
type Err struct {
	reason  int
	message string
}

func (e Err) Error() string {
	return e.message
}
