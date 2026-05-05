package tv

import (
	"os"
	"path/filepath"
	"strings"
)

// FileDialogFlag controls which action buttons appear in a FileDialog.
type FileDialogFlag int

const (
	FdOpenButton    FileDialogFlag = 1 << iota
	FdOKButton
	FdReplaceButton
	FdClearButton
	FdHelpButton
)

// inputDetection classifies the text in the FileInputLine for Valid().
type inputDetection int

const (
	inputFilename  inputDetection = iota
	inputWildcard
	inputDirectory
)

// FileInputLine wraps an InputLine with wildcard support and CmFileFocused
// auto-fill behaviour.
type FileInputLine struct {
	*InputLine
	wildcard string
}

// HandleEvent intercepts EvBroadcast+CmFileFocused to auto-fill the line
// with the focused filename when the FileInputLine does NOT have focus.
// All other events are delegated to the embedded InputLine.
func (fil *FileInputLine) HandleEvent(event *Event) {
	if event.What == EvBroadcast && event.Command == CmFileFocused && !fil.HasState(SfSelected) {
		entry, ok := event.Info.(*FileEntry)
		if ok && entry != nil {
			if entry.IsDir {
				fil.SetText(entry.Name + "/" + fil.wildcard)
			} else {
				fil.SetText(entry.Name)
			}
			event.Clear()
			return
		}
		// Info was nil or wrong type — don't auto-fill, delegate
	}
	fil.InputLine.HandleEvent(event)
}

// Clear empties the text content.
func (fil *FileInputLine) Clear() {
	fil.SetText("")
}

// FileDialog is a modal dialog for browsing files and directories.
type FileDialog struct {
	*Dialog
	fileList   *FileList
	fileInfo   *FileInfoPane
	fileInput  *FileInputLine
	wildcard   string
	resultPath string
	initialDir string
}

// NewFileDialogInDir creates a FileDialog in the given directory with the
// specified wildcard, title, and action button flags.
func NewFileDialogInDir(dir, wildcard, title string, flags FileDialogFlag) *FileDialog {
	if wildcard == "" {
		wildcard = "*"
	}

	fd := &FileDialog{
		Dialog:     NewDialog(NewRect(0, 0, 52, 20), title),
		wildcard:   wildcard,
		initialDir: dir,
	}
	fd.SetSelf(fd)

	// --- Row 0: Label "File ~N~ame:" at (1, 0, 12, 1)
	nameLabel := NewLabel(NewRect(1, 0, 12, 1), "File ~n~ame", nil)
	fd.Insert(nameLabel)

	// --- Row 1: FileInputLine at (1, 1, 32, 1)
	il := NewInputLine(NewRect(1, 1, 32, 1), 256)
	il.SetText(wildcard)

	fil := &FileInputLine{
		InputLine: il,
		wildcard:  wildcard,
	}
	fd.fileInput = fil

	// --- Row 1: History at (33, 1, 1, 1)
	history := NewHistory(NewRect(33, 1, 1, 1), il, 0)
	fd.Insert(history)

	// --- Row 3: Label "~F~iles:" at (1, 3, 10, 1)
	filesLabel := NewLabel(NewRect(1, 3, 10, 1), "~F~iles", nil)
	fd.Insert(filesLabel)

	// --- Rows 4-13 (y=4, h=10): FileList at (1, 4, 35, 10)
	fl := NewFileList(NewRect(1, 4, 35, 10))
	fl.ReadDirectory(dir, wildcard)
	fd.fileList = fl
	fd.Insert(fl)

	// --- Rows 14-15 (y=14, h=2): FileInfoPane at (1, 14, 35, 2)
	fip := NewFileInfoPane(NewRect(1, 14, 35, 2))
	fd.fileInfo = fip
	fd.Insert(fip)

	// --- Buttons on right (x=38, w=12), starting y=1, space 3 apart
	buttonDefs := buildFileDialogButtons(flags)
	buttonX := 38
	buttonW := 12
	buttonSpacing := 3 // y increment per button

	for i, bd := range buttonDefs {
		y := 1 + i*buttonSpacing
		var opts []ButtonOption
		if i == 0 {
			opts = append(opts, WithDefault())
		}
		btn := NewButton(NewRect(buttonX, y, buttonW, 2), bd.label, bd.cmd, opts...)
		fd.Insert(btn)
	}

	// Insert FileInputLine AFTER buttons. Insert makes it focused (via
	// selectChild). Since the first button has bfDefault=true, losing
	// focus does NOT reset its amDefault — IsDefault() stays true.
	// FileInputLine being focused prevents auto-fill on CmFileFocused
	// broadcasts (the user is typing, auto-fill would be disruptive).
	fd.Insert(fil)

	return fd
}

type fileDialogButtonDef struct {
	label string
	cmd   CommandCode
}

func buildFileDialogButtons(flags FileDialogFlag) []fileDialogButtonDef {
	var defs []fileDialogButtonDef

	if flags&FdOpenButton != 0 {
		defs = append(defs, fileDialogButtonDef{"Open", CmFileOpen})
	}
	if flags&FdOKButton != 0 {
		defs = append(defs, fileDialogButtonDef{"OK", CmOK})
	}
	if flags&FdReplaceButton != 0 {
		defs = append(defs, fileDialogButtonDef{"Replace", CmFileReplace})
	}
	if flags&FdClearButton != 0 {
		defs = append(defs, fileDialogButtonDef{"Clear", CmFileClear})
	}
	if flags&FdHelpButton != 0 {
		defs = append(defs, fileDialogButtonDef{"Help", CmUser + 100})
	}

	// If no action buttons specified, default to OK
	if len(defs) == 0 {
		defs = append(defs, fileDialogButtonDef{"OK", CmOK})
	}

	// Cancel always appended at the end
	defs = append(defs, fileDialogButtonDef{"Cancel", CmCancel})

	return defs
}

// HandleEvent intercepts commands before delegating to Dialog.
func (fd *FileDialog) HandleEvent(event *Event) {
	if event.What == EvCommand && event.Command == CmFileClear {
		fd.fileInput.Clear()
		fd.resultPath = ""
		event.Clear()
		return
	}
	fd.Dialog.HandleEvent(event)
	if event.What == EvCommand && (event.Command == CmFileOpen || event.Command == CmFileReplace) {
		event.Command = CmOK
		if !fd.Valid(CmOK) {
			event.Clear()
		}
	}
}

// FileName returns the resolved file path.
func (fd *FileDialog) FileName() string {
	return fd.resultPath
}

// Valid performs custom validation for the file dialog.
func (fd *FileDialog) Valid(cmd CommandCode) bool {
	if cmd == CmCancel {
		return true
	}
	if cmd != CmOK {
		return true
	}

	text := strings.TrimSpace(fd.fileInput.Text())
	if text == "" {
		return false
	}

	detect := fd.detectInput(text)
	switch detect {
	case inputWildcard:
		fd.wildcard = text
		fd.fileInput.wildcard = text
		fd.fileList.ReadDirectory(fd.fileList.Dir(), text)
		// Broadcast CmFileFilter to owner's children
		if owner := fd.Owner(); owner != nil {
			ev := &Event{What: EvBroadcast, Command: CmFileFilter}
			owner.HandleEvent(ev)
		}
		return false

	case inputDirectory:
		fd.fileList.ReadDirectory(text, fd.wildcard)
		fd.fileInput.wildcard = fd.wildcard
		return false

	case inputFilename:
		// Resolve to absolute path
		if filepath.IsAbs(text) {
			fd.resultPath = text
		} else {
			fd.resultPath = filepath.Join(fd.fileList.Dir(), text)
		}
		return true
	}

	return false
}

// detectInput classifies the text as wildcard, directory, or filename.
func (fd *FileDialog) detectInput(text string) inputDetection {
	// Check for wildcard characters first
	if strings.ContainsAny(text, "*?") {
		return inputWildcard
	}

	// Check if it's an existing directory. Resolve relative paths
	// against the dialog's current directory, not the process CWD.
	resolved := text
	if !filepath.IsAbs(text) {
		resolved = filepath.Join(fd.fileList.Dir(), text)
	}
	if info, err := os.Stat(resolved); err == nil && info.IsDir() {
		return inputDirectory
	}

	return inputFilename
}
