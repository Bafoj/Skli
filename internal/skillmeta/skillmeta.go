package skillmeta

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// Metadata contiene campos relevantes del frontmatter de SKILL.md.
type Metadata struct {
	Name        string
	Description string
}

// ParseFile lee metadata desde un SKILL.md.
func ParseFile(skillFile string, maxLines int) (Metadata, error) {
	if maxLines <= 0 {
		maxLines = 40
	}

	f, err := os.Open(skillFile)
	if err != nil {
		return Metadata{}, err
	}
	defer f.Close()

	meta := Metadata{}
	inFrontmatter := false
	lineCount := 0

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lineCount++
		line := strings.TrimSpace(scanner.Text())

		if line == "---" {
			if !inFrontmatter {
				inFrontmatter = true
				continue
			}
			break
		}

		if inFrontmatter {
			if strings.HasPrefix(line, "name:") {
				meta.Name = strings.TrimSpace(strings.TrimPrefix(line, "name:"))
			}
			if strings.HasPrefix(line, "description:") {
				meta.Description = strings.TrimSpace(strings.TrimPrefix(line, "description:"))
			}
		}

		if lineCount >= maxLines {
			break
		}
	}

	if err := scanner.Err(); err != nil {
		return Metadata{}, err
	}

	return meta, nil
}

// ParseDir lee metadata desde <skillDir>/SKILL.md y usa el nombre de carpeta como fallback.
func ParseDir(skillDir string, maxLines int) (Metadata, error) {
	meta, err := ParseFile(filepath.Join(skillDir, "SKILL.md"), maxLines)
	if err != nil {
		return Metadata{}, err
	}
	if meta.Name == "" {
		meta.Name = filepath.Base(skillDir)
	}
	return meta, nil
}
