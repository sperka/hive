package styles

import "github.com/charmbracelet/lipgloss"

var (
	ColorGreen  lipgloss.Color
	ColorYellow lipgloss.Color
	ColorGray   lipgloss.Color
	ColorCyan   lipgloss.Color
	ColorWhite  lipgloss.Color
	ColorDim    lipgloss.Color
	ColorBg     lipgloss.Color
	ColorBgDark lipgloss.Color
)

var (
	IconActive    string
	IconRunning   string
	IconReady     string
	IconCompleted string
	IconFailed    string
	IconIdle      string
)

var (
	OuterBox   lipgloss.Style
	Panel      lipgloss.Style
	PanelTitle lipgloss.Style
	Selected   lipgloss.Style
	Normal     lipgloss.Style
	Dim        lipgloss.Style
	Logo       lipgloss.Style
	Stats      lipgloss.Style
	Help       lipgloss.Style
	HelpKey    lipgloss.Style
	HelpDesc   lipgloss.Style
	Label      lipgloss.Style
	Value      lipgloss.Style
)

func ApplyTheme(t Theme) {
	ColorGreen = t.Green
	ColorYellow = t.Yellow
	ColorGray = t.Gray
	ColorCyan = t.Cyan
	ColorWhite = t.White
	ColorDim = t.Dim
	ColorBg = t.Bg
	ColorBgDark = t.BgDark

	IconActive = lipgloss.NewStyle().Foreground(ColorGreen).Render("●")
	IconRunning = lipgloss.NewStyle().Foreground(ColorYellow).Render("◐")
	IconReady = lipgloss.NewStyle().Foreground(ColorCyan).Render("◌")
	IconCompleted = lipgloss.NewStyle().Foreground(ColorGreen).Render("✓")
	IconFailed = lipgloss.NewStyle().Foreground(ColorYellow).Render("✕")
	IconIdle = lipgloss.NewStyle().Foreground(ColorGray).Render("○")

	OuterBox = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorGray).
		Padding(1, 2)

	Panel = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorGray).
		Padding(1, 2)

	PanelTitle = lipgloss.NewStyle().
		Bold(true).
		Foreground(ColorCyan)

	Selected = lipgloss.NewStyle().
		Bold(true).
		Foreground(ColorWhite).
		Background(ColorDim)

	Normal = lipgloss.NewStyle().
		Foreground(ColorWhite)

	Dim = lipgloss.NewStyle().
		Foreground(ColorDim)

	Logo = lipgloss.NewStyle().
		Foreground(ColorCyan).
		Bold(true)

	Stats = lipgloss.NewStyle().
		Foreground(ColorGray)

	Help = lipgloss.NewStyle().
		Foreground(ColorWhite).
		Background(ColorBgDark).
		Padding(0, 2)

	HelpKey = lipgloss.NewStyle().
		Foreground(ColorCyan).
		Background(ColorBgDark).
		Bold(true)

	HelpDesc = lipgloss.NewStyle().
		Foreground(ColorWhite).
		Background(ColorBgDark)

	Label = lipgloss.NewStyle().
		Foreground(ColorGray).
		Width(10)

	Value = lipgloss.NewStyle().
		Foreground(ColorWhite)
}

func init() {
	ApplyTheme(TokyoNight)
}
