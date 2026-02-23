package shared

import "skli/internal/gitrepo"

// Mensajes de navegación
type NavigateToScanningMsg struct{ URL string }
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
type NavigateToProgressMsg struct {
	TempDir         string
	RemoteURL       string
	SkillsRoot      string
	ConfigLocalPath string
	CommitHash      string
	Selected        []gitrepo.SkillInfo
}
type NavigateToConfigMsg struct{}
type NavigateToManageRemotesMsg struct{}
type NavigateToInputRemoteMsg struct{}
type NavigateToDoneMsg struct {
	ConfigMode bool
	LocalPath  string
}
type NavigateToErrorMsg struct{ Err error }
type NavigateToManageMsg struct{}
type QuitMsg struct{}

// Mensajes de estado
type RemotesUpdatedMsg struct {
	Remotes []string
}

type ConfigSavedMsg struct {
	LocalPath    string
	Remotes      []string
	NavigateBack bool // If true, navigate to config screen after saving
}

// Mensajes de resultado de operaciones
type ScanResultMsg struct {
	Result    gitrepo.ScanResult
	RemoteURL string
	Err       error
}

type DownloadResultMsg struct {
	Err error
}
