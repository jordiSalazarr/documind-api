package delete

import "errors"

var ErrProjectNotFound = errors.New("project not found")

type Command struct {
	ID string
}
