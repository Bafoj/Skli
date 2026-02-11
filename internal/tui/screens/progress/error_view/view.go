package error_view

import (
	"fmt"
	"skli/internal/tui/shared"
)

func View(errorMessage string) string {
	return shared.ErrorStyle.Render(fmt.Sprintf("✘ Error: %s", errorMessage)) +
		shared.HelpStyle.Render("\nPress 'r' to retry or 'q' to quit")
}
