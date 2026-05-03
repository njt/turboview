package tv

import "unicode"

type ValidatorWithAutoFill interface {
	Validator
	IsValidInputAutoFill(s *string, noAutoFill bool) bool
}

type PicResult int

const (
	prComplete   PicResult = iota
	prIncomplete
	prEmpty
	prError
)

type PictureValidator struct {
	pic      string
	autoFill bool
}

var _ ValidatorWithAutoFill = (*PictureValidator)(nil)

func NewPictureValidator(pic string, autoFill bool) ValidatorWithAutoFill {
	return &PictureValidator{pic: pic, autoFill: autoFill}
}

func (pv *PictureValidator) Error() {}

func (pv *PictureValidator) IsValid(s string) bool {
	pic := []rune(pv.pic)
	input := []rune(s)
	result, _, _, _ := pv.scan(pic, 0, input, 0, false)
	return result == prComplete || result == prEmpty
}

func (pv *PictureValidator) IsValidInput(s string, noAutoFill bool) bool {
	pic := []rune(pv.pic)
	input := []rune(s)
	result, _, _, _ := pv.scan(pic, 0, input, 0, false)
	return result != prError
}

func (pv *PictureValidator) IsValidInputAutoFill(s *string, noAutoFill bool) bool {
	pic := []rune(pv.pic)
	input := []rune(*s)
	modify := !noAutoFill
	result, _, _, outInput := pv.scan(pic, 0, input, 0, modify)
	if modify {
		*s = string(outInput)
	}
	return result != prError
}

// scan walks picture and input in parallel.
// Returns (result, final picIdx, final inIdx, possibly-modified input).
func (pv *PictureValidator) scan(pic []rune, picIdx int, input []rune, inIdx int, modify bool) (PicResult, int, int, []rune) {
	for {
		if picIdx >= len(pic) {
			if inIdx >= len(input) {
				if inIdx == 0 && len(input) == 0 {
					return prEmpty, picIdx, inIdx, input
				}
				return prComplete, picIdx, inIdx, input
			}
			return prError, picIdx, inIdx, input
		}

		ch := pic[picIdx]

		switch ch {
		case '#':
			if inIdx >= len(input) {
				return prIncomplete, picIdx, inIdx, input
			}
			if !unicode.IsDigit(input[inIdx]) {
				return prError, picIdx, inIdx, input
			}
			picIdx++
			inIdx++

		case '?':
			if inIdx >= len(input) {
				return prIncomplete, picIdx, inIdx, input
			}
			if !unicode.IsLetter(input[inIdx]) {
				return prError, picIdx, inIdx, input
			}
			picIdx++
			inIdx++

		case '&':
			if inIdx >= len(input) {
				return prIncomplete, picIdx, inIdx, input
			}
			if !unicode.IsLetter(input[inIdx]) {
				return prError, picIdx, inIdx, input
			}
			if modify {
				input[inIdx] = unicode.ToUpper(input[inIdx])
			}
			picIdx++
			inIdx++

		case '!':
			if inIdx >= len(input) {
				return prIncomplete, picIdx, inIdx, input
			}
			if modify {
				input[inIdx] = unicode.ToUpper(input[inIdx])
			}
			picIdx++
			inIdx++

		case '@':
			if inIdx >= len(input) {
				return prIncomplete, picIdx, inIdx, input
			}
			picIdx++
			inIdx++

		case ';':
			picIdx++
			if picIdx >= len(pic) {
				break
			}
			literal := pic[picIdx]
			if inIdx >= len(input) {
				return prIncomplete, picIdx, inIdx, input
			}
			if input[inIdx] != literal {
				return prError, picIdx, inIdx, input
			}
			picIdx++
			inIdx++

		case '[':
			picIdx++
			savedInIdx := inIdx
			savedInput := make([]rune, len(input))
			copy(savedInput, input)

			result, newPicIdx, newInIdx, newInput := pv.scan(pic, picIdx, input, inIdx, modify)

			if result == prError {
				input = savedInput
				inIdx = savedInIdx
				picIdx = pv.findClose(pic, picIdx, ']')
			} else if result == prIncomplete {
				// Optional group partially matched — treat as OK, skip rest of group
				input = newInput
				inIdx = newInIdx
				picIdx = pv.findClose(pic, newPicIdx, ']')
			} else {
				input = newInput
				inIdx = newInIdx
				picIdx = newPicIdx
			}

		case '{':
			picIdx++
			result, newPicIdx, newInIdx, newInput := pv.scan(pic, picIdx, input, inIdx, modify)
			input = newInput
			inIdx = newInIdx
			picIdx = newPicIdx
			if result == prError {
				return prError, picIdx, inIdx, input
			}

		case ']', '}':
			picIdx++
			if inIdx >= len(input) {
				if inIdx == 0 && len(input) == 0 {
					return prEmpty, picIdx, inIdx, input
				}
				return prComplete, picIdx, inIdx, input
			}
			return prComplete, picIdx, inIdx, input

		default:
			if modify && pv.autoFill {
				if inIdx >= len(input) || input[inIdx] != ch {
					newInput := make([]rune, len(input)+1)
					copy(newInput, input[:inIdx])
					newInput[inIdx] = ch
					copy(newInput[inIdx+1:], input[inIdx:])
					input = newInput
				}
				inIdx++
				picIdx++
			} else {
				if inIdx >= len(input) {
					return prIncomplete, picIdx, inIdx, input
				}
				if input[inIdx] != ch {
					return prError, picIdx, inIdx, input
				}
				picIdx++
				inIdx++
			}
		}
	}
}

// findClose finds the matching closing bracket, accounting for nesting.
func (pv *PictureValidator) findClose(pic []rune, start int, close rune) int {
	var open rune
	if close == ']' {
		open = '['
	} else {
		open = '{'
	}
	depth := 1
	i := start
	for i < len(pic) && depth > 0 {
		if pic[i] == ';' {
			i += 2
			continue
		}
		if pic[i] == open {
			depth++
		} else if pic[i] == close {
			depth--
		}
		i++
	}
	return i
}
