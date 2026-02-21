package cfg

import (
	"errors"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type ConfigFile struct {
	BasePath    string                   `yaml:"-"`
	LogFilePath string                   `yaml:"log_file"`
	Jobs        []JobConfig              `yaml:"jobs,omitempty"`
	Services    map[string]ServiceConfig `yaml:"services,omitempty"`
}

type DependencyConfig struct {
	Target            string `yaml:"target"`
	RestartWithTarget bool   `yaml:"restart_with_target"`
	WaitTargetStarted bool   `yaml:"wait_target_restarted"`
}

type ServiceConfig struct {
	TaskConfig `yaml:"task_config,inline"`

	AutoRestart  bool               `yaml:"auto_restart"`
	Dependencies []DependencyConfig `yaml:"dependencies"`
}

type JobConfig struct {
	Name     string       `yaml:"name"`
	Steps    []StepConfig `yaml:"steps"`
	RunAfter []CmdConfig  `yaml:"run_after,omitempty"`
}

type StepConfig struct {
	Name      string       `yaml:"name"`
	Tasks     []TaskConfig `yaml:"tasks"`
	RunBefore []CmdConfig  `yaml:"run_before,omitempty"`
	RunAfter  []CmdConfig  `yaml:"run_after,omitempty"`
}

type TaskConfig struct {
	CmdConfig `yaml:"cmd_config,inline"`

	Name      string      `yaml:"name"`
	RunBefore []CmdConfig `yaml:"run_before,omitempty"`
	RunAfter  []CmdConfig `yaml:"run_after,omitempty"`
}

type CmdConfig struct {
	Cmd  string `yaml:"cmd"`
	Path string `yaml:"path,omitempty"`
	// TODO: add a FailedWhen: a template calculated with the exit code, the stdout and stderr
}

// findConfig tries to find a configuration file. If no path is given in argument, it tries to find a bolt.yml file in the parent directories of the current working directory.
func findConfig(givenPath string) (ret string, err error) {
	if len(givenPath) > 0 {
		if filepath.IsAbs(givenPath) {
			ret = givenPath
		} else {
			ret, err = filepath.Abs(givenPath)
			if err != nil {
				return "", newConfigError("An error occured calculating an absolute path: %s", err)
			}
		}

		stat, err := os.Stat(ret)
		if err != nil {
			return "", newConfigError("An error occured when checking the path \"%s\":\n%s", ret, err)
		}
		if stat.IsDir() {
			return "", newConfigError("The path \"%s\" is a directory", ret)
		}

		return ret, nil
	}

	curDir, err := os.Getwd()
	if err != nil {
		return "", newConfigError("Failed to read the current working directory: %s", err)
	}
	for {
		stat, err := os.Stat(filepath.Join(curDir, "bolt.yml"))
		if err == nil && !stat.IsDir() {
			return filepath.Join(curDir, "bolt.yml"), nil
		}

		if curDir == filepath.Dir(curDir) {
			// We have reached the root directory
			break
		}

		// Go to the parent directory
		curDir = filepath.Dir(curDir)
	}

	return "", newConfigError("No configuration file could be found")
}

func validateCommands(configs []CmdConfig) error {
	for cmdIdx, cmd := range configs {
		if len(cmd.Cmd) == 0 {
			return newConfigError("The task #%d has no command declared", cmdIdx)
		}
	}

	return nil
}

func validateJobConfig(job *JobConfig) error {
	if len(job.Steps) == 0 {
		return newConfigError("No step is declared in the job \"%s\"", job.Name)
	}

	stepNames := make(map[string]bool)
	for stepIdx, step := range job.Steps {
		if len(step.Name) == 0 {
			return newConfigError("The step #%d in the job \"%s\" has no name declared", stepIdx, job.Name)
		}

		// Check all the step names are unique
		if _, ok := stepNames[step.Name]; ok {
			return newConfigError("There are multiple steps named \"%s\" in the job \"%s\"", step.Name, job.Name)
		}
		stepNames[step.Name] = true

		// Check the hooks
		if err := validateCommands(step.RunBefore); err != nil {
			return newConfigError("The step \"%s\" in the job \"%s\" has invalid run_before hooks: %s", step.Name, job.Name, err)
		}
		if err := validateCommands(step.RunAfter); err != nil {
			return newConfigError("The step \"%s\" in the job \"%s\" has invalid run_after hooks: %s", step.Name, job.Name, err)
		}

		// Check all the tasks. Check that within a step, the names are unique
		taskNames := make(map[string]bool)
		if len(step.Tasks) == 0 {
			return newConfigError("The step \"%s\" in the job \"%s\" has no task declared", step.Name, job.Name)
		}
		for taskIdx, task := range step.Tasks {
			if _, ok := taskNames[task.Name]; ok {
				return newConfigError("There are multiple tasks named \"%s\" in the step \"%s\"in the job \"%s\"", task.Name, step.Name, job.Name)
			}
			taskNames[task.Name] = true

			if err := validateTaskConfig(task); err != nil {
				if len(task.Name) > 0 {
					return newConfigError("The task \"%s\" in the step \"%s\" in the job \"%s\" is invalid: %s", task.Name, step.Name, job.Name, err)
				}
				return newConfigError("The task #%d in the step \"%s\" in the job \"%s\" is invalid: %s", taskIdx, step.Name, job.Name, err)
			}
		}
	}

	return nil
}

func validateTaskConfig(task TaskConfig) error {
	if len(task.Name) == 0 {
		return newConfigError("No name is declared")
	}

	if err := validateCommands(task.RunBefore); err != nil {
		return newConfigError("The run_before hooks are invalid: %s", err)
	}
	if err := validateCommands(task.RunAfter); err != nil {
		return newConfigError("The run_after hooks are invalid: %s", err)
	}
	if len(task.Cmd) == 0 {
		return newConfigError("No command is declared")
	}

	return nil
}

func validateConfig(cfg *ConfigFile) error {
	if len(cfg.Jobs) == 0 && len(cfg.Services) == 0 {
		return errors.New("No job and no service is declared in the configuration")
	}

	jobNames := make(map[string]bool)
	for jobIdx, job := range cfg.Jobs {
		if len(job.Name) == 0 {
			return newConfigError("The job #%d has no name declared", jobIdx)
		}

		// Check all the job names are unique
		if _, exists := jobNames[job.Name]; exists {
			return newConfigError("There are multiple jobs named \"%s\"", job.Name)
		}
		jobNames[job.Name] = true

		if err := validateJobConfig(&job); err != nil {
			return err
		}
	}

	for serviceId, service := range cfg.Services {
		if len(serviceId) == 0 {
			return newConfigError("The key of a service is not defined")
		}

		if err := validateTaskConfig(service.TaskConfig); err != nil {
			return newConfigError("The service \"%s\" is invalid: %s", serviceId, err)
		}
	}

	return nil
}

// parseConfig reads the content of the file at the absolute path given in argument, and extracts its yaml content into a ConfigFile.
func parseConfig(fileContent []byte) (*ConfigFile, error) {
	output := ConfigFile{}
	if err := yaml.Unmarshal(fileContent, &output); err != nil {
		return nil, newConfigError("The file could not be parsed from YAML: %s", err.Error())
	}

	if err := validateConfig(&output); err != nil {
		return nil, newConfigError("The configuration is invalid: %s", err)
	}

	return &output, nil
}

func FindAndParseConfig(givenPath string) (*ConfigFile, error) {
	configPath, err := findConfig(givenPath)
	if err != nil {
		return nil, err
	}

	fileContent, err := os.ReadFile(configPath)
	if err != nil {
		return nil, newConfigError("The contents of the file \"%s\" could not be read: %s", configPath, err)
	}

	config, err := parseConfig(fileContent)
	if err != nil {
		return nil, err
	}

	config.BasePath = filepath.Dir(configPath)

	return config, nil
}
