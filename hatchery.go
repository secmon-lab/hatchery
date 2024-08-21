package hatchery

type Hatchery struct {
	pipelines map[PipelineID]*Pipeline
}

type Option func(*Hatchery)

func New(opts ...Option) *Hatchery {
	h := &Hatchery{}
	for _, opt := range opts {
		opt(h)
	}
	return h
}
