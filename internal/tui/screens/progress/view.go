package progress

import (
	"skli/internal/tui/screens/progress/done"
	"skli/internal/tui/screens/progress/downloading"
	"skli/internal/tui/screens/progress/error_view"
)

func (s ProgressScreen) View() string {
	switch s.State {
	case StateDownloading:
		return downloading.View(s.Spinner, s.ConfigLocalPath)
	case StateDone:
		return done.View(s.ConfigMode, s.ConfigLocalPath)
	case StateError:
		return error_view.View(s.ErrorMessage)
	}
	return ""
}
