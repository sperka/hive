package components

import (
	"fmt"

	"github.com/mdipanjan/hive/internal/styles"
)

var Version = "dev"

func RenderHelp() string {
	themeName := styles.GetCurrentTheme().Name
	return RenderHelpBar([]HelpItem{
		{Key: "n", Desc: "new"},
		{Key: "enter", Desc: "attach"},
		{Key: "d", Desc: "delete"},
		{Key: "r", Desc: "rename"},
		{Key: "/", Desc: "search"},
		{Key: "t", Desc: fmt.Sprintf("theme [%s]", themeName)},
		{Key: "?", Desc: "help"},
		{Key: "q", Desc: "quit"},
	})
}
