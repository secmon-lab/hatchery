package hatchery

import "errors"

var (
	ErrStreamConflicted = errors.New("pipeline id conflicted")
	ErrStreamNotFound   = errors.New("pipeline not found")
)
