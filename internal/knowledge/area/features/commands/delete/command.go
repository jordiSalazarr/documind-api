package delete

import "errors"

var ErrAreaNotFound = errors.New("area not found")

type Command struct {
	ID string
}
