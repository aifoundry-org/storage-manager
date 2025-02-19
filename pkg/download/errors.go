package download

import "fmt"

var _ error = &ErrUnsupportedScheme{}

type ErrUnsupportedScheme struct {
	Scheme string
}

func (e *ErrUnsupportedScheme) Error() string {
	return fmt.Sprintf("unknown scheme %s", e.Scheme)
}
