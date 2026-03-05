package cfg

type ConfigError struct {
	translatedMessage string
}

func (e *ConfigError) Error() string {
	return e.translatedMessage
}

func newConfigError(key string, args ...any) *ConfigError {
	return &ConfigError{
		translatedMessage: GetI18nPrinter().Sprintf(key, args...),
	}
}
