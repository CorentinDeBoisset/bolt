package pkg

import (
	"os"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

var p *message.Printer

func GetI18nPrinter() *message.Printer {
	if p == nil {
		langEnv := os.Getenv("LANG")
		parsedLanguage, _ := language.Parse(langEnv)
		if parsedLanguage == language.Und {
			parsedLanguage = language.English
		}

		p = message.NewPrinter(parsedLanguage)
	}

	return p
}
