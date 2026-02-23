package uploading

import (
	"fmt"
	"skli/internal/tui/shared"
)

func View(msg string) string {
	return fmt.Sprintf("\n  Uploading skill...\n\n  %s\n\n  %s",
		shared.InfoStyle.Render(msg),
		shared.HelpStyle.Render("esc/enter: back"),
	)
}
