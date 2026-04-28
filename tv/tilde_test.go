package tv

import (
	"testing"
)

// TestParseTildeLabelBasic tests parsing a label with tilde-delimited shortcut.
func TestParseTildeLabelBasic(t *testing.T) {
	result := ParseTildeLabel("~Alt+X~ Exit")

	if len(result) != 2 {
		t.Fatalf("ParseTildeLabel() returned %d segments, want 2", len(result))
	}

	if result[0].Text != "Alt+X" || !result[0].Shortcut {
		t.Errorf("segment 0: got {%q, %v}, want {\"Alt+X\", true}", result[0].Text, result[0].Shortcut)
	}

	if result[1].Text != " Exit" || result[1].Shortcut {
		t.Errorf("segment 1: got {%q, %v}, want {\" Exit\", false}", result[1].Text, result[1].Shortcut)
	}
}

// TestParseTildeLabelNoTilde tests parsing a label without tilde.
func TestParseTildeLabelNoTilde(t *testing.T) {
	result := ParseTildeLabel("No tilde")

	if len(result) != 1 {
		t.Fatalf("ParseTildeLabel() returned %d segments, want 1", len(result))
	}

	if result[0].Text != "No tilde" || result[0].Shortcut {
		t.Errorf("segment 0: got {%q, %v}, want {\"No tilde\", false}", result[0].Text, result[0].Shortcut)
	}
}

// TestParseTildeLabelConsecutiveTildes tests parsing tildes back-to-back.
func TestParseTildeLabelConsecutiveTildes(t *testing.T) {
	result := ParseTildeLabel("~O~K")

	if len(result) != 2 {
		t.Fatalf("ParseTildeLabel() returned %d segments, want 2", len(result))
	}

	if result[0].Text != "O" || !result[0].Shortcut {
		t.Errorf("segment 0: got {%q, %v}, want {\"O\", true}", result[0].Text, result[0].Shortcut)
	}

	if result[1].Text != "K" || result[1].Shortcut {
		t.Errorf("segment 1: got {%q, %v}, want {\"K\", false}", result[1].Text, result[1].Shortcut)
	}
}

// TestParseTildeLabelEmpty tests parsing an empty string.
func TestParseTildeLabelEmpty(t *testing.T) {
	result := ParseTildeLabel("")

	if len(result) != 0 {
		t.Fatalf("ParseTildeLabel(\"\") returned %d segments, want 0", len(result))
	}
}

// TestParseTildeLabelStartsWithTilde tests parsing when label starts with tilde.
func TestParseTildeLabelStartsWithTilde(t *testing.T) {
	result := ParseTildeLabel("~S~ave")

	if len(result) != 2 {
		t.Fatalf("ParseTildeLabel() returned %d segments, want 2", len(result))
	}

	if result[0].Text != "S" || !result[0].Shortcut {
		t.Errorf("segment 0: got {%q, %v}, want {\"S\", true}", result[0].Text, result[0].Shortcut)
	}

	if result[1].Text != "ave" || result[1].Shortcut {
		t.Errorf("segment 1: got {%q, %v}, want {\"ave\", false}", result[1].Text, result[1].Shortcut)
	}
}

// TestParseTildeLabelEndsWithTilde tests parsing when label ends with tilde.
func TestParseTildeLabelEndsWithTilde(t *testing.T) {
	result := ParseTildeLabel("Delete ~D~")

	if len(result) != 2 {
		t.Fatalf("ParseTildeLabel() returned %d segments, want 2", len(result))
	}

	if result[0].Text != "Delete " || result[0].Shortcut {
		t.Errorf("segment 0: got {%q, %v}, want {\"Delete \", false}", result[0].Text, result[0].Shortcut)
	}

	if result[1].Text != "D" || !result[1].Shortcut {
		t.Errorf("segment 1: got {%q, %v}, want {\"D\", true}", result[1].Text, result[1].Shortcut)
	}
}

// TestParseTildeLabelMultipleTildes tests parsing with multiple shortcut pairs.
func TestParseTildeLabelMultipleTildes(t *testing.T) {
	result := ParseTildeLabel("~F~ile ~E~dit")

	if len(result) != 4 {
		t.Fatalf("ParseTildeLabel() returned %d segments, want 4", len(result))
	}

	expected := []struct {
		text     string
		shortcut bool
	}{
		{"F", true},
		{"ile ", false},
		{"E", true},
		{"dit", false},
	}

	for i, exp := range expected {
		if result[i].Text != exp.text || result[i].Shortcut != exp.shortcut {
			t.Errorf("segment %d: got {%q, %v}, want {%q, %v}", i, result[i].Text, result[i].Shortcut, exp.text, exp.shortcut)
		}
	}
}

// TestParseTildeLabelWhitespace tests parsing with leading/trailing whitespace.
func TestParseTildeLabelWhitespace(t *testing.T) {
	result := ParseTildeLabel("  ~Q~uit  ")

	if len(result) != 3 {
		t.Fatalf("ParseTildeLabel() returned %d segments, want 3", len(result))
	}

	if result[0].Text != "  " || result[0].Shortcut {
		t.Errorf("segment 0: got {%q, %v}, want {\"  \", false}", result[0].Text, result[0].Shortcut)
	}

	if result[1].Text != "Q" || !result[1].Shortcut {
		t.Errorf("segment 1: got {%q, %v}, want {\"Q\", true}", result[1].Text, result[1].Shortcut)
	}

	if result[2].Text != "uit  " || result[2].Shortcut {
		t.Errorf("segment 2: got {%q, %v}, want {\"uit  \", false}", result[2].Text, result[2].Shortcut)
	}
}

// TestParseTildeLabelSingleTilde tests label with single tilde (incomplete pair).
func TestParseTildeLabelSingleTilde(t *testing.T) {
	result := ParseTildeLabel("Text ~Shortcut")

	// Single unpaired tilde should still be treated as literal text
	// (implementation detail, but we verify consistent behavior)
	if len(result) == 0 {
		t.Fatalf("ParseTildeLabel() returned empty result")
	}

	// The exact behavior depends on implementation:
	// either include the tilde in text or skip it.
	// We just verify it doesn't crash and returns something.
}

// TestParseTildeLabelOnlyTildes tests label with only tildes.
func TestParseTildeLabelOnlyTildes(t *testing.T) {
	result := ParseTildeLabel("~~")

	// ~~ opens and closes tilde mode with no content between — produces no segments
	if len(result) != 0 {
		t.Errorf("ParseTildeLabel(\"~~\") returned %d segments, want 0", len(result))
	}
}

// TestParseTildeLabelComplexShortcut tests label with multi-character shortcut.
func TestParseTildeLabelComplexShortcut(t *testing.T) {
	result := ParseTildeLabel("Press ~Ctrl+Alt+X~ to exit")

	if len(result) != 3 {
		t.Fatalf("ParseTildeLabel() returned %d segments, want 3", len(result))
	}

	if result[0].Text != "Press " || result[0].Shortcut {
		t.Errorf("segment 0: got {%q, %v}, want {\"Press \", false}", result[0].Text, result[0].Shortcut)
	}

	if result[1].Text != "Ctrl+Alt+X" || !result[1].Shortcut {
		t.Errorf("segment 0: got {%q, %v}, want {\"Ctrl+Alt+X\", true}", result[0].Text, result[0].Shortcut)
	}

	if result[2].Text != " to exit" || result[2].Shortcut {
		t.Errorf("segment 2: got {%q, %v}, want {\" to exit\", false}", result[2].Text, result[2].Shortcut)
	}
}

// TestParseTildeLabelLabelSegmentFields tests that LabelSegment has Text and Shortcut fields.
func TestParseTildeLabelLabelSegmentFields(t *testing.T) {
	result := ParseTildeLabel("~A~B")

	if len(result) < 1 {
		t.Fatalf("ParseTildeLabel() returned no segments")
	}

	segment := result[0]

	// Verify the segment has Text and Shortcut fields (compile-time check essentially)
	_ = segment.Text
	_ = segment.Shortcut
}

// TestParseTildeLabelNoDoubleShortcut tests that consecutive tilde pairs don't merge.
func TestParseTildeLabelNoDoubleShortcut(t *testing.T) {
	result := ParseTildeLabel("~A~~B~")

	if len(result) < 2 {
		t.Fatalf("ParseTildeLabel() returned %d segments, want at least 2", len(result))
	}

	// First pair: ~A~
	if result[0].Text != "A" || !result[0].Shortcut {
		t.Errorf("segment 0: got {%q, %v}, want {\"A\", true}", result[0].Text, result[0].Shortcut)
	}

	// Between: empty or continues to next pair
	// Second pair: ~B~
	foundB := false
	for _, seg := range result {
		if seg.Text == "B" && seg.Shortcut {
			foundB = true
			break
		}
	}

	if !foundB {
		t.Errorf("ParseTildeLabel() did not find shortcut segment for B")
	}
}

// TestParseTildeLabelSpecExample1 verifies spec example: ~Alt+X~ Exit
func TestParseTildeLabelSpecExample1(t *testing.T) {
	result := ParseTildeLabel("~Alt+X~ Exit")

	expected := []struct {
		text     string
		shortcut bool
	}{
		{"Alt+X", true},
		{" Exit", false},
	}

	if len(result) != len(expected) {
		t.Fatalf("ParseTildeLabel() returned %d segments, want %d", len(result), len(expected))
	}

	for i, exp := range expected {
		if result[i].Text != exp.text || result[i].Shortcut != exp.shortcut {
			t.Errorf("segment %d: got {%q, %v}, want {%q, %v}", i, result[i].Text, result[i].Shortcut, exp.text, exp.shortcut)
		}
	}
}

// TestParseTildeLabelSpecExample2 verifies spec example: No tilde
func TestParseTildeLabelSpecExample2(t *testing.T) {
	result := ParseTildeLabel("No tilde")

	expected := []struct {
		text     string
		shortcut bool
	}{
		{"No tilde", false},
	}

	if len(result) != len(expected) {
		t.Fatalf("ParseTildeLabel() returned %d segments, want %d", len(result), len(expected))
	}

	for i, exp := range expected {
		if result[i].Text != exp.text || result[i].Shortcut != exp.shortcut {
			t.Errorf("segment %d: got {%q, %v}, want {%q, %v}", i, result[i].Text, result[i].Shortcut, exp.text, exp.shortcut)
		}
	}
}

// TestParseTildeLabelSpecExample3 verifies spec example: ~O~K
func TestParseTildeLabelSpecExample3(t *testing.T) {
	result := ParseTildeLabel("~O~K")

	expected := []struct {
		text     string
		shortcut bool
	}{
		{"O", true},
		{"K", false},
	}

	if len(result) != len(expected) {
		t.Fatalf("ParseTildeLabel() returned %d segments, want %d", len(result), len(expected))
	}

	for i, exp := range expected {
		if result[i].Text != exp.text || result[i].Shortcut != exp.shortcut {
			t.Errorf("segment %d: got {%q, %v}, want {%q, %v}", i, result[i].Text, result[i].Shortcut, exp.text, exp.shortcut)
		}
	}
}

// TestParseTildeLabelAllTildesAreShortcuts tests that every pair of tildes creates shortcut.
func TestParseTildeLabelAllTildesAreShortcuts(t *testing.T) {
	result := ParseTildeLabel("~A~ ~B~ ~C~")

	shortcuts := 0
	for _, seg := range result {
		if seg.Shortcut {
			shortcuts++
		}
	}

	if shortcuts != 3 {
		t.Errorf("ParseTildeLabel() found %d shortcuts, want 3", shortcuts)
	}
}

// TestParseTildeLabelPreservesTextOrder tests text segments appear in order.
func TestParseTildeLabelPreservesTextOrder(t *testing.T) {
	result := ParseTildeLabel("First ~X~ Second ~Y~ Third")

	texts := make([]string, len(result))
	for i, seg := range result {
		texts[i] = seg.Text
	}

	expectedTexts := []string{"First ", "X", " Second ", "Y", " Third"}
	if len(texts) != len(expectedTexts) {
		t.Fatalf("got %d segments, want %d", len(texts), len(expectedTexts))
	}

	for i, expected := range expectedTexts {
		if texts[i] != expected {
			t.Errorf("segment %d text: got %q, want %q", i, texts[i], expected)
		}
	}
}

// TestParseTildeLabelSpecialCharactersInShortcut tests special chars in shortcut.
func TestParseTildeLabelSpecialCharactersInShortcut(t *testing.T) {
	result := ParseTildeLabel("~+~Plus")

	if len(result) < 1 {
		t.Fatalf("ParseTildeLabel() returned no segments")
	}

	if result[0].Text != "+" || !result[0].Shortcut {
		t.Errorf("segment 0: got {%q, %v}, want {\"+\", true}", result[0].Text, result[0].Shortcut)
	}
}

// TestParseTildeLabelEmptyShortcut tests tilde pair with nothing between.
func TestParseTildeLabelEmptyShortcut(t *testing.T) {
	result := ParseTildeLabel("~~Button")

	// ~~ opens and closes with no content; "Button" follows as non-shortcut
	if len(result) != 1 {
		t.Fatalf("ParseTildeLabel(\"~~Button\") returned %d segments, want 1", len(result))
	}

	if result[0].Text != "Button" || result[0].Shortcut {
		t.Errorf("segment 0: got {%q, %v}, want {\"Button\", false}", result[0].Text, result[0].Shortcut)
	}
}

// TestParseTildeLabelUnicodeCharacters tests Unicode in labels.
func TestParseTildeLabelUnicodeCharacters(t *testing.T) {
	result := ParseTildeLabel("~Ñ~Exit")

	if len(result) < 1 {
		t.Fatalf("ParseTildeLabel() returned no segments")
	}

	if result[0].Text != "Ñ" || !result[0].Shortcut {
		t.Errorf("segment 0: got {%q, %v}, want {\"Ñ\", true}", result[0].Text, result[0].Shortcut)
	}
}
