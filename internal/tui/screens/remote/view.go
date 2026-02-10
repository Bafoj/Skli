package remote

import (
	"skli/internal/tui/screens/remote/input"
	"skli/internal/tui/screens/remote/input_new"
	"skli/internal/tui/screens/remote/managing"
	"skli/internal/tui/screens/remote/selecting"
)

func (s RemoteScreen) View() string {
	switch s.State {
	case StateSelecting:
		return selecting.View(s.List)
	case StateInput:
		return input.View(s.TextInput, len(s.Remotes) > 0)
	case StateInputNew:
		return input_new.View(s.TextInput)
	case StateManaging:
		return managing.View(s.List)
	}
	return ""
}
