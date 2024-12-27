package chrisdown

import (
	"bytes"
	"html"
	"regexp"
	"strconv"
	"strings"
)

// Helper function to handle inline formatting
func handleInlineFormatting(text string) string {
	// Images (must come before links)
	text = regexp.MustCompile(`!\[([^]]+)]\(([^)]+)\)`).ReplaceAllString(text, `<img src="$2" alt="$1">`)

	// Strong emphasis (bold)
	text = regexp.MustCompile(`\*\*(.+?)\*\*`).ReplaceAllString(text, "<strong>$1</strong>")
	text = regexp.MustCompile(`__(.+?)__`).ReplaceAllString(text, "<strong>$1</strong>")

	// Emphasis (italic)
	text = regexp.MustCompile(`\*(.+?)\*`).ReplaceAllString(text, "<em>$1</em>")
	text = regexp.MustCompile(`_(.+?)_`).ReplaceAllString(text, "<em>$1</em>")

	// Code
	text = regexp.MustCompile("`([^`]+)`").ReplaceAllString(text, "<code>$1</code>")

	// Links (must come after images)
	text = regexp.MustCompile(`\[([^]]+)]\(([^)]+)\)`).ReplaceAllString(text, `<a href="$2">$1</a>`)

	// Strikethrough
	text = regexp.MustCompile(`~~(.+?)~~`).ReplaceAllString(text, "<del>$1</del>")

	return text
}

// Config holds configuration options for the Markdown renderer
type Config struct {
	ImageBaseURL string // Base URL to prepend to relative image paths
}

// RenderMarkdown converts Markdown to HTML with the given configuration
func RenderMarkdown(input string, cfg Config) string {
	var buffer bytes.Buffer
	lines := strings.Split(input, "\n")

	state := struct {
		inUnorderedList bool
		inOrderedList   bool
		inCodeBlock     bool
		codeLanguage    string
		inParagraph     bool
		listDepth       int
	}{}

	// Pre-process images if base URL is provided
	if cfg.ImageBaseURL != "" {
		input = regexp.MustCompile(`!\[([^]]*)]\(([^)]+)\)`).ReplaceAllStringFunc(input, func(match string) string {
			parts := regexp.MustCompile(`!\[([^]]*)]\(([^)]+)\)`).FindStringSubmatch(match)
			if len(parts) > 2 && !strings.HasPrefix(parts[2], "http") && !strings.HasPrefix(parts[2], "data:") {
				parts[2] = cfg.ImageBaseURL + "/" + strings.TrimLeft(parts[2], "/")
				return `![` + parts[1] + `](` + parts[2] + `)`
			}
			return match
		})
		lines = strings.Split(input, "\n")
	}

	for i := 0; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])

		// Handle empty lines
		if line == "" {
			if state.inParagraph {
				buffer.WriteString("</p>\n")
				state.inParagraph = false
			}
			continue
		}

		// Handle Lists
		if listMatch := regexp.MustCompile(`^(\s*)([-*+]|\d+\.)\s+(.+)$`).FindStringSubmatch(line); listMatch != nil {
			indent := len(listMatch[1])
			marker := listMatch[2]
			content := listMatch[3]

			if state.inParagraph {
				buffer.WriteString("</p>\n")
				state.inParagraph = false
			}

			isOrdered := regexp.MustCompile(`^\d+\.$`).MatchString(marker)
			newDepth := indent/2 + 1

			// Close lists if needed
			for state.listDepth > newDepth {
				if state.inOrderedList {
					buffer.WriteString("</ol>\n")
				} else {
					buffer.WriteString("</ul>\n")
				}
				state.listDepth--
			}

			// Open new lists if needed
			for state.listDepth < newDepth {
				if isOrdered {
					buffer.WriteString("<ol>\n")
				} else {
					buffer.WriteString("<ul>\n")
				}
				state.listDepth++
			}

			// Handle task lists
			if taskMatch := regexp.MustCompile(`^\[([ xX])]\s+(.+)$`).FindStringSubmatch(content); taskMatch != nil {
				checked := taskMatch[1] != " "
				content = taskMatch[2]
				buffer.WriteString(`<li><input type="checkbox" disabled`)
				if checked {
					buffer.WriteString(` checked`)
				}
				buffer.WriteString(`> ` + handleInlineFormatting(html.EscapeString(content)) + "</li>\n")
			} else {
				buffer.WriteString("<li>" + handleInlineFormatting(html.EscapeString(content)) + "</li>\n")
			}
			continue
		}

		// Handle other elements (kept same as before...)
		if strings.HasPrefix(line, "```") {
			if state.inCodeBlock {
				buffer.WriteString("</code></pre>\n")
				state.inCodeBlock = false
				state.codeLanguage = ""
			} else {
				language := strings.TrimPrefix(line, "```")
				state.codeLanguage = strings.TrimSpace(language)
				buffer.WriteString("<pre><code")
				if state.codeLanguage != "" {
					buffer.WriteString(` class="language-` + html.EscapeString(state.codeLanguage) + `"`)
				}
				buffer.WriteString(">\n")
				state.inCodeBlock = true
			}
			continue
		}

		if state.inCodeBlock {
			buffer.WriteString(html.EscapeString(line) + "\n")
			continue
		}

		// Handle Headings
		if match := regexp.MustCompile(`^(#{1,6})\s+(.+?)(?:\s+{#([^}]+)})?$`).FindStringSubmatch(line); match != nil {
			level := len(match[1])
			text := html.EscapeString(match[2])
			id := match[3]

			if state.inParagraph {
				buffer.WriteString("</p>\n")
				state.inParagraph = false
			}

			buffer.WriteString("<h" + strconv.Itoa(level))
			if id != "" {
				buffer.WriteString(` id="` + id + `"`)
			}
			buffer.WriteString(">" + handleInlineFormatting(text) + "</h" + strconv.Itoa(level) + ">\n")
			continue
		}

		// Handle regular paragraphs
		if !state.inParagraph {
			buffer.WriteString("<p>")
			state.inParagraph = true
		} else {
			buffer.WriteString(" ")
		}
		buffer.WriteString(handleInlineFormatting(html.EscapeString(line)))
	}

	// Close any remaining open tags
	if state.inParagraph {
		buffer.WriteString("</p>\n")
	}
	for state.listDepth > 0 {
		if state.inOrderedList {
			buffer.WriteString("</ol>\n")
		} else {
			buffer.WriteString("</ul>\n")
		}
		state.listDepth--
	}
	if state.inCodeBlock {
		buffer.WriteString("</code></pre>\n")
	}

	return buffer.String()
}
