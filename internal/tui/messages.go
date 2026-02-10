package tui

import "skli/internal/gitrepo"

// Mensajes de navegación entre pantallas
type NavigateToInputRemoteMsg struct{}
type NavigateToScanningMsg struct {
	URL        string
	SkillsRoot string
}
type NavigateToSkillsMsg struct {
	Skills     []gitrepo.SkillInfo
	TempDir    string
	RemoteURL  string
	SkillsRoot string
	CommitHash string
}
type NavigateToEditorMsg struct {
	Skills     []Skill
	TempDir    string
	RemoteURL  string
	SkillsRoot string
	CommitHash string
}
type NavigateToConfigMsg struct{}
type NavigateToManageRemotesMsg struct{}
type NavigateToProgressMsg struct {
	TempDir         string
	RemoteURL       string
	SkillsRoot      string
	ConfigLocalPath string
	CommitHash      string
	Selected        []gitrepo.SkillInfo
}
type NavigateToDoneMsg struct {
	ConfigMode bool
	LocalPath  string
}
type NavigateToErrorMsg struct {
	Err error
}
type NavigateToManageMsg struct{}
type QuitMsg struct{}

// Mensajes de configuración
type ConfigSavedMsg struct{}
type RemotesUpdatedMsg struct {
	Remotes []string
}

// Mensajes de operaciones async
type ScanResultMsg struct {
	Result    gitrepo.ScanResult
	RemoteURL string
	Err       error
}

type DownloadResultMsg struct {
	Err error
}
