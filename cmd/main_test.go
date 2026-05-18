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
	"path/filepath"
	"strings"
	"testing"
	"time"
)

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

func TestApplyEnableWebhooksEnv_FlagWinsOverEnv(t *testing.T) {
	cfg := &flagConfig{enableWebhooks: false}
	logCount := 0
	logf := func(msg string, kv ...any) { logCount++ }

	applyEnableWebhooksEnv(cfg, "true", true, logf)

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

	applyEnableWebhooksEnv(cfg, "true", false, logf)

	if !cfg.enableWebhooks {
		t.Errorf("env var 'true' with no flag should set enableWebhooks=true")
	}
}

func TestApplyEnableWebhooksEnv_EnvAppliedWhenFlagUnset_False(t *testing.T) {
	cfg := &flagConfig{enableWebhooks: true}
	logf := func(string, ...any) {}

	applyEnableWebhooksEnv(cfg, "false", false, logf)

	if cfg.enableWebhooks {
		t.Errorf("env var 'false' with no flag should set enableWebhooks=false")
	}
}

func TestApplyEnableWebhooksEnv_LogsEvenWhenFlagWins(t *testing.T) {
	cfg := &flagConfig{enableWebhooks: true}
	logCount := 0
	var firstMsg string
	logf := func(msg string, kv ...any) {
		logCount++
		if logCount == 1 {
			firstMsg = msg
		}
	}

	applyEnableWebhooksEnv(cfg, "true", true, logf)

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
			applyEnableWebhooksEnv(cfg, v, false, func(string, ...any) {})
			if !cfg.enableWebhooks {
				t.Errorf("envVal=%q should enable webhooks", v)
			}
		})
	}
	for _, v := range falsyValues {
		t.Run("falsy/"+v, func(t *testing.T) {
			cfg := &flagConfig{enableWebhooks: true}
			applyEnableWebhooksEnv(cfg, v, false, func(string, ...any) {})
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
			applyEnableWebhooksEnv(cfg, v, false, logf)
			if cfg.enableWebhooks {
				t.Errorf("envVal=%q is not a valid boolean and must not flip the default; got enableWebhooks=true", v)
			}
			if !warned {
				t.Errorf("envVal=%q is invalid and must produce an 'ignoring' log line", v)
			}
		})
	}
}
