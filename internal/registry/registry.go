package registry

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Project struct {
	Path         string `json:"path"`
	Name         string `json:"name"`
	ServiceCount int    `json:"service_count"`
	LastAccessed string `json:"last_accessed"`
}

type ProjectsFile struct {
	Projects []Project `json:"projects"`
}

func projectsFilePath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".muster", "projects.json")
}

// Load reads the projects registry from ~/.muster/projects.json.
func Load() (*ProjectsFile, error) {
	path := projectsFilePath()

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &ProjectsFile{}, nil
		}
		return nil, err
	}

	var pf ProjectsFile
	if err := json.Unmarshal(data, &pf); err != nil {
		return nil, err
	}

	return &pf, nil
}

// Prune removes entries where neither muster.json nor deploy.json exists.
// Returns the number of entries removed.
func Prune() (int, error) {
	pf, err := Load()
	if err != nil {
		return 0, err
	}

	var kept []Project
	removed := 0

	for _, p := range pf.Projects {
		musterPath := filepath.Join(p.Path, "muster.json")
		deployPath := filepath.Join(p.Path, "deploy.json")

		if fileExists(musterPath) || fileExists(deployPath) {
			kept = append(kept, p)
		} else {
			removed++
		}
	}

	if removed > 0 {
		pf.Projects = kept
		if err := Save(pf); err != nil {
			return 0, err
		}
	}

	return removed, nil
}

// Save atomically writes the projects file.
func Save(pf *ProjectsFile) error {
	path := projectsFilePath()

	data, err := json.MarshalIndent(pf, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')

	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0600); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// HasConfig checks if a directory contains muster.json or deploy.json.
func HasConfig(dir string) bool {
	return fileExists(filepath.Join(dir, "muster.json")) ||
		fileExists(filepath.Join(dir, "deploy.json"))
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
