package servicemgmt

import (
	"testing"

	"github.com/corentindeboisset/bolt/pkg/cfg"
	"github.com/stretchr/testify/require"
)

func TestNewOrchestrator(t *testing.T) {

	configText := `
services:
  serviceA:
    name: Service A
    cmd: "echo 'A'"
    dependencies:
      - target: serviceB

  serviceB:
    name: Z Service B
    cmd: echo 'B'
    dependencies:
      - target: serviceC

  serviceC:
    name: Service C
    cmd: echo 'C'
`
	config, err := cfg.ParseConfig([]byte(configText))
	require.Nil(t, err)
	orchestrator, err := NewOrchestrator(".", config.Services)
	require.Nil(t, err)
	require.Len(t, orchestrator.ServiceList, 3)
	require.Equal(t, orchestrator.ServiceList["serviceC"].Id, "serviceC")
	require.Len(t, orchestrator.ServiceList["serviceA"].Dependencies, 1)
	require.Len(t, orchestrator.ServiceList["serviceC"].Dependencies, 0)
	require.Equal(t, orchestrator.SortedServices()[2].Config.Name, "Z Service B")
}

func TestOrchestratorErrors(t *testing.T) {
	t.Parallel()

	autodependencyError := `
services:
  serviceA:
    name: Service A
    cmd: "echo 'A'"
    dependencies:
      - target: serviceA
`
	config, err := cfg.ParseConfig([]byte(autodependencyError))
	require.Nil(t, err)
	_, err = NewOrchestrator(".", config.Services)
	require.ErrorContains(t, err, "the service \"serviceA\" cannot be dependent on itself")

	invalidDependencyError := `
services:
  serviceA:
    name: Service A
    cmd: "echo 'A'"
    dependencies:
      - target: zozo
`
	config, err = cfg.ParseConfig([]byte(invalidDependencyError))
	require.Nil(t, err)
	_, err = NewOrchestrator(".", config.Services)
	require.ErrorContains(t, err, "the dependency target \"zozo\" of the service \"serviceA\" does not exist")

	doubleDependencyError := `
services:
  serviceA:
    name: Service A
    cmd: "echo 'A'"
    dependencies:
      - target: serviceB
      - target: serviceB

  serviceB:
    name: Service B
    cmd: echo 'B'
`
	config, err = cfg.ParseConfig([]byte(doubleDependencyError))
	require.Nil(t, err)
	_, err = NewOrchestrator(".", config.Services)
	require.ErrorContains(t, err, "the service \"serviceA\" have two dependencies on \"serviceB\"")

	circularDependencyError := `
services:
  serviceA:
    name: Service A
    cmd: "echo 'A'"
    dependencies:
      - target: serviceB

  serviceB:
    name: Service B
    cmd: echo 'B'
    dependencies:
      - target: serviceC

  serviceC:
    name: Service C
    cmd: echo 'C'
    dependencies:
      - target: serviceA
`
	config, err = cfg.ParseConfig([]byte(circularDependencyError))
	require.Nil(t, err)
	_, err = NewOrchestrator(".", config.Services)
	require.ErrorContains(t, err, "a circular dependency have been detected between the services")
}
