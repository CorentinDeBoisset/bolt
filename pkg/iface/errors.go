package iface

import (
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/term"
	"github.com/corentindeboisset/bolt/pkg/cfg"
)

type FormattableError interface {
	Error() string
	Format() string
}

var errorTitleStyle = lipgloss.NewStyle().Bold(true).Padding(1, 0)

var errorBodyStyle = lipgloss.NewStyle().
	Padding(1, 2).
	Border(lipgloss.DoubleBorder(), true).
	BorderForeground(ErrorColor).
	Foreground(ErrorColor)

func PrintError(e FormattableError) {
	termWidth, _, err := term.GetSize(os.Stdout.Fd())
	if err != nil {
		termWidth = 40
	}

	i18n := cfg.GetI18nPrinter()

	header := errorTitleStyle.
		Width(termWidth).
		Render(i18n.Sprintf("💥 The command failed:"))

	body := errorBodyStyle.
		Width(termWidth - errorBodyStyle.GetHorizontalBorderSize()).
		Render(e.Format())

	fmt.Print(lipgloss.JoinVertical(lipgloss.Left, header, body, ""))
}
