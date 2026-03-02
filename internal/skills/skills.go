package skills

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Skill represents an installed skill from skill.json.
type Skill struct {
	Name        string   `json:"name"`
	Version     string   `json:"version"`
	Description string   `json:"description"`
	Author      string   `json:"author"`
	Hooks       []string `json:"hooks"`
	Enabled     bool     // derived from .enabled file presence
	Scope       string   // "project" or "global"
}

// ScanDir reads all skills from a directory and returns them.
func ScanDir(dir, scope string) []Skill {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}

	var skills []Skill
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		skillDir := filepath.Join(dir, entry.Name())
		manifestPath := filepath.Join(skillDir, "skill.json")

		data, err := os.ReadFile(manifestPath)
		if err != nil {
			continue
		}

		var s Skill
		if err := json.Unmarshal(data, &s); err != nil {
			continue
		}

		// Check .enabled file
		if _, err := os.Stat(filepath.Join(skillDir, ".enabled")); err == nil {
			s.Enabled = true
		}

		s.Scope = scope
		if s.Name == "" {
			s.Name = entry.Name()
		}

		skills = append(skills, s)
	}
	return skills
}

// ScanAll returns skills from both project and global dirs.
// Project skills take priority (global dupes skipped).
func ScanAll(projectDir string) []Skill {
	var all []Skill
	seen := map[string]bool{}

	// Project skills first
	if projectDir != "" {
		projectSkillsDir := filepath.Join(projectDir, ".muster", "skills")
		for _, s := range ScanDir(projectSkillsDir, "project") {
			all = append(all, s)
			seen[s.Name] = true
		}
	}

	// Global skills (skip duplicates)
	home, err := os.UserHomeDir()
	if err == nil {
		globalSkillsDir := filepath.Join(home, ".muster", "skills")
		for _, s := range ScanDir(globalSkillsDir, "global") {
			if !seen[s.Name] {
				all = append(all, s)
			}
		}
	}

	return all
}
