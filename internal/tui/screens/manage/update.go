package manage

import (
	"fmt"
	"strings"

	"skli/internal/db"
	"skli/internal/tui/screens/manage/commands"
	"skli/internal/tui/screens/manage/delegates"
	"skli/internal/tui/shared"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

func (s ManageScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch s.State {
	case StateList:
		return s.updateList(msg)
	case StateConfirm:
		return s.updateConfirm(msg)
	case StateSelectingRemote:
		return s.updateSelectingRemote(msg)
	case StateInputRemote:
		return s.updateInputRemote(msg)
	case StateUploading:
		return s.updateUploading(msg)
	}
	return s, nil
}

func (s ManageScreen) updateList(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.List.SetSize(msg.Width, msg.Height-4)
		return s, nil

	case commands.DeleteSkillsMsg:
		if msg.Err != nil {
			s.Msg = fmt.Sprintf("Error: %v", msg.Err)
			return s, nil
		}
		s.Msg = fmt.Sprintf("Removed %d skill(s)", len(msg.Deleted))
		refreshed, _ := NewManageScreen(s.ConfigRemotes, s.Mode)
		refreshed.Msg = s.Msg
		return refreshed, nil

	case tea.KeyMsg:
		switch s.Mode {
		case ModeList:
			switch msg.String() {
			case "esc", "q":
				return s, tea.Quit
			}

		case ModeRemove:
			switch msg.String() {
			case " ":
				s = s.toggleSelectedCurrent()
				return s, nil
			case "enter":
				selected := s.selectedSkills()
				if len(selected) == 0 {
					s.Msg = "Select at least one skill to remove"
					return s, nil
				}
				return s, commands.DeleteSkillsCmd(selected)
			case "esc", "q":
				return s, tea.Quit
			}

		case ModeUpload:
			switch msg.String() {
			case " ":
				s = s.toggleSelectedCurrent()
				return s, nil
			case "enter":
				selected := s.selectedSkills()
				if len(selected) == 0 {
					s.Msg = "Select at least one skill to upload"
					return s, nil
				}
				s.State = StateUploading
				s.Msg = fmt.Sprintf("Uploading %d skill(s) to %s...", len(selected), s.TargetRemote)
				return s, commands.UploadSkillsCmd(selected, s.TargetRemote)
			case "esc":
				s.State = StateSelectingRemote
				return s, nil
			case "q":
				return s, tea.Quit
			}

		default:
			switch msg.String() {
			case "enter", "d", "backspace":
				if item, ok := s.List.SelectedItem().(InstalledSkillItem); ok && item.Skill != nil {
					sk := item.Skill.Skill
					s.ToDelete = &sk
					s.State = StateConfirm
					s.ConfirmCursor = 1
					return s, nil
				}
			case "esc", "q":
				return s, tea.Quit
			case "u":
				if item, ok := s.List.SelectedItem().(InstalledSkillItem); ok && item.Skill != nil {
					s.SelectedSkill = &item.Skill.Skill
					var items []list.Item
					seen := make(map[string]bool)

					if item.Skill.Skill.RemoteRepo != "" {
						items = append(items, remoteItem{
							url:         item.Skill.Skill.RemoteRepo,
							displayName: fmt.Sprintf("%s (Origin)", item.Skill.Skill.RemoteRepo),
						})
						seen[item.Skill.Skill.RemoteRepo] = true
					}

					for _, r := range s.ConfigRemotes {
						if !seen[r] {
							items = append(items, remoteItem{url: r})
							seen[r] = true
						}
					}

					items = append(items, customURLItem{})

					delegate := delegates.NewRemoteDelegate()
					l := list.New(items, delegate, 60, 14)
					l.Title = "Select target repository"
					l.SetShowStatusBar(false)
					l.SetFilteringEnabled(false)
					l.Styles.Title = shared.TitleStyle
					s.RemoteList = l

					s.State = StateSelectingRemote
					return s, nil
				}
			}
		}
	}

	var cmd tea.Cmd
	s.List, cmd = s.List.Update(msg)
	return s, cmd
}

func (s ManageScreen) updateSelectingRemote(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.RemoteList.SetSize(msg.Width, msg.Height-4)
		return s, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			if s.Mode == ModeUpload {
				return s, tea.Quit
			}
			s.State = StateList
			return s, nil
		case "enter":
			selected := s.RemoteList.SelectedItem()
			if selected == nil {
				return s, nil
			}
			switch item := selected.(type) {
			case remoteItem:
				if s.Mode == ModeUpload {
					s.TargetRemote = item.url
					s.State = StateList
					s.Msg = fmt.Sprintf("Target selected: %s", item.url)
					return s, nil
				}

				s.State = StateUploading
				s.Msg = fmt.Sprintf("Starting upload of '%s' to %s...", s.SelectedSkill.Name, item.url)
				return s, commands.UploadSkillsCmd([]db.InstalledSkill{*s.SelectedSkill}, item.url)
			case customURLItem:
				s.State = StateInputRemote
				s.RemoteInput.Focus()
				s.RemoteInput.SetValue("")
				return s, textinput.Blink
			}
		}
	}
	var cmd tea.Cmd
	s.RemoteList, cmd = s.RemoteList.Update(msg)
	return s, cmd
}

func (s ManageScreen) updateInputRemote(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			if s.Mode == ModeUpload {
				s.State = StateSelectingRemote
				return s, nil
			}
			s.State = StateSelectingRemote
			return s, nil
		case "enter":
			url := strings.TrimSpace(s.RemoteInput.Value())
			if url != "" {
				if s.Mode == ModeUpload {
					s.TargetRemote = url
					s.State = StateList
					s.Msg = fmt.Sprintf("Target selected: %s", url)
					return s, nil
				}
				s.State = StateUploading
				s.Msg = fmt.Sprintf("Starting upload of '%s' to %s...", s.SelectedSkill.Name, url)
				return s, commands.UploadSkillsCmd([]db.InstalledSkill{*s.SelectedSkill}, url)
			}
		}
	}
	var cmd tea.Cmd
	s.RemoteInput, cmd = s.RemoteInput.Update(msg)
	return s, cmd
}

func (s ManageScreen) updateConfirm(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "left", "h", "right", "l", "tab":
			if s.ConfirmCursor == 0 {
				s.ConfirmCursor = 1
			} else {
				s.ConfirmCursor = 0
			}
		case "y", "Y":
			s.ConfirmCursor = 0
			if s.ToDelete != nil {
				return s, commands.DeleteSkillsCmd([]db.InstalledSkill{*s.ToDelete})
			}
		case "n", "N", "esc":
			s.State = StateList
			s.ToDelete = nil
			return s, nil
		case "enter":
			if s.ConfirmCursor == 0 {
				if s.ToDelete != nil {
					return s, commands.DeleteSkillsCmd([]db.InstalledSkill{*s.ToDelete})
				}
			} else {
				s.State = StateList
				s.ToDelete = nil
				return s, nil
			}
		}
	case commands.DeleteSkillsMsg:
		if msg.Err != nil {
			s.State = StateList
			s.ToDelete = nil
			s.Msg = fmt.Sprintf("Error: %v", msg.Err)
			return s, nil
		}
		refreshed, _ := NewManageScreen(s.ConfigRemotes, s.Mode)
		refreshed.Msg = fmt.Sprintf("Removed: %s", strings.Join(msg.Deleted, ", "))
		return refreshed, nil
	}
	return s, nil
}

func (s ManageScreen) updateUploading(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case commands.UploadSkillsMsg:
		var lines []string
		okCount := 0
		for _, result := range msg.Results {
			if result.Err != nil {
				lines = append(lines, fmt.Sprintf("✘ %s: %v", result.SkillName, result.Err))
				continue
			}
			okCount++
			lines = append(lines, fmt.Sprintf("✔ %s: %s", result.SkillName, result.PRURL))
		}
		s.Msg = strings.Join(lines, "\n")
		if s.Mode == ModeUpload {
			refreshed, _ := NewManageScreen(s.ConfigRemotes, s.Mode)
			refreshed.TargetRemote = s.TargetRemote
			refreshed.State = StateList
			refreshed.Msg = s.Msg
			if okCount > 0 {
				refreshed.Msg = fmt.Sprintf("Uploaded %d skill(s)\n%s", okCount, s.Msg)
			}
			return refreshed, nil
		}
		return s, nil
	case tea.KeyMsg:
		if msg.String() == "esc" || msg.String() == "enter" {
			s.State = StateList
			if s.Mode == ModeUpload {
				return s, tea.Quit
			}
			return s, nil
		}
	}
	return s, nil
}
