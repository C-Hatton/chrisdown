package chrisdown

import (
	"bytes"
	"html"
	"regexp"
	"strconv"
	"strings"
)

// Helper function to handle inline formatting (unchanged)
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
		inParagraph  bool
		inCodeBlock  bool
		codeLanguage string
		listStack    []struct {
			isOrdered bool
			depth     int
		}
		currentListDepth int
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
		indent := len(lines[i]) - len(strings.TrimLeft(lines[i], " "))

		// Handle empty lines
		if line == "" {
			if state.inParagraph {
				buffer.WriteString("</p>\n")
				state.inParagraph = false
			}
			continue
		}

		// Handle Lists
		if listMatch := regexp.MustCompile(`^([-*+]|\d+\.|[a-z]\.|[ivxIVX]+\.)\s+(.+)$`).FindStringSubmatch(line); listMatch != nil {
			marker := listMatch[1]
			content := listMatch[2]
			depth := indent/2 + 1

			if state.inParagraph {
				buffer.WriteString("</p>\n")
				state.inParagraph = false
			}

			isOrdered := regexp.MustCompile(`^(\d+\.|[a-z]\.|[ivxIVX]+\.)$`).MatchString(marker)

			// Close deeper lists
			for len(state.listStack) > 0 && state.listStack[len(state.listStack)-1].depth > depth {
				if state.listStack[len(state.listStack)-1].isOrdered {
					buffer.WriteString("</ol>\n")
				} else {
					buffer.WriteString("</ul>\n")
				}
				state.listStack = state.listStack[:len(state.listStack)-1]
			}

			// If we're at a new depth or switching list types, start a new list
			if len(state.listStack) == 0 ||
				state.listStack[len(state.listStack)-1].depth < depth ||
				(state.listStack[len(state.listStack)-1].depth == depth &&
					state.listStack[len(state.listStack)-1].isOrdered != isOrdered) {

				if isOrdered {
					buffer.WriteString("<ol style=\"list-style-type: ")
					switch depth % 3 {
					case 1:
						buffer.WriteString("decimal")
					case 2:
						buffer.WriteString("lower-alpha")
					case 0:
						buffer.WriteString("upper-roman")
					}
					buffer.WriteString("\">\n")
				} else {
					buffer.WriteString("<ul style=\"list-style-type: ")
					switch depth % 3 {
					case 1:
						buffer.WriteString("circle")
					case 2:
						buffer.WriteString("disc")
					case 0:
						buffer.WriteString("square")
					}
					buffer.WriteString("\">\n")
				}
				state.listStack = append(state.listStack, struct {
					isOrdered bool
					depth     int
				}{isOrdered, depth})
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

			// If next line is empty or has less indent, close appropriate lists
			if i == len(lines)-1 ||
				len(strings.TrimSpace(lines[i+1])) == 0 ||
				(len(lines[i+1])-len(strings.TrimLeft(lines[i+1], " "))) < indent {
				currentDepth := depth
				for len(state.listStack) > 0 && state.listStack[len(state.listStack)-1].depth >= currentDepth {
					if state.listStack[len(state.listStack)-1].isOrdered {
						buffer.WriteString("</ol>\n")
					} else {
						buffer.WriteString("</ul>\n")
					}
					state.listStack = state.listStack[:len(state.listStack)-1]
				}
			}
			continue
		}

		// Handle other elements (unchanged)
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
	for len(state.listStack) > 0 {
		if state.listStack[len(state.listStack)-1].isOrdered {
			buffer.WriteString("</ol>\n")
		} else {
			buffer.WriteString("</ul>\n")
		}
		state.listStack = state.listStack[:len(state.listStack)-1]
	}
	if state.inCodeBlock {
		buffer.WriteString("</code></pre>\n")
	}

	return buffer.String()
}
