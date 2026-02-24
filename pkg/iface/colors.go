package iface

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/lucasb-eyer/go-colorful"
	"github.com/muesli/termenv"
)

type Theme struct {
	NoticeableSurfaceStyle        lipgloss.Style
	AccentSurfaceStyle            lipgloss.Style
	HighlightSurfaceStyle         lipgloss.Style
	InvertedHighlightSurfaceStyle lipgloss.Style
	InvertedAccentSurfaceStyle    lipgloss.Style

	SeparatorColor           lipgloss.TerminalColor
	BlurredOutputBorderColor lipgloss.TerminalColor
	FocusedOutputBorderColor lipgloss.TerminalColor
}

// The base palette is available here: https://coolors.co/palette/264653-2a9d8f-e9c46a-f4a261-e76f51

var ErrorColor = lipgloss.Color("1")

func LoadTheme() Theme {
	// bgColor is the actual color of the terminal's background
	bgColor := termenv.ConvertToRGB(termenv.DefaultOutput().BackgroundColor())
	bgH, bgS, bgL := bgColor.Hsl()

	var noticeableSurfaceColor, accentSurfaceColor, highlightSurfaceColor colorful.Color
	var invertedAccentSurfaceColor, invertedHighlightSurfaceColor colorful.Color
	var bodyColorOnNoticeable, bodyColorOnAccent, bodyColorOnHighlight colorful.Color
	var bodyColorOnInvertedAccent, bodyColorOnInvertedHighlight colorful.Color
	var separatorCol, focusedOutputBorderColor colorful.Color

	if bgL < 0.22 {
		noticeableSurfaceColor = colorful.Hsl(bgH, bgS, 0.15+0.2*bgL).Clamped()
		bodyColorOnNoticeable = colorful.Hsl(43, 0.58, 0.8+0.2*bgL).Clamped()

		accentSurfaceColor = colorful.Hsl(92, 0.37, 0.14+0.15*bgL)
		bodyColorOnAccent = colorful.Hsl(43, 0.58, 0.75)

		highlightSurfaceColor = colorful.Hsl(92, 0.37, 0.15+0.3*bgL)
		bodyColorOnHighlight = colorful.Hsl(43, 1.0, 0.95)

		invertedAccentSurfaceColor = colorful.Hsl(26.7, 0.65, 0.55)
		bodyColorOnInvertedAccent = colorful.Hsl(0, 0, 0.07)

		invertedHighlightSurfaceColor = colorful.Hsl(26.7, 0.95, 0.95)
		bodyColorOnInvertedHighlight = colorful.Hsl(0, 0, 0.1)

		separatorCol = colorful.Hsl(0, 0, 0.4)
		focusedOutputBorderColor = colorful.Hsl(26, 0.87, 0.55)
	} else {
		noticeableSurfaceColor = colorful.Hsl(bgH, bgS, 0.95*bgL).Clamped()
		bodyColorOnNoticeable = colorful.Hsl(43, 0.58, 0.15*bgL).Clamped()

		accentSurfaceColor = colorful.Hsl(92, 0.27, 0.7)
		bodyColorOnAccent = colorful.Hsl(0, 0, 0.1)

		highlightSurfaceColor = colorful.Hsl(92, 0.27, 0.55)
		bodyColorOnHighlight = colorful.Hsl(43, 0.58, 0.1)

		invertedAccentSurfaceColor = colorful.Hsl(26, 0.87, 0.6+0.2*bgL)
		bodyColorOnInvertedAccent = colorful.Hsl(0, 0, 0.1)

		invertedHighlightSurfaceColor = colorful.Hsl(26, 0.9, 0.15+0.05*bgL)
		bodyColorOnInvertedHighlight = colorful.Hsl(0, 0, 0.95)

		separatorCol = colorful.Hsl(0, 0, 0.15)
		focusedOutputBorderColor = colorful.Hsl(26, 0.87, 0.67)
	}

	// Convert them all to lipgloss colors
	lgNoticeableColor := lipgloss.Color(noticeableSurfaceColor.Hex())
	lgBodyColorOnNoticeable := lipgloss.Color(bodyColorOnNoticeable.Hex())

	lgAccentSurfaceColor := lipgloss.Color(accentSurfaceColor.Hex())
	lgBodyColorOnAccent := lipgloss.Color(bodyColorOnAccent.Hex())

	lgHighlightSurfaceColor := lipgloss.Color(highlightSurfaceColor.Hex())
	lgBodyColorOnHighlight := lipgloss.Color(bodyColorOnHighlight.Hex())

	lgInvertedHighlightSurfaceColor := lipgloss.Color(invertedHighlightSurfaceColor.Hex())
	lgBodyColorOnInvertedHighlight := lipgloss.Color(bodyColorOnInvertedHighlight.Hex())

	lgInvertedAccentSurfaceColor := lipgloss.Color(invertedAccentSurfaceColor.Hex())
	lgBodyColorOnInvertedAccent := lipgloss.Color(bodyColorOnInvertedAccent.Hex())

	return Theme{
		NoticeableSurfaceStyle: lipgloss.NewStyle().
			Background(lgNoticeableColor).
			BorderBackground(lgNoticeableColor).
			MarginBackground(lgNoticeableColor).
			Foreground(lgBodyColorOnNoticeable),

		AccentSurfaceStyle: lipgloss.NewStyle().
			Background(lgAccentSurfaceColor).
			BorderBackground(lgAccentSurfaceColor).
			MarginBackground(lgAccentSurfaceColor).
			Foreground(lgBodyColorOnAccent),

		HighlightSurfaceStyle: lipgloss.NewStyle().
			Background(lgHighlightSurfaceColor).
			BorderBackground(lgHighlightSurfaceColor).
			MarginBackground(lgHighlightSurfaceColor).
			Foreground(lgBodyColorOnHighlight),

		InvertedHighlightSurfaceStyle: lipgloss.NewStyle().
			Background(lgInvertedHighlightSurfaceColor).
			BorderBackground(lgInvertedHighlightSurfaceColor).
			MarginBackground(lgInvertedHighlightSurfaceColor).
			Foreground(lgBodyColorOnInvertedHighlight),

		InvertedAccentSurfaceStyle: lipgloss.NewStyle().
			Background(lgInvertedAccentSurfaceColor).
			BorderBackground(lgInvertedAccentSurfaceColor).
			MarginBackground(lgInvertedAccentSurfaceColor).
			Foreground(lgBodyColorOnInvertedAccent),

		SeparatorColor:           lipgloss.Color(separatorCol.Hex()),
		FocusedOutputBorderColor: lipgloss.Color(focusedOutputBorderColor.Hex()),
		BlurredOutputBorderColor: lipgloss.Color("#808080"),
	}
}
