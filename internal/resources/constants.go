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

// Port constants matching Java Constants.java
const (
	// MasterPort is the port for master-worker communication.
	MasterPort = 5557
	// MasterBindPort is the secondary port for master binding.
	MasterBindPort = 5558
	// WebUIPort is the Locust web UI port.
	WebUIPort = 8089
	// WorkerPort is the port exposed by worker nodes.
	WorkerPort = 8080
	// DefaultMetricsExporterPort is the default port for the Prometheus metrics exporter.
	DefaultMetricsExporterPort = 9646
)

// Mount path constants
const (
	// DefaultMountPath is the default path where ConfigMap is mounted.
	DefaultMountPath = "/lotest/src"
	// LibMountPath is the path where the lib ConfigMap is mounted.
	LibMountPath = "/opt/locust/lib"
)

// Label constants
const (
	// LabelTestName is the label key for the performance test name.
	LabelTestName = "performance-test-name"
	// LabelPodName is the label key for the pod name (used as service selector).
	LabelPodName = "performance-test-pod-name"
	// LabelManagedBy is the label key indicating the managing operator.
	LabelManagedBy = "managed-by"
	// ManagedByValue is the value for the managed-by label.
	ManagedByValue = "locust-k8s-operator"
	// LabelApp is the app label key.
	LabelApp = "app"
)

// Prometheus annotation constants
const (
	// AnnotationPrometheusScrape enables Prometheus scraping.
	AnnotationPrometheusScrape = "prometheus.io/scrape"
	// AnnotationPrometheusPath specifies the metrics endpoint path.
	AnnotationPrometheusPath = "prometheus.io/path"
	// AnnotationPrometheusPort specifies the metrics port.
	AnnotationPrometheusPort = "prometheus.io/port"
	// MetricsEndpointPath is the path for the metrics endpoint.
	MetricsEndpointPath = "/metrics"
)

// Job constants
const (
	// BackoffLimit is the number of retries before marking a job as failed.
	BackoffLimit = 0
	// MasterReplicaCount is the fixed replica count for master (always 1).
	MasterReplicaCount = 1
)

// Container constants
const (
	// MetricsExporterContainerName is the name of the metrics exporter sidecar.
	MetricsExporterContainerName = "locust-metrics-exporter"
	// LibVolumeName is the name of the lib volume.
	LibVolumeName = "locust-lib"
)

// Exporter environment variable constants
const (
	// ExporterURIEnvVar is the environment variable for the exporter URI.
	ExporterURIEnvVar = "LOCUST_EXPORTER_URI"
	// ExporterPortEnvVar is the environment variable for the exporter listen address.
	ExporterPortEnvVar = "LOCUST_EXPORTER_WEB_LISTEN_ADDRESS"
)

// Kafka environment variable constants
const (
	// EnvKafkaBootstrapServers is the Kafka bootstrap servers env var name.
	EnvKafkaBootstrapServers = "KAFKA_BOOTSTRAP_SERVERS"
	// EnvKafkaSecurityEnabled is the Kafka security enabled env var name.
	EnvKafkaSecurityEnabled = "KAFKA_SECURITY_ENABLED"
	// EnvKafkaSecurityProtocol is the Kafka security protocol env var name.
	EnvKafkaSecurityProtocol = "KAFKA_SECURITY_PROTOCOL_CONFIG"
	// EnvKafkaSaslMechanism is the Kafka SASL mechanism env var name.
	EnvKafkaSaslMechanism = "KAFKA_SASL_MECHANISM"
	// EnvKafkaSaslJaasConfig is the Kafka SASL JAAS config env var name.
	EnvKafkaSaslJaasConfig = "KAFKA_SASL_JAAS_CONFIG"
	// EnvKafkaUsername is the Kafka username env var name.
	EnvKafkaUsername = "KAFKA_USERNAME"
	// EnvKafkaPassword is the Kafka password env var name.
	EnvKafkaPassword = "KAFKA_PASSWORD"
)

// Service constants
const (
	// ProtocolTCP is the TCP protocol string.
	ProtocolTCP = "TCP"
	// PortNamePrefix is the prefix for port names.
	PortNamePrefix = "port"
	// MetricsPortName is the name for the metrics port.
	MetricsPortName = "prometheus-metrics"
)

// Node affinity constants
const (
	// DefaultNodeMatchExpressionOperator is the default operator for node selector requirements.
	DefaultNodeMatchExpressionOperator = "In"
)
