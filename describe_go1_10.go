// +build go1.10

package describe

import (
	"strings"
)

type describer struct {
	indentStep     int
	currentIndent  int
	stringBuilder  strings.Builder
	referenceNames map[uintptr]int
	seenReferences map[uintptr]bool
}
