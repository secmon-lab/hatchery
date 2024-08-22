package types

// SecretString is a type to handle secret string such as API key, password, etc. The string is not shown in JSON or YAML output and logging output.
type SecretString struct {
	s string `masq:"secret" json:"-" yaml:"-"`
}

// NewSecretString creates SecretString instance with given string.
func NewSecretString(s string) SecretString {
	return SecretString{s: s}
}

// UnsafeString returns the string value of SecretString. This method should be used only for using the secret value to authorize external service.
func (s SecretString) UnsafeString() string {
	return s.s
}
