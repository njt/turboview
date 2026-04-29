package tv

func InputBox(owner Container, title, prompt, defaultValue string) (string, CommandCode) {
	promptRunes := []rune(prompt)
	titleRunes := []rune(title)

	contentW := len(promptRunes) + 6
	if tw := len(titleRunes) + 6; tw > contentW {
		contentW = tw
	}
	dialogW := contentW
	if dialogW < 30 {
		dialogW = 30
	}
	if dialogW > 60 {
		dialogW = 60
	}
	dialogH := 7
	innerW := dialogW - 2

	ob := owner.Bounds()
	dx := (ob.Width() - dialogW) / 2
	dy := (ob.Height() - dialogH) / 2
	if dx < 0 {
		dx = 0
	}
	if dy < 0 {
		dy = 0
	}

	dlg := NewDialog(NewRect(dx, dy, dialogW, dialogH), title)

	input := NewInputLine(NewRect(1, 1, innerW, 1), 255)
	input.SetText(defaultValue)

	lbl := NewLabel(NewRect(1, 0, innerW, 1), prompt, input)
	dlg.Insert(lbl)
	dlg.Insert(input)

	btnW := 12
	btnGap := 2
	totalBtnW := 2*btnW + btnGap
	startX := (innerW - totalBtnW) / 2
	if startX < 0 {
		startX = 0
	}

	okBtn := NewButton(NewRect(startX, 3, btnW, 2), "OK", CmOK, WithDefault())
	cancelBtn := NewButton(NewRect(startX+btnW+btnGap, 3, btnW, 2), "Cancel", CmCancel)
	dlg.Insert(okBtn)
	dlg.Insert(cancelBtn)
	dlg.SetFocusedChild(input)

	result := owner.ExecView(dlg)
	if result == CmCancel {
		return "", CmCancel
	}
	return input.Text(), result
}
