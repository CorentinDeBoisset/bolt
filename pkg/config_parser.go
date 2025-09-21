package pkg

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type ConfigFile struct {
	basePath    string      `yaml:"-"`
	LogFilePath string      `yaml:"log_file"`
	Jobs        []JobConfig `yaml:"jobs"`
}

type JobConfig struct {
	Name  string       `yaml:"name"`
	Steps []StepConfig `yaml:"steps"`
}

type StepConfig struct {
	Name      string       `yaml:"name"`
	Tasks     []TaskConfig `yaml:"tasks"`
	RunBefore []CmdConfig  `yaml:"run_before,omitempty"`
	RunAfter  []CmdConfig  `yaml:"run_after,omitempty"`
}

type TaskConfig struct {
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

func validateConfig(cfg *ConfigFile) error {
	if len(cfg.Jobs) == 0 {
		return errors.New("no job is declared")
	}

	jobNames := make(map[string]bool)
	for jobIdx, job := range cfg.Jobs {
		if len(job.Name) == 0 {
			return fmt.Errorf("the job #%d has no name declared", jobIdx)
		}

		// Check all the job names are unique
		if _, exists := jobNames[job.Name]; exists {
			return fmt.Errorf("there are multiple jobs named \"%s\"", job.Name)
		}
		jobNames[job.Name] = true

		if len(job.Steps) == 0 {
			return fmt.Errorf("no step is declared in the job \"%s\"", job.Name)
		}

		stepNames := make(map[string]bool)
		for stepIdx, step := range job.Steps {
			if len(step.Name) == 0 {
				return fmt.Errorf("the step #%d in the job \"%s\" has no name declared", stepIdx, job.Name)
			}

			// Check all the step names are unique
			if _, ok := stepNames[step.Name]; ok {
				return fmt.Errorf("there are multiple steps named \"%s\" in the job \"%s\"", step.Name, job.Name)
			}
			stepNames[step.Name] = true

			// Check the hooks
			if err := validateCommands(step.RunBefore); err != nil {
				return fmt.Errorf("the step \"%s\" in the job \"%s\" has invalid run_before hooks: %w", step.Name, job.Name, err)
			}
			if err := validateCommands(step.RunAfter); err != nil {
				return fmt.Errorf("the step \"%s\" in the job \"%s\" has invalid run_after hooks: %w", step.Name, job.Name, err)
			}

			// Check all the tasks. Check that within a step, the names are unique
			taskNames := make(map[string]bool)
			if len(step.Tasks) == 0 {
				return fmt.Errorf("the step \"%s\" in the job \"%s\" has no task declared", step.Name, job.Name)
			}
			for taskIdx, task := range step.Tasks {
				if len(task.Name) == 0 {
					return fmt.Errorf("the task #%d in the step \"%s\" in the job \"%s\" has no name declared", taskIdx, step.Name, job.Name)
				}
				if _, ok := taskNames[task.Name]; ok {
					return fmt.Errorf("there are multiple tasks named \"%s\" in the step \"%s\"in the job \"%s\"", task.Name, step.Name, job.Name)
				}
				taskNames[task.Name] = true

				if err := validateCommands(task.RunBefore); err != nil {
					return fmt.Errorf("the task \"%s\" in the step \"%s\" in the job \"%s\" has invalid run_before hooks: %w", step.Name, job.Name, task.Name, err)
				}
				if err := validateCommands(task.RunAfter); err != nil {
					return fmt.Errorf("the task \"%s\" in the step \"%s\" in the job \"%s\" has invalid run_after hooks: %w", step.Name, job.Name, task.Name, err)
				}
				if err := validateCommands(task.Background); err != nil {
					return fmt.Errorf("the task \"%s\" in the step \"%s\" in the job \"%s\" has invalid background tasks: %w", step.Name, job.Name, task.Name, err)
				}
				if len(task.Cmd) == 0 {
					return fmt.Errorf("the task \"%s\" in the step \"%s\" in the job \"%s\" has no command declared", step.Name, job.Name, task.Name)
				}
			}
		}
	}

	return nil
}

// parseConfig reads the content of the file at the absolute path given in argument, and extracts its yaml content into a ConfigFile.
func parseConfig(fileContent []byte) (*ConfigFile, error) {
	output := ConfigFile{}
	if err := yaml.Unmarshal(fileContent, &output); err != nil {
		return nil, fmt.Errorf("the file could not be parsed from YAML: %w", err)
	}

	if err := validateConfig(&output); err != nil {
		return nil, fmt.Errorf("the configuration is invalid: %w", err)
	}

	return &output, nil
}

func findAndParseConfig(givenPath string) (*ConfigFile, error) {
	configPath, err := findConfig(givenPath)
	if err != nil {
		return nil, err
	}

	fileContent, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("the contents of the file \"%s\" could not be read: %w", configPath, err)
	}

	config, err := parseConfig(fileContent)
	if err != nil {
		return nil, err
	}

	config.basePath = filepath.Dir(configPath)

	return config, nil
}
