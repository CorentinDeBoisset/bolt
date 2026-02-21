package pkg

import "github.com/charmbracelet/lipgloss"

// The original palette is available here: https://coolors.co/palette/264653-2a9d8f-e9c46a-f4a261-e76f51

const refCharcoalBlue11 = "#122126"
const refCharcoalBlue14 = "#162931"
const refCharcoalBlue24 = "#264653"
const refCharcoalBlue34 = "#366577"
const refCharcoalBlue44 = "#46829a"
const refCharcoalBlue55 = "#619fb7"
const refCharcoalBlue66 = "#88b6c9"
const refCharcoalBlue77 = "#aeceda"
const refCharcoalBlue83 = "#c3dae4"
const refCharcoalBlue9 = "#eaf2f5"

const refVerdigris1 = "#0c2c29"
const refVerdigris2 = "#15514a"
const refVerdigris3 = "#1f756b"
const refVerdigris4 = "#2a9d8f"
const refVerdigris5 = "#36c9b8"
const refVerdigris6 = "#62d5c8"
const refVerdigris7 = "#8ee1d7"
const refVerdigris8 = "#bbede7"
const refVerdigris9 = "#d3f3f0"

const refJasmine1 = "#312507"
const refJasmine2 = "#624a0f"
const refJasmine3 = "#926f16"
const refJasmine4 = "#c3941d"
const refJasmine5 = "#e1b137"
const refJasmine6 = "#e9c46a"
const refJasmine7 = "#f0d799"
const refJasmine8 = "#f4e3b8"
const refJasmine9 = "#f8ecce"

const refSandyBrown1 = "#341a04"
const refSandyBrown2 = "#693307"
const refSandyBrown3 = "#9d4d0b"
const refSandyBrown4 = "#d2660f"
const refSandyBrown5 = "#f08228"
const refSandyBrown6 = "#f4a261"
const refSandyBrown7 = "#f7bf91"
const refSandyBrown8 = "#fad3b3"
const refSandyBrown9 = "#fbe1cb"

const refBurntPeach1 = "#310f07"
const refBurntPeach2 = "#5e1d0d"
const refBurntPeach3 = "#8b2b13"
const refBurntPeach4 = "#b83919"
const refBurntPeach5 = "#e14923"
const refBurntPeach6 = "#e76f51"
const refBurntPeach7 = "#ee9781"
const refBurntPeach8 = "#f5c0b3"
const refBurntPeach9 = "#f9dad2"

var BackgroundColor = lipgloss.AdaptiveColor{Light: refJasmine7, Dark: refCharcoalBlue11}
var NoticeableBackgroundColor = lipgloss.AdaptiveColor{Light: refJasmine6, Dark: refCharcoalBlue14}
var AccentBackgroundColor = lipgloss.AdaptiveColor{Light: refJasmine6, Dark: refCharcoalBlue34}
var HighlightBackgroundColor = lipgloss.AdaptiveColor{Light: refJasmine4, Dark: refCharcoalBlue44}

var BodyColor = lipgloss.AdaptiveColor{Light: refSandyBrown2, Dark: refSandyBrown8}

var SecondaryTextColor = lipgloss.AdaptiveColor{Light: refJasmine8, Dark: refCharcoalBlue14}

var SeparatorColor = lipgloss.AdaptiveColor{Light: refVerdigris4, Dark: refSandyBrown3}

var SpinnerColor = lipgloss.AdaptiveColor{Light: refVerdigris4, Dark: refBurntPeach5}

var BaseAppStyle = lipgloss.NewStyle().
	Background(BackgroundColor).
	BorderBackground(BackgroundColor).
	MarginBackground(BackgroundColor)
