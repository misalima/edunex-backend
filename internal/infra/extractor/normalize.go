package extractor

import (
	"regexp"
	"strings"
)

var (
	horizontalSpaceRe = regexp.MustCompile(`[ \t]+`)
	verticalSpaceRe   = regexp.MustCompile(`\n{3,}`)
)

// normalizeText removes extra whitespace while preserving paragraph structure.
func normalizeText(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")

	s = horizontalSpaceRe.ReplaceAllString(s, " ")

	lines := strings.Split(s, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimSpace(line)
	}
	s = strings.Join(lines, "\n")

	s = verticalSpaceRe.ReplaceAllString(s, "\n\n")

	return strings.TrimSpace(s)
}
