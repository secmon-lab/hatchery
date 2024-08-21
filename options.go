package hatchery

func WithPipeline(id PipelineID, src Source, dst Destination) Option {
	return func(h *Hatchery) {
		p := &Pipeline{
			src: src,
			dst: dst,
		}
		h.pipelines[p.ID()] = p
	}
}
