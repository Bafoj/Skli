package selecting_remote

import (
	"github.com/charmbracelet/bubbles/list"
)

func View(remoteList list.Model) string {
	return "\n" + remoteList.View()
}
