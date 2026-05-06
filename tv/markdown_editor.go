package tv

// MarkdownEditor is a markdown-aware text editor that combines the editing
// capabilities of Editor with live markdown parsing via parseMarkdown.
type MarkdownEditor struct {
	*Editor
	blocks      []mdBlock
	sourceCache string
	showSource  bool
}

// NewMarkdownEditor creates a new MarkdownEditor with the given bounds.
func NewMarkdownEditor(bounds Rect) *MarkdownEditor {
	me := &MarkdownEditor{}
	me.Editor = NewEditor(bounds)
	me.Editor.SetOptions(OfSelectable|OfFirstClick, true)
	me.Editor.SetGrowMode(GfGrowHiX | GfGrowHiY)
	me.Editor.SetSelf(me)
	me.blocks = []mdBlock{}
	me.reparse()
	return me
}

// SetText overrides Editor.SetText: calls the embedded Editor.SetText,
// then calls reparse() to populate blocks.
func (me *MarkdownEditor) SetText(s string) {
	me.Editor.SetText(s)
	me.reparse()
}

// Text delegates to Editor.Text() (inherited from Memo).
func (me *MarkdownEditor) Text() string {
	return me.Editor.Text()
}

// ShowSource returns the current source toggle state.
func (me *MarkdownEditor) ShowSource() bool {
	return me.showSource
}

// SetShowSource sets the source toggle state.
func (me *MarkdownEditor) SetShowSource(on bool) {
	me.showSource = on
}

// reparse joins Memo.lines into a string, runs parseMarkdown, and stores the
// result in blocks. It is a no-op if the source text has not changed since
// the last parse.
func (me *MarkdownEditor) reparse() {
	src := me.Editor.Text()
	if src == me.sourceCache {
		return
	}
	me.sourceCache = src
	if src == "" {
		me.blocks = []mdBlock{}
		return
	}
	me.blocks = parseMarkdown(src)
	if me.blocks == nil {
		me.blocks = []mdBlock{}
	}
}
