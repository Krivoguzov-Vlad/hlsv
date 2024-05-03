package highlighter

type Highlighter interface {
	Highlight(s string) string
}
