package tv

// input_box.go — InputBox: a convenience dialog combining a prompt Label,
// an InputLine, and OK/Cancel buttons.
//
// Spec:
//   - InputBox(owner Container, title, prompt, defaultValue string) (string, CommandCode)
//   - Dialog width: max(len(prompt)+6, len(title)+6, 30), capped at 60
//   - Dialog height: 7 (frame + prompt row + gap + input row + gap + button row + frame)
//   - Centered in owner's bounds
//   - The InputLine is pre-filled with defaultValue and has the cursor at the end
//   - The user can Tab between InputLine, OK, and Cancel
//   - On CmOK, returns the InputLine's current text and CmOK
//   - On CmCancel, returns empty string and CmCancel

func InputBox(owner Container, title, prompt, defaultValue string) (string, CommandCode) {
	// ---------------------------------------------------------------------------
	// Compute dialog dimensions.
	// ---------------------------------------------------------------------------
	const minW = 30
	const maxW = 60
	const dlgH = 7

	promptLen := len([]rune(prompt))
	titleLen := len([]rune(title))

	dlgW := minW
	if promptLen+6 > dlgW {
		dlgW = promptLen + 6
	}
	if titleLen+6 > dlgW {
		dlgW = titleLen + 6
	}
	if dlgW > maxW {
		dlgW = maxW
	}

	// ---------------------------------------------------------------------------
	// Center in owner bounds.
	// ---------------------------------------------------------------------------
	ob := owner.Bounds()
	dx := (ob.Width() - dlgW) / 2
	dy := (ob.Height() - dlgH) / 2
	// NewRect(x, y, width, height) — third arg is width, not x2.
	dlgBounds := NewRect(dx, dy, dlgW, dlgH)

	// ---------------------------------------------------------------------------
	// Build the dialog and its children.
	// ---------------------------------------------------------------------------
	//
	// Client area is (dlgW-2) × (dlgH-2) = (dlgW-2) × 5.
	// Row 0: prompt label
	// Row 1: (blank)
	// Row 2: input line
	// Row 3: (blank)
	// Row 4: OK + Cancel buttons
	clientW := dlgW - 2

	d := NewDialog(dlgBounds, title)

	// Prompt label (row 0 in client area, full width, no link).
	promptLabel := NewLabel(NewRect(0, 0, clientW, 1), prompt, nil)
	d.Insert(promptLabel)

	// Input line (row 2 in client area, full width).
	il := NewInputLine(NewRect(0, 2, clientW, 1), 0)
	il.SetText(defaultValue)
	d.Insert(il)

	// OK and Cancel buttons (row 4).
	btnW := 10
	gap := 2
	totalBtnW := btnW + gap + btnW
	btnX := (clientW - totalBtnW) / 2
	if btnX < 0 {
		btnX = 0
	}
	okBtn := NewButton(NewRect(btnX, 4, btnX+btnW, 5), "OK", CmOK)
	cancelBtn := NewButton(NewRect(btnX+btnW+gap, 4, btnX+btnW+gap+btnW, 5), "Cancel", CmCancel)
	d.Insert(okBtn)
	d.Insert(cancelBtn)

	// Start with focus on the InputLine.
	d.SetFocusedChild(il)

	// ---------------------------------------------------------------------------
	// Run the modal loop.
	// ---------------------------------------------------------------------------
	code := owner.ExecView(d)

	if code == CmOK {
		return il.Text(), CmOK
	}
	return "", CmCancel
}
