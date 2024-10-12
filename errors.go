package hatchery

import "errors"

var (
	ErrStreamConflicted = errors.New("stream id conflicted")
	ErrNoStreamFound    = errors.New("no stream found")
	ErrInvalidStream    = errors.New("invalid stream")
)
