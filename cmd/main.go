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
	"crypto/tls"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/certwatcher"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/metrics/filters"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	locustv1 "github.com/AbdelrhmanHamouda/locust-k8s-operator/api/v1"
	locustv2 "github.com/AbdelrhmanHamouda/locust-k8s-operator/api/v2"
	"github.com/AbdelrhmanHamouda/locust-k8s-operator/internal/config"
	"github.com/AbdelrhmanHamouda/locust-k8s-operator/internal/controller"
	// +kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

const (
	envEnableWebhooks = "ENABLE_WEBHOOKS"

	// Throttles the "still waiting for certs" log line so a default 2-minute
	// wait does not produce 120 identical lines.
	certWaitLogInterval = 10 * time.Second
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(locustv1.AddToScheme(scheme))
	utilruntime.Must(locustv2.AddToScheme(scheme))
	// +kubebuilder:scaffold:scheme
}

func main() {
	flags, flagWasSet := parseFlags()
	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&flags.zapOpts), coloredConsoleEncoder()))

	// Empty/unset is ignored so chart users with blank `extraEnvs` entries
	// don't accidentally trip the deprecation path.
	if envVal := os.Getenv(envEnableWebhooks); envVal != "" {
		applyEnableWebhooksEnv(flags, envVal, flagWasSet, setupLog.Info)
	}

	if flags.enableWebhooks && flags.webhookCertPath == "" {
		setupLog.Error(nil, "--webhook-cert-path is required when --enable-webhooks=true "+
			"(the chart default is /tmp/k8s-webhook-server/serving-certs); "+
			"or set --enable-webhooks=false to run without admission webhooks")
		os.Exit(1)
	}

	tlsOpts := configureTLS(flags.enableHTTP2)

	// Leave WebhookServer unset when disabled: controller-runtime constructs
	// a dormant default that never starts unless GetWebhookServer() is called.
	var (
		webhookServer      webhook.Server
		webhookCertWatcher *certwatcher.CertWatcher
	)
	if flags.enableWebhooks {
		var err error
		webhookServer, webhookCertWatcher, err = setupWebhookServer(flags, tlsOpts)
		if err != nil {
			setupLog.Error(err, "unable to setup webhook server")
			os.Exit(1)
		}
	}
	metricsServerOptions, metricsCertWatcher := setupMetricsServer(flags, tlsOpts)

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		Metrics:                metricsServerOptions,
		WebhookServer:          webhookServer,
		HealthProbeBindAddress: flags.probeAddr,
		LeaderElection:         flags.enableLeaderElection,
		LeaderElectionID:       "locust-k8s-operator.locust.io",
		// LeaderElectionReleaseOnCancel defines if the leader should step down voluntarily
		// when the Manager ends. This requires the binary to immediately end when the
		// Manager is stopped, otherwise, this setting is unsafe. Setting this significantly
		// speeds up voluntary leader transitions as the new leader don't have to wait
		// LeaseDuration time first.
		//
		// In the default scaffold provided, the program ends immediately after
		// the manager stops, so would be fine to enable this option. However,
		// if you are doing or is intended to do any operation such as perform cleanups
		// after the manager stops then its usage might be unsafe.
		// LeaderElectionReleaseOnCancel: true,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	if err := registerControllersAndWebhooks(mgr, flags.enableWebhooks); err != nil {
		setupLog.Error(err, "failed to setup controllers")
		os.Exit(1)
	}

	if metricsCertWatcher != nil {
		setupLog.Info("Adding metrics certificate watcher to manager")
		if err := mgr.Add(metricsCertWatcher); err != nil {
			setupLog.Error(err, "unable to add metrics certificate watcher to manager")
			os.Exit(1)
		}
	}

	if webhookCertWatcher != nil {
		setupLog.Info("Adding webhook certificate watcher to manager")
		if err := mgr.Add(webhookCertWatcher); err != nil {
			setupLog.Error(err, "unable to add webhook certificate watcher to manager")
			os.Exit(1)
		}
	}

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}
	// Calling mgr.GetWebhookServer() is only safe in the enabled branch:
	// the call itself adds the webhook server as a manager runnable.
	if flags.enableWebhooks {
		if err := mgr.AddReadyzCheck("webhook", mgr.GetWebhookServer().StartedChecker()); err != nil {
			setupLog.Error(err, "unable to set up webhook ready check")
			os.Exit(1)
		}
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}

type flagConfig struct {
	metricsAddr            string
	metricsCertPath        string
	metricsCertName        string
	metricsCertKey         string
	webhookCertPath        string
	webhookCertName        string
	webhookCertKey         string
	webhookCertWaitTimeout time.Duration
	probeAddr              string
	enableLeaderElection   bool
	enableWebhooks         bool
	secureMetrics          bool
	enableHTTP2            bool
	zapOpts                zap.Options
}

// parseFlags returns the parsed configuration plus a bool indicating whether
// --enable-webhooks was explicitly set on the command line, so the env-var
// fallback (applied later, once the logger is up) can honour
// "explicit flag wins over env var".
func parseFlags() (*flagConfig, bool) {
	cfg := &flagConfig{}

	flag.StringVar(&cfg.metricsAddr, "metrics-bind-address", "0", "The address the metrics endpoint binds to. "+
		"Use :8443 for HTTPS or :8080 for HTTP, or leave as 0 to disable the metrics service.")
	flag.StringVar(&cfg.probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&cfg.enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.BoolVar(&cfg.secureMetrics, "metrics-secure", true,
		"If set, the metrics endpoint is served securely via HTTPS. Use --metrics-secure=false to use HTTP instead.")
	flag.BoolVar(&cfg.enableWebhooks, "enable-webhooks", false,
		"Enable conversion and validation webhooks. When true, --webhook-cert-path is required. "+
			"Default false matches the Helm chart default (webhook.enabled).")
	flag.StringVar(&cfg.webhookCertPath, "webhook-cert-path", "", "The directory that contains the webhook certificate.")
	flag.StringVar(&cfg.webhookCertName, "webhook-cert-name", "tls.crt", "The name of the webhook certificate file.")
	flag.StringVar(&cfg.webhookCertKey, "webhook-cert-key", "tls.key", "The name of the webhook key file.")
	flag.DurationVar(&cfg.webhookCertWaitTimeout, "webhook-cert-wait-timeout", 2*time.Minute,
		"Maximum time to wait for webhook cert files to appear before failing. "+
			"Set to 0 to wait forever (legacy behaviour).")
	flag.StringVar(&cfg.metricsCertPath, "metrics-cert-path", "",
		"The directory that contains the metrics server certificate.")
	flag.StringVar(&cfg.metricsCertName, "metrics-cert-name", "tls.crt",
		"The name of the metrics server certificate file.")
	flag.StringVar(&cfg.metricsCertKey, "metrics-cert-key", "tls.key", "The name of the metrics server key file.")
	flag.BoolVar(&cfg.enableHTTP2, "enable-http2", false,
		"If set, HTTP/2 will be enabled for the metrics and webhook servers")

	cfg.zapOpts = zap.Options{
		Development: false,
	}
	cfg.zapOpts.BindFlags(flag.CommandLine)
	flag.Parse()

	var flagWasSet bool
	flag.Visit(func(f *flag.Flag) {
		if f.Name == "enable-webhooks" {
			flagWasSet = true
		}
	})

	return cfg, flagWasSet
}

// applyEnableWebhooksEnv resolves the deprecated ENABLE_WEBHOOKS env var
// against cfg. Explicit --enable-webhooks always wins; only standard
// strconv.ParseBool forms set cfg.enableWebhooks. Anything else is ignored
// with a warning — the prior `envVal != "false"` check silently enabled
// webhooks on every typo.
func applyEnableWebhooksEnv(cfg *flagConfig, envVal string, flagWasSet bool, logf func(msg string, keysAndValues ...any)) {
	logf("ENABLE_WEBHOOKS env var is deprecated, use --enable-webhooks=true|false",
		"removalRelease", "v2.3.0")
	if flagWasSet {
		return
	}
	parsed, err := strconv.ParseBool(envVal)
	if err != nil {
		logf("ignoring ENABLE_WEBHOOKS: not a valid boolean", "value", envVal)
		return
	}
	cfg.enableWebhooks = parsed
}

// configureTLS creates TLS options based on HTTP/2 setting
func configureTLS(enableHTTP2 bool) []func(*tls.Config) {
	var tlsOpts []func(*tls.Config)

	tlsOpts = append(tlsOpts, func(c *tls.Config) {
		c.MinVersion = tls.VersionTLS12
	})

	// if the enable-http2 flag is false (the default), http/2 should be disabled
	// due to its vulnerabilities. More specifically, disabling http/2 will
	// prevent from being vulnerable to the HTTP/2 Stream Cancellation and
	// Rapid Reset CVEs. For more information see:
	// - https://github.com/advisories/GHSA-qppj-fm5r-hxr3
	// - https://github.com/advisories/GHSA-4374-p667-p6c8
	if !enableHTTP2 {
		disableHTTP2 := func(c *tls.Config) {
			setupLog.Info("disabling http/2")
			c.NextProtos = []string{"http/1.1"}
		}
		tlsOpts = append(tlsOpts, disableHTTP2)
	}

	return tlsOpts
}

// setupWebhookServer creates a webhook server with an optional certificate
// watcher. Only call when webhooks are enabled.
func setupWebhookServer(
	flags *flagConfig,
	tlsOpts []func(*tls.Config),
) (webhook.Server, *certwatcher.CertWatcher, error) {
	webhookTLSOpts := tlsOpts
	var webhookCertWatcher *certwatcher.CertWatcher

	if len(flags.webhookCertPath) > 0 {
		certFile := filepath.Join(flags.webhookCertPath, flags.webhookCertName)
		keyFile := filepath.Join(flags.webhookCertPath, flags.webhookCertKey)

		setupLog.Info("Waiting for webhook certificate files",
			"cert", certFile, "key", keyFile,
			"timeout", flags.webhookCertWaitTimeout)

		if err := waitForWebhookCerts(certFile, keyFile, flags.webhookCertWaitTimeout); err != nil {
			return nil, nil, fmt.Errorf("webhook certificate files not available: %w", err)
		}

		var err error
		webhookCertWatcher, err = certwatcher.New(certFile, keyFile)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to initialize webhook certificate watcher: %w", err)
		}

		webhookTLSOpts = append(webhookTLSOpts, func(config *tls.Config) {
			config.GetCertificate = webhookCertWatcher.GetCertificate
		})
	}

	return webhook.NewServer(webhook.Options{
		TLSOpts: webhookTLSOpts,
	}), webhookCertWatcher, nil
}

// waitForWebhookCerts polls until both cert files exist or the timeout elapses.
// timeout==0 means wait forever (legacy behaviour, opt-in via flag).
func waitForWebhookCerts(certFile, keyFile string, timeout time.Duration) error {
	deadline := time.Time{}
	if timeout > 0 {
		deadline = time.Now().Add(timeout)
	}
	start := time.Now()
	lastLog := time.Time{}
	for {
		_, certErr := os.Stat(certFile)
		_, keyErr := os.Stat(keyFile)
		if certErr == nil && keyErr == nil {
			return nil
		}
		if !deadline.IsZero() && time.Now().After(deadline) {
			return fmt.Errorf("webhook cert files (%s, %s) not present after %s — "+
				"is cert-manager installed and is the Certificate resource Ready? "+
				"Check `kubectl describe certificate -A` and `kubectl -n cert-manager logs deploy/cert-manager`",
				certFile, keyFile, timeout)
		}
		if lastLog.IsZero() || time.Since(lastLog) >= certWaitLogInterval {
			setupLog.Info("Webhook certificate files not ready, polling",
				"elapsed", time.Since(start).Round(time.Second),
				"timeout", timeout)
			lastLog = time.Now()
		}
		time.Sleep(time.Second)
	}
}

// setupMetricsServer creates metrics server options with optional certificate watcher
func setupMetricsServer(flags *flagConfig, tlsOpts []func(*tls.Config)) (metricsserver.Options, *certwatcher.CertWatcher) { //nolint:lll
	// Metrics endpoint is enabled in 'config/default/kustomization.yaml'. The Metrics options configure the server.
	// More info:
	// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.21.0/pkg/metrics/server
	// - https://book.kubebuilder.io/reference/metrics.html
	metricsServerOptions := metricsserver.Options{
		BindAddress:   flags.metricsAddr,
		SecureServing: flags.secureMetrics,
		TLSOpts:       tlsOpts,
	}

	if flags.secureMetrics {
		// FilterProvider is used to protect the metrics endpoint with authn/authz.
		// These configurations ensure that only authorized users and service accounts
		// can access the metrics endpoint. The RBAC are configured in 'config/rbac/kustomization.yaml'. More info:
		// https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.21.0/pkg/metrics/filters#WithAuthenticationAndAuthorization
		metricsServerOptions.FilterProvider = filters.WithAuthenticationAndAuthorization
	}

	var metricsCertWatcher *certwatcher.CertWatcher

	// If the certificate is not specified, controller-runtime will automatically
	// generate self-signed certificates for the metrics server. While convenient for development and testing,
	// this setup is not recommended for production.
	//
	// TODO(user): If you enable certManager, uncomment the following lines:
	// - [METRICS-WITH-CERTS] at config/default/kustomization.yaml to generate and use certificates
	// managed by cert-manager for the metrics server.
	// - [PROMETHEUS-WITH-CERTS] at config/prometheus/kustomization.yaml for TLS certification.
	if len(flags.metricsCertPath) > 0 {
		setupLog.Info("Initializing metrics certificate watcher using provided certificates",
			"metrics-cert-path", flags.metricsCertPath,
			"metrics-cert-name", flags.metricsCertName,
			"metrics-cert-key", flags.metricsCertKey)

		var err error
		metricsCertWatcher, err = certwatcher.New(
			filepath.Join(flags.metricsCertPath, flags.metricsCertName),
			filepath.Join(flags.metricsCertPath, flags.metricsCertKey),
		)
		if err != nil {
			setupLog.Error(err, "Failed to initialize metrics certificate watcher")
			os.Exit(1)
		}

		metricsServerOptions.TLSOpts = append(metricsServerOptions.TLSOpts, func(config *tls.Config) {
			config.GetCertificate = metricsCertWatcher.GetCertificate
		})
	}

	return metricsServerOptions, metricsCertWatcher
}

// registerControllersAndWebhooks registers the LocustTest reconciler and,
// when enableWebhooks is true, the v1 conversion + v2 validation webhooks.
//
// When enableWebhooks is false the body MUST NOT call SetupWebhookWithManager
// or mgr.GetWebhookServer() — either call adds the webhook server as a
// manager runnable, after which the manager tries to load TLS certs from the
// default temp dir.
func registerControllersAndWebhooks(mgr ctrl.Manager, enableWebhooks bool) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load operator configuration: %w", err)
	}
	setupLog.Info("Operator configuration loaded",
		"ttlSecondsAfterFinished", cfg.TTLSecondsAfterFinished,
		"metricsExporterImage", cfg.MetricsExporterImage,
		"affinityInjection", cfg.EnableAffinityCRInjection,
		"tolerationsInjection", cfg.EnableTolerationsCRInjection)

	if err := (&controller.LocustTestReconciler{
		Client:   mgr.GetClient(),
		Scheme:   mgr.GetScheme(),
		Config:   cfg,
		Recorder: mgr.GetEventRecorderFor("locusttest-controller"),
	}).SetupWithManager(mgr); err != nil {
		return fmt.Errorf("unable to create controller LocustTest: %w", err)
	}

	if enableWebhooks {
		if err := (&locustv1.LocustTest{}).SetupWebhookWithManager(mgr); err != nil {
			return fmt.Errorf("unable to create webhook LocustTest v1: %w", err)
		}
		if err := (&locustv2.LocustTest{}).SetupWebhookWithManager(mgr); err != nil {
			return fmt.Errorf("unable to create webhook LocustTest v2: %w", err)
		}
	}
	// +kubebuilder:scaffold:builder

	return nil
}
