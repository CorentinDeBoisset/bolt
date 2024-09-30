package pkg

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigParse(t *testing.T) {
	sampleConfig := `
steps:
  - name: first_step
    run_before:
      - name: my command
        cmd: my-command
    jobs:
      - name: echoes
        cmd: echo 12 && echo 13

      - name: sleep
        cmd: sleep 5
    run_after:
      - name: other command
        cmd: other-command

  - name: second_step
    jobs:
      - name: echoes 1
        cmd: echo 12 && echo 13
`

	config, err := parseConfig([]byte(sampleConfig))
	assert.Nil(t, err)
	assert.Len(t, config.Steps, 2)

	assert.Len(t, config.Steps[0].RunBefore, 1)
	assert.Equal(t, config.Steps[0].RunBefore[0].Cmd, "my-command")
	assert.Len(t, config.Steps[0].Jobs, 2)
	assert.Equal(t, config.Steps[0].Jobs[1].Cmd, "sleep 5")
	assert.Len(t, config.Steps[0].RunAfter, 1)

	assert.Len(t, config.Steps[1].RunBefore, 0)
	assert.Len(t, config.Steps[1].Jobs, 1)
	assert.Len(t, config.Steps[1].RunAfter, 0)
}

func TestConfigErrors(t *testing.T) {
	invalidYaml := `this is some plaintext`
	_, err := parseConfig([]byte(invalidYaml))
	assert.ErrorContains(t, err, "the file could not be parsed from YAML")

	emptyConfig := `
steps: []
`
	_, err = parseConfig([]byte(emptyConfig))
	assert.ErrorContains(t, err, "no step is declared")

	noStepName := `
steps:
    - jobs:
        - name: firstjob
          cmd: echo 1
`
	_, err = parseConfig([]byte(noStepName))
	assert.ErrorContains(t, err, "the step #0 has no name declared")

	duplicateStepNameConfig := `
steps:
    - name: firststep
      jobs:
        - name: firstjob
          cmd: echo 1

    - name: firststep
      jobs:
        - name: firstjob
          cmd: echo 1
`
	_, err = parseConfig([]byte(duplicateStepNameConfig))
	assert.ErrorContains(t, err, "there are multiple steps named \"firststep\"")

	duplicateJobNames := `
steps:
    - name: firststep
      jobs:
        - name: firstjob
          cmd: echo 1

        - name: firstjob
          cmd: echo 2
`
	_, err = parseConfig([]byte(duplicateJobNames))
	assert.ErrorContains(t, err, "there are multiple jobs named \"firstjob\" in the step \"firststep\"")
}
