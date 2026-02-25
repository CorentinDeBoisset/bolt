package iface

import (
	"bytes"
	"regexp"

	"github.com/charmbracelet/x/ansi"
)

type sequence struct {
	content []byte
	length  int // The number of skipped bytes
	visible bool
}

func prepareSequence(input []byte) []sequence {
	var state byte
	sequences := make([]sequence, 0)
	for len(input) > 0 {
		seq, width, n, newState := ansi.DecodeSequence(input, state, nil)

		sequences = append(sequences, sequence{content: seq, length: n, visible: width > 0})

		state = newState
		input = input[n:]
	}

	return sequences
}

func findAllInSequence(r *regexp.Regexp, sequences []sequence) [][]int {
	filteredContent := make([]byte, 0)
	for _, seq := range sequences {
		if seq.visible {
			filteredContent = append(filteredContent, seq.content...)
		}
	}

	currentSeqIdx := 0
	totalLen := 0
	visibleLen := 0

	// skippedCount := 0
	matchesOnFiltered := r.FindAllIndex(filteredContent, -1)
	if matchesOnFiltered == nil {
		return nil
	}

	originalMatches := make([][]int, len(matchesOnFiltered))
	for matchIdx, match := range matchesOnFiltered {
		// Go forward as long as a skipped range is before the start of the match
		for currentSeqIdx < len(sequences) && visibleLen < match[0] {
			totalLen += sequences[currentSeqIdx].length
			if sequences[currentSeqIdx].visible {
				visibleLen += sequences[currentSeqIdx].length
			}
			currentSeqIdx++
		}

		originalStartIdx := totalLen

		// Go forward as long as skipped ranges are before the end of the match
		for currentSeqIdx < len(sequences) && visibleLen < match[1] {
			totalLen += sequences[currentSeqIdx].length
			if sequences[currentSeqIdx].visible {
				visibleLen += sequences[currentSeqIdx].length
			}
			currentSeqIdx++
		}

		originalEndIdx := totalLen

		originalMatches[matchIdx] = []int{originalStartIdx, originalEndIdx}
	}

	return originalMatches
}

func DecorateCmdOutput(r *regexp.Regexp, content []byte, highLightIdx int, theme Theme) ([]byte, []int) {
	sequences := prepareSequence(content)
	matches := findAllInSequence(r, sequences)
	if matches == nil {
		return content, nil
	}

	output := make([]byte, 0)
	currentOffset := 0
	currentMatchIdx := 0
	lineNb := 0
	matchLines := make([]int, len(matches))
	for _, sequence := range sequences {
		if bytes.Equal(sequence.content, []byte("\n")) {
			lineNb++
		}

		for currentMatchIdx < len(matches) && currentOffset >= matches[currentMatchIdx][1] {
			matchLines[currentMatchIdx] = lineNb
			currentMatchIdx++
		}

		if currentMatchIdx < len(matches) && currentOffset >= matches[currentMatchIdx][0] && sequence.visible {
			if currentMatchIdx == highLightIdx {
				output = append(output, []byte(theme.InvertedHighlightSurfaceStyle.Render(string(sequence.content)))...)
			} else {
				output = append(output, []byte(theme.InvertedAccentSurfaceStyle.Render(string(sequence.content)))...)
			}
		} else {
			output = append(output, sequence.content...)
		}

		currentOffset += sequence.length
	}

	return output, matchLines
}
