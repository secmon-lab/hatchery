package hatchery

import "errors"

var (
	ErrPipelineConflicted = errors.New("pipeline id conflicted")
	ErrPipelineNotFound   = errors.New("pipeline not found")
)
