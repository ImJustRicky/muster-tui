package auth

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type TUIConfig struct {
	Token string `json:"token"`
}

// TokenConfigPath returns the path to ~/.muster-tui/config.json.
func TokenConfigPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".muster-tui", "config.json")
}

// LoadToken checks MUSTER_TOKEN env var first, then reads from config file.
func LoadToken() (string, error) {
	if token := os.Getenv("MUSTER_TOKEN"); token != "" {
		return token, nil
	}

	data, err := os.ReadFile(TokenConfigPath())
	if err != nil {
		return "", err
	}

	var cfg TUIConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return "", err
	}

	return cfg.Token, nil
}

// SaveToken writes the token to ~/.muster-tui/config.json with restricted permissions.
func SaveToken(token string) error {
	path := TokenConfigPath()

	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}

	data, err := json.Marshal(TUIConfig{Token: token})
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600)
}
