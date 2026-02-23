package confirm

import (
	"fmt"
	"skli/internal/db"
	"skli/internal/tui/shared"
)

func View(toDelete *db.InstalledSkill, confirmCursor int) string {
	if toDelete == nil {
		return ""
	}

	var yes, no string
	if confirmCursor == 0 {
		yes = shared.SuccessStyle.Render(shared.SelectorDot(true) + " Yes")
		no = shared.DimStyle.Render(shared.SelectorDot(false) + " No")
	} else {
		yes = shared.DimStyle.Render(shared.SelectorDot(false) + " Yes")
		no = shared.ErrorStyle.Render(shared.SelectorDot(true) + " No")
	}

	return fmt.Sprintf(
		"\n  %s\n\n  Skill: %s\n  Path:  %s\n\n  %s    %s",
		shared.ErrorPopup("This action will delete the skill"),
		shared.ErrorStyle.Render(toDelete.Name),
		shared.DimStyle.Render(toDelete.Path),
		yes,
		no,
	)
}
