package highlighter

type multi struct {
	highlighters []Highlighter
}

func Multi(highlighters ...Highlighter) Highlighter {
	return &multi{highlighters: highlighters}
}

func (p *multi) Highlight(s string) string {
	for _, h := range p.highlighters {
		s = h.Highlight(s)
	}
	return s
}
