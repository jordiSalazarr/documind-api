package deletetype

import (
	"errors"
)

var ErrRelationTypeNotFound = errors.New("relation type not found")

type Command struct{ ID string }
