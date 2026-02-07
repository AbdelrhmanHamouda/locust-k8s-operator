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

// Package testdata provides test fixtures and helpers for unit tests.
package testdata

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"

	locustv1 "github.com/AbdelrhmanHamouda/locust-k8s-operator/api/v1"
	locustv2 "github.com/AbdelrhmanHamouda/locust-k8s-operator/api/v2"
)

// LoadLocustTest loads a v1 LocustTest from a JSON fixture file.
func LoadLocustTest(filename string) (*locustv1.LocustTest, error) {
	_, currentFile, _, _ := runtime.Caller(0)
	testdataDir := filepath.Dir(currentFile)

	data, err := os.ReadFile(filepath.Join(testdataDir, filename))
	if err != nil {
		return nil, err
	}

	var lt locustv1.LocustTest
	if err := json.Unmarshal(data, &lt); err != nil {
		return nil, err
	}

	return &lt, nil
}

// MustLoadLocustTest loads a v1 LocustTest from a JSON fixture file and panics on error.
// Useful in tests where fixture loading should never fail.
func MustLoadLocustTest(filename string) *locustv1.LocustTest {
	lt, err := LoadLocustTest(filename)
	if err != nil {
		panic(err)
	}
	return lt
}

// LoadV2Fixture loads a v2 LocustTest from a JSON fixture file.
func LoadV2Fixture(filename string) (*locustv2.LocustTest, error) {
	_, currentFile, _, _ := runtime.Caller(0)
	testdataDir := filepath.Dir(currentFile)

	data, err := os.ReadFile(filepath.Join(testdataDir, filename))
	if err != nil {
		return nil, err
	}

	var lt locustv2.LocustTest
	if err := json.Unmarshal(data, &lt); err != nil {
		return nil, err
	}

	return &lt, nil
}

// MustLoadV2Fixture loads a v2 LocustTest from a JSON fixture file and panics on error.
// Useful in tests where fixture loading should never fail.
func MustLoadV2Fixture(filename string) *locustv2.LocustTest {
	lt, err := LoadV2Fixture(filename)
	if err != nil {
		panic(err)
	}
	return lt
}
