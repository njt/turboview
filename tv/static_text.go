package tv

import "github.com/gdamore/tcell/v2"

var _ Widget = (*StaticText)(nil)

type StaticText struct {
	BaseView
	text string
}

func NewStaticText(bounds Rect, text string) *StaticText {
	st := &StaticText{text: text}
	st.SetBounds(bounds)
	st.SetState(SfVisible, true)
	st.SetSelf(st)
	return st
}

func (st *StaticText) Text() string     { return st.text }
func (st *StaticText) SetText(t string) { st.text = t }

func (st *StaticText) Draw(buf *DrawBuffer) {
	w := st.Bounds().Width()
	h := st.Bounds().Height()
	if w <= 0 || h <= 0 {
		return
	}

	style := tcell.StyleDefault
	if cs := st.ColorScheme(); cs != nil {
		style = cs.LabelNormal
	}

	x, y := 0, 0
	words := splitWords(st.text)
	for _, word := range words {
		runes := []rune(word)
		if x > 0 && x+len(runes) > w {
			x = 0
			y++
			if y >= h {
				return
			}
		}
		for _, r := range runes {
			if r == '\n' {
				x = 0
				y++
				if y >= h {
					return
				}
				continue
			}
			if x < w {
				buf.WriteChar(x, y, r, style)
			}
			x++
		}
		if x < w {
			x++ // space after word
		}
	}
}

func splitWords(s string) []string {
	var words []string
	current := ""
	for _, r := range s {
		if r == ' ' {
			if current != "" {
				words = append(words, current)
				current = ""
			}
		} else {
			current += string(r)
		}
	}
	if current != "" {
		words = append(words, current)
	}
	return words
}
