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
		Usage:     "skill manager",
		UsageText: "skli [command] [arguments]",
		CommandNotFound: func(ctx context.Context, c *cli.Command, s string) {
			fmt.Println(errorStyle.Render(fmt.Sprintf("✘ unknown command: %s", s)))
			fmt.Println("Use --help to see available commands.")
		},
		Commands: []*cli.Command{
			{
				Name:      "add",
				Usage:     "install skills from a repo or open the TUI selector",
				ArgsUsage: "[git-repo-path]",
				Action: func(_ context.Context, cmd *cli.Command) error {
					if cmd.NArg() > 1 {
						return cli.Exit("usage: skli add [git-repo-path]", 1)
					}
					return service.Add(cmd.Args().First())
				},
			},
			{
				Name:      "rm",
				Usage:     "remove installed skills",
				ArgsUsage: "[skill-name]",
				Action: func(_ context.Context, cmd *cli.Command) error {
					if cmd.NArg() > 1 {
						return cli.Exit("usage: skli rm [skill-name]", 1)
					}
					if cmd.NArg() == 1 {
						skill, err := service.RemoveByName(cmd.Args().First())
						if err != nil {
							return err
						}
						fmt.Println(successStyle.Render(fmt.Sprintf("✔ skill removed: %s (%s)", skill.Name, skill.Path)))
						return nil
					}
					return service.RemoveTUI()
				},
			},
			{
				Name:  "sync",
				Usage: "sync installed skills with their source repo",
				Action: func(_ context.Context, cmd *cli.Command) error {
					if cmd.NArg() != 0 {
						return cli.Exit("usage: skli sync", 1)
					}
					return renderSync(service)
				},
			},
			{
				Name:  "list",
				Usage: "list installed skills and local skills found in ./skills",
				Action: func(_ context.Context, cmd *cli.Command) error {
					if cmd.NArg() != 0 {
						return cli.Exit("usage: skli list", 1)
					}
					return service.ListTUI()
				},
			},
			{
				Name:      "upload",
				Usage:     "upload local skills to a target repo",
				ArgsUsage: "[git-dest-repo-path] [local-skill-path]",
				Action: func(_ context.Context, cmd *cli.Command) error {
					if cmd.NArg() != 0 && cmd.NArg() != 2 {
						return cli.Exit("usage: skli upload [git-dest-repo-path] [local-skill-path]", 1)
					}
					if cmd.NArg() == 2 {
						result, err := service.UploadDirect(cmd.Args().Get(0), cmd.Args().Get(1))
						if err != nil {
							return err
						}
						fmt.Println(infoStyle.Render(fmt.Sprintf("🔄 Uploading %s to %s...", result.Skill.Name, cmd.Args().Get(0))))
						fmt.Println(successStyle.Render("✔ PR created"))
						fmt.Println(dimStyle.Render(result.PRURL))
						return nil
					}
					return service.UploadTUI()
				},
			},
			{
				Name:  "config",
				Usage: "open skli configuration",
				Action: func(_ context.Context, cmd *cli.Command) error {
					if cmd.NArg() != 0 {
						return cli.Exit("usage: skli config", 1)
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
			return cli.Exit("unknown command. Use --help to see available commands.", 1)
		},
	}
}

func renderSync(service app.Service) error {
	fmt.Println(infoStyle.Render("🔄 Syncing skills..."))
	fmt.Println()

	summary, err := service.SyncAll()
	if err != nil {
		return err
	}

	if len(summary.Results) == 0 {
		fmt.Println(infoStyle.Render("ℹ There are no installed skills to sync (skli.lock is empty)."))
		fmt.Println(infoStyle.Render("  Use 'skli add' to install skills first."))
		return nil
	}

	for _, r := range summary.Results {
		if r.Error != nil {
			fmt.Println(errorStyle.Render(fmt.Sprintf("  ✘ %s: %v", r.SkillName, r.Error)))
		} else if r.Updated {
			fmt.Println(successStyle.Render(fmt.Sprintf("  ✔ %s updated", r.SkillName)))
		} else if r.Skipped {
			fmt.Println(dimStyle.Render(fmt.Sprintf("  ○ %s unchanged", r.SkillName)))
		}
	}

	fmt.Println()
	if summary.Errors > 0 {
		fmt.Println(errorStyle.Render(fmt.Sprintf("Completed with %d errors. %d updated, %d unchanged.", summary.Errors, summary.Updated, summary.Skipped)))
	} else if summary.Updated == 0 {
		fmt.Println(successStyle.Render(fmt.Sprintf("✔ All skills are up to date (%d checked).", summary.Skipped)))
	} else {
		fmt.Println(successStyle.Render(fmt.Sprintf("✔ %d skills updated, %d unchanged.", summary.Updated, summary.Skipped)))
	}
	return nil
}
