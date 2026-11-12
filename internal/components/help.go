package components

import (
	"fmt"

	"github.com/mdipanjan/hive/internal/styles"
)

var Version = "dev"

func RenderHelp(widths ...int) string {
	theme := fmt.Sprintf("theme [%s]", styles.GetCurrentTheme().Name)
	if len(widths) > 0 {
		theme = "theme"
	}
	items := []HelpItem{
		{Key: "n", Desc: "new"},
		{Key: "enter", Desc: "attach"},
		{Key: "d", Desc: "delete"},
		{Key: "r", Desc: "rename"},
		{Key: "/", Desc: "search"},
		{Key: "t", Desc: theme},
		{Key: "?", Desc: "help"},
		{Key: "q", Desc: "quit"},
	}
	if len(widths) > 0 {
		items = append([]HelpItem{{Key: "j/k", Desc: "select"}}, items...)
	}
	return RenderHelpBar(items, widths...)
}
