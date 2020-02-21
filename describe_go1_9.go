// +build !go1.10

package describe

type describer struct {
	indentStep     int
	currentIndent  int
	stringBuilder  stringBuilder
	referenceNames map[uintptr]int
	seenReferences map[uintptr]bool
}

type stringBuilder struct {
	buffer []byte
}

func (this *stringBuilder) WriteString(value string) {
	this.buffer = append(this.buffer, []byte(value)...)
}

func (this *stringBuilder) String() string {
	return string(this.buffer)
}

func (this *stringBuilder) Reset() {
	this.buffer = this.buffer[:0]
}
