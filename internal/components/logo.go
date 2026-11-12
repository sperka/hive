package components

import "github.com/mdipanjan/hive/internal/styles"

func RenderLogo() string {
	return styles.Logo.Bold(true).Italic(false).Render("HIVE")
}
