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
	"sync/atomic"
	"testing"
	"time"

	"sigs.k8s.io/controller-runtime/pkg/healthz"
)

// fakeReadyzAdder is a test double for ctrl.Manager's healthz/readyz wiring.
// It records every (name, kind) tuple so tests can assert exactly which checks
// were registered. The "no GetWebhookServer call" invariant from #317 is
// asserted by ensuring the recorder never sees ("webhook", readyz) when
// enableWebhooks=false.
type fakeReadyzAdder struct {
	healthChecks []string
	readyChecks  []string
	addErr       error
}

func (f *fakeReadyzAdder) AddHealthzCheck(name string, _ healthz.Checker) error {
	if f.addErr != nil {
		return f.addErr
	}
	f.healthChecks = append(f.healthChecks, name)
	return nil
}

func (f *fakeReadyzAdder) AddReadyzCheck(name string, _ healthz.Checker) error {
	if f.addErr != nil {
		return f.addErr
	}
	f.readyChecks = append(f.readyChecks, name)
	return nil
}

// sentinelChecker returns nil; used to verify nil vs non-nil behaviour without
// invoking real webhook plumbing.
func sentinelChecker(_ *http.Request) error { return nil }

// -----------------------------------------------------------------------------
// addHealthChecks — the regression-prevention surface
// -----------------------------------------------------------------------------

// TestAddHealthChecks_DisabledSkipsWebhookCheck pins the structural invariant
// behind issue #317: when webhooks are disabled, the readyz registration MUST
// NOT include the "webhook" check. If it does, the manager calls
// GetWebhookServer() (the side-effect that adds the webhook server as a
// runnable), which then tries to load TLS certs from the default temp dir and
// crashes the operator on Helm-default installs.
func TestAddHealthChecks_DisabledSkipsWebhookCheck(t *testing.T) {
	rec := &fakeReadyzAdder{}

	if err := addHealthChecks(rec, false, healthz.Checker(sentinelChecker)); err != nil {
		t.Fatalf("addHealthChecks returned error: %v", err)
	}

	for _, name := range rec.readyChecks {
		if name == "webhook" {
			t.Fatalf("readyz check %q must not be registered when webhooks are disabled "+
				"(see #317 — registering it forces the webhook server to start, which "+
				"requires certs that the chart deliberately doesn't mount)", name)
		}
	}
	if !contains(rec.readyChecks, "readyz") {
		t.Errorf("expected baseline readyz check, got %v", rec.readyChecks)
	}
	if !contains(rec.healthChecks, "healthz") {
		t.Errorf("expected baseline healthz check, got %v", rec.healthChecks)
	}
}

// TestAddHealthChecks_DisabledIgnoresNonNilChecker is defense in depth: even
// when a webhook checker is supplied (e.g., a future caller forgets to gate
// the construction), addHealthChecks itself must drop it when enableWebhooks
// is false.
func TestAddHealthChecks_DisabledIgnoresNonNilChecker(t *testing.T) {
	rec := &fakeReadyzAdder{}

	if err := addHealthChecks(rec, false, healthz.Checker(sentinelChecker)); err != nil {
		t.Fatalf("addHealthChecks returned error: %v", err)
	}

	for _, name := range rec.readyChecks {
		if name == "webhook" {
			t.Fatalf("addHealthChecks must drop the webhook checker when enableWebhooks=false")
		}
	}
}

// TestAddHealthChecks_EnabledRegistersWebhookCheck preserves the original
// motivation of commit 9ae57702: when webhooks ARE enabled and a started
// checker is supplied, the readyz block must include "webhook" so the pod's
// Endpoints don't go green before admission requests will succeed.
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

// TestAddHealthChecks_EnabledNilCheckerSkipped covers the edge where webhooks
// are flagged on but no checker was constructed (e.g., manager creation
// failed earlier). The function should not panic and should not register a
// nil checker.
func TestAddHealthChecks_EnabledNilCheckerSkipped(t *testing.T) {
	rec := &fakeReadyzAdder{}

	if err := addHealthChecks(rec, true, nil); err != nil {
		t.Fatalf("addHealthChecks returned error: %v", err)
	}

	for _, name := range rec.readyChecks {
		if name == "webhook" {
			t.Fatalf("addHealthChecks must not register a nil webhook checker")
		}
	}
}

// TestAddHealthChecks_PropagatesErrors confirms the function surfaces errors
// from the underlying readyz registration.
func TestAddHealthChecks_PropagatesErrors(t *testing.T) {
	want := errors.New("boom")
	rec := &fakeReadyzAdder{addErr: want}

	err := addHealthChecks(rec, false, nil)
	if !errors.Is(err, want) {
		t.Errorf("expected wrapped error %v, got %v", want, err)
	}
}

// -----------------------------------------------------------------------------
// waitForWebhookCerts — bounded poll, prevents silent hang
// -----------------------------------------------------------------------------

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

// -----------------------------------------------------------------------------
// applyEnableWebhooksEnv — deprecation alias precedence
// -----------------------------------------------------------------------------

func TestApplyEnableWebhooksEnv_FlagWinsOverEnv(t *testing.T) {
	cfg := &flagConfig{enableWebhooks: false} // flag was set to false
	logCount := int32(0)
	logf := func(msg string, kv ...any) { atomic.AddInt32(&logCount, 1) }

	applyEnableWebhooksEnv(cfg, "true", true, logf)

	if cfg.enableWebhooks {
		t.Errorf("flag must win over env var, got enableWebhooks=true")
	}
	if atomic.LoadInt32(&logCount) != 1 {
		t.Errorf("expected exactly one deprecation log, got %d", logCount)
	}
}

func TestApplyEnableWebhooksEnv_EnvAppliedWhenFlagUnset_True(t *testing.T) {
	cfg := &flagConfig{enableWebhooks: false}
	logf := func(string, ...any) {}

	applyEnableWebhooksEnv(cfg, "true", false, logf)

	if !cfg.enableWebhooks {
		t.Errorf("env var 'true' with no flag should set enableWebhooks=true")
	}
}

func TestApplyEnableWebhooksEnv_EnvAppliedWhenFlagUnset_False(t *testing.T) {
	cfg := &flagConfig{enableWebhooks: true} // imagine the binary default flipped
	logf := func(string, ...any) {}

	applyEnableWebhooksEnv(cfg, "false", false, logf)

	if cfg.enableWebhooks {
		t.Errorf("env var 'false' with no flag should set enableWebhooks=false")
	}
}

func TestApplyEnableWebhooksEnv_LogsEvenWhenFlagWins(t *testing.T) {
	cfg := &flagConfig{enableWebhooks: true}
	logCount := int32(0)
	logf := func(msg string, kv ...any) {
		atomic.AddInt32(&logCount, 1)
		if !strings.Contains(msg, "deprecated") {
			t.Errorf("deprecation log must mention 'deprecated', got %q", msg)
		}
	}

	applyEnableWebhooksEnv(cfg, "true", true, logf)

	if atomic.LoadInt32(&logCount) != 1 {
		t.Errorf("user with stale env var should still see the deprecation warning, got %d logs", logCount)
	}
}

// -----------------------------------------------------------------------------
// Helpers
// -----------------------------------------------------------------------------

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
