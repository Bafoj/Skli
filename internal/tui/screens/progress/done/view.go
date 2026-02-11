package done

import (
	"fmt"
	"skli/internal/tui/shared"
)

func View(configMode bool, configLocalPath string) string {
	var msg string
	if configMode {
		msg = shared.SuccessStyle.Render("✔ Configuration saved successfully!")
	} else {
		msg = shared.SuccessStyle.Render(fmt.Sprintf("✔ Skills installed successfully in ./%s/!", configLocalPath))
	}
	return msg + shared.HelpStyle.Render("\nPress any key to quit")
}
