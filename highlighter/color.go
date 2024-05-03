package highlighter

import "github.com/fatih/color"

type Color struct {
	color *color.Color
}

func NewColor(color *color.Color) *Color {
	return &Color{color: color}
}

func (c *Color) Highlight(s string) string {
	return c.color.Sprintf(s)
}
