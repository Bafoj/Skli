package main

import (
	"flag"
	"fmt"
	"os"

	"skli/internal/db"
	"skli/internal/sync"
	"skli/internal/tui"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	successStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00")).Bold(true)
	errorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000")).Bold(true)
	infoStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#7D56F4"))
)

func main() {
	syncFlag := flag.Bool("sync", false, "Sincronizar todos los skills instalados desde sus repos de origen")
	flag.Parse()

	if *syncFlag {
		runSync()
		return
	}

	// Modo interactivo normal
	p := tea.NewProgram(tui.InitialModel(), tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Printf("Ocurrió un error: %v", err)
		os.Exit(1)
	}
}

func runSync() {
	fmt.Println(infoStyle.Render("🔄 Sincronizando skills..."))

	database, err := db.InitDB()
	if err != nil {
		fmt.Println(errorStyle.Render(fmt.Sprintf("✘ Error inicializando base de datos: %v", err)))
		os.Exit(1)
	}
	defer database.Close()

	results, err := sync.SyncAllSkills(database)
	if err != nil {
		fmt.Println(errorStyle.Render(fmt.Sprintf("✘ Error sincronizando: %v", err)))
		os.Exit(1)
	}

	if len(results) == 0 {
		fmt.Println(infoStyle.Render("ℹ No hay skills instalados para sincronizar."))
		fmt.Println(infoStyle.Render("  Usa 'skli' para instalar skills primero."))
		return
	}

	fmt.Println()
	updated := 0
	errors := 0

	for _, r := range results {
		if r.Error != nil {
			fmt.Println(errorStyle.Render(fmt.Sprintf("  ✘ %s: %v", r.SkillName, r.Error)))
			errors++
		} else if r.Updated {
			fmt.Println(successStyle.Render(fmt.Sprintf("  ✔ %s actualizado", r.SkillName)))
			updated++
		}
	}

	fmt.Println()
	if errors > 0 {
		fmt.Println(errorStyle.Render(fmt.Sprintf("Completado con %d errores. %d skills actualizados.", errors, updated)))
	} else {
		fmt.Println(successStyle.Render(fmt.Sprintf("✔ %d skills actualizados correctamente.", updated)))
	}
}
