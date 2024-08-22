package hatchery

import "context"

type PipelineID string

type Pipeline struct {
	id  PipelineID
	src Source
	dst Destination
}

func (p *Pipeline) ID() PipelineID {
	return p.id
}

func (p *Pipeline) Run(ctx context.Context) error {
	return p.src.Load(ctx, p.dst)
}
