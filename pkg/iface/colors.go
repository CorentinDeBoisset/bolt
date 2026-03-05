package iface

import (
	"image/color"
	"os"
	"sync"

	"charm.land/lipgloss/v2"
	"github.com/lucasb-eyer/go-colorful"
)

type HelpTheme struct {
	BaseTitle  lipgloss.Style
	ErrorTitle lipgloss.Style

	Command    lipgloss.Style
	SubCommand lipgloss.Style
	DimmedArg  lipgloss.Style
	Flag       lipgloss.Style

	Codeblock           lipgloss.Style
	CodeblockBase       lipgloss.Style
	CodeblockCommand    lipgloss.Style
	CodeblockSubCommand lipgloss.Style
	CodeblockDimmedArg  lipgloss.Style
	CodeblockFlag       lipgloss.Style
}

type Theme struct {
	NoticeableSurfaceStyle                 lipgloss.Style
	UnfocusedHighlightSurfaceStyle         lipgloss.Style
	HighlightSurfaceStyle                  lipgloss.Style
	InvertedHighlightSurfaceStyle          lipgloss.Style
	InvertedUnfocusedHighlightSurfaceStyle lipgloss.Style

	SeparatorColor           color.Color
	BlurredOutputBorderColor color.Color
	FocusedOutputBorderColor color.Color
}

// This is the actual color of the background of the terminal
var BackgroundColor = sync.OnceValue(func() colorful.Color {
	rawColor, err := lipgloss.BackgroundColor(os.Stdin, os.Stderr)
	if err != nil || rawColor == nil {
		return colorful.Hsl(0, 0, 0)
	}
	bgColor, ok := colorful.MakeColor(rawColor)
	if !ok {
		return colorful.Hsl(0, 0, 0)
	}
	return bgColor
})

// The base palette is available here: https://coolors.co/palette/264653-2a9d8f-e9c46a-f4a261-e76f51

var ErrorColor = lipgloss.Color("1")

func LoadTheme() Theme {
	bgH, bgS, bgL := BackgroundColor().Hsl()

	var noticeableSurfaceColor, unfocusedHighlightSurfaceColor, highlightSurfaceColor colorful.Color
	var invertedUnfocusedHighlightSurfaceColor, invertedHighlightSurfaceColor colorful.Color
	var bodyColorOnNoticeable, bodyColorOnUnfocusedHighlight, bodyColorOnHighlight colorful.Color
	var bodyColorOnInvertedUnfocusedHighlight, bodyColorOnInvertedHighlight colorful.Color
	var separatorCol, focusedOutputBorderColor colorful.Color

	if bgL < 0.22 {
		noticeableSurfaceColor = colorful.Hsl(bgH, bgS, 0.15+0.2*bgL).Clamped()
		bodyColorOnNoticeable = colorful.Hsl(43, 0.58, 0.8+0.2*bgL).Clamped()

		unfocusedHighlightSurfaceColor = colorful.Hsl(92, 0.20, 0.14+0.15*bgL).Clamped()
		bodyColorOnUnfocusedHighlight = colorful.Hsl(43, 0.58, 0.75)

		highlightSurfaceColor = colorful.Hsl(92, 0.37, 0.15+0.3*bgL).Clamped()
		bodyColorOnHighlight = colorful.Hsl(43, 1.0, 0.95)

		invertedUnfocusedHighlightSurfaceColor = colorful.Hsl(26.7, 0.65, 0.55)
		bodyColorOnInvertedUnfocusedHighlight = colorful.Hsl(0, 0, 0.07)

		invertedHighlightSurfaceColor = colorful.Hsl(26.7, 0.95, 0.95)
		bodyColorOnInvertedHighlight = colorful.Hsl(0, 0, 0.1)

		separatorCol = colorful.Hsl(0, 0, 0.4)
		focusedOutputBorderColor = colorful.Hsl(26, 0.87, 0.55)
	} else {
		noticeableSurfaceColor = colorful.Hsl(bgH, 0.05+0.4*bgS, 0.95*bgL).Clamped()
		bodyColorOnNoticeable = colorful.Hsl(43, 0.58, 0.15*bgL).Clamped()

		unfocusedHighlightSurfaceColor = colorful.Hsl(92, 0.27, 0.85*bgL).Clamped()
		bodyColorOnUnfocusedHighlight = colorful.Hsl(0, 0, 0.1)

		highlightSurfaceColor = colorful.Hsl(92, 0.27, 0.55)
		bodyColorOnHighlight = colorful.Hsl(43, 0.58, 0.1)

		invertedUnfocusedHighlightSurfaceColor = colorful.Hsl(26, 0.87, 0.6+0.2*bgL).Clamped()
		bodyColorOnInvertedUnfocusedHighlight = colorful.Hsl(0, 0, 0.1)

		invertedHighlightSurfaceColor = colorful.Hsl(26, 0.9, 0.15+0.05*bgL).Clamped()
		bodyColorOnInvertedHighlight = colorful.Hsl(0, 0, 0.95)

		separatorCol = colorful.Hsl(0, 0, 0.15)
		focusedOutputBorderColor = colorful.Hsl(26, 0.87, 0.67)
	}

	// Convert them all to lipgloss colors
	lgNoticeableColor := lipgloss.Color(noticeableSurfaceColor.Hex())
	lgBodyColorOnNoticeable := lipgloss.Color(bodyColorOnNoticeable.Hex())

	lgUnfocusedHighlightSurfaceColor := lipgloss.Color(unfocusedHighlightSurfaceColor.Hex())
	lgBodyColorOnUnfocusedHighlight := lipgloss.Color(bodyColorOnUnfocusedHighlight.Hex())

	lgHighlightSurfaceColor := lipgloss.Color(highlightSurfaceColor.Hex())
	lgBodyColorOnHighlight := lipgloss.Color(bodyColorOnHighlight.Hex())

	lgInvertedHighlightSurfaceColor := lipgloss.Color(invertedHighlightSurfaceColor.Hex())
	lgBodyColorOnInvertedHighlight := lipgloss.Color(bodyColorOnInvertedHighlight.Hex())

	lgInvertedUnfocusedHighlightSurfaceColor := lipgloss.Color(invertedUnfocusedHighlightSurfaceColor.Hex())
	lgBodyColorOnInvertedUnfocusedHighlight := lipgloss.Color(bodyColorOnInvertedUnfocusedHighlight.Hex())

	return Theme{
		NoticeableSurfaceStyle: lipgloss.NewStyle().
			Background(lgNoticeableColor).
			BorderBackground(lgNoticeableColor).
			Foreground(lgBodyColorOnNoticeable),

		UnfocusedHighlightSurfaceStyle: lipgloss.NewStyle().
			Background(lgUnfocusedHighlightSurfaceColor).
			BorderBackground(lgUnfocusedHighlightSurfaceColor).
			Foreground(lgBodyColorOnUnfocusedHighlight),

		HighlightSurfaceStyle: lipgloss.NewStyle().
			Background(lgHighlightSurfaceColor).
			BorderBackground(lgHighlightSurfaceColor).
			Foreground(lgBodyColorOnHighlight),

		InvertedHighlightSurfaceStyle: lipgloss.NewStyle().
			Background(lgInvertedHighlightSurfaceColor).
			BorderBackground(lgInvertedHighlightSurfaceColor).
			Foreground(lgBodyColorOnInvertedHighlight),

		InvertedUnfocusedHighlightSurfaceStyle: lipgloss.NewStyle().
			Background(lgInvertedUnfocusedHighlightSurfaceColor).
			BorderBackground(lgInvertedUnfocusedHighlightSurfaceColor).
			Foreground(lgBodyColorOnInvertedUnfocusedHighlight),

		SeparatorColor:           lipgloss.Color(separatorCol.Hex()),
		FocusedOutputBorderColor: lipgloss.Color(focusedOutputBorderColor.Hex()),
		BlurredOutputBorderColor: lipgloss.Color("#808080"),
	}
}

func LoadHelpTheme() HelpTheme {
	bgH, bgS, bgL := BackgroundColor().Hsl()
	lightDark := lipgloss.LightDark(bgL < 0.5)

	var codeblockSurface colorful.Color
	var codeblockForeground colorful.Color

	if bgL < 0.22 {
		codeblockSurface = colorful.Hsl(bgH, bgS, 0.15+0.2*bgL).Clamped()
		codeblockForeground = colorful.Hsl(43, 0.58, 0.8+0.2*bgL).Clamped()
	} else {
		codeblockSurface = colorful.Hsl(bgH, bgS, 0.95*bgL).Clamped()
		codeblockForeground = colorful.Hsl(43, 0.58, 0.15*bgL).Clamped()
	}

	lgCodeblockSurface := lipgloss.Color(codeblockSurface.Hex())

	baseCodeBlock := lipgloss.NewStyle().Background(lgCodeblockSurface)

	// TODO: use real colors instead of boilerplate. The commented stuff was a first iteration.
	//
	// titleStyle := baseTitleStyle.Foreground(lipgloss.AdaptiveColor{Dark: "#52901b", Light: "#72a443"})
	// cmdStyle := lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Dark: "#52901b", Light: "#72a443"})
	// subCmdStyle := lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Dark: "#bc7025", Light: "#bc702"}) // FIXME colors
	// dimmedArgStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#808080"))
	// codeblockStyle := theme.NoticeableSurfaceStyle.Width(width()-2).Padding(1, 2).Margin(0, 1)
	// // highlightedArgStyle := lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Dark: "#5b6ce9", Light: "#838ee2"}) // FIXME colors

	return HelpTheme{
		BaseTitle:  lipgloss.NewStyle().Foreground(lipgloss.Color("#508000")),
		ErrorTitle: lipgloss.NewStyle().Background(lipgloss.Red).Foreground(lightDark(lipgloss.BrightWhite, lipgloss.Black)),

		Command:    lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000")),
		SubCommand: lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFF00")),
		DimmedArg:  lipgloss.NewStyle().Foreground(lipgloss.Color("#808080")),
		Flag:       lipgloss.NewStyle().Foreground(lipgloss.Color("#508000")),

		Codeblock:           baseCodeBlock.Padding(1, 2).Margin(0, 1),
		CodeblockBase:       baseCodeBlock.Foreground(lipgloss.Color(codeblockForeground.Hex())),
		CodeblockCommand:    baseCodeBlock.Foreground(lipgloss.Color("#FF0000")),
		CodeblockSubCommand: baseCodeBlock.Foreground(lipgloss.Color("#AFAF00")),
		CodeblockDimmedArg:  baseCodeBlock.Foreground(lipgloss.Color("#808080")),
		CodeblockFlag:       baseCodeBlock.Foreground(lipgloss.Color("#508000")),
	}
}
