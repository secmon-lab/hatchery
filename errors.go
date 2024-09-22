package hatchery

import "errors"

var (
	ErrStreamConflicted = errors.New("stream id conflicted")
	ErrStreamNotFound   = errors.New("stream not found")
	ErrInvalidStream    = errors.New("invalid stream")
)
