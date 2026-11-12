package styles

import "testing"

func TestLoadedThemeIsCurrentAndCyclesFromSelection(t *testing.T) {
	before := GetCurrentTheme()
	t.Cleanup(func() { ApplyTheme(before) })

	for _, tc := range []struct {
		key  string
		name string
		next string
	}{
		{key: "nord", name: "Nord", next: "Gruvbox"},
		{key: "zinc-dark", name: "Zinc Dark", next: "Tokyo Night"},
	} {
		t.Run(tc.key, func(t *testing.T) {
			ApplyTheme(GetThemeByKey(tc.key))
			if got := GetCurrentTheme().Name; got != tc.name {
				t.Errorf("loaded theme name = %q, want %q", got, tc.name)
			}
			if got := NextTheme().Name; got != tc.next {
				t.Errorf("next theme = %q, want %q", got, tc.next)
			}
			if got := GetCurrentTheme().Name; got != tc.next {
				t.Errorf("current theme after cycling = %q, want %q", got, tc.next)
			}
		})
	}
}
