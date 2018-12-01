package assembler

type sourceLine struct {
	data     string
	filename string
	orgLinum int
}

type state struct {
	source []sourceLine
}
