/*
Copyright 2026.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v2

// Condition types for LocustTest.
const (
	// ConditionTypeReady indicates all resources are created and ready.
	ConditionTypeReady = "Ready"

	// ConditionTypeWorkersConnected indicates workers have connected to master.
	ConditionTypeWorkersConnected = "WorkersConnected"

	// ConditionTypeTestCompleted indicates the test has finished.
	ConditionTypeTestCompleted = "TestCompleted"

	// ConditionTypeSpecDrifted indicates the CR spec was modified after creation but changes are ignored.
	ConditionTypeSpecDrifted = "SpecDrifted"

	// ConditionTypePodsHealthy indicates whether pods are healthy and running.
	ConditionTypePodsHealthy = "PodsHealthy"
)

// Condition reasons for Ready condition.
const (
	ReasonResourcesCreating = "ResourcesCreating"
	ReasonResourcesCreated  = "ResourcesCreated"
	ReasonResourcesFailed   = "ResourcesFailed"
)

// Condition reasons for WorkersConnected condition.
const (
	ReasonWaitingForWorkers   = "WaitingForWorkers"
	ReasonAllWorkersConnected = "AllWorkersConnected"
	ReasonWorkersMissing      = "WorkersMissing"
)

// Condition reasons for TestCompleted condition.
const (
	ReasonTestInProgress = "TestInProgress"
	ReasonTestSucceeded  = "TestSucceeded"
	ReasonTestFailed     = "TestFailed"
)

// Condition reasons for SpecDrifted condition.
const (
	ReasonSpecChangeIgnored = "SpecChangeIgnored"
)

// Condition reasons for PodsHealthy condition.
const (
	ReasonPodsStarting       = "PodsStarting"
	ReasonPodsHealthy        = "PodsHealthy"
	ReasonPodImagePullError  = "ImagePullError"
	ReasonPodConfigError     = "ConfigurationError"
	ReasonPodSchedulingError = "SchedulingError"
	ReasonPodCrashLoop       = "CrashLoopBackOff"
	ReasonPodInitError       = "InitializationError"
)

// Phase represents the current lifecycle phase of a LocustTest.
type Phase string

// Phase constants for LocustTest status.
const (
	PhasePending   Phase = "Pending"
	PhaseRunning   Phase = "Running"
	PhaseSucceeded Phase = "Succeeded"
	PhaseFailed    Phase = "Failed"
)
