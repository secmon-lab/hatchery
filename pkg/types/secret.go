package hatchery

type SecretString struct {
	v string `masq:"secret" json:"-" yaml:"-"`
}

func NewSecretString(v string) SecretString {
	return SecretString{v: v}
}

func (s SecretString) UnsafeString() string {
	return s.v
}
