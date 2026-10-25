package ship

import "errors"

// The errors of reading an imported model, by their codes.
var (
	ErrModelUnreadable = errors.New("ship.model-unreadable")
	ErrModelTooBig     = errors.New("ship.model-too-big")
)
