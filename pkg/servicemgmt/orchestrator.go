package servicemgmt

import (
	"context"
	"fmt"
	"log"
	"maps"
	"slices"
	"strings"
	"sync"

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
	Id string

	BasePath     string
	Config       cfg.ServiceConfig
	Dependencies []*ServiceDependency

	ctx    context.Context
	cancel context.CancelCauseFunc
	Output cmdrunr.SafeBuffer

	openPending bool

	State           ServiceState
	StateMtx        sync.Mutex
	DoneCond        *sync.Cond
	StartupOverCond *sync.Cond
}

type Orchestrator struct {
	ctx         context.Context
	cancel      context.CancelCauseFunc
	jobsDone    chan any
	BasePath    string
	ServiceList map[string]*ManagedService
}

func NewOrchestrator(basePath string, serviceConfigList map[string]cfg.ServiceConfig) (*Orchestrator, error) {
	serviceList := make(map[string]*ManagedService)
	dependencyConfigs := make(map[string]*[]cfg.DependencyConfig)
	for serviceId, serviceConfig := range serviceConfigList {
		name := serviceConfig.Name
		if len(name) == 0 {
			name = serviceId
		}

		serviceList[serviceId] = &ManagedService{
			Id:           serviceId,
			BasePath:     basePath,
			Config:       serviceConfig,
			Dependencies: make([]*ServiceDependency, 0),
			State:        SERVICE_OFF,
		}
		serviceList[serviceId].DoneCond = sync.NewCond(&serviceList[serviceId].StateMtx)
		serviceList[serviceId].StartupOverCond = sync.NewCond(&serviceList[serviceId].StateMtx)
		dependencyConfigs[serviceId] = &serviceConfig.Dependencies
	}

	// Build the dependency graph from the configuration
	for serviceId, service := range serviceList {
		for _, dependencyConfig := range *dependencyConfigs[serviceId] {
			if dependencyConfig.Target == serviceId {
				return nil, fmt.Errorf("the service \"%s\" cannot be dependent on itself", service.Id)
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
				return nil, fmt.Errorf("the dependency target \"%s\" of the service \"%s\" does not exist", dependencyConfig.Target, service.Id)
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

	ctx, cancel := context.WithCancelCause(context.Background())

	return &Orchestrator{BasePath: basePath, ServiceList: serviceList, ctx: ctx, cancel: cancel}, nil
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

// Returns the list of services sorted by case-insensitive alphabetical order
func (o *Orchestrator) SortedServices() []*ManagedService {
	return slices.SortedFunc(maps.Values(o.ServiceList), func(a, b *ManagedService) int {
		return strings.Compare(strings.ToLower(a.Config.Name), strings.ToLower(b.Config.Name))
	})
}

// Kill all services, wait for all process to end and return.
// If you call it in a goroutine, you can should a channel as an argument that will be closed once the shutdown is complete.
func (o *Orchestrator) Shutdown(done chan any) {
	o.cancel(cmdrunr.PlannedKill)

	// Wait for all services to be off
	for _, service := range o.ServiceList {
		service.StateMtx.Lock()

		for service.IsInExecution() {
			service.DoneCond.Wait()
		}

		service.StateMtx.Unlock()
	}

	if done != nil {
		close(done)
	}
}

// Calls start on a given service
func (o *Orchestrator) StartService(id string, outputWidth, outputHeight int) {
	for _, service := range o.ServiceList {
		if service.Id == id {
			service.Start(o.ctx, outputWidth, outputHeight)
			return
		}
	}
}

// Calls kill on a given service
func (o *Orchestrator) KillService(id string, appOnly bool) {
	for _, service := range o.ServiceList {
		if service.Id == id {
			service.Kill(appOnly)
			return
		}
	}
}

// Calls restart on a given service
func (o *Orchestrator) RestartService(id string, appOnly bool, outputWidth, outputHeight int) {
	for _, service := range o.ServiceList {
		if service.Id == id {
			service.Restart(o.ctx, appOnly, outputWidth, outputHeight)
			return
		}
	}
}

// Calls open on a given service
func (o *Orchestrator) OpenService(id string) {
	for _, service := range o.ServiceList {
		if service.Id == id {
			service.Open()
			return
		}
	}
}

// Check if the service is running.
// Be careful, you should lock the service's StateMtx before calling this method
func (s *ManagedService) IsInExecution() bool {
	return s.State == SERVICE_RUNNING || s.State == SERVICE_STARTING
}

// This method starts all dependencies, waits for them to be up, then starts the service and returns
// Be careful since this can block the routine for a while when waiting
func (s *ManagedService) Start(baseCtx context.Context, outputWidth, outputHeight int) {
	s.StateMtx.Lock()
	if s.IsInExecution() {
		log.Printf("The service %s is already started", s.Id)
		s.StateMtx.Unlock()
		return
	}
	s.StateMtx.Unlock()

	for _, dependency := range s.Dependencies {
		dependency.Target.Start(baseCtx, outputWidth, outputHeight)
	}

	// Wait for hard dependencies to pass their healthcheck
	var wg sync.WaitGroup
	for _, dependency := range s.Dependencies {
		if dependency.WaitTargetStarted {
			wg.Go(func() {
				dependency.Target.StateMtx.Lock()
				defer dependency.Target.StateMtx.Unlock()

				for dependency.Target.State == SERVICE_STARTING {
					dependency.Target.StartupOverCond.Wait()
				}
			})
		}
	}
	wg.Wait()

	s.StateMtx.Lock()
	defer s.StateMtx.Unlock()

	if s.IsInExecution() {
		return
	}

	s.ctx, s.cancel = context.WithCancelCause(baseCtx)
	s.State = SERVICE_STARTING

	go func() {
		log.Printf("Starting the service %s", s.Id)

		healthcheckDone := make(chan any)

		// Start the healthcheck routine
		go func() {
			defer close(healthcheckDone)

			if s.Config.Healthcheck.Port > 0 {
				if cmdrunr.WaitForPort(s.ctx, s.Config.Healthcheck.Port) {
					log.Printf("Healthcheck status for service %s on Port %d ok", s.Id, s.Config.Healthcheck.Port)
					s.onStartupSuccess()
				} else {
					log.Printf("Healthcheck failed for service %s", s.Id)
				}
			} else {
				s.onStartupSuccess()
			}

			// TODO: add other healthchecks (file created, text in output...)

			log.Printf("Healthcheck finished for the service %s", s.Id)
		}()

		// Start the service and wait for it to finish
		ok := cmdrunr.RunCommand(s.ctx, s.BasePath, s.Config.Path, s.Config.Cmd, &s.Output, outputWidth, outputHeight)

		log.Printf("The service %s has finished running", s.Id)

		// Ensure the healthcheck routine is finished
		s.cancel(cmdrunr.PlannedKill)
		<-healthcheckDone

		// Post-run updates
		s.StateMtx.Lock()
		defer s.StateMtx.Unlock()

		if ok {
			s.State = SERVICE_OFF
		} else {
			s.State = SERVICE_ERROR
		}

		s.DoneCond.Broadcast()
	}()
}

// Marks the service as running if it was in the "starting" phase
func (s *ManagedService) onStartupSuccess() {
	s.StateMtx.Lock()
	defer s.StateMtx.Unlock()

	if s.State == SERVICE_STARTING {
		s.State = SERVICE_RUNNING
	}

	s.StartupOverCond.Broadcast()
}

// Trigger the kill message, then waits for the process to have finished (and if appOnly=false, same for all dependencies)
func (s *ManagedService) Kill(appOnly bool) {
	s.StateMtx.Lock()
	if !s.IsInExecution() {
		s.StateMtx.Unlock()
		return
	}
	s.StateMtx.Unlock()

	s.cancel(cmdrunr.PlannedKill)

	// Wait for the service to be done and broadcast it
	s.StateMtx.Lock()
	for s.IsInExecution() {
		s.DoneCond.Wait()
	}
	s.StateMtx.Unlock()

	if !appOnly {
		for _, dependency := range s.Dependencies {
			// This is a tad inefficient, since we wait for every
			// kill to be over before passing to the next service.
			dependency.Target.Kill(appOnly)
		}
	}
}

// Kill the service properly, then run the start sequence
func (s *ManagedService) Restart(baseCtx context.Context, appOnly bool, outputWidth, outputHeight int) {
	s.Kill(appOnly)
	s.Start(baseCtx, outputWidth, outputHeight)
}

// Run the os-specific command to open the service target
// It may wait until the service is not in "starting" anymore
func (s *ManagedService) Open() {
	if len(s.Config.OpenTarget) == 0 {
		return
	}
	if s.openPending {
		return
	}

	s.openPending = true

	s.StateMtx.Lock()
	defer s.StateMtx.Unlock()

	for s.IsInExecution() && s.State != SERVICE_RUNNING {
		s.StartupOverCond.Wait()
	}

	if s.State == SERVICE_RUNNING {
		SystemOpen(s.Config.OpenTarget)
	}

	s.openPending = false
}
