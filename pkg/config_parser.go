package pkg

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type CiConfig struct {
	basePath string       `yaml:"-"`
	Steps    []StepConfig `yaml:"steps"`
}

type StepConfig struct {
	Name      string      `yaml:"name"`
	Jobs      []JobConfig `yaml:"jobs"`
	RunBefore []CmdConfig `yaml:"run_before,omitempty"`
	RunAfter  []CmdConfig `yaml:"run_after,omitempty"`
}

type JobConfig struct {
	Name       string      `yaml:"name"`
	Cmd        string      `yaml:"cmd"`
	Path       string      `yaml:"path,omitempty"`
	RunBefore  []CmdConfig `yaml:"run_before,omitempty"`
	Background []CmdConfig `yaml:"background,omitempty"`
	RunAfter   []CmdConfig `yaml:"run_after,omitempty"`
}

type CmdConfig struct {
	Cmd  string `yaml:"cmd"`
	Path string `yaml:"path,omitempty"`
	// TODO: add a FailedWhen: a template calculated with the exit code, the stdout and stderr
}

// findConfig tries to find a configuration file. If no path is given in argument, it tries to find a localci.yml file in the parent directories of the current working directory.
func findConfig(givenPath string) (ret string, err error) {
	if len(givenPath) > 0 {
		if filepath.IsAbs(givenPath) {
			ret = givenPath
		} else {
			ret, err = filepath.Abs(givenPath)
			if err != nil {
				return "", fmt.Errorf("an error occured calculating an absolute path: %w", err)
			}
		}

		stat, err := os.Stat(ret)
		if err != nil {
			return "", fmt.Errorf("an error occured when checking the path \"%s\": %w", ret, err)
		}
		if stat.IsDir() {
			return "", fmt.Errorf("the path \"%s\" is a directory", ret)
		}

		return ret, nil
	}

	curDir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get the current working directory: %w", err)
	}
	for {
		stat, err := os.Stat(filepath.Join(curDir, "localci.yml"))
		if err == nil && !stat.IsDir() {
			return filepath.Join(curDir, "localci.yml"), nil
		}

		if curDir == filepath.Dir(curDir) {
			// We have reached the root directory
			break
		}

		// Go to the parent directory
		curDir = filepath.Dir(curDir)
	}

	return "", errors.New("no configuration file could be found")
}

func validateCommands(configs []CmdConfig) error {
	for cmdIdx, cmd := range configs {
		if len(cmd.Cmd) == 0 {
			return fmt.Errorf("the task #%d has no command declared", cmdIdx)
		}
	}

	return nil
}

func validateConfig(cfg *CiConfig) error {
	if len(cfg.Steps) == 0 {
		return errors.New("no step is declared")
	}
	// TODO: check that names are unique
	for stepIdx, step := range cfg.Steps {
		if len(step.Name) == 0 {
			return fmt.Errorf("the step #%d has no name declared", stepIdx)
		}
		if len(step.Jobs) == 0 {
			return fmt.Errorf("the step \"%s\" has no job declared", step.Name)
		}
		if err := validateCommands(step.RunBefore); err != nil {
			return fmt.Errorf("the step \"%s\" has invalid run_before hooks: %w", step.Name, err)
		}
		if err := validateCommands(step.RunAfter); err != nil {
			return fmt.Errorf("the step \"%s\" has invalid run_after hooks: %w", step.Name, err)
		}
		for jobIdx, job := range step.Jobs {
			if len(job.Name) == 0 {
				return fmt.Errorf("the job #%d in the step \"%s\" has no name declared", jobIdx, step.Name)
			}
			if err := validateCommands(job.RunBefore); err != nil {
				return fmt.Errorf("the job \"%s:%s\" has invalid run_before hooks: %w", step.Name, job.Name, err)
			}
			if err := validateCommands(job.RunAfter); err != nil {
				return fmt.Errorf("the job \"%s:%s\" has invalid run_after hooks: %w", step.Name, job.Name, err)
			}
			if err := validateCommands(job.Background); err != nil {
				return fmt.Errorf("the job \"%s:%s\" has invalid background tasks: %w", step.Name, job.Name, err)
			}
			if len(job.Cmd) == 0 {
				return fmt.Errorf("the job \"%s:%s\" has no command declared", step.Name, job.Name)
			}
		}
	}

	return nil
}

// parseConfig reads the content of the file at the absolute path given in argument, and extracts its yaml content into a CiConfig.
func parseConfig(path string) (*CiConfig, error) {
	fileContent, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("the contents of the file \"%s\" could not be read: %w", path, err)
	}

	output := CiConfig{}
	if err := yaml.Unmarshal(fileContent, &output); err != nil {
		return nil, fmt.Errorf("an error occured when reading the configuration file: %w", err)
	}

	output.basePath = filepath.Dir(path)

	if err := validateConfig(&output); err != nil {
		return nil, fmt.Errorf("the configuration is invalid: %w", err)
	}

	return &output, nil
}

func findAndParseConfig(givenPath string) (*CiConfig, error) {
	configPath, err := findConfig(givenPath)
	if err != nil {
		return nil, err
	}

	config, err := parseConfig(configPath)
	if err != nil {
		return nil, err
	}
	return config, nil
}
