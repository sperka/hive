package components

import (
	"strings"

	"github.com/mdipanjan/hive/internal/styles"
)

type FormData struct {
	Tool       int
	CustomTool string
	Path       string
	Name       string
	Focus      int
	Button     int
	Error      string
}

func RenderNewDialog(tools []string, form FormData) string {
	var content strings.Builder

	content.WriteString(styles.PanelTitle.Render("NEW SESSION"))
	content.WriteString("\n\n")

	toolLabel := styles.Label.Render("TOOL")
	choices := append(append([]string{}, tools...), "Custom...")
	toolValue := renderToolSelector(choices, form.Tool, form.Focus == 0)
	content.WriteString(toolLabel + toolValue + "\n\n")
	if form.Tool == len(tools) {
		content.WriteString(styles.Label.Render("CUSTOM") + form.CustomTool + "\n\n")
	}

	pathLabel := styles.Label.Render("PATH")
	pathValue := RenderTextInput(form.Path, form.Focus == 2, 30)
	content.WriteString(pathLabel + pathValue + "\n\n")

	nameLabel := styles.Label.Render("NAME")
	namePlaceholder := form.Name
	if namePlaceholder == "" {
		namePlaceholder = "(auto-generate)"
	}
	nameValue := RenderTextInput(namePlaceholder, form.Focus == 3, 30)
	content.WriteString(nameLabel + nameValue + "\n\n")

	if form.Error != "" {
		content.WriteString(styles.Normal.Render("Error: "+form.Error) + "\n\n")
	}
	content.WriteString(RenderButtons([]string{"Create", "Cancel"}, form.Button, form.Focus == 4, 70))

	dialogStyle := styles.Panel.Width(70).Padding(1, 2)
	return dialogStyle.Render(content.String())
}

func renderToolSelector(tools []string, selected int, focused bool) string {
	var parts []string

	for i, tool := range tools {
		var icon string
		if i == selected {
			icon = styles.IconRunning + " "
		} else {
			icon = styles.IconIdle + " "
		}

		if i == selected && focused {
			parts = append(parts, styles.Selected.Render(icon+tool))
		} else if i == selected {
			parts = append(parts, styles.Value.Render(icon+tool))
		} else {
			parts = append(parts, styles.Dim.Render(icon+tool))
		}
	}

	return strings.Join(parts, "  ")
}
