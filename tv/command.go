package tv

type CommandCode int

const (
	CmQuit    CommandCode = iota + 1
	CmClose
	CmOK
	CmCancel
	CmYes
	CmNo
	CmMenu
	CmResize
	CmZoom
	CmTile
	CmCascade
	CmNext
	CmPrev
	CmUser CommandCode = 1000
)
