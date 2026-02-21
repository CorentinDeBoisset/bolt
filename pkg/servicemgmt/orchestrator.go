package servicemgmt

import (
	"fmt"

	"github.com/corentindeboisset/bolt/pkg/cfg"
	"github.com/corentindeboisset/bolt/pkg/cmdrunr"
)

type ServiceState int

const (
	SERVICE_OFF ServiceState = iota
	SERVICE_STARTING
	SERVICE_FAILED_DEPENDENCY
	SERVICE_RUNNING
	SERVICE_ERROR
)

type ServiceDependency struct {
	Target            *ManagedService
	RestartWithTarget bool
	WaitTargetStarted bool
}

type ManagedService struct {
	Id           string
	Name         string
	Dependencies []*ServiceDependency
	AutoRestart  bool

	State  ServiceState
	Output cmdrunr.SafeBuffer
}

type Orchestrator struct {
	ServiceList map[string]*ManagedService
}

func NewOrchestrator(serviceConfigList map[string]cfg.ServiceConfig) (*Orchestrator, error) {
	serviceList := make(map[string]*ManagedService)
	dependencyConfigs := make(map[string]*[]cfg.DependencyConfig)
	for serviceId, serviceConfig := range serviceConfigList {
		name := serviceConfig.Name
		if len(name) == 0 {
			name = serviceId
		}

		serviceList[serviceId] = &ManagedService{
			Id:           serviceId,
			Name:         serviceConfig.Name,
			Dependencies: make([]*ServiceDependency, 0),
			AutoRestart:  serviceConfig.AutoRestart,
			State:        SERVICE_OFF,
		}
		dependencyConfigs[serviceId] = &serviceConfig.Dependencies
	}

	// Build the dependency graph from the configuration
	for serviceId, service := range serviceList {
		for _, dependencyConfig := range *dependencyConfigs[serviceId] {
			if dependencyConfig.Target == serviceId {
				return nil, fmt.Errorf("the service \"%s\" cannot be dependent on itself", service.Name)
			}

			foundDependency := false
			for targetCandidateId, targetCandidate := range serviceList {
				if dependencyConfig.Target == targetCandidateId {
					dependency := &ServiceDependency{
						Target:            targetCandidate,
						RestartWithTarget: dependencyConfig.RestartWithTarget,
						WaitTargetStarted: dependencyConfig.WaitTargetStarted,
					}
					service.Dependencies = append(service.Dependencies, dependency)
					foundDependency = true
				}
			}
			if !foundDependency {
				return nil, fmt.Errorf("the dependency target \"%s\" of the service \"%s\" does not exist", dependencyConfig.Target, serviceId)
			}
		}

		// Check there are no double-dependencies
		for idxA, dependencyA := range service.Dependencies {
			for idxB, dependencyB := range service.Dependencies {
				if idxA != idxB && dependencyA.Target.Id == dependencyB.Target.Id {
					return nil, fmt.Errorf("the service \"%s\" have two dependencies on \"%s\"", service.Id, dependencyA.Target.Id)
				}
			}
		}
	}

	if !isDependencyGraphValid(serviceList) {
		return nil, fmt.Errorf("a circular dependency have been detected between the services")
	}

	return &Orchestrator{serviceList}, nil
}

// Use Kahn's algorithm to ensure the dependency graph has no cycle.
// Implementation insipired from https://github.com/amwolff/gorder/blob/master/gorder.go (MIT licence)
func isDependencyGraphValid(serviceList map[string]*ManagedService) bool {
	indegrees := make(map[string]int)
	for serviceId := range serviceList {
		for _, dependency := range serviceList[serviceId].Dependencies {
			indegrees[dependency.Target.Id]++
		}
	}

	queue := make([]string, 0)
	for serviceId := range serviceList {
		if _, ok := indegrees[serviceId]; !ok {
			queue = append(queue, serviceId)
		}
	}

	for len(queue) > 0 {
		serviceId := queue[len(queue)-1]
		queue = queue[:(len(queue) - 1)]
		for _, dependency := range serviceList[serviceId].Dependencies {
			indegrees[dependency.Target.Id]--
			if indegrees[dependency.Target.Id] == 0 {
				queue = append(queue, dependency.Target.Id)
			}
		}
	}

	for _, indegree := range indegrees {
		if indegree > 0 {
			return false
		}
	}

	return true
}

// func (o *Orchestrator) StartService(id string) {
// 	for _, service := range o.ServiceList {
// 		if service.Id == id {
// 			service.Start()
// 			return
// 		}
// 	}
// }

// func (s *ManagedService) Start() {
// }
