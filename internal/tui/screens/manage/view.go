package manage

import (
	"skli/internal/tui/screens/manage/confirm"
	"skli/internal/tui/screens/manage/input_remote"
	"skli/internal/tui/screens/manage/list_view"
	"skli/internal/tui/screens/manage/selecting_remote"
	"skli/internal/tui/screens/manage/uploading"
)

func (s ManageScreen) View() string {
	switch s.State {
	case StateList:
		return list_view.View(s.List, s.Skills, s.Msg)
	case StateConfirm:
		return confirm.View(s.ToDelete, s.ConfirmCursor)
	case StateSelectingRemote:
		return selecting_remote.View(s.RemoteList)
	case StateInputRemote:
		return input_remote.View(s.RemoteInput)
	case StateUploading:
		return uploading.View(s.Msg)
	}
	return ""
}
