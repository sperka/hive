package graphics

import (
	"bytes"
	"encoding/base64"
	"image/color"
	"image/png"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi/kitty"
)

func TestMarkKeepsOriginalStaggeredHexagonSpacing(t *testing.T) {
	img := drawMark(color.NRGBA{R: 255, A: 255})
	for row, cells := range []string{
		"  x x x ",
		" x x x x",
		"  x x x ",
	} {
		for column := 0; column < 8; column++ {
			ink := false
			for y := row * 32; y < (row+1)*32; y++ {
				for x := column * 18; x < (column+1)*18; x++ {
					_, _, _, alpha := img.At(x, y).RGBA()
					ink = ink || alpha > 0
				}
			}
			want := cells[column] == 'x'
			if ink != want {
				t.Errorf("row %d column %d has ink=%v, want %v for original 3-4-3 spacing",
					row, column, ink, want)
			}
		}
	}
}

func TestGraphicsRequiresExplicitOptIn(t *testing.T) {
	for _, tt := range []struct {
		name, mode, passthrough string
		tmux, want              bool
	}{
		{name: "default retains original glyphs"},
		{name: "auto retains original glyphs", mode: "auto"},
		{name: "disabled", mode: "off"},
		{name: "explicit", mode: "kitty", want: true},
		{name: "tmux off", mode: "kitty", tmux: true, passthrough: "off"},
		{name: "tmux on", mode: "kitty", tmux: true, passthrough: "on", want: true},
		{name: "tmux all", mode: "kitty", tmux: true, passthrough: "all", want: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if got := graphicsEnabled(tt.mode, tt.tmux, tt.passthrough); got != tt.want {
				t.Fatalf("enabled = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMarkEncodesPNGAndPlaceholderGeometry(t *testing.T) {
	mark := &Mark{id: 0x123456, enabled: true}
	view := mark.Render("#7dcfff")
	if w, h := lipgloss.Size(view); w != 8 || h != 3 {
		t.Fatalf("mark occupies %dx%d, want 8x3", w, h)
	}
	if strings.Count(view, string(kitty.Placeholder)) != 24 {
		t.Fatal("missing image placeholder cells")
	}
	var payload string
	first := true
	for _, part := range strings.Split(view, "\x1b\\") {
		if !strings.HasPrefix(part, "\x1b_G") {
			continue
		}
		fields := strings.SplitN(strings.TrimPrefix(part, "\x1b_G"), ";", 2)
		var options kitty.Options
		if err := options.UnmarshalText([]byte(fields[0])); err != nil {
			t.Fatal(err)
		}
		if first {
			if options.ID != 0x123456 || !options.VirtualPlacement || options.Quite != 2 ||
				options.Columns != 8 || options.Rows != 3 || options.Format != kitty.PNG {
				t.Fatalf("wrong protocol options: %+v", options)
			}
			first = false
		}
		if len(fields[1]) > 4096 {
			t.Fatal("graphics payload exceeds protocol chunk size")
		}
		payload += fields[1]
	}
	raw, err := base64.StdEncoding.DecodeString(payload)
	if err != nil {
		t.Fatal(err)
	}
	image, err := png.Decode(bytes.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}
	if image.Bounds().Dx() != 144 || image.Bounds().Dy() != 96 {
		t.Fatalf("unexpected image bounds: %v", image.Bounds())
	}
	_, _, _, alpha := image.At(0, 0).RGBA()
	if alpha != 0 {
		t.Fatal("mark background must be transparent")
	}
	if mark.Render("#7dcfff") != view {
		t.Fatal("unchanged mark produces a different transmission on every frame")
	}
	if mark.Render("#ff0000") == view {
		t.Fatal("mark failed to follow the theme color")
	}
}

func TestMarkUsesTmuxPassthroughAndOnlyDeletesItsOwnImage(t *testing.T) {
	mark := &Mark{id: 42, enabled: true, tmux: true}
	if !strings.Contains(mark.Render("#7dcfff"), "\x1bPtmux;\x1b\x1b_G") {
		t.Fatal("tmux image transmission was not wrapped")
	}
	cleanup := mark.Close()
	if !strings.Contains(cleanup, "d=I") || !strings.Contains(cleanup, "i=42") ||
		strings.Contains(cleanup, "d=A") {
		t.Fatalf("cleanup must delete only this image: %q", cleanup)
	}
	var disabled *Mark
	if disabled.Render("#ffffff") != "" || disabled.Close() != "" {
		t.Fatal("disabled graphics emitted terminal sequences")
	}
}
