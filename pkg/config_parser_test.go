package pkg

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigParse(t *testing.T) {
	t.Parallel()

	sampleConfig := `
jobs:
  - name: My first job
    steps:
      - name: first_step
        run_before:
          - name: my command
            cmd: my-command
        tasks:
          - name: echoes
            cmd: echo 12 && echo 13

          - name: sleep
            cmd: sleep 5
        run_after:
          - name: other command
            cmd: other-command

      - name: second_step
        tasks:
          - name: echoes 1
            cmd: echo 12 && echo 13
`

	config, err := parseConfig([]byte(sampleConfig))
	assert.Nil(t, err)
	assert.Len(t, config.Jobs, 1)
	assert.Equal(t, config.Jobs[0].Name, "My first job")

	assert.Len(t, config.Jobs[0].Steps, 2)

	assert.Len(t, config.Jobs[0].Steps[0].RunBefore, 1)
	assert.Equal(t, config.Jobs[0].Steps[0].RunBefore[0].Cmd, "my-command")
	assert.Len(t, config.Jobs[0].Steps[0].Tasks, 2)
	assert.Equal(t, config.Jobs[0].Steps[0].Tasks[1].Cmd, "sleep 5")
	assert.Len(t, config.Jobs[0].Steps[0].RunAfter, 1)

	assert.Len(t, config.Jobs[0].Steps[1].RunBefore, 0)
	assert.Len(t, config.Jobs[0].Steps[1].Tasks, 1)
	assert.Len(t, config.Jobs[0].Steps[1].RunAfter, 0)
}

func TestConfigErrors(t *testing.T) {
	t.Parallel()

	invalidYaml := `this is some plaintext`
	_, err := parseConfig([]byte(invalidYaml))
	assert.ErrorContains(t, err, "the file could not be parsed from YAML")

	emptyConfig := `
jobs: []
`
	_, err = parseConfig([]byte(emptyConfig))
	assert.ErrorContains(t, err, "no job is declared")

	noJobName := `
jobs:
  - steps: []
`
	_, err = parseConfig([]byte(noJobName))
	assert.ErrorContains(t, err, "the job #0 has no name declared")

	emptyJobConfig := `
jobs:
  - name: My first job
    steps: []
`
	_, err = parseConfig([]byte(emptyJobConfig))
	assert.ErrorContains(t, err, "no step is declared in the job \"My first job\"")

	noStepName := `
jobs:
  - name: My first job
    steps:
        - tasks:
            - name: first_task
              cmd: echo 1
`
	_, err = parseConfig([]byte(noStepName))
	assert.ErrorContains(t, err, "the step #0 in the job \"My first job\" has no name declared")

	duplicateStepNameConfig := `
jobs:
  - name: My first job
    steps:
        - name: first_step
          tasks:
            - name: first_task
              cmd: echo 1

        - name: first_step
          tasks:
            - name: first_task
              cmd: echo 1
`
	_, err = parseConfig([]byte(duplicateStepNameConfig))
	assert.ErrorContains(t, err, "there are multiple steps named \"first_step\" in the job \"My first job\"")

	duplicateTasksNames := `
jobs:
  - name: My first job
    steps:
        - name: first_step
          tasks:
            - name: first_task
              cmd: echo 1

            - name: first_task
              cmd: echo 2
`
	_, err = parseConfig([]byte(duplicateTasksNames))
	assert.ErrorContains(t, err, "there are multiple tasks named \"first_task\" in the step \"first_step\" in the job \"My first job\"")
}
