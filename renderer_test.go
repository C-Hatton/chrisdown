package chrisdown

import (
	"io/ioutil"
	"os"
	"testing"
)

// TestRenderMarkdownBasic tests the RenderMarkdown function with predefined input and output.
func TestRenderMarkdownBasic(t *testing.T) {
	config := Config{
		ImageBaseURL: "https://example.com/images",
	}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "basic heading",
			input:    "# Heading\n",
			expected: "<h1>Heading</h1>\n",
		},
		{
			name:     "subheading",
			input:    "## Subheading\n",
			expected: "<h2>Subheading</h2>\n",
		},
		{
			name:     "unordered list",
			input:    "- Item 1\n- Item 2\n",
			expected: "<ul>\n<li>Item 1</li>\n<li>Item 2</li>\n</ul>\n",
		},
		{
			name:     "paragraph with formatting",
			input:    "This is **bold** text and this is *italic* text.\n",
			expected: "<p>This is <strong>bold</strong> text and this is <em>italic</em> text.</p>\n",
		},
		{
			name:     "relative image",
			input:    "![alt text](/image.png)\n",
			expected: "<p><img src=\"https://example.com/images/image.png\" alt=\"alt text\"></p>\n",
		},
		{
			name:     "absolute image",
			input:    "![alt text](https://other.com/image.png)\n",
			expected: "<p><img src=\"https://other.com/image.png\" alt=\"alt text\"></p>\n",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			output := RenderMarkdown(test.input, config)
			if output != test.expected {
				t.Errorf("RenderMarkdown(%q) = %q; want %q", test.input, output, test.expected)
			}
		})
	}
}

// TestMarkdownFileCLI simulates testing the markdown renderer using files.
func TestMarkdownFileCLI(t *testing.T) {
	// Create temporary input and output files
	inputFile, err := ioutil.TempFile("", "test_input_*.md")
	if err != nil {
		t.Fatal(err)
	}
	defer func(name string) {
		err := os.Remove(name)
		if err != nil {
			t.Fatal(err)
		}
	}(inputFile.Name())

	outputFile, err := ioutil.TempFile("", "test_output_*.html")
	if err != nil {
		t.Fatal(err)
	}
	defer func(name string) {
		err := os.Remove(name)
		if err != nil {
			t.Fatal(err)
		}
	}(outputFile.Name())

	// Write markdown content to the input file
	markdownContent := "# Title\n\nThis is **bold** text and this is *italic* text."
	expectedHTML := "<h1>Title</h1>\n<p>This is <strong>bold</strong> text and this is <em>italic</em> text.</p>\n"

	_, err = inputFile.WriteString(markdownContent)
	if err != nil {
		t.Fatal(err)
	}
	err = inputFile.Close()
	if err != nil {
		t.Fatal(err)
	}

	// Run the CLI simulation
	RunMarkdownRenderer(inputFile.Name(), outputFile.Name())

	// Read the output file
	outputContent, err := ioutil.ReadFile(outputFile.Name())
	if err != nil {
		t.Fatal(err)
	}

	// Compare the output content with the expected HTML
	if string(outputContent) != expectedHTML {
		t.Errorf("Output mismatch.\nGot: %q\nExpected: %q", string(outputContent), expectedHTML)
	}
}

// RunMarkdownRenderer simulates the ./chrisdown CLI functionality.
func RunMarkdownRenderer(inputPath, outputPath string) {
	// Read the markdown input file
	inputContent, err := ioutil.ReadFile(inputPath)
	if err != nil {
		panic(err)
	}

	config := Config{
		ImageBaseURL: "https://example.com/images",
	}

	// Render the markdown
	outputContent := RenderMarkdown(string(inputContent), config)

	// Write the HTML output file
	err = ioutil.WriteFile(outputPath, []byte(outputContent), 0644)
	if err != nil {
		panic(err)
	}
}
