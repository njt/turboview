package tv

import (
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	east "github.com/yuin/goldmark/extension/ast"
	"github.com/yuin/goldmark/text"
)

type mdBlockKind int

const (
	blockParagraph mdBlockKind = iota
	blockHeader
	blockCodeBlock
	blockBulletList
	blockNumberList
	blockBlockquote
	blockTable
	blockHRule
	blockDefList
	blockCheckList
)

type mdRunStyle int

const (
	runNormal mdRunStyle = iota
	runBold
	runItalic
	runBoldItalic
	runCode
	runLink
	runStrikethrough
)

type mdBlock struct {
	kind     mdBlockKind
	level    int
	runs     []mdRun
	language string
	code     []string
	items    []mdItem
	children []mdBlock
	headers  [][]mdRun
	rows     [][][]mdRun
}

type mdRun struct {
	text  string
	style mdRunStyle
	url   string
}

type mdItem struct {
	runs     []mdRun
	children []mdBlock
	checked  *bool
	term     []mdRun
}

func parseMarkdown(src string) []mdBlock {
	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			extension.DefinitionList,
		),
	)
	source := []byte(src)
	reader := text.NewReader(source)
	doc := md.Parser().Parse(reader)
	return walkBlocks(doc, source)
}

func walkBlocks(parent ast.Node, source []byte) []mdBlock {
	var blocks []mdBlock
	for child := parent.FirstChild(); child != nil; child = child.NextSibling() {
		blocks = append(blocks, convertNode(child, source)...)
	}
	return blocks
}

func convertNode(node ast.Node, source []byte) []mdBlock {
	switch n := node.(type) {
	case *ast.Paragraph:
		return []mdBlock{{kind: blockParagraph, runs: collectInlineRuns(n, source, runNormal)}}

	case *ast.Heading:
		return []mdBlock{{kind: blockHeader, level: n.Level, runs: collectInlineRuns(n, source, runNormal)}}

	case *ast.FencedCodeBlock:
		lang := ""
		if n.Info != nil {
			lang = strings.TrimSpace(string(n.Info.Text(source)))
			if idx := strings.IndexByte(lang, ' '); idx >= 0 {
				lang = lang[:idx]
			}
		}
		var lines []string
		for i := 0; i < n.Lines().Len(); i++ {
			seg := n.Lines().At(i)
			line := string(seg.Value(source))
			line = strings.TrimRight(line, "\n")
			lines = append(lines, line)
		}
		return []mdBlock{{kind: blockCodeBlock, language: lang, code: lines}}

	case *ast.CodeBlock:
		var lines []string
		for i := 0; i < n.Lines().Len(); i++ {
			seg := n.Lines().At(i)
			line := strings.TrimRight(string(seg.Value(source)), "\n")
			lines = append(lines, line)
		}
		return []mdBlock{{kind: blockCodeBlock, code: lines}}

	case *ast.List:
		return convertList(n, source)

	case *ast.Blockquote:
		return []mdBlock{{kind: blockBlockquote, children: walkBlocks(n, source)}}

	case *east.Table:
		return []mdBlock{convertTable(n, source)}

	case *ast.ThematicBreak:
		return []mdBlock{{kind: blockHRule}}

	case *east.DefinitionList:
		return []mdBlock{convertDefList(n, source)}
	}
	return nil
}

func convertList(list *ast.List, source []byte) []mdBlock {
	// Determine if this is an ordered list
	baseKind := blockBulletList
	if list.IsOrdered() {
		baseKind = blockNumberList
	}

	// Check if ANY list item has a TaskCheckBox (to detect mixed lists)
	hasAnyCheck := false
	hasAnyNonCheck := false
	for itemNode := list.FirstChild(); itemNode != nil; itemNode = itemNode.NextSibling() {
		if li, ok := itemNode.(*ast.ListItem); ok {
			if itemHasCheckbox(li) {
				hasAnyCheck = true
			} else {
				hasAnyNonCheck = true
			}
		}
	}

	// If mixed: split into separate lists
	if hasAnyCheck && hasAnyNonCheck {
		var blocks []mdBlock
		var curItems []mdItem
		curIsCheck := false
		first := true

		for itemNode := list.FirstChild(); itemNode != nil; itemNode = itemNode.NextSibling() {
			if li, ok := itemNode.(*ast.ListItem); ok {
				isCheck := itemHasCheckbox(li)
				if first {
					curIsCheck = isCheck
				}
				if isCheck != curIsCheck {
					// Emit current group
					kind := baseKind
					if curIsCheck {
						kind = blockCheckList
					}
					b := mdBlock{kind: kind, items: curItems}
					if list.IsOrdered() {
						b.level = list.Start
					}
					blocks = append(blocks, b)
					curItems = nil
					curIsCheck = isCheck
				}
				curItems = append(curItems, convertListItem(li, source, curIsCheck))
				first = false
			}
		}
		// Emit final group
		if len(curItems) > 0 {
			kind := baseKind
			if curIsCheck {
				kind = blockCheckList
			}
			b := mdBlock{kind: kind, items: curItems}
			if list.IsOrdered() {
				b.level = list.Start
			}
			blocks = append(blocks, b)
		}
		return blocks
	}

	// Uniform list (all check or all non-check)
	kind := baseKind
	if hasAnyCheck {
		kind = blockCheckList
	}

	var items []mdItem
	for child := list.FirstChild(); child != nil; child = child.NextSibling() {
		if li, ok := child.(*ast.ListItem); ok {
			items = append(items, convertListItem(li, source, hasAnyCheck))
		}
	}

	b := mdBlock{kind: kind, items: items}
	if list.IsOrdered() {
		b.level = list.Start
	}
	return []mdBlock{b}
}

// itemHasCheckbox checks if a ListItem contains a TaskCheckBox.
func itemHasCheckbox(li *ast.ListItem) bool {
	for c := li.FirstChild(); c != nil; c = c.NextSibling() {
		if tb, ok := c.(*ast.TextBlock); ok {
			for gc := tb.FirstChild(); gc != nil; gc = gc.NextSibling() {
				if _, ok := gc.(*east.TaskCheckBox); ok {
					return true
				}
			}
		}
	}
	return false
}

func convertListItem(li *ast.ListItem, source []byte, isCheckList bool) mdItem {
	item := mdItem{}

	for child := li.FirstChild(); child != nil; child = child.NextSibling() {
		switch c := child.(type) {
		case *ast.TextBlock:
			runs := collectInlineRuns(c, source, runNormal)
			if isCheckList {
				for gc := c.FirstChild(); gc != nil; gc = gc.NextSibling() {
					if cb, ok := gc.(*east.TaskCheckBox); ok {
						checked := cb.IsChecked
						item.checked = &checked
						break
					}
				}
			}
			item.runs = append(item.runs, runs...)
		case *ast.Paragraph:
			item.runs = append(item.runs, collectInlineRuns(c, source, runNormal)...)
		default:
			item.children = append(item.children, convertNode(child, source)...)
		}
	}
	return item
}

func convertTable(table *east.Table, source []byte) mdBlock {
	b := mdBlock{kind: blockTable}
	for child := table.FirstChild(); child != nil; child = child.NextSibling() {
		switch c := child.(type) {
		case *east.TableHeader:
			// TableHeader has TableCell nodes as direct children (no TableRow wrapper)
			b.headers = convertTableCells(c, source)
		case *east.TableRow:
			// Body rows are direct children of Table
			b.rows = append(b.rows, convertTableCells(c, source))
		}
	}
	return b
}

func convertTableCells(row ast.Node, source []byte) [][]mdRun {
	var cells [][]mdRun
	for cell := row.FirstChild(); cell != nil; cell = cell.NextSibling() {
		if tc, ok := cell.(*east.TableCell); ok {
			cells = append(cells, collectInlineRuns(tc, source, runNormal))
		}
	}
	return cells
}

func convertDefList(dl *east.DefinitionList, source []byte) mdBlock {
	b := mdBlock{kind: blockDefList}
	var currentItem *mdItem
	for child := dl.FirstChild(); child != nil; child = child.NextSibling() {
		switch c := child.(type) {
		case *east.DefinitionTerm:
			if currentItem != nil {
				b.items = append(b.items, *currentItem)
			}
			currentItem = &mdItem{
				term: collectInlineRuns(c, source, runNormal),
			}
		case *east.DefinitionDescription:
			if currentItem == nil {
				currentItem = &mdItem{}
			}
			for gc := c.FirstChild(); gc != nil; gc = gc.NextSibling() {
				switch gcc := gc.(type) {
				case *ast.TextBlock:
					currentItem.runs = append(currentItem.runs, collectInlineRuns(gcc, source, runNormal)...)
				case *ast.Paragraph:
					currentItem.runs = append(currentItem.runs, collectInlineRuns(gcc, source, runNormal)...)
				default:
					currentItem.children = append(currentItem.children, convertNode(gc, source)...)
				}
			}
		}
	}
	if currentItem != nil {
		b.items = append(b.items, *currentItem)
	}
	return b
}

func collectInlineRuns(node ast.Node, source []byte, parentStyle mdRunStyle) []mdRun {
	var runs []mdRun
	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		switch c := child.(type) {
		case *ast.Text:
			t := string(c.Text(source))
			if c.SoftLineBreak() {
				t += " "
			}
			if t != "" {
				runs = append(runs, mdRun{text: t, style: parentStyle})
			}
		case *ast.String:
			t := string(c.Value)
			if t != "" {
				runs = append(runs, mdRun{text: t, style: parentStyle})
			}
		case *ast.CodeSpan:
			var buf strings.Builder
			for gc := c.FirstChild(); gc != nil; gc = gc.NextSibling() {
				if t, ok := gc.(*ast.Text); ok {
					buf.Write(t.Text(source))
				}
			}
			if buf.Len() > 0 {
				runs = append(runs, mdRun{text: buf.String(), style: runCode})
			}
		case *ast.Emphasis:
			innerStyle := runItalic
			if c.Level == 2 {
				innerStyle = runBold
			}
			if parentStyle == runBold && innerStyle == runItalic {
				innerStyle = runBoldItalic
			} else if parentStyle == runItalic && innerStyle == runBold {
				innerStyle = runBoldItalic
			}
			runs = append(runs, collectInlineRuns(c, source, innerStyle)...)
		case *ast.Link:
			var buf strings.Builder
			for gc := c.FirstChild(); gc != nil; gc = gc.NextSibling() {
				if t, ok := gc.(*ast.Text); ok {
					buf.Write(t.Text(source))
				}
			}
			runs = append(runs, mdRun{
				text:  buf.String(),
				style: runLink,
				url:   string(c.Destination),
			})
		case *ast.Image:
			alt := string(c.Text(source))
			if alt == "" {
				alt = "image"
			}
			runs = append(runs, mdRun{text: "[IMG: " + alt + "]", style: runCode})
		case *ast.AutoLink:
			url := string(c.URL(source))
			runs = append(runs, mdRun{text: url, style: runLink, url: url})
		case *east.Strikethrough:
			runs = append(runs, collectInlineRuns(c, source, runStrikethrough)...)
		case *east.TaskCheckBox:
			// handled at list item level
		default:
			runs = append(runs, collectInlineRuns(child, source, parentStyle)...)
		}
	}
	return mergeRuns(runs)
}

func mergeRuns(runs []mdRun) []mdRun {
	if len(runs) <= 1 {
		return runs
	}
	merged := []mdRun{runs[0]}
	for _, r := range runs[1:] {
		last := &merged[len(merged)-1]
		if last.style == r.style && last.url == r.url {
			last.text += r.text
		} else {
			merged = append(merged, r)
		}
	}
	return merged
}
