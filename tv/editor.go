package tv

type Editor struct {
	*Memo
	modified bool
}

func (e *Editor) Modified() bool {
	return e.modified
}
