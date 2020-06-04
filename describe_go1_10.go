// +build go1.10

package describe

import (
	"strings"

	"github.com/kstenerud/go-duplicates"
)

type describer struct {
	indentStep     int
	currentIndent  int
	stringBuilder  strings.Builder
	referenceNames map[duplicates.TypedPointer]int
	seenReferences map[duplicates.TypedPointer]bool
}
