package config

// ConfigScreen es el modelo para la pantalla de configuración
type ConfigScreen struct {
	Cursor          int
	ConfigLocalPath string
	Remotes         []string
}

// NewConfigScreen crea una nueva pantalla de configuración
func NewConfigScreen(configLocalPath string, remotes []string) ConfigScreen {
	return ConfigScreen{
		ConfigLocalPath: configLocalPath,
		Remotes:         remotes,
	}
}

// UpdateConfigPath actualiza el path de la configuración
func (s *ConfigScreen) UpdateConfigPath(path string) {
	s.ConfigLocalPath = path
}

// UpdateRemotes actualiza la lista de remotes
func (s *ConfigScreen) UpdateRemotes(remotes []string) {
	s.Remotes = remotes
}
