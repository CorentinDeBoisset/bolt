package iface

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/lucasb-eyer/go-colorful"
	"github.com/muesli/termenv"
)

type Theme struct {
	NoticeableSurfaceStyle lipgloss.Style
	AccentSurfaceStyle     lipgloss.Style
	HighlightSurfaceStyle  lipgloss.Style
	SeparatorColor         lipgloss.Color
}

// The base palette is available here: https://coolors.co/palette/264653-2a9d8f-e9c46a-f4a261-e76f51

var ErrorColor = lipgloss.Color("1")

var (
	FocusedOutputBorderColor = lipgloss.AdaptiveColor{
		Dark:  "#f08128",
		Light: "#f4a261",
	}

	BlurredOutputBorderColor = lipgloss.Color("#808080")
)

func LoadTheme() Theme {
	// bgColor is the actual color of the terminal's background
	bgColor := termenv.ConvertToRGB(termenv.DefaultOutput().BackgroundColor())
	bgH, bgS, bgL := bgColor.Hsl()

	var noticeableSurfaceColor, accentSurfaceColor, highlightSurfaceColor colorful.Color
	var bodyColorOnAccent, bodyColorOnHighlight colorful.Color
	var separatorCol colorful.Color

	if bgL < 0.5 {
		noticeableSurfaceColor = colorful.Hsl(bgH, bgS, 0.1+0.9*bgL).Clamped()
	} else {
		noticeableSurfaceColor = colorful.Hsl(bgH, bgS, 0.95*bgL).Clamped()
	}

	if bgL < 0.22 {
		accentSurfaceColor = colorful.Hsl(197, 0.37, 0.17)
		highlightSurfaceColor = colorful.Hsl(197, 0.37, 0.22)
		bodyColorOnAccent = colorful.Hsl(0, 0, 0.95)
		bodyColorOnHighlight = colorful.Hsl(0, 0, 0.95)
		separatorCol = colorful.Hsl(0, 0, 0.4)
	} else if bgL < 0.45 {
		accentSurfaceColor = colorful.Hsl(197, 0.37, 0.45)
		highlightSurfaceColor = colorful.Hsl(197, 0.37, 0.50)
		bodyColorOnAccent = colorful.Hsl(0, 0, 0.95)
		bodyColorOnHighlight = colorful.Hsl(0, 0, 0.95)
		separatorCol = colorful.Hsl(0, 0, 0.95)
	} else {
		accentSurfaceColor = colorful.Hsl(197, 0.7, 0.6)
		highlightSurfaceColor = colorful.Hsl(197, 0.7, 0.75)
		bodyColorOnAccent = colorful.Hsl(0, 0, 0.1)
		bodyColorOnHighlight = colorful.Hsl(0, 0, 0.1)
		separatorCol = colorful.Hsl(0, 0, 0.15)
	}

	// Convert them all to
	lgNoticeableColor := lipgloss.Color(noticeableSurfaceColor.Hex())
	lgAccentSurfaceColor := lipgloss.Color(accentSurfaceColor.Hex())
	lgHighlightSurfaceColor := lipgloss.Color(highlightSurfaceColor.Hex())

	lgBodyColorOnAccent := lipgloss.Color(bodyColorOnAccent.Hex())
	lgBodyColorOnHighlight := lipgloss.Color(bodyColorOnHighlight.Hex())

	return Theme{
		NoticeableSurfaceStyle: lipgloss.NewStyle().
			Background(lgNoticeableColor).
			BorderBackground(lgNoticeableColor).
			MarginBackground(lgNoticeableColor),

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

		SeparatorColor: lipgloss.Color(separatorCol.Hex()),
	}
}
