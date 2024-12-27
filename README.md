Usage:

	// Declare images link
	config := Config{
		ImageBaseURL: "https://example.com/images",
	}

	// Render the markdown
	outputContent := chrisdown.RenderMarkdown(string(inputContent), chrisdown.Config(config))

Example Input & Output:
```
# title

- circle
    - square
        - disc

1. decimal
   1. lower-alpha
      1. decimal

```
```
<h1>title</h1>
<ul style="list-style-type: circle">
<li>circle</li>
<ul style="list-style-type: square">
<li>square</li>
<ul style="list-style-type: disc">
<li>disc</li>
</ul>
</ul>
<ol style="list-style-type: decimal">
<li>decimal</li>
<ol style="list-style-type: lower-alpha">
<li>lower-alpha</li>
<ol style="list-style-type: decimal">
<li>decimal</li>
</ol>
</ol>
</ol>
</ul>
```

You may want to add some CSS to render the HTML nicely.
