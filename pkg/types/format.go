package types

type DataFormat string

const (
	FmtJSON  DataFormat = "json"
	FmtJSONL DataFormat = "jsonl"
	FmtYAML  DataFormat = "yaml"
)

func (x DataFormat) Ext() string {
	if x == "" {
		return "log"
	}
	return string(x)
}
