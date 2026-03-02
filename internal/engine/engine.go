package engine

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

type Engine struct {
	BinaryPath string
	Token      string
}

// NewEngine finds muster on PATH and validates it with --version.
func NewEngine(token string) (*Engine, error) {
	path, err := exec.LookPath("muster")
	if err != nil {
		return nil, fmt.Errorf("muster not found on PATH: %w", err)
	}

	cmd := exec.Command(path, "--version")
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("muster --version failed: %w", err)
	}

	return &Engine{BinaryPath: path, Token: token}, nil
}

func (e *Engine) run(args ...string) *exec.Cmd {
	cmd := exec.Command(e.BinaryPath, args...)
	cmd.Env = append(cmd.Environ(), "MUSTER_TOKEN="+e.Token)
	return cmd
}

// runInDir runs a muster command with the working directory set to dir.
func (e *Engine) runInDir(dir string, args ...string) *exec.Cmd {
	cmd := e.run(args...)
	cmd.Dir = dir
	return cmd
}

// StatusInDir runs muster status --json in the given project directory.
func (e *Engine) StatusInDir(dir string) (*StatusResult, error) {
	cmd := e.runInDir(dir, "status", "--json")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("muster status failed: %w", err)
	}

	var result StatusResult
	if err := json.Unmarshal(out, &result); err != nil {
		return nil, fmt.Errorf("parsing status output: %w", err)
	}

	return &result, nil
}

// DeployInDir runs muster deploy --json in the given project directory.
func (e *Engine) DeployInDir(dir string, services []string, dryRun bool) (<-chan DeployEvent, error) {
	args := []string{"deploy", "--json"}
	if dryRun {
		args = append(args, "--dry-run")
	}
	args = append(args, services...)

	cmd := e.runInDir(dir, args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("creating stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("starting deploy: %w", err)
	}

	ch := make(chan DeployEvent)
	go func() {
		defer close(ch)
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			var ev DeployEvent
			if err := json.Unmarshal(scanner.Bytes(), &ev); err != nil {
				continue
			}
			ch <- ev
		}
		cmd.Wait()
	}()

	return ch, nil
}

type ServiceStatus struct {
	Name       string `json:"name"`
	Status     string `json:"status"`
	HealthType string `json:"health_type"`
	Detail     string `json:"detail"`
}

type StatusResult struct {
	Project  string                   `json:"project"`
	Services map[string]ServiceStatus `json:"services"`
}

// Status runs muster status --json and returns parsed results.
func (e *Engine) Status() (*StatusResult, error) {
	cmd := e.run("status", "--json")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("muster status failed: %w", err)
	}

	var result StatusResult
	if err := json.Unmarshal(out, &result); err != nil {
		return nil, fmt.Errorf("parsing status output: %w", err)
	}

	return &result, nil
}

type DeployEvent struct {
	Event    string `json:"event"`
	Service  string `json:"service"`
	Name     string `json:"name"`
	Status   string `json:"status"`
	Line     string `json:"line"`
	LogFile  string `json:"log_file"`
	Hook     string `json:"hook"`
	Index    int    `json:"index"`
	Total    int    `json:"total"`
	ExitCode int    `json:"exit_code"`
	HookLines int   `json:"hook_lines"`
	DryRun   bool   `json:"dry_run"`
}

// Deploy runs muster deploy --json and streams NDJSON events on the returned channel.
func (e *Engine) Deploy(services []string, dryRun bool) (<-chan DeployEvent, error) {
	args := []string{"deploy", "--json"}
	if dryRun {
		args = append(args, "--dry-run")
	}
	args = append(args, services...)

	cmd := e.run(args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("creating stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("starting deploy: %w", err)
	}

	ch := make(chan DeployEvent)
	go func() {
		defer close(ch)
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			var ev DeployEvent
			if err := json.Unmarshal(scanner.Bytes(), &ev); err != nil {
				continue
			}
			ch <- ev
		}
		cmd.Wait()
	}()

	return ch, nil
}

type HistoryEvent struct {
	Timestamp string `json:"timestamp"`
	Service   string `json:"service"`
	Action    string `json:"action"`
	Status    string `json:"status"`
	Commit    string `json:"commit"`
}

// History runs muster history --json and returns parsed events.
func (e *Engine) History(all bool, service string) ([]HistoryEvent, error) {
	args := []string{"history", "--json"}
	if all {
		args = append(args, "--all")
	}
	if service != "" {
		args = append(args, service)
	}

	cmd := e.run(args...)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("muster history failed: %w", err)
	}

	var events []HistoryEvent
	if err := json.Unmarshal(out, &events); err != nil {
		return nil, fmt.Errorf("parsing history output: %w", err)
	}

	return events, nil
}

type DoctorCheck struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Detail string `json:"detail"`
}

type DoctorResult struct {
	Pass     int           `json:"pass"`
	Warnings int           `json:"warnings"`
	Failures int           `json:"failures"`
	Checks   []DoctorCheck `json:"checks"`
}

// Doctor runs muster doctor --json and returns parsed results.
func (e *Engine) Doctor(fix bool) (*DoctorResult, error) {
	args := []string{"doctor", "--json"}
	if fix {
		args = append(args, "--fix")
	}

	cmd := e.run(args...)
	out, err := cmd.Output()
	if err != nil {
		// Doctor may exit non-zero on failures, still try to parse
		if len(out) == 0 {
			return nil, fmt.Errorf("muster doctor failed: %w", err)
		}
	}

	var result DoctorResult
	if err := json.Unmarshal(out, &result); err != nil {
		return nil, fmt.Errorf("parsing doctor output: %w", err)
	}

	return &result, nil
}

// RunRaw starts a muster command and streams stdout/stderr lines on a channel.
// Used for commands without --json (rollback, logs, cleanup).
func (e *Engine) RunRaw(args []string) (<-chan string, error) {
	cmd := e.run(args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("creating stdout pipe: %w", err)
	}
	cmd.Stderr = cmd.Stdout // merge stderr into stdout

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("starting command: %w", err)
	}

	ch := make(chan string)
	go func() {
		defer close(ch)
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			ch <- scanner.Text()
		}
		cmd.Wait()
	}()

	return ch, nil
}

// Version returns the muster CLI version string.
func (e *Engine) Version() (string, error) {
	cmd := e.run("--version")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}
