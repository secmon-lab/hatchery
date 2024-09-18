package secret

import "strings"

type String struct {
	v string
}

func NewString(v string) String {
	return String{v: v}
}

func (x String) Unsafe() string {
	return x.v
}

func (x String) String() string {
	return strings.Repeat("*", len(x.v))
}