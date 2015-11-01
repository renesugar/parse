package parse // import "github.com/tdewolff/parse"

type SyntaxError struct {
	msg  string
	Line int
}

func NewSyntaxError(msg string, line int) *SyntaxError {
	return &SyntaxError{msg, line}
}

func (e *SyntaxError) Error() string {
	return e.msg
}
