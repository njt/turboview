package tv

import (
	"fmt"
	"unicode/utf8"

	"github.com/gdamore/tcell/v2"
)

// FileInfoPane displays a single-line summary of a file entry:
// filename on the left, size/date/time on the right.
type FileInfoPane struct {
	BaseView
	entry *FileEntry
}

// NewFileInfoPane creates a FileInfoPane with the given bounds.
// It is visible and receives post-process broadcast events,
// but is not selectable.
func NewFileInfoPane(bounds Rect) *FileInfoPane {
	p := &FileInfoPane{}
	p.SetBounds(bounds)
	p.SetState(SfVisible, true)
	p.SetOptions(OfPostProcess, true)
	p.SetSelf(p)
	return p
}

// SetEntry stores the file entry to display. Nil is allowed and clears the pane.
func (p *FileInfoPane) SetEntry(e *FileEntry) {
	p.entry = e
}

// HandleEvent listens for EvBroadcast events with Command == CmFileFocused.
// It extracts the *FileEntry from event.Info and stores it.
// The event is never cleared (broadcasts are meant to reach all listeners).
func (p *FileInfoPane) HandleEvent(event *Event) {
	if event.What != EvBroadcast {
		return
	}
	if event.Command != CmFileFocused {
		return
	}

	entry, ok := event.Info.(*FileEntry)
	if !ok || entry == nil {
		return
	}

	p.entry = entry
}

// Draw renders the file info pane. It fills the background with spaces in
// ListNormal style, then writes the filename on the left and size/date/time
// on the right. Safe to call with nil ColorScheme and nil entry.
func (p *FileInfoPane) Draw(buf *DrawBuffer) {
	if p.entry == nil {
		return
	}

	width := p.Bounds().Width()
	if width < 4 {
		return
	}

	// Resolve style, falling back to StyleDefault when ColorScheme is nil.
	scheme := p.ColorScheme()
	style := tcell.StyleDefault
	if scheme != nil {
		style = scheme.ListNormal
	}

	sub := buf.SubBuffer(p.Bounds())

	// Fill the entire row with spaces in ListNormal style.
	for x := 0; x < width; x++ {
		sub.WriteChar(x, 0, ' ', style)
	}

	// --- Right side: size + date + time (computed first so left side
	//     can truncate based on remaining space) ----------------------------

	var sizeStr string
	if p.entry.IsDir {
		sizeStr = "<DIR>"
	} else {
		sizeStr = commaFormat(p.entry.Size)
	}

	dateStr := p.entry.ModTime.Format("Jan  2, 2006")
	timeStr := p.entry.ModTime.Format("3:04pm")

	rightText := sizeStr + "  " + dateStr + "  " + timeStr
	rightPos := width - len(rightText)
	if rightPos < 0 {
		rightPos = 0
	}

	// --- Left side: "  " + filename ---------------------------------------

	displayName := p.entry.Name
	if p.entry.IsDir {
		displayName += "/"
	}

	// Truncate with "~" if the name exceeds half the pane width, but also
	// ensure the left text does not overlap the right text area.
	maxChars := width / 2
	// The right side occupies len(rightText) characters ending at the right
	// edge.  The left text starts at column 0 with "  " (2-char prefix).
	// So the name body touches the right text at column len(rightText) + 2.
	availableForName := width - len(rightText) - 2
	if availableForName < maxChars {
		maxChars = availableForName
	}
	if maxChars < 1 {
		maxChars = 1
	}

	if utf8.RuneCountInString(displayName) > maxChars {
		runes := []rune(displayName)
		runes = runes[:maxChars-1]
		runes = append(runes, '~')
		displayName = string(runes)
	}

	leftText := "  " + displayName
	sub.WriteStr(0, 0, leftText, style)
	sub.WriteStr(rightPos, 0, rightText, style)
}

// commaFormat formats an int64 with thousand-separator commas.
// Examples: 0 -> "0", 1234 -> "1,234", 1000000 -> "1,000,000".
func commaFormat(n int64) string {
	if n == 0 {
		return "0"
	}

	isNeg := n < 0
	if isNeg {
		n = -n
	}

	s := fmt.Sprintf("%d", n)
	var result []byte
	for i := 0; i < len(s); i++ {
		if i > 0 && (len(s)-i)%3 == 0 {
			result = append(result, ',')
		}
		result = append(result, s[i])
	}

	if isNeg {
		return "-" + string(result)
	}
	return string(result)
}
