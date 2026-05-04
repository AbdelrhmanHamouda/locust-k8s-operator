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

package main

import (
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"sigs.k8s.io/controller-runtime/pkg/healthz"
)

// fakeReadyzAdder records each registration so tests can assert exactly
// which checks were added. Errors are injected per check name so each
// registration call site can be exercised independently.
type fakeReadyzAdder struct {
	healthChecks []string
	readyChecks  []string
	healthErr    map[string]error
	readyErr     map[string]error
}

func (f *fakeReadyzAdder) AddHealthzCheck(name string, _ healthz.Checker) error {
	if err, ok := f.healthErr[name]; ok {
		return err
	}
	f.healthChecks = append(f.healthChecks, name)
	return nil
}

func (f *fakeReadyzAdder) AddReadyzCheck(name string, _ healthz.Checker) error {
	if err, ok := f.readyErr[name]; ok {
		return err
	}
	f.readyChecks = append(f.readyChecks, name)
	return nil
}

func sentinelChecker(_ *http.Request) error { return nil }

func assertNoWebhookReadyz(t *testing.T, rec *fakeReadyzAdder, ctx string) {
	t.Helper()
	for _, name := range rec.readyChecks {
		if name == "webhook" {
			t.Fatalf("%s: webhook readyz must not be registered "+
				"(adding it triggers GetWebhookServer, which loads TLS certs "+
				"the chart deliberately does not mount)", ctx)
		}
	}
}

func TestAddHealthChecks_DisabledSkipsWebhookCheck(t *testing.T) {
	rec := &fakeReadyzAdder{}

	if err := addHealthChecks(rec, false, healthz.Checker(sentinelChecker)); err != nil {
		t.Fatalf("addHealthChecks returned error: %v", err)
	}

	assertNoWebhookReadyz(t, rec, "webhooks disabled")
	if !contains(rec.readyChecks, "readyz") {
		t.Errorf("expected baseline readyz check, got %v", rec.readyChecks)
	}
	if !contains(rec.healthChecks, "healthz") {
		t.Errorf("expected baseline healthz check, got %v", rec.healthChecks)
	}
}

func TestAddHealthChecks_EnabledRegistersWebhookCheck(t *testing.T) {
	rec := &fakeReadyzAdder{}

	if err := addHealthChecks(rec, true, healthz.Checker(sentinelChecker)); err != nil {
		t.Fatalf("addHealthChecks returned error: %v", err)
	}

	if !contains(rec.readyChecks, "webhook") {
		t.Errorf("expected readyz check %q to be registered when webhooks are enabled, got %v",
			"webhook", rec.readyChecks)
	}
}

func TestAddHealthChecks_EnabledNilCheckerSkipped(t *testing.T) {
	rec := &fakeReadyzAdder{}

	if err := addHealthChecks(rec, true, nil); err != nil {
		t.Fatalf("addHealthChecks returned error: %v", err)
	}

	assertNoWebhookReadyz(t, rec, "nil checker")
}

func TestAddHealthChecks_PropagatesHealthzError(t *testing.T) {
	want := errors.New("healthz boom")
	rec := &fakeReadyzAdder{healthErr: map[string]error{"healthz": want}}

	err := addHealthChecks(rec, false, nil)
	if !errors.Is(err, want) {
		t.Errorf("expected wrapped error %v, got %v", want, err)
	}
}

func TestAddHealthChecks_PropagatesReadyzError(t *testing.T) {
	want := errors.New("readyz boom")
	rec := &fakeReadyzAdder{readyErr: map[string]error{"readyz": want}}

	err := addHealthChecks(rec, false, nil)
	if !errors.Is(err, want) {
		t.Errorf("expected wrapped error %v, got %v", want, err)
	}
}

func TestAddHealthChecks_PropagatesWebhookReadyzError(t *testing.T) {
	want := errors.New("webhook readyz boom")
	rec := &fakeReadyzAdder{readyErr: map[string]error{"webhook": want}}

	err := addHealthChecks(rec, true, healthz.Checker(sentinelChecker))
	if !errors.Is(err, want) {
		t.Errorf("expected wrapped error %v, got %v", want, err)
	}
}

func TestWaitForWebhookCerts_TimesOutWhenAbsent(t *testing.T) {
	dir := t.TempDir()
	certFile := filepath.Join(dir, "tls.crt")
	keyFile := filepath.Join(dir, "tls.key")

	start := time.Now()
	err := waitForWebhookCerts(certFile, keyFile, 100*time.Millisecond)
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
	if !strings.Contains(err.Error(), "not present") {
		t.Errorf("error should mention missing files, got: %v", err)
	}
	// Lower bound: timeout must actually wait (allow scheduler slop).
	if elapsed < 80*time.Millisecond {
		t.Errorf("returned too quickly: %v", elapsed)
	}
	// Upper bound: must give up; allow generous headroom for CI.
	if elapsed > 3*time.Second {
		t.Errorf("hung past timeout: %v", elapsed)
	}
}

func TestWaitForWebhookCerts_ReturnsImmediatelyWhenPresent(t *testing.T) {
	dir := t.TempDir()
	certFile := filepath.Join(dir, "tls.crt")
	keyFile := filepath.Join(dir, "tls.key")
	mustWriteFile(t, certFile, "cert")
	mustWriteFile(t, keyFile, "key")

	start := time.Now()
	err := waitForWebhookCerts(certFile, keyFile, 5*time.Second)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("expected nil error when files exist, got: %v", err)
	}
	if elapsed > 500*time.Millisecond {
		t.Errorf("returned too slowly when files were already present: %v", elapsed)
	}
}

func TestWaitForWebhookCerts_PollsUntilFilesAppear(t *testing.T) {
	dir := t.TempDir()
	certFile := filepath.Join(dir, "tls.crt")
	keyFile := filepath.Join(dir, "tls.key")

	// Write the files asynchronously after a short delay; the wait must
	// observe them and return cleanly before the timeout fires.
	go func() {
		time.Sleep(150 * time.Millisecond)
		_ = os.WriteFile(certFile, []byte("cert"), 0o600)
		_ = os.WriteFile(keyFile, []byte("key"), 0o600)
	}()

	err := waitForWebhookCerts(certFile, keyFile, 5*time.Second)
	if err != nil {
		t.Fatalf("expected wait to succeed when files appear, got: %v", err)
	}
}

func TestApplyEnableWebhooksEnv_FlagWinsOverEnv(t *testing.T) {
	cfg := &flagConfig{enableWebhooks: false, enableWebhooksFlagSet: true}
	logCount := 0
	logf := func(msg string, kv ...any) { logCount++ }

	applyEnableWebhooksEnv(cfg, "true", logf)

	if cfg.enableWebhooks {
		t.Errorf("flag must win over env var, got enableWebhooks=true")
	}
	if logCount != 1 {
		t.Errorf("expected exactly one deprecation log, got %d", logCount)
	}
}

func TestApplyEnableWebhooksEnv_EnvAppliedWhenFlagUnset_True(t *testing.T) {
	cfg := &flagConfig{enableWebhooks: false}
	logf := func(string, ...any) {}

	applyEnableWebhooksEnv(cfg, "true", logf)

	if !cfg.enableWebhooks {
		t.Errorf("env var 'true' with no flag should set enableWebhooks=true")
	}
}

func TestApplyEnableWebhooksEnv_EnvAppliedWhenFlagUnset_False(t *testing.T) {
	cfg := &flagConfig{enableWebhooks: true}
	logf := func(string, ...any) {}

	applyEnableWebhooksEnv(cfg, "false", logf)

	if cfg.enableWebhooks {
		t.Errorf("env var 'false' with no flag should set enableWebhooks=false")
	}
}

func TestApplyEnableWebhooksEnv_LogsEvenWhenFlagWins(t *testing.T) {
	cfg := &flagConfig{enableWebhooks: true, enableWebhooksFlagSet: true}
	logCount := 0
	var firstMsg string
	logf := func(msg string, kv ...any) {
		logCount++
		if logCount == 1 {
			firstMsg = msg
		}
	}

	applyEnableWebhooksEnv(cfg, "true", logf)

	if logCount != 1 {
		t.Errorf("user with stale env var should still see the deprecation warning, got %d logs", logCount)
	}
	if !strings.Contains(firstMsg, "deprecated") {
		t.Errorf("deprecation log must mention 'deprecated', got %q", firstMsg)
	}
}

func TestApplyEnableWebhooksEnv_ParsesViaStrconvParseBool(t *testing.T) {
	truthyValues := []string{"1", "t", "T", "TRUE", "true", "True"}
	falsyValues := []string{"0", "f", "F", "FALSE", "false", "False"}
	invalidValues := []string{"yes", "no", "on", "off", "True_typo", "True ", "  true"}

	for _, v := range truthyValues {
		t.Run("truthy/"+v, func(t *testing.T) {
			cfg := &flagConfig{enableWebhooks: false}
			applyEnableWebhooksEnv(cfg, v, func(string, ...any) {})
			if !cfg.enableWebhooks {
				t.Errorf("envVal=%q should enable webhooks", v)
			}
		})
	}
	for _, v := range falsyValues {
		t.Run("falsy/"+v, func(t *testing.T) {
			cfg := &flagConfig{enableWebhooks: true}
			applyEnableWebhooksEnv(cfg, v, func(string, ...any) {})
			if cfg.enableWebhooks {
				t.Errorf("envVal=%q should disable webhooks", v)
			}
		})
	}
	for _, v := range invalidValues {
		t.Run("invalid/"+v, func(t *testing.T) {
			cfg := &flagConfig{enableWebhooks: false}
			var warned bool
			logf := func(msg string, _ ...any) {
				if strings.Contains(msg, "ignoring") {
					warned = true
				}
			}
			applyEnableWebhooksEnv(cfg, v, logf)
			if cfg.enableWebhooks {
				t.Errorf("envVal=%q is not a valid boolean and must not flip the default; got enableWebhooks=true", v)
			}
			if !warned {
				t.Errorf("envVal=%q is invalid and must produce an 'ignoring' log line", v)
			}
		})
	}
}

func TestValidateFlags_RejectsEnabledWebhooksWithoutCertPath(t *testing.T) {
	cfg := &flagConfig{enableWebhooks: true, webhookCertPath: ""}

	err := validateFlags(cfg)
	if err == nil {
		t.Fatal("expected error when --enable-webhooks=true without --webhook-cert-path, got nil")
	}
	for _, want := range []string{"--webhook-cert-path", "--enable-webhooks"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error message must mention %q to be actionable, got: %v", want, err)
		}
	}
}

func TestValidateFlags_AcceptsEnabledWebhooksWithCertPath(t *testing.T) {
	cfg := &flagConfig{enableWebhooks: true, webhookCertPath: "/tmp/k8s-webhook-server/serving-certs"}

	if err := validateFlags(cfg); err != nil {
		t.Errorf("validateFlags should accept webhooks=true with cert path, got: %v", err)
	}
}

func TestValidateFlags_AcceptsDisabledWebhooksRegardlessOfCertPath(t *testing.T) {
	cases := []*flagConfig{
		{enableWebhooks: false, webhookCertPath: ""},
		{enableWebhooks: false, webhookCertPath: "/tmp/whatever"},
	}
	for _, cfg := range cases {
		if err := validateFlags(cfg); err != nil {
			t.Errorf("validateFlags should accept webhooks=false (cert-path=%q), got: %v",
				cfg.webhookCertPath, err)
		}
	}
}

func contains(xs []string, want string) bool {
	for _, x := range xs {
		if x == want {
			return true
		}
	}
	return false
}

func mustWriteFile(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
