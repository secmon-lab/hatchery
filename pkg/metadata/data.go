package metadata

import (
	"time"

	"github.com/secmon-lab/hatchery/pkg/types"
)

type MetaData struct {
	timestamp  *time.Time
	seq        int
	format     types.DataFormat
	schemaHint string
	slug       string
}

func (m MetaData) Timestamp() time.Time {
	if m.timestamp == nil {
		return time.Now()
	}
	return *m.timestamp
}

func (m MetaData) Seq() int                 { return m.seq }
func (m MetaData) Format() types.DataFormat { return m.format }
func (m MetaData) SchemaHint() string       { return m.schemaHint }
func (m MetaData) Slug() string             { return m.slug }

func New(options ...Option) MetaData {
	var md MetaData
	for _, opt := range options {
		opt(&md)
	}
	return md
}

type Option func(*MetaData)

// WithTimestamp sets timestamp to MetaData. If it's not set, it uses the current time.
func WithTimestamp(ts time.Time) Option {
	return func(m *MetaData) {
		m.timestamp = &ts
	}
}

// WithSeq sets sequence number to MetaData. It's used to identify the order of data.
func WithSeq(seq int) Option {
	return func(m *MetaData) {
		m.seq = seq
	}
}

func WithFormat(f types.DataFormat) Option {
	return func(md *MetaData) {
		md.format = f
	}
}

func WithSchemaHint(hint string) Option {
	return func(md *MetaData) {
		md.schemaHint = hint
	}
}

func WithSlug(slug string) Option {
	return func(md *MetaData) {
		md.slug = slug
	}
}
