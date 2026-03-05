package iface

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"sync"
	"unicode"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/charmbracelet/x/term"
	"github.com/corentindeboisset/tera/pkg/cfg"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var width = sync.OnceValue(func() int {
	w, _, err := term.GetSize(os.Stdout.Fd())
	if err != nil {
		return 80
	}
	return w
})

func RenderError(err error) {
	i18n := cfg.GetI18nPrinter()

	errPrefix := lipgloss.NewStyle().
		MarginBottom(1).
		Background(lipgloss.Color("1")).
		Foreground(lipgloss.AdaptiveColor{Dark: "15", Light: "0"}).
		Padding(0, 1).
		Render(i18n.Sprintf("ERROR"))

	_, _ = fmt.Fprintf(os.Stderr, "\n%s\n%s\n\n", errPrefix, err.Error())
}

func renderCommandPath(c *cobra.Command, cmdStyle, subcmdStyle lipgloss.Style) string {
	if c.HasParent() {
		return renderCommandPath(c.Parent(), cmdStyle, subcmdStyle) + subcmdStyle.Render(" "+c.Name())
	}
	return cmdStyle.Render(c.DisplayName())
}

func renderParamTable(rows [][]string, totalWidth int) string {
	firstColWidth := 0
	for _, row := range rows {
		if len(row) > 0 {
			firstColWidth = max(firstColWidth, lipgloss.Width(row[0]))
		}
	}

	firstColStyle := lipgloss.NewStyle().Padding(0, 2).Width(firstColWidth + 4) // Include the padding into the width()
	secondColStyle := lipgloss.NewStyle().Width(totalWidth - firstColWidth - 4) // Same

	return table.New().
		Border(lipgloss.HiddenBorder()).
		BorderTop(false).BorderBottom(false).BorderLeft(false).BorderRight(false).
		BorderColumn(false).BorderRow(false).
		StyleFunc(func(_, col int) lipgloss.Style {
			if col == 0 {
				return firstColStyle
			}
			return secondColStyle
		}).
		Rows(rows...).
		Render()
}

func renderCommands(cmd *cobra.Command, theme HelpTheme, width int) string {
	i18n := cfg.GetI18nPrinter()

	fragments := make([]string, 0)
	hasGroups := false
	for _, group := range cmd.Groups() {
		rows := make([][]string, 0, len(cmd.Commands()))
		for _, subCmd := range cmd.Commands() {
			if subCmd.GroupID == group.ID && (subCmd.IsAvailableCommand() || subCmd.Name() == "help") {
				subCmdSynopsis := theme.SubCommand.Render(subCmd.Name())
				if subCmd.HasAvailableSubCommands() {
					subCmdSynopsis += theme.DimmedArg.Render(" " + i18n.Sprintf("[command]"))
				}
				if subCmd.HasAvailableFlags() {
					subCmdSynopsis += theme.DimmedArg.Render(" " + i18n.Sprintf("[--flags]"))
				}
				rows = append(rows, []string{subCmdSynopsis, subCmd.Short})
			}
		}
		if len(rows) > 0 {
			hasGroups = true
			fragments = append(
				fragments,
				i18n.Sprintf("%s:", group.Title),
				"",
				renderParamTable(rows, width),
			)
		}
	}
	if hasGroups {
		fragments = append(fragments, i18n.Sprintf("Additional Commands:"), "")
	}
	rows := make([][]string, 0, len(cmd.Commands()))
	for _, subCmd := range cmd.Commands() {
		if len(subCmd.GroupID) == 0 && (subCmd.IsAvailableCommand() || subCmd.Name() == "help") {
			subCmdSynopsis := theme.SubCommand.Render(subCmd.Name())
			if subCmd.HasAvailableSubCommands() {
				subCmdSynopsis += theme.DimmedArg.Render(" " + i18n.Sprintf("[command]"))
			}
			if subCmd.HasAvailableFlags() {
				subCmdSynopsis += theme.DimmedArg.Render(" " + i18n.Sprintf("[--flags]"))
			}
			rows = append(rows, []string{subCmdSynopsis, subCmd.Short})
		}
	}
	if len(rows) > 0 {
		fragments = append(fragments, renderParamTable(rows, width))
	}

	return strings.Join(fragments, "\n")
}

func renderFlags(flags *pflag.FlagSet, theme HelpTheme, width int) string {
	i18n := cfg.GetI18nPrinter()

	rows := [][]string{}

	flags.VisitAll(func(f *pflag.Flag) {
		if f.Hidden {
			return
		}

		var firstPart string
		if f.Shorthand != "" && f.ShorthandDeprecated == "" {
			firstPart = theme.Flag.Render(fmt.Sprintf("-%s, --%s", f.Shorthand, f.Name))
		} else {
			firstPart = theme.Flag.Render(fmt.Sprintf("    --%s", f.Name))
		}

		secondPart := f.Usage
		if f.DefValue != "" {
			switch f.Value.Type() {
			case "bool", "boolfunc":
				if f.DefValue != "false" && f.DefValue != "" {
					secondPart += i18n.Sprintf(" (default %s)", f.DefValue)
				}
			case "duration":
				if f.DefValue != "0s" {
					secondPart += i18n.Sprintf(" (default %s)", f.DefValue)
				}
			case "int", "int8", "int32", "int64", "uint", "uint8", "uint16", "uint32", "uint64", "count", "float32", "float64":
				if f.DefValue != "0" {
					secondPart += i18n.Sprintf(" (default %s)", f.DefValue)
				}
			case "string", "ip", "ipMask", "ipNet":
				if f.DefValue != "" {
					secondPart += i18n.Sprintf(" (default %q)", f.DefValue)
				}
			case "boolSlice", "intSlice", "stringSlice", "stingArray":
				if f.DefValue != "[]" {
					secondPart += i18n.Sprintf(" (default %s)", f.DefValue)
				}
			default:
				if f.DefValue != "false" && f.DefValue != "<nil>" && f.DefValue != "" && f.DefValue != "0" && f.DefValue != "[]" {
					secondPart += i18n.Sprintf(" (default %s)", f.DefValue)
				}
			}
		}

		rows = append(rows, []string{firstPart, secondPart})
	})

	return renderParamTable(rows, width)
}

func RenderUsage(cmd *cobra.Command) error {
	i18n := cfg.GetI18nPrinter()

	cmd.InitDefaultHelpFlag()
	cmd.InitDefaultVersionFlag()

	theme := LoadHelpTheme()
	output := cmd.OutOrStderr()

	fragments := make([]string, 0)

	usageFragments := make([]string, 0)
	if cmd.Runnable() {
		cmdSuffix := strings.TrimLeftFunc(strings.Replace(cmd.Use, cmd.Name(), "", 1), unicode.IsSpace)
		cmdSynopsis := renderCommandPath(cmd, theme.CodeblockCommand, theme.CodeblockSubCommand)
		if len(cmdSuffix) > 0 {
			cmdSynopsis += theme.CodeblockDimmedArg.Render(" " + cmdSuffix)
		}

		flagRegexp := regexp.MustCompile(`\[(((--[a-zA-Z-]+|-[a-zA-Z])( [^\[\]]+)?)|flags)\]`)
		if !cmd.DisableFlagsInUseLine && cmd.HasAvailableFlags() && !flagRegexp.MatchString(cmdSynopsis) {
			cmdSynopsis += theme.CodeblockDimmedArg.Render(" " + i18n.Sprintf("[--flags]"))
		}

		usageFragments = append(usageFragments, cmdSynopsis)
	}

	if cmd.HasAvailableSubCommands() {
		subCommandUsage := renderCommandPath(cmd, theme.CodeblockCommand, theme.CodeblockSubCommand)
		subCommandUsage += theme.CodeblockDimmedArg.Render(" " + i18n.Sprintf("[command]"))

		usageFragments = append(usageFragments, subCommandUsage)
	}

	if len(usageFragments) > 0 {
		synopsisContent := lipgloss.JoinVertical(lipgloss.Left, usageFragments...)
		sypopsisWidth := min(max(lipgloss.Width(synopsisContent), 80), 120)
		fragments = append(
			fragments,
			theme.BaseTitle.Render(i18n.Sprintf("USAGE")),
			theme.Codeblock.
				Margin(1, 1).
				Width(sypopsisWidth+theme.Codeblock.GetHorizontalPadding()+theme.Codeblock.GetHorizontalBorderSize()).
				Render(synopsisContent),
		)
	}

	if len(cmd.Aliases) > 0 {
		style := theme.SubCommand
		if !cmd.HasParent() {
			style = theme.Command
		}
		nameAndAliases := make([]string, 0)
		nameAndAliases = append(nameAndAliases, style.Render(cmd.Name()))
		for _, alias := range cmd.Aliases {
			nameAndAliases = append(nameAndAliases, style.Render(alias))
		}
		aliasBlock := lipgloss.JoinVertical(lipgloss.Left, nameAndAliases...)

		fragments = append(
			fragments,
			theme.BaseTitle.Render(i18n.Sprintf("ALIASES")),
			lipgloss.NewStyle().Margin(1, 2).Render(aliasBlock),
		)
	}

	if len(cmd.Example) > 0 {
		fragments = append(
			fragments,
			theme.BaseTitle.Render(i18n.Sprintf("EXAMPLE")),
			theme.Codeblock.Margin(1, 1).Render(cmd.Example),
		)
	}

	if cmd.HasAvailableSubCommands() {
		fragments = append(
			fragments,
			theme.BaseTitle.Render(i18n.Sprintf("COMMANDS")),
			"",
			renderCommands(cmd, theme, width()),
			"",
		)
	}

	if cmd.HasAvailableFlags() {
		fragments = append(
			fragments,
			theme.BaseTitle.Render("FLAGS"),
			"",
			renderFlags(cmd.Flags(), theme, width()),
			"",
		)
	}

	additionalHelp := make([][]string, 0, len(cmd.Commands()))
	for _, subcmd := range cmd.Commands() {
		if subcmd.IsAdditionalHelpTopicCommand() {
			additionalHelp = append(additionalHelp, []string{renderCommandPath(subcmd, theme.Command, theme.SubCommand), subcmd.Short})
		}
	}
	if len(additionalHelp) > 0 {
		fragments = append(
			fragments,
			theme.BaseTitle.Render(i18n.Sprintf("ADDITIONNAL HELP")),
			"",
			renderParamTable(additionalHelp, width()),
			"",
		)
	}

	lines := strings.Lines(strings.TrimRightFunc(lipgloss.JoinVertical(lipgloss.Left, fragments...), unicode.IsSpace))
	for line := range lines {
		_, _ = fmt.Fprintln(output, strings.TrimRightFunc(line, unicode.IsSpace))
	}

	return nil
}

func RenderHelp(cmd *cobra.Command, _ []string) {
	// TODO: Add a sexy header

	output := cmd.OutOrStdout()

	usage := cmd.Long
	if len(usage) == 0 {
		usage = cmd.Short
	}
	usage = strings.TrimRightFunc(usage, unicode.IsSpace)
	if len(usage) > 0 {
		fmt.Fprintln(output, usage)
		fmt.Fprintln(output)
	}

	if cmd.Runnable() || cmd.HasSubCommands() {
		fmt.Fprintln(output, cmd.UsageString())
	}
}
