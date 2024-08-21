package hatchery

type PipelineID string

type Pipeline struct {
	id  PipelineID
	src Source
	dst Destination
}

func (p *Pipeline) ID() PipelineID {
	return p.id
}
