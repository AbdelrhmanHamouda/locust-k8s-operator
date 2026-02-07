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

package resources

import (
	"strings"
	"testing"

	locustv2 "github.com/AbdelrhmanHamouda/locust-k8s-operator/api/v2"
	"github.com/go-logr/logr"
	"github.com/stretchr/testify/assert"
	"k8s.io/utils/ptr"
)

const testCommandSeed = "locust -f /lotest/src/test.py"
const testMasterHost = "my-test-master"

// Helper to create a default MasterSpec for testing
func testMasterSpec() *locustv2.MasterSpec {
	return &locustv2.MasterSpec{
		Command:   testCommandSeed,
		Autostart: ptr.To(true),
		Autoquit:  &locustv2.AutoquitConfig{Enabled: true, Timeout: 60},
	}
}

func TestBuildMasterCommand(t *testing.T) {
	workerReplicas := int32(5)
	masterSpec := testMasterSpec()

	cmd := BuildMasterCommand(masterSpec, workerReplicas, false, logr.Discard())

	// Verify all expected flags are present
	assert.Contains(t, cmd, "locust")
	assert.Contains(t, cmd, "-f")
	assert.Contains(t, cmd, "/lotest/src/test.py")
	assert.Contains(t, cmd, "--master")
	assert.Contains(t, cmd, "--master-port=5557")
	assert.Contains(t, cmd, "--expect-workers=5")
	assert.Contains(t, cmd, "--autostart")
	assert.Contains(t, cmd, "--autoquit")
	assert.Contains(t, cmd, "60")
	assert.Contains(t, cmd, "--enable-rebalancing")
	assert.Contains(t, cmd, "--only-summary")
}

func TestBuildMasterCommand_SplitsCorrectly(t *testing.T) {
	masterSpec := &locustv2.MasterSpec{
		Command:   "locust   -f   /lotest/src/test.py",
		Autostart: ptr.To(true),
	}
	workerReplicas := int32(3)

	cmd := BuildMasterCommand(masterSpec, workerReplicas, false, logr.Discard())

	// strings.Fields handles multiple spaces correctly
	assert.Equal(t, "locust", cmd[0])
	assert.Equal(t, "-f", cmd[1])
	assert.Equal(t, "/lotest/src/test.py", cmd[2])
}

func TestBuildWorkerCommand(t *testing.T) {
	masterHost := testMasterHost

	cmd := BuildWorkerCommand(testCommandSeed, masterHost, false, nil, logr.Discard())

	// Verify all expected flags are present
	assert.Contains(t, cmd, "locust")
	assert.Contains(t, cmd, "-f")
	assert.Contains(t, cmd, "/lotest/src/test.py")
	assert.Contains(t, cmd, "--worker")
	assert.Contains(t, cmd, "--master-port=5557")
	assert.Contains(t, cmd, "--master-host="+testMasterHost)
}

func TestBuildWorkerCommand_MasterHostCorrect(t *testing.T) {
	masterHost := "team-a-load-test-master"

	cmd := BuildWorkerCommand(testCommandSeed, masterHost, false, nil, logr.Discard())

	// Find the master-host flag
	found := false
	for _, arg := range cmd {
		if arg == "--master-host=team-a-load-test-master" {
			found = true
			break
		}
	}
	assert.True(t, found, "master-host flag should contain the correct master host")
}

// ===== OTel Flag Tests =====

func TestBuildMasterCommand_OTelDisabled(t *testing.T) {
	workerReplicas := int32(3)
	masterSpec := testMasterSpec()

	cmd := BuildMasterCommand(masterSpec, workerReplicas, false, logr.Discard())

	// --otel flag should NOT be present
	assert.NotContains(t, cmd, "--otel")
}

func TestBuildMasterCommand_OTelEnabled(t *testing.T) {
	workerReplicas := int32(3)
	masterSpec := testMasterSpec()

	cmd := BuildMasterCommand(masterSpec, workerReplicas, true, logr.Discard())

	// --otel flag should be present
	assert.Contains(t, cmd, "--otel")
}

func TestBuildMasterCommand_OTelFlagPosition(t *testing.T) {
	workerReplicas := int32(3)
	masterSpec := testMasterSpec()

	cmd := BuildMasterCommand(masterSpec, workerReplicas, true, logr.Discard())

	// Find positions of --otel and --master
	otelIndex := -1
	masterIndex := -1
	for i, arg := range cmd {
		if arg == "--otel" {
			otelIndex = i
		}
		if arg == "--master" {
			masterIndex = i
		}
	}

	// --otel should appear before --master
	assert.NotEqual(t, -1, otelIndex, "--otel flag should be present")
	assert.NotEqual(t, -1, masterIndex, "--master flag should be present")
	assert.Less(t, otelIndex, masterIndex, "--otel should appear before --master")
}

// ===== Autostart/Autoquit Tests =====

func TestBuildMasterCommand_AutostartDisabled(t *testing.T) {
	masterSpec := &locustv2.MasterSpec{
		Command:   testCommandSeed,
		Autostart: ptr.To(false),
	}

	cmd := BuildMasterCommand(masterSpec, 3, false, logr.Discard())

	assert.NotContains(t, cmd, "--autostart")
}

func TestBuildMasterCommand_AutostartDefault(t *testing.T) {
	// When Autostart is nil, default to true
	masterSpec := &locustv2.MasterSpec{
		Command: testCommandSeed,
	}

	cmd := BuildMasterCommand(masterSpec, 3, false, logr.Discard())

	assert.Contains(t, cmd, "--autostart")
}

func TestBuildMasterCommand_AutoquitDisabled(t *testing.T) {
	masterSpec := &locustv2.MasterSpec{
		Command:   testCommandSeed,
		Autostart: ptr.To(true),
		Autoquit:  &locustv2.AutoquitConfig{Enabled: false},
	}

	cmd := BuildMasterCommand(masterSpec, 3, false, logr.Discard())

	assert.NotContains(t, cmd, "--autoquit")
}

func TestBuildMasterCommand_AutoquitCustomTimeout(t *testing.T) {
	masterSpec := &locustv2.MasterSpec{
		Command:   testCommandSeed,
		Autostart: ptr.To(true),
		Autoquit:  &locustv2.AutoquitConfig{Enabled: true, Timeout: 120},
	}

	cmd := BuildMasterCommand(masterSpec, 3, false, logr.Discard())

	assert.Contains(t, cmd, "--autoquit")
	assert.Contains(t, cmd, "120")
	assert.NotContains(t, cmd, "60")
}

func TestBuildMasterCommand_AutoquitDefault(t *testing.T) {
	// When Autoquit is nil, default to enabled with 60s timeout
	masterSpec := &locustv2.MasterSpec{
		Command: testCommandSeed,
	}

	cmd := BuildMasterCommand(masterSpec, 3, false, logr.Discard())

	assert.Contains(t, cmd, "--autoquit")
	assert.Contains(t, cmd, "60")
}

func TestBuildWorkerCommand_OTelDisabled(t *testing.T) {
	masterHost := testMasterHost

	cmd := BuildWorkerCommand(testCommandSeed, masterHost, false, nil, logr.Discard())

	// --otel flag should NOT be present
	assert.NotContains(t, cmd, "--otel")
}

func TestBuildWorkerCommand_OTelEnabled(t *testing.T) {
	masterHost := testMasterHost

	cmd := BuildWorkerCommand(testCommandSeed, masterHost, true, nil, logr.Discard())

	// --otel flag should be present
	assert.Contains(t, cmd, "--otel")
}

func TestBuildWorkerCommand_OTelFlagPosition(t *testing.T) {
	masterHost := testMasterHost

	cmd := BuildWorkerCommand(testCommandSeed, masterHost, true, nil, logr.Discard())

	// Find positions of --otel and --worker
	otelIndex := -1
	workerIndex := -1
	for i, arg := range cmd {
		if arg == "--otel" {
			otelIndex = i
		}
		if arg == "--worker" {
			workerIndex = i
		}
	}

	// --otel should appear before --worker
	assert.NotEqual(t, -1, otelIndex, "--otel flag should be present")
	assert.NotEqual(t, -1, workerIndex, "--worker flag should be present")
	assert.Less(t, otelIndex, workerIndex, "--otel should appear before --worker")
}

// ===== ExtraArgs Tests =====

func TestBuildMasterCommand_WithExtraArgs(t *testing.T) {
	masterSpec := testMasterSpec()
	masterSpec.ExtraArgs = []string{"--csv=results", "--users=100"}

	cmd := BuildMasterCommand(masterSpec, 3, false, logr.Discard())

	// ExtraArgs should be present in command
	assert.Contains(t, cmd, "--csv=results")
	assert.Contains(t, cmd, "--users=100")

	// ExtraArgs should come after operator-managed flags
	onlySummaryIndex := -1
	csvIndex := -1
	for i, arg := range cmd {
		if arg == "--only-summary" {
			onlySummaryIndex = i
		}
		if arg == "--csv=results" {
			csvIndex = i
		}
	}
	assert.Greater(t, csvIndex, onlySummaryIndex, "extraArgs should come after --only-summary")
}

func TestBuildMasterCommand_WithExtraArgsNil(t *testing.T) {
	masterSpec := testMasterSpec()
	masterSpec.ExtraArgs = nil

	cmd := BuildMasterCommand(masterSpec, 3, false, logr.Discard())

	// Command should be identical to behavior without extraArgs
	assert.Contains(t, cmd, "--master")
	assert.Contains(t, cmd, "--only-summary")
}

func TestBuildMasterCommand_WithExtraArgsEmpty(t *testing.T) {
	masterSpec := testMasterSpec()
	masterSpec.ExtraArgs = []string{}

	cmd := BuildMasterCommand(masterSpec, 3, false, logr.Discard())

	// Command should be identical to behavior without extraArgs
	assert.Contains(t, cmd, "--master")
	assert.Contains(t, cmd, "--only-summary")
}

func TestBuildMasterCommand_WithConflictingExtraArgs(t *testing.T) {
	masterSpec := testMasterSpec()
	masterSpec.ExtraArgs = []string{"--master-port=9999"}

	cmd := BuildMasterCommand(masterSpec, 3, false, logr.Discard())

	// Command should contain both operator flag and user flag
	// User flag comes last, so it wins per POSIX behavior
	assert.Contains(t, cmd, "--master-port=5557")
	assert.Contains(t, cmd, "--master-port=9999")

	// Find indices to verify order
	operatorIndex := -1
	userIndex := -1
	for i, arg := range cmd {
		if arg == "--master-port=5557" {
			operatorIndex = i
		}
		if arg == "--master-port=9999" {
			userIndex = i
		}
	}
	assert.Greater(t, userIndex, operatorIndex, "user's flag should come after operator's flag")
}

func TestBuildWorkerCommand_WithExtraArgs(t *testing.T) {
	extraArgs := []string{"--csv=results"}

	cmd := BuildWorkerCommand(testCommandSeed, testMasterHost, false, extraArgs, logr.Discard())

	// ExtraArgs should be present in command
	assert.Contains(t, cmd, "--csv=results")

	// ExtraArgs should come after operator-managed flags
	masterHostFlag := "--master-host=" + testMasterHost
	masterHostIndex := -1
	csvIndex := -1
	for i, arg := range cmd {
		if arg == masterHostFlag {
			masterHostIndex = i
		}
		if arg == "--csv=results" {
			csvIndex = i
		}
	}
	assert.Greater(t, csvIndex, masterHostIndex, "extraArgs should come after --master-host")
}

func TestBuildWorkerCommand_WithConflictingExtraArgs(t *testing.T) {
	extraArgs := []string{"--worker", "--master-host=evil"}

	cmd := BuildWorkerCommand(testCommandSeed, testMasterHost, false, extraArgs, logr.Discard())

	// Command should contain both operator flags and user flags
	assert.Contains(t, cmd, "--worker")
	assert.Contains(t, cmd, "--master-host="+testMasterHost)
	assert.Contains(t, cmd, "--master-host=evil")
}

func TestDetectFlagConflicts_WithConflict(t *testing.T) {
	extraArgs := []string{"--master-port=9999"}

	conflicts := detectFlagConflicts(extraArgs)

	assert.Equal(t, []string{"--master-port=9999"}, conflicts)
}

func TestDetectFlagConflicts_WithoutConflict(t *testing.T) {
	extraArgs := []string{"--csv=results"}

	conflicts := detectFlagConflicts(extraArgs)

	assert.Empty(t, conflicts)
}

func TestDetectFlagConflicts_WithOtelConflict(t *testing.T) {
	extraArgs := []string{"--otel"}

	conflicts := detectFlagConflicts(extraArgs)

	assert.Equal(t, []string{"--otel"}, conflicts)
}

func TestDetectFlagConflicts_MultipleConflicts(t *testing.T) {
	extraArgs := []string{"--master-port=9999", "--csv=results", "--worker"}

	conflicts := detectFlagConflicts(extraArgs)

	assert.Len(t, conflicts, 2)
	assert.Contains(t, conflicts, "--master-port=9999")
	assert.Contains(t, conflicts, "--worker")
}

// TestBuildMasterCommand_LogsWarningOnConflict verifies that conflict warnings are logged
func TestBuildMasterCommand_LogsWarningOnConflict(t *testing.T) {
	// Use a test logger that captures log output
	var logBuffer strings.Builder
	logger := logr.New(&testLogSink{writer: &logBuffer})

	masterSpec := testMasterSpec()
	masterSpec.ExtraArgs = []string{"--master-port=9999"}

	BuildMasterCommand(masterSpec, 3, false, logger)

	// Verify that warning was logged (exact message match not required, just presence)
	logOutput := logBuffer.String()
	assert.Contains(t, logOutput, "extraArgs override operator-managed flags", "should log conflict warning")
}

func TestBuildWorkerCommand_LogsWarningOnConflict(t *testing.T) {
	var logBuffer strings.Builder
	logger := logr.New(&testLogSink{writer: &logBuffer})

	extraArgs := []string{"--worker"}

	BuildWorkerCommand(testCommandSeed, testMasterHost, false, extraArgs, logger)

	logOutput := logBuffer.String()
	assert.Contains(t, logOutput, "extraArgs override operator-managed flags", "should log conflict warning")
}

// testLogSink is a simple logr.LogSink implementation for testing
type testLogSink struct {
	writer *strings.Builder
}

func (t *testLogSink) Init(info logr.RuntimeInfo) {}

func (t *testLogSink) Enabled(level int) bool {
	return true
}

func (t *testLogSink) Info(level int, msg string, keysAndValues ...interface{}) {
	t.writer.WriteString(msg)
	for i := 0; i < len(keysAndValues); i += 2 {
		if i+1 < len(keysAndValues) {
			t.writer.WriteString(" ")
			t.writer.WriteString(keysAndValues[i].(string))
			t.writer.WriteString("=")
		}
	}
}

func (t *testLogSink) Error(err error, msg string, keysAndValues ...interface{}) {
	t.writer.WriteString("ERROR: ")
	t.writer.WriteString(msg)
}

func (t *testLogSink) WithValues(keysAndValues ...interface{}) logr.LogSink {
	return t
}

func (t *testLogSink) WithName(name string) logr.LogSink {
	return t
}
