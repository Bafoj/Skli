package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"skli/internal/config"
	"skli/internal/sync"
	"skli/internal/tui"
)

var (
	successStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00")).Bold(true)
	errorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000")).Bold(true)
	infoStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#7D56F4"))
	dimStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))
)

func main() {
	syncFlag := flag.Bool("sync", false, "Sincronizar todos los skills instalados desde sus repos de origen")
	pathFlag := flag.String("path", "skills", "Directorio dentro del repo donde buscar skills")
	flag.Parse()

	// Cargar configuración global
	cfg, _ := config.LoadConfig()

	// Manejar comandos
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "config":
			runConfig(cfg)
			return
		case "sync":
			runSync()
			return
		}
	}

	if *syncFlag {
		runSync()
		return
	}

	// Obtener el repo URL de los argumentos posicionales si existe
	initialURL := ""
	if args := flag.Args(); len(args) > 0 {
		initialURL = args[0]
	}

	// Modo interactivo normal
	p := tea.NewProgram(tui.InitialModel(initialURL, *pathFlag, cfg.LocalPath, false, cfg.Remotes), tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Printf("Ocurrió un error: %v", err)
		os.Exit(1)
	}
}

func runSync() {
	// ... (rest of the function)
	fmt.Println(infoStyle.Render("🔄 Sincronizando skills..."))
	fmt.Println()

	results, err := sync.SyncAllSkills()
	if err != nil {
		fmt.Println(errorStyle.Render(fmt.Sprintf("✘ Error sincronizando: %v", err)))
		os.Exit(1)
	}

	if len(results) == 0 {
		fmt.Println(infoStyle.Render("ℹ No hay skills instalados para sincronizar (skli.lock vacío)."))
		fmt.Println(infoStyle.Render("  Usa 'skli' para instalar skills primero."))
		return
	}

	updated := 0
	skipped := 0
	errors := 0

	for _, r := range results {
		if r.Error != nil {
			fmt.Println(errorStyle.Render(fmt.Sprintf("  ✘ %s: %v", r.SkillName, r.Error)))
			errors++
		} else if r.Updated {
			fmt.Println(successStyle.Render(fmt.Sprintf("  ✔ %s actualizado", r.SkillName)))
			updated++
		} else if r.Skipped {
			fmt.Println(dimStyle.Render(fmt.Sprintf("  ○ %s sin cambios", r.SkillName)))
			skipped++
		}
	}

	fmt.Println()
	if errors > 0 {
		fmt.Println(errorStyle.Render(fmt.Sprintf("Completado con %d errores. %d actualizados, %d sin cambios.", errors, updated, skipped)))
	} else if updated == 0 {
		fmt.Println(successStyle.Render(fmt.Sprintf("✔ Todos los skills están actualizados (%d verificados).", skipped)))
	} else {
		fmt.Println(successStyle.Render(fmt.Sprintf("✔ %d skills actualizados, %d sin cambios.", updated, skipped)))
	}
}

func runConfig(cfg config.Config) {
	p := tea.NewProgram(tui.InitialModel("", "skills", cfg.LocalPath, true, cfg.Remotes), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Ocurrió un error en la configuración: %v", err)
		os.Exit(1)
	}
}
