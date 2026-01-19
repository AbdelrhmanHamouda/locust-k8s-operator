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
	"testing"

	"github.com/stretchr/testify/assert"
)

const testCommandSeed = "locust -f /lotest/src/test.py"

func TestBuildMasterCommand(t *testing.T) {
	workerReplicas := int32(5)

	cmd := BuildMasterCommand(testCommandSeed, workerReplicas)

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
	seedWithExtraSpaces := "locust   -f   /lotest/src/test.py"
	workerReplicas := int32(3)

	cmd := BuildMasterCommand(seedWithExtraSpaces, workerReplicas)

	// strings.Fields handles multiple spaces correctly
	assert.Equal(t, "locust", cmd[0])
	assert.Equal(t, "-f", cmd[1])
	assert.Equal(t, "/lotest/src/test.py", cmd[2])
}

func TestBuildWorkerCommand(t *testing.T) {
	masterHost := "my-test-master"

	cmd := BuildWorkerCommand(testCommandSeed, masterHost)

	// Verify all expected flags are present
	assert.Contains(t, cmd, "locust")
	assert.Contains(t, cmd, "-f")
	assert.Contains(t, cmd, "/lotest/src/test.py")
	assert.Contains(t, cmd, "--worker")
	assert.Contains(t, cmd, "--master-port=5557")
	assert.Contains(t, cmd, "--master-host=my-test-master")
}

func TestBuildWorkerCommand_MasterHostCorrect(t *testing.T) {
	masterHost := "team-a-load-test-master"

	cmd := BuildWorkerCommand(testCommandSeed, masterHost)

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
