package tui

type Screen string

type Overlay string

const (
	ScreenList Screen = "list"
	ScreenNew  Screen = "new"

	OverlayNone          Overlay = ""
	OverlayHelp          Overlay = "help"
	OverlaySearch        Overlay = "search"
	OverlayDeleteConfirm Overlay = "delete-confirm"
	OverlayRename        Overlay = "rename"
	OverlayPathPicker    Overlay = "path-picker"
)

type AppState struct {
	Screen  Screen
	Overlay Overlay
}

func NewAppState() AppState {
	return AppState{Screen: ScreenList, Overlay: OverlayNone}
}

func (s AppState) ShowingHelp() bool { return s.Overlay == OverlayHelp }
func (s AppState) Searching() bool   { return s.Overlay == OverlaySearch }
func (s AppState) ConfirmingDelete() bool {
	return s.Overlay == OverlayDeleteConfirm
}
func (s AppState) PickingPath() bool     { return s.Overlay == OverlayPathPicker }
func (s AppState) RenamingSession() bool { return s.Overlay == OverlayRename }
func (s AppState) CreatingSession() bool {
	return s.Screen == ScreenNew && s.Overlay == OverlayNone
}

func (s *AppState) ShowHelp()            { s.Overlay = OverlayHelp }
func (s *AppState) Search()              { s.Overlay = OverlaySearch }
func (s *AppState) ConfirmDelete()       { s.Overlay = OverlayDeleteConfirm }
func (s *AppState) RenameSession()       { s.Overlay = OverlayRename }
func (s *AppState) PickPath()            { s.Overlay = OverlayPathPicker }
func (s *AppState) CloseOverlay()        { s.Overlay = OverlayNone }
func (s *AppState) StartNewSession()     { s.Screen, s.Overlay = ScreenNew, OverlayNone }
func (s *AppState) ReturnToSessionList() { s.Screen, s.Overlay = ScreenList, OverlayNone }
