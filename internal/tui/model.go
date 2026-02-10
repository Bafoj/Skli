package tui

import "skli/internal/gitrepo"

// Editor representa un editor soportado
type Editor struct {
	Name string
	Path string
}

// Editors lista de editores soportados
var Editors = []Editor{
	{Name: "Windsurf", Path: ".windsurf/skills"},
	{Name: "Antigravity", Path: ".antigravity/skills"},
	{Name: "Cursor", Path: ".cursor/skills"},
	{Name: "VSCode", Path: ".vscode/skills"},
	{Name: "OpenCode", Path: ".opencode/skills"},
	{Name: "Custom", Path: ""},
}

// Skill representa una habilidad encontrada en el repositorio
type Skill struct {
	Info     gitrepo.SkillInfo
	Selected bool
}
