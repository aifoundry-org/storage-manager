package cache

import "fmt"

var _ error = &NotFoundError{}

type NotFoundError struct {
	Key string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("not found %s", e.Key)
}
