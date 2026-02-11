package app

import (
	tea "github.com/charmbracelet/bubbletea"

	"skli/internal/config"
	"skli/internal/db"
	"skli/internal/gitrepo"
	"skli/internal/skills"
	sklisync "skli/internal/sync"
	"skli/internal/tui"
	"skli/internal/tui/screens/manage"
)

// Service encapsula casos de uso de la app.
type Service struct {
	cfg config.Config
}

func NewService(cfg config.Config) Service {
	return Service{cfg: cfg}
}

func (s Service) Add(initialURL string) error {
	return s.runTUI(initialURL, false, manage.ModeNone)
}

func (s Service) RemoveTUI() error {
	return s.runTUI("", false, manage.ModeRemove)
}

func (s Service) UploadTUI() error {
	return s.runTUI("", false, manage.ModeUpload)
}

func (s Service) ConfigTUI() error {
	return s.runTUI("", true, manage.ModeNone)
}

func (s Service) RemoveByName(name string) (db.InstalledSkill, error) {
	return skills.DeleteByName(name, skills.DefaultRoot)
}

type UploadResult struct {
	Skill db.InstalledSkill
	PRURL string
}

func (s Service) UploadDirect(targetRepo, localSkillPath string) (UploadResult, error) {
	skill, err := skills.PrepareLocalForUpload(localSkillPath)
	if err != nil {
		return UploadResult{}, err
	}

	prURL, err := gitrepo.UploadSkill(skill, targetRepo)
	if err != nil {
		return UploadResult{}, err
	}

	return UploadResult{
		Skill: skill,
		PRURL: prURL,
	}, nil
}

type SyncSummary struct {
	Results []sklisync.SyncResult
	Updated int
	Skipped int
	Errors  int
}

func (s Service) SyncAll() (SyncSummary, error) {
	results, err := sklisync.SyncAllSkills()
	if err != nil {
		return SyncSummary{}, err
	}

	summary := SyncSummary{Results: results}
	for _, r := range results {
		if r.Error != nil {
			summary.Errors++
			continue
		}
		if r.Updated {
			summary.Updated++
			continue
		}
		if r.Skipped {
			summary.Skipped++
		}
	}

	return summary, nil
}

func (s Service) runTUI(initialURL string, configMode bool, manageMode manage.Mode) error {
	p := tea.NewProgram(
		tui.NewRootModel(initialURL, skills.DefaultRoot, s.cfg.LocalPath, configMode, manageMode, s.cfg.Remotes),
		tea.WithAltScreen(),
	)
	_, err := p.Run()
	return err
}
