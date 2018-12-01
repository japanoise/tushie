package assembler

type sourceLine struct {
	data     string
	filename string
	orgLinum int
}

type labelData struct {
	name string
	addr uint64
}

type state struct {
	source []sourceLine
	labels map[string]labelData
}
