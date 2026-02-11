package main

import (
	"context"
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/urfave/cli/v3"

	"skli/internal/app"
	"skli/internal/config"
)

var (
	successStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00")).Bold(true)
	errorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000")).Bold(true)
	infoStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#7D56F4"))
	dimStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))
)

func main() {
	cfg, _ := config.LoadConfig()
	service := app.NewService(cfg)

	if err := buildCLI(service).Run(context.Background(), os.Args); err != nil {
		fmt.Println(errorStyle.Render(fmt.Sprintf("✘ %v", err)))
		os.Exit(1)
	}
}

func buildCLI(service app.Service) *cli.Command {
	return &cli.Command{
		Name:      "skli",
		Usage:     "gestor de skills",
		UsageText: "skli [comando] [argumentos]",
		CommandNotFound: func(ctx context.Context, c *cli.Command, s string) {
			fmt.Println(errorStyle.Render(fmt.Sprintf("✘ comando desconocido: %s", s)))
			fmt.Println("Usa --help para ver comandos.")
		},
		Commands: []*cli.Command{
			{
				Name:      "add",
				Usage:     "instala skills desde un repo o abre el selector TUI",
				ArgsUsage: "[git-repo-path]",
				Action: func(_ context.Context, cmd *cli.Command) error {
					if cmd.NArg() > 1 {
						return cli.Exit("uso: skli add [git-repo-path]", 1)
					}
					return service.Add(cmd.Args().First())
				},
			},
			{
				Name:      "rm",
				Usage:     "elimina skills instalados",
				ArgsUsage: "[skill-name]",
				Action: func(_ context.Context, cmd *cli.Command) error {
					if cmd.NArg() > 1 {
						return cli.Exit("uso: skli rm [skill-name]", 1)
					}
					if cmd.NArg() == 1 {
						skill, err := service.RemoveByName(cmd.Args().First())
						if err != nil {
							return err
						}
						fmt.Println(successStyle.Render(fmt.Sprintf("✔ skill eliminado: %s (%s)", skill.Name, skill.Path)))
						return nil
					}
					return service.RemoveTUI()
				},
			},
			{
				Name:  "sync",
				Usage: "sincroniza skills instalados con su repo origen",
				Action: func(_ context.Context, cmd *cli.Command) error {
					if cmd.NArg() != 0 {
						return cli.Exit("uso: skli sync", 1)
					}
					return renderSync(service)
				},
			},
			{
				Name:  "list",
				Usage: "lista skills instaladas y skills locales detectadas en ./skills",
				Action: func(_ context.Context, cmd *cli.Command) error {
					if cmd.NArg() != 0 {
						return cli.Exit("uso: skli list", 1)
					}
					return service.ListTUI()
				},
			},
			{
				Name:      "upload",
				Usage:     "sube skills locales a un repo destino",
				ArgsUsage: "[git-dest-repo-path] [local-skill-path]",
				Action: func(_ context.Context, cmd *cli.Command) error {
					if cmd.NArg() != 0 && cmd.NArg() != 2 {
						return cli.Exit("uso: skli upload [git-dest-repo-path] [local-skill-path]", 1)
					}
					if cmd.NArg() == 2 {
						result, err := service.UploadDirect(cmd.Args().Get(0), cmd.Args().Get(1))
						if err != nil {
							return err
						}
						fmt.Println(infoStyle.Render(fmt.Sprintf("🔄 Subiendo %s a %s...", result.Skill.Name, cmd.Args().Get(0))))
						fmt.Println(successStyle.Render("✔ PR creado"))
						fmt.Println(dimStyle.Render(result.PRURL))
						return nil
					}
					return service.UploadTUI()
				},
			},
			{
				Name:  "config",
				Usage: "abre la configuracion de skli",
				Action: func(_ context.Context, cmd *cli.Command) error {
					if cmd.NArg() != 0 {
						return cli.Exit("uso: skli config", 1)
					}
					return service.ConfigTUI()
				},
			},
		},
		Action: func(_ context.Context, cmd *cli.Command) error {
			if cmd.NArg() == 0 {
				cli.ShowRootCommandHelp(cmd)
				return nil
			}
			return cli.Exit("comando desconocido. Usa --help para ver comandos.", 1)
		},
	}
}

func renderSync(service app.Service) error {
	fmt.Println(infoStyle.Render("🔄 Sincronizando skills..."))
	fmt.Println()

	summary, err := service.SyncAll()
	if err != nil {
		return err
	}

	if len(summary.Results) == 0 {
		fmt.Println(infoStyle.Render("ℹ No hay skills instalados para sincronizar (skli.lock vacio)."))
		fmt.Println(infoStyle.Render("  Usa 'skli add' para instalar skills primero."))
		return nil
	}

	for _, r := range summary.Results {
		if r.Error != nil {
			fmt.Println(errorStyle.Render(fmt.Sprintf("  ✘ %s: %v", r.SkillName, r.Error)))
		} else if r.Updated {
			fmt.Println(successStyle.Render(fmt.Sprintf("  ✔ %s actualizado", r.SkillName)))
		} else if r.Skipped {
			fmt.Println(dimStyle.Render(fmt.Sprintf("  ○ %s sin cambios", r.SkillName)))
		}
	}

	fmt.Println()
	if summary.Errors > 0 {
		fmt.Println(errorStyle.Render(fmt.Sprintf("Completado con %d errores. %d actualizados, %d sin cambios.", summary.Errors, summary.Updated, summary.Skipped)))
	} else if summary.Updated == 0 {
		fmt.Println(successStyle.Render(fmt.Sprintf("✔ Todos los skills estan actualizados (%d verificados).", summary.Skipped)))
	} else {
		fmt.Println(successStyle.Render(fmt.Sprintf("✔ %d skills actualizados, %d sin cambios.", summary.Updated, summary.Skipped)))
	}
	return nil
}
