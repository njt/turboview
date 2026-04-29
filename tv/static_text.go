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

	lines := wrapText(st.text, w)
	for y, line := range lines {
		if y >= h {
			break
		}
		centered := false
		runes := []rune(line)
		if len(runes) > 0 && runes[0] == '\x03' {
			centered = true
			runes = runes[1:]
		}
		x := 0
		if centered {
			x = (w - len(runes)) / 2
			if x < 0 {
				x = 0
			}
		}
		for _, r := range runes {
			if x < w {
				buf.WriteChar(x, y, r, style)
			}
			x++
		}
	}
}

// wrapText splits text into display lines, preserving \x03 centering prefix
// per line and honoring the given width for word-wrapping.
func wrapText(text string, width int) []string {
	rawLines := splitOnNewlines(text)
	var result []string
	for _, raw := range rawLines {
		prefix := ""
		content := raw
		if len(raw) > 0 && raw[0] == '\x03' {
			prefix = "\x03"
			content = raw[1:]
		}
		words := splitWords(content)
		if len(words) == 0 {
			result = append(result, prefix)
			continue
		}
		line := prefix
		lineLen := 0
		for _, word := range words {
			wLen := len([]rune(word))
			if lineLen > 0 && lineLen+1+wLen > width {
				result = append(result, line)
				line = word
				lineLen = wLen
			} else {
				if lineLen > 0 {
					line += " "
					lineLen++
				}
				line += word
				lineLen += wLen
			}
		}
		result = append(result, line)
	}
	return result
}

// splitOnNewlines splits s on '\n' characters, returning at least one element.
func splitOnNewlines(s string) []string {
	var lines []string
	current := ""
	for _, r := range s {
		if r == '\n' {
			lines = append(lines, current)
			current = ""
		} else {
			current += string(r)
		}
	}
	lines = append(lines, current)
	return lines
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
