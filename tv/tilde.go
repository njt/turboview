package tv

type LabelSegment struct {
	Text     string
	Shortcut bool
}

func ParseTildeLabel(label string) []LabelSegment {
	var segments []LabelSegment
	inTilde := false
	current := ""
	for _, r := range label {
		if r == '~' {
			if current != "" {
				segments = append(segments, LabelSegment{Text: current, Shortcut: inTilde})
				current = ""
			}
			inTilde = !inTilde
			continue
		}
		current += string(r)
	}
	if current != "" {
		segments = append(segments, LabelSegment{Text: current, Shortcut: inTilde})
	}
	return segments
}
