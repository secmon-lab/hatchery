package metadata

import "time"

type MetaData struct {
	timestamp *time.Time
	seq       int
}

func (m MetaData) Timestamp() time.Time {
	if m.timestamp == nil {
		return time.Now()
	}
	return *m.timestamp
}

func (m MetaData) Seq() int {
	return m.seq
}

func New(options ...Option) MetaData {
	md := &MetaData{}
	for _, opt := range options {
		opt(md)
	}
	return *md
}

type Option func(*MetaData)

func WithTimestamp(ts time.Time) Option {
	return func(m *MetaData) {
		m.timestamp = &ts
	}
}

func WithSeq(seq int) Option {
	return func(m *MetaData) {
		m.seq = seq
	}
}
