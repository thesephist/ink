package main

const (
	ErrUnknown = 0
	ErrSyntax  = 1
	ErrRuntime = 2
	ErrSystem  = 40
	ErrAssert  = 100
)

type Err struct {
	reason  int
	message string
}

func (e Err) Error() string {
	return e.message
}
