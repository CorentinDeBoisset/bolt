package cfg

import "fmt"

// All errors must enforce the iface.FormattableError interface

type ConfigError struct {
	message           string
	translatedMessage string
}

func (e *ConfigError) Error() string {
	return e.message
}

func (e *ConfigError) Format() string {
	return e.translatedMessage
}

func newConfigError(message string, args ...any) *ConfigError {
	return &ConfigError{
		message:           fmt.Sprintf(message, args...),
		translatedMessage: GetI18nPrinter().Sprintf(message, args...),
	}
}
