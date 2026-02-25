package scrollbar

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func RenderScrollbar(scrollHeight, windowHeight int, offset float64, frameStyle lipgloss.Style) string {
	fragments := make([]string, 0)

	surroundingBorder := frameStyle.GetBorderStyle()

	if frameStyle.GetBorderTop() {
		fragments = append(fragments, surroundingBorder.TopRight)
	}

	if windowHeight <= 3 || scrollHeight <= windowHeight {
		// No scrollbar
		for range windowHeight {
			fragments = append(fragments, surroundingBorder.Right)
		}
	} else {
		offset = min(max(offset, 0.), 1.)

		scrollZoneHeight := windowHeight

		scrollBarHeight := max(3, ((scrollZoneHeight * scrollZoneHeight) / scrollHeight))
		scrollBarOffset := int(offset*float64(scrollZoneHeight-scrollBarHeight) + 0.4)

		for range scrollBarOffset {
			fragments = append(fragments, surroundingBorder.Right)
		}

		fragments = append(fragments, lipgloss.DoubleBorder().MiddleTop)
		for range scrollBarHeight - 2 {
			fragments = append(fragments, lipgloss.DoubleBorder().Right)
		}
		fragments = append(fragments, lipgloss.DoubleBorder().MiddleBottom)

		for range scrollZoneHeight - scrollBarOffset - scrollBarHeight {
			fragments = append(fragments, surroundingBorder.Right)
		}
	}

	if frameStyle.GetBorderBottom() {
		fragments = append(fragments, surroundingBorder.BottomRight)
	}

	return lipgloss.NewStyle().
		Foreground(frameStyle.GetBorderRightForeground()).
		Background(frameStyle.GetBorderRightBackground()).
		Render(strings.Join(fragments, "\n"))
}
