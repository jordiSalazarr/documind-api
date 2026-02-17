package deleterelation

import (
	"errors"
)

var ErrRelationNotFound = errors.New("relation not found")

type Command struct{ ID string }
