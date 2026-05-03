// tv/edit_window.go
package tv

import "path/filepath"

type EditWindow struct {
	*Window
	editor    *Editor
	indicator *Indicator
}

const editWindowIndicatorWidth = 14

func NewEditWindow(bounds Rect, filename string, opts ...WindowOption) *EditWindow {
	title := "Untitled"
	if filename != "" {
		title = filepath.Base(filename)
	}

	w := bounds.Width()
	h := bounds.Height()
	if w < 24 {
		w = 24
	}
	if h < 6 {
		h = 6
	}
	bounds = NewRect(bounds.A.X, bounds.A.Y, w, h)

	win := NewWindow(bounds, title, opts...)
	ew := &EditWindow{Window: win}

	clientW := max(w-2, 0)
	clientH := max(h-2, 0)

	// Layout: Editor fills top rows. Bottom row shared by Indicator (left) and HScroll (right).
	// VScroll on right edge, above bottom row.
	bottomY := clientH - 1

	vScroll := NewScrollBar(NewRect(clientW-1, 0, 1, bottomY), Vertical)
	vScroll.SetGrowMode(GfGrowLoX | GfGrowHiX | GfGrowHiY)

	indW := editWindowIndicatorWidth
	if indW > clientW-2 {
		indW = clientW - 2
	}
	hScrollX := indW
	hScrollW := max(clientW-1-hScrollX, 0)
	hScroll := NewScrollBar(NewRect(hScrollX, bottomY, hScrollW, 1), Horizontal)
	hScroll.SetGrowMode(GfGrowLoY | GfGrowHiY | GfGrowHiX)

	editorW := max(clientW-1, 0)
	editorH := max(bottomY, 0)
	editor := NewEditor(NewRect(0, 0, editorW, editorH))
	editor.SetVScrollBar(vScroll)
	editor.SetHScrollBar(hScroll)
	editor.SetGrowMode(GfGrowHiX | GfGrowHiY)

	ind := NewIndicator(NewRect(0, bottomY, indW, 1))
	ind.SetGrowMode(GfGrowLoY | GfGrowHiY)

	ew.editor = editor
	ew.indicator = ind

	win.Insert(editor)
	win.Insert(vScroll)
	win.Insert(hScroll)
	win.Insert(ind)

	if filename != "" {
		editor.LoadFile(filename)
	}

	return ew
}

func (ew *EditWindow) Editor() *Editor { return ew.editor }

func (ew *EditWindow) HandleEvent(event *Event) {
	if event.What == EvCommand && event.Command == CmClose {
		if !ew.Valid(CmClose) {
			event.Clear()
			return
		}
	}
	ew.Window.HandleEvent(event)
}

func (ew *EditWindow) Valid(cmd CommandCode) bool {
	if cmd != CmClose && cmd != CmQuit {
		return true
	}
	if !ew.editor.Modified() {
		return true
	}
	result := MessageBox(ew, "Confirm",
		"Save changes to "+ew.Title()+"?",
		MbYes|MbNo|MbCancel)
	switch result {
	case CmYes:
		if ew.editor.FileName() == "" {
			return false
		}
		err := ew.editor.SaveFile(ew.editor.FileName())
		return err == nil
	case CmNo:
		return true
	default:
		return false
	}
}
