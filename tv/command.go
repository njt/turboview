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

	CmDefault
	CmGrabDefault
	CmReleaseDefault
	CmReceivedFocus
	CmReleasedFocus

	CmScrollBarClicked
	CmScrollBarChanged

	CmSelectWindowNum
	CmRecordHistory

	CmFind
	CmReplace
	CmSearchAgain
	CmIndicatorUpdate

	CmFileOpen
	CmFileReplace
	CmFileClear
	CmFileFocused
	CmFileFilter

	CmColorForegroundChanged
	CmColorBackgroundChanged
	CmNewColorGroup
	CmNewColorIndex

	CmUser CommandCode = 1000
)
