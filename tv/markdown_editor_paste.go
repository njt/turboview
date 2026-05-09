package tv

import (
	"regexp"
	"strings"
)

// handlePaste inserts the current package-level clipboard content into the
// editor. When forcePlain is true, the text is always inserted verbatim.
// Otherwise, HTML content is detected and converted to markdown.
//
// Selection is replaced if active; otherwise the text is inserted at the
// cursor position.
func (me *MarkdownEditor) handlePaste(forcePlain bool) {
	if clipboard == "" {
		return
	}

	text := clipboard

	// HTML detection and conversion (unless forced plain)
	if !forcePlain && looksLikeHTML(text) {
		text = htmlToMarkdown(text)
	}

	// Replace selection if active, otherwise insert at cursor
	if me.Memo.HasSelection() {
		me.Memo.deleteSelection()
	}
	me.Memo.insertText(text)
}

// looksLikeHTML returns true if the string appears to be HTML content.
func looksLikeHTML(s string) bool {
	s = strings.TrimSpace(s)
	return strings.HasPrefix(s, "<") && (strings.Contains(s, "</") || strings.Contains(s, "/>"))
}

var htmlLinkRE = regexp.MustCompile(`<a[^>]*href="([^"]*)"[^>]*>([^<]*)</a>`)

// htmlToMarkdown converts common HTML tags to their markdown equivalents.
func htmlToMarkdown(s string) string {
	// Convert <a> tags before the replacer handles other tags
	s = htmlLinkRE.ReplaceAllString(s, "[$2]($1)")

	r := strings.NewReplacer(
		"<h1>", "# ", "</h1>", "",
		"<h2>", "## ", "</h2>", "",
		"<h3>", "### ", "</h3>", "",
		"<h4>", "#### ", "</h4>", "",
		"<h5>", "##### ", "</h5>", "",
		"<h6>", "###### ", "</h6>", "",
		"<strong>", "**", "</strong>", "**",
		"<b>", "**", "</b>", "**",
		"<em>", "*", "</em>", "*",
		"<i>", "*", "</i>", "*",
		"<code>", "`", "</code>", "`",
		"<pre>", "```\n", "</pre>", "\n```",
		"<li>", "- ", "</li>", "",
		"<p>", "", "</p>", "\n\n",
		"<br>", "\n", "<br/>", "\n", "<br />", "\n",
		"&lt;", "<", "&gt;", ">", "&amp;", "&",
		"&quot;", "\"", "&#34;", "\"",
		"&apos;", "'", "&#39;", "'",
	)
	return r.Replace(s)
}
