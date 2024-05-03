package highlighter

type pointer struct {
	pointer string
}

func Pointer(p string) Highlighter {
	return &pointer{pointer: p}
}

func (p *pointer) Highlight(s string) string {
	return p.pointer + s
}
