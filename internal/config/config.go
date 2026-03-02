package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type HealthConfig struct {
	Enabled  *bool  `json:"enabled"`
	Type     string `json:"type"`
	Endpoint string `json:"endpoint"`
	Timeout  int    `json:"timeout"`
}

type K8sConfig struct {
	Deployment string `json:"deployment"`
	Namespace  string `json:"namespace"`
}

type RemoteConfig struct {
	Enabled      bool   `json:"enabled"`
	Host         string `json:"host"`
	User         string `json:"user"`
	Port         int    `json:"port"`
	IdentityFile string `json:"identity_file"`
	ProjectDir   string `json:"project_dir"`
}

type ServiceConfig struct {
	Name       string        `json:"name"`
	Health     HealthConfig  `json:"health"`
	K8s        *K8sConfig    `json:"k8s,omitempty"`
	Remote     *RemoteConfig `json:"remote,omitempty"`
	SkipDeploy bool          `json:"skip_deploy"`
}

type DeployConfig struct {
	Project     string                   `json:"project"`
	Services    map[string]ServiceConfig `json:"services"`
	DeployOrder []string                 `json:"deploy_order"`
}

type GlobalSettings struct {
	ColorMode            string `json:"color_mode"`
	LogRetentionDays     int    `json:"log_retention_days"`
	DefaultStack         string `json:"default_stack"`
	DefaultHealthTimeout int    `json:"default_health_timeout"`
	UpdateCheck          string `json:"update_check"`
}

// FindConfig walks up from the current working directory looking for deploy.json.
func FindConfig() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		path := filepath.Join(dir, "deploy.json")
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", os.ErrNotExist
		}
		dir = parent
	}
}

// LoadDeploy parses a deploy.json file into a DeployConfig.
func LoadDeploy(path string) (*DeployConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg DeployConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	// Default Health.Enabled to true if not explicitly set
	for key, svc := range cfg.Services {
		if svc.Health.Enabled == nil {
			enabled := true
			svc.Health.Enabled = &enabled
			cfg.Services[key] = svc
		}
	}

	return &cfg, nil
}

func defaultGlobalSettings() *GlobalSettings {
	return &GlobalSettings{
		ColorMode:            "auto",
		LogRetentionDays:     7,
		DefaultStack:         "bare",
		DefaultHealthTimeout: 10,
		UpdateCheck:          "on",
	}
}

// LoadGlobal parses ~/.muster/settings.json. Returns defaults if the file is missing.
func LoadGlobal() (*GlobalSettings, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return defaultGlobalSettings(), nil
	}

	path := filepath.Join(home, ".muster", "settings.json")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return defaultGlobalSettings(), nil
		}
		return nil, err
	}

	settings := defaultGlobalSettings()
	if err := json.Unmarshal(data, settings); err != nil {
		return nil, err
	}

	return settings, nil
}
