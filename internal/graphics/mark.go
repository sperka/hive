// Package graphics owns the optional terminal image used by the dashboard header.
package graphics

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"image"
	"image/color"
	"math"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/ansi/kitty"
	"golang.org/x/term"
)

const (
	markColumns = 8
	markRows    = 3
)

type Mark struct {
	id        int
	enabled   bool
	tmux      bool
	color     string
	rendering string
}

func NewMark() *Mark {
	if !term.IsTerminal(int(os.Stdout.Fd())) {
		return nil
	}
	mode := strings.ToLower(os.Getenv("HIVE_GRAPHICS"))
	// Keep the original font-rendered hexagons unless an image is requested.
	if mode != "kitty" {
		return nil
	}
	insideTmux := os.Getenv("TMUX") != ""
	passthrough := ""
	if insideTmux {
		output, _ := exec.Command("tmux", "show-option", "-gv", "allow-passthrough").Output()
		passthrough = strings.TrimSpace(string(output))
	}
	if !graphicsEnabled(mode, insideTmux, passthrough) {
		return nil
	}
	var id [3]byte
	if _, err := rand.Read(id[:]); err != nil {
		return nil
	}
	value := int(id[0])<<16 | int(id[1])<<8 | int(id[2])
	if value == 0 {
		return nil
	}
	return &Mark{id: value, enabled: true, tmux: insideTmux}
}

func graphicsEnabled(mode string, tmux bool, passthrough string) bool {
	return mode == "kitty" && (!tmux || passthrough == "on" || passthrough == "all")
}

func (m *Mark) Render(themeColor string) string {
	if m == nil || !m.enabled {
		return ""
	}
	if m.rendering != "" && m.color == themeColor {
		return m.rendering
	}
	rgb, err := strconv.ParseUint(strings.TrimPrefix(themeColor, "#"), 16, 24)
	if err != nil {
		rgb = 0x7dcfff
	}
	img := drawMark(color.NRGBA{R: uint8(rgb >> 16), G: uint8(rgb >> 8), B: uint8(rgb), A: 255})
	options := &kitty.Options{
		Action: kitty.TransmitAndPut, ID: m.id, Quite: 2, Format: kitty.PNG,
		Transmission: kitty.Direct, Chunk: true, VirtualPlacement: true,
		DoNotMoveCursor: true, Columns: markColumns, Rows: markRows,
	}
	if m.tmux {
		options.ChunkFormatter = ansi.TmuxPassthrough
	}
	var output bytes.Buffer
	if err := kitty.EncodeGraphics(&output, img, options); err != nil {
		return ""
	}
	// The image ID is encoded in truecolor. Placeholders follow ordinary text
	// repainting, so overlays and resizes do not leave an image at old coordinates.
	for row := 0; row < markRows; row++ {
		if row > 0 {
			output.WriteByte('\n')
		}
		fmt.Fprintf(&output, "\x1b[38;2;%d;%d;%dm", m.id>>16, (m.id>>8)&255, m.id&255)
		for column := 0; column < markColumns; column++ {
			output.WriteRune(kitty.Placeholder)
			output.WriteRune(kitty.Diacritic(row))
			output.WriteRune(kitty.Diacritic(column))
		}
		output.WriteString("\x1b[39m")
	}
	m.color, m.rendering = themeColor, output.String()
	return m.rendering
}

func (m *Mark) Close() string {
	if m == nil || !m.enabled {
		return ""
	}
	sequence := ansi.KittyGraphics(nil, "a=d", "d=I", fmt.Sprintf("i=%d", m.id), "q=2")
	if m.tmux {
		return ansi.TmuxPassthrough(sequence)
	}
	return sequence
}

type point struct{ x, y float64 }

func drawMark(tint color.NRGBA) image.Image {
	img := image.NewNRGBA(image.Rect(0, 0, markColumns*18, markRows*32))
	// Match the original 3-4-3 text mark, including its leading padding.
	var centers []point
	for row, count := range []int{3, 4, 3} {
		start := markColumns/2 - count + 1
		for column := start; column < start+2*count; column += 2 {
			centers = append(centers, point{(float64(column) + 0.5) * 18, (float64(row) + 0.5) * 32})
		}
	}
	for _, center := range centers {
		var vertices [6]point
		for i := range vertices {
			angle := math.Pi/6 + float64(i)*math.Pi/3
			vertices[i] = point{center.x + math.Cos(angle)*8, center.y + math.Sin(angle)*8}
		}
		for y := int(center.y) - 9; y <= int(center.y)+9; y++ {
			for x := int(center.x) - 9; x <= int(center.x)+9; x++ {
				var alpha float64
				for sy := 0; sy < 4; sy++ {
					for sx := 0; sx < 4; sx++ {
						p := point{float64(x) + (float64(sx)+0.5)/4, float64(y) + (float64(sy)+0.5)/4}
						distance := math.Inf(1)
						for i, a := range vertices {
							b := vertices[(i+1)%len(vertices)]
							cross := (b.x-a.x)*(p.y-a.y) - (b.y-a.y)*(p.x-a.x)
							distance = math.Min(distance, cross/math.Hypot(b.x-a.x, b.y-a.y))
						}
						if distance >= 0 && distance < 1.2 {
							alpha += 235
						}
					}
				}
				if alpha > 0 {
					pixel := tint
					pixel.A = uint8(alpha / 16)
					img.SetNRGBA(x, y, pixel)
				}
			}
		}
	}
	return img
}
