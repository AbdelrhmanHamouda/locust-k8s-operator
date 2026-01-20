{{/*
=============================================================================
LOCUST K8S OPERATOR - HELM TEMPLATE HELPERS
=============================================================================
This file contains reusable template helpers for the Locust K8s Operator chart.

Sections:
  1. Standard Naming Helpers - Chart name, fullname, labels
  2. Backward Compatibility Helpers - Map old value paths to new paths
  3. Environment Variables Helper - Generate env vars for the operator

For users upgrading from v1.x charts:
  The backward compatibility helpers allow you to continue using old value
  paths (e.g., config.loadGenerationPods.resource.cpuRequest) while we
  recommend migrating to the new paths (e.g., locustPods.resources.requests.cpu).
=============================================================================
*/}}

{{/*
=============================================================================
SECTION 1: Standard Naming Helpers
=============================================================================
*/}}

{{/*
Expand the name of the chart.
Used for: container names, label values
Override with: .Values.nameOverride
*/}}
{{- define "locust-k8s-operator.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
Used for: deployment name, service account name, RBAC resources
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "locust-k8s-operator.fullname" -}}
{{- $name := default .Chart.Name }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
Used for: helm.sh/chart label to track which chart version deployed the resources
*/}}
{{- define "locust-k8s-operator.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels applied to all resources.
Includes: chart info, selector labels, version, and managed-by
*/}}
{{- define "locust-k8s-operator.labels" -}}
helm.sh/chart: {{ include "locust-k8s-operator.chart" . }}
{{ include "locust-k8s-operator.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels used for pod selection.
These labels are used by Deployments to select their pods and by Services to route traffic.
Must remain consistent across upgrades to avoid orphaned pods.
*/}}
{{- define "locust-k8s-operator.selectorLabels" -}}
app.kubernetes.io/name: {{ include "locust-k8s-operator.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use.
If serviceAccount.create is true, uses fullname or custom name.
If serviceAccount.create is false, uses the specified name or "default".
*/}}
{{- define "locust-k8s-operator.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "locust-k8s-operator.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
=============================================================================
SECTION 2: Backward Compatibility Helpers
=============================================================================
These helpers map old v1.x chart value paths to new v2.x paths.
This allows users upgrading from v1.x to continue using their existing
values.yaml files while migrating to the new structure.

Priority order:
  1. New path (locustPods.*, kafka.*, etc.)
  2. Old path (config.loadGenerationPods.*, etc.)
  3. Default value

Example - both of these work:
  # New (recommended)
  locustPods:
    resources:
      requests:
        cpu: 500m

  # Old (deprecated, still works)
  config:
    loadGenerationPods:
      resource:
        cpuRequest: 500m
=============================================================================
*/}}

{{/*
Pod CPU Request - new path with fallback to old path
*/}}
{{- define "locust.podCpuRequest" -}}
{{- if .Values.locustPods }}
{{- .Values.locustPods.resources.requests.cpu | default "250m" }}
{{- else if .Values.config }}
{{- .Values.config.loadGenerationPods.resource.cpuRequest | default "250m" }}
{{- else }}
{{- "250m" }}
{{- end }}
{{- end }}

{{/*
Pod Memory Request - new path with fallback to old path
*/}}
{{- define "locust.podMemRequest" -}}
{{- if .Values.locustPods }}
{{- .Values.locustPods.resources.requests.memory | default "128Mi" }}
{{- else if .Values.config }}
{{- .Values.config.loadGenerationPods.resource.memRequest | default "128Mi" }}
{{- else }}
{{- "128Mi" }}
{{- end }}
{{- end }}

{{/*
Pod Ephemeral Storage Request - new path with fallback to old path
*/}}
{{- define "locust.podEphemeralRequest" -}}
{{- if .Values.locustPods }}
{{- .Values.locustPods.resources.requests.ephemeralStorage | default "30M" }}
{{- else if .Values.config }}
{{- .Values.config.loadGenerationPods.resource.ephemeralRequest | default "30M" }}
{{- else }}
{{- "30M" }}
{{- end }}
{{- end }}

{{/*
Pod CPU Limit - new path with fallback to old path
*/}}
{{- define "locust.podCpuLimit" -}}
{{- if .Values.locustPods }}
{{- .Values.locustPods.resources.limits.cpu | default "1000m" }}
{{- else if .Values.config }}
{{- .Values.config.loadGenerationPods.resource.cpuLimit | default "1000m" }}
{{- else }}
{{- "1000m" }}
{{- end }}
{{- end }}

{{/*
Pod Memory Limit - new path with fallback to old path
*/}}
{{- define "locust.podMemLimit" -}}
{{- if .Values.locustPods }}
{{- .Values.locustPods.resources.limits.memory | default "1024Mi" }}
{{- else if .Values.config }}
{{- .Values.config.loadGenerationPods.resource.memLimit | default "1024Mi" }}
{{- else }}
{{- "1024Mi" }}
{{- end }}
{{- end }}

{{/*
Pod Ephemeral Storage Limit - new path with fallback to old path
*/}}
{{- define "locust.podEphemeralLimit" -}}
{{- if .Values.locustPods }}
{{- .Values.locustPods.resources.limits.ephemeralStorage | default "50M" }}
{{- else if .Values.config }}
{{- .Values.config.loadGenerationPods.resource.ephemeralLimit | default "50M" }}
{{- else }}
{{- "50M" }}
{{- end }}
{{- end }}

{{/*
Affinity Injection - new path with fallback to old path
*/}}
{{- define "locust.affinityInjection" -}}
{{- if .Values.locustPods }}
{{- .Values.locustPods.affinityInjection | default true }}
{{- else if .Values.config }}
{{- .Values.config.loadGenerationPods.affinity.enableCrInjection | default true }}
{{- else }}
{{- true }}
{{- end }}
{{- end }}

{{/*
Tolerations Injection - new path with fallback to old path
*/}}
{{- define "locust.tolerationsInjection" -}}
{{- if .Values.locustPods }}
{{- .Values.locustPods.tolerationsInjection | default true }}
{{- else if .Values.config }}
{{- .Values.config.loadGenerationPods.taintTolerations.enableCrInjection | default true }}
{{- else }}
{{- true }}
{{- end }}
{{- end }}

{{/*
TTL Seconds After Finished - new path with fallback to old path
*/}}
{{- define "locust.ttlSecondsAfterFinished" -}}
{{- if .Values.locustPods }}
{{- .Values.locustPods.ttlSecondsAfterFinished | default "" }}
{{- else if .Values.config }}
{{- .Values.config.loadGenerationJobs.ttlSecondsAfterFinished | default "" }}
{{- else }}
{{- "" }}
{{- end }}
{{- end }}

{{/*
Metrics Exporter Image - new path with fallback to old path
*/}}
{{- define "locust.metricsExporterImage" -}}
{{- if .Values.locustPods }}
{{- .Values.locustPods.metricsExporter.image | default "containersol/locust_exporter:v0.5.0" }}
{{- else if .Values.config }}
{{- .Values.config.loadGenerationPods.metricsExporter.image | default "containersol/locust_exporter:v0.5.0" }}
{{- else }}
{{- "containersol/locust_exporter:v0.5.0" }}
{{- end }}
{{- end }}

{{/*
Metrics Exporter Port - new path with fallback to old path
*/}}
{{- define "locust.metricsExporterPort" -}}
{{- if .Values.locustPods }}
{{- .Values.locustPods.metricsExporter.port | default 9646 }}
{{- else if .Values.config }}
{{- .Values.config.loadGenerationPods.metricsExporter.port | default 9646 }}
{{- else }}
{{- 9646 }}
{{- end }}
{{- end }}

{{/*
Metrics Exporter Pull Policy - new path with fallback to old path
*/}}
{{- define "locust.metricsExporterPullPolicy" -}}
{{- if .Values.locustPods }}
{{- .Values.locustPods.metricsExporter.pullPolicy | default "IfNotPresent" }}
{{- else if .Values.config }}
{{- .Values.config.loadGenerationPods.metricsExporter.pullPolicy | default "IfNotPresent" }}
{{- else }}
{{- "IfNotPresent" }}
{{- end }}
{{- end }}

{{/*
Metrics Exporter CPU Request
*/}}
{{- define "locust.metricsExporterCpuRequest" -}}
{{- if .Values.locustPods }}
{{- .Values.locustPods.metricsExporter.resources.requests.cpu | default "100m" }}
{{- else if .Values.config }}
{{- .Values.config.loadGenerationPods.metricsExporter.resource.cpuRequest | default "100m" }}
{{- else }}
{{- "100m" }}
{{- end }}
{{- end }}

{{/*
Metrics Exporter Memory Request
*/}}
{{- define "locust.metricsExporterMemRequest" -}}
{{- if .Values.locustPods }}
{{- .Values.locustPods.metricsExporter.resources.requests.memory | default "64Mi" }}
{{- else if .Values.config }}
{{- .Values.config.loadGenerationPods.metricsExporter.resource.memRequest | default "64Mi" }}
{{- else }}
{{- "64Mi" }}
{{- end }}
{{- end }}

{{/*
Metrics Exporter Ephemeral Request
*/}}
{{- define "locust.metricsExporterEphemeralRequest" -}}
{{- if .Values.locustPods }}
{{- .Values.locustPods.metricsExporter.resources.requests.ephemeralStorage | default "30M" }}
{{- else if .Values.config }}
{{- .Values.config.loadGenerationPods.metricsExporter.resource.ephemeralRequest | default "30M" }}
{{- else }}
{{- "30M" }}
{{- end }}
{{- end }}

{{/*
Metrics Exporter CPU Limit
*/}}
{{- define "locust.metricsExporterCpuLimit" -}}
{{- if .Values.locustPods }}
{{- .Values.locustPods.metricsExporter.resources.limits.cpu | default "250m" }}
{{- else if .Values.config }}
{{- .Values.config.loadGenerationPods.metricsExporter.resource.cpuLimit | default "250m" }}
{{- else }}
{{- "250m" }}
{{- end }}
{{- end }}

{{/*
Metrics Exporter Memory Limit
*/}}
{{- define "locust.metricsExporterMemLimit" -}}
{{- if .Values.locustPods }}
{{- .Values.locustPods.metricsExporter.resources.limits.memory | default "128Mi" }}
{{- else if .Values.config }}
{{- .Values.config.loadGenerationPods.metricsExporter.resource.memLimit | default "128Mi" }}
{{- else }}
{{- "128Mi" }}
{{- end }}
{{- end }}

{{/*
Metrics Exporter Ephemeral Limit
*/}}
{{- define "locust.metricsExporterEphemeralLimit" -}}
{{- if .Values.locustPods }}
{{- .Values.locustPods.metricsExporter.resources.limits.ephemeralStorage | default "50M" }}
{{- else if .Values.config }}
{{- .Values.config.loadGenerationPods.metricsExporter.resource.ephemeralLimit | default "50M" }}
{{- else }}
{{- "50M" }}
{{- end }}
{{- end }}

{{/*
Kafka Bootstrap Servers - new path with fallback to old path
*/}}
{{- define "locust.kafkaBootstrapServers" -}}
{{- if .Values.kafka }}
{{- .Values.kafka.bootstrapServers | default "localhost:9092" }}
{{- else if .Values.config }}
{{- .Values.config.loadGenerationPods.kafka.bootstrapServers | default "localhost:9092" }}
{{- else }}
{{- "localhost:9092" }}
{{- end }}
{{- end }}

{{/*
Kafka Security Enabled - new path with fallback to old path
*/}}
{{- define "locust.kafkaSecurityEnabled" -}}
{{- if .Values.kafka }}
{{- .Values.kafka.security.enabled | default false }}
{{- else if .Values.config }}
{{- .Values.config.loadGenerationPods.kafka.acl.enabled | default false }}
{{- else }}
{{- false }}
{{- end }}
{{- end }}

{{/*
=============================================================================
SECTION 3: Environment Variables Helper
=============================================================================
Generates all environment variables for the operator container.
These env vars configure how the operator creates Locust test pods.

The operator reads these at startup and uses them as defaults when
creating master/worker Jobs for LocustTest CRs.

Categories:
  - Resource limits: CPU, memory, ephemeral storage for Locust pods
  - Feature flags: Affinity injection, tolerations injection
  - Metrics exporter: Sidecar container configuration
  - Kafka: Message queue configuration (deprecated)
=============================================================================
*/}}
{{- define "locust-k8s-operator.envVars" -}}
# Webhook configuration
# Controls whether the operator registers conversion and validation webhooks
- name: ENABLE_WEBHOOKS
  value: {{ .Values.webhook.enabled | quote }}
# Resource limits for Locust test pods (master and workers)
# These define the default resources when not specified in the LocustTest CR
- name: POD_CPU_REQUEST
  value: {{ include "locust.podCpuRequest" . | quote }}
- name: POD_MEM_REQUEST
  value: {{ include "locust.podMemRequest" . | quote }}
- name: POD_EPHEMERAL_REQUEST
  value: {{ include "locust.podEphemeralRequest" . | quote }}
- name: POD_CPU_LIMIT
  value: {{ include "locust.podCpuLimit" . | quote }}
- name: POD_MEM_LIMIT
  value: {{ include "locust.podMemLimit" . | quote }}
- name: POD_EPHEMERAL_LIMIT
  value: {{ include "locust.podEphemeralLimit" . | quote }}
# Feature flags for pod scheduling
# When enabled, the operator injects affinity/tolerations from the CR into pods
- name: ENABLE_AFFINITY_CR_INJECTION
  value: {{ include "locust.affinityInjection" . | quote }}
- name: ENABLE_TAINT_TOLERATIONS_CR_INJECTION
  value: {{ include "locust.tolerationsInjection" . | quote }}
# Metrics exporter sidecar configuration
# This Prometheus exporter runs alongside the Locust master to expose metrics
# Note: Not used when OpenTelemetry is enabled (OTel replaces the sidecar)
- name: METRICS_EXPORTER_IMAGE
  value: {{ include "locust.metricsExporterImage" . | quote }}
- name: METRICS_EXPORTER_PORT
  value: {{ include "locust.metricsExporterPort" . | quote }}
- name: METRICS_EXPORTER_IMAGE_PULL_POLICY
  value: {{ include "locust.metricsExporterPullPolicy" . | quote }}
- name: METRICS_EXPORTER_CPU_REQUEST
  value: {{ include "locust.metricsExporterCpuRequest" . | quote }}
- name: METRICS_EXPORTER_MEM_REQUEST
  value: {{ include "locust.metricsExporterMemRequest" . | quote }}
- name: METRICS_EXPORTER_EPHEMERAL_REQUEST
  value: {{ include "locust.metricsExporterEphemeralRequest" . | quote }}
- name: METRICS_EXPORTER_CPU_LIMIT
  value: {{ include "locust.metricsExporterCpuLimit" . | quote }}
- name: METRICS_EXPORTER_MEM_LIMIT
  value: {{ include "locust.metricsExporterMemLimit" . | quote }}
- name: METRICS_EXPORTER_EPHEMERAL_LIMIT
  value: {{ include "locust.metricsExporterEphemeralLimit" . | quote }}
# Job TTL - automatically clean up completed Jobs after this many seconds
# If not set, Jobs remain until manually deleted or CR is deleted
{{- $ttl := include "locust.ttlSecondsAfterFinished" . }}
{{- if $ttl }}
- name: JOB_TTL_SECONDS_AFTER_FINISHED
  value: {{ $ttl | quote }}
{{- end }}
# Kafka configuration (DEPRECATED - kept for backward compatibility)
# Consider using OpenTelemetry for metrics export instead
- name: KAFKA_BOOTSTRAP_SERVERS
  value: {{ include "locust.kafkaBootstrapServers" . | quote }}
- name: KAFKA_SECURITY_ENABLED
  value: {{ include "locust.kafkaSecurityEnabled" . | quote }}
{{- if or (and .Values.kafka .Values.kafka.security.enabled) (and .Values.config .Values.config.loadGenerationPods.kafka.acl.enabled) }}
- name: KAFKA_SECURITY_PROTOCOL_CONFIG
  value: {{ .Values.kafka.security.protocol | default .Values.config.loadGenerationPods.kafka.acl.protocol | default "SASL_PLAINTEXT" | quote }}
- name: KAFKA_SASL_MECHANISM
  value: {{ .Values.kafka.security.saslMechanism | default .Values.config.loadGenerationPods.kafka.sasl.mechanism | default "SCRAM-SHA-512" | quote }}
{{- if .Values.kafka.security.jaasConfig }}
- name: KAFKA_SASL_JAAS_CONFIG
  value: {{ .Values.kafka.security.jaasConfig | quote }}
{{- else if .Values.config }}
- name: KAFKA_SASL_JAAS_CONFIG
  value: {{ .Values.config.loadGenerationPods.kafka.sasl.jaas.config | quote }}
{{- end }}
{{- if .Values.kafka.credentials.secretName }}
- name: KAFKA_USERNAME
  valueFrom:
    secretKeyRef:
      name: {{ .Values.kafka.credentials.secretName }}
      key: {{ .Values.kafka.credentials.usernameKey | default "username" }}
- name: KAFKA_PASSWORD
  valueFrom:
    secretKeyRef:
      name: {{ .Values.kafka.credentials.secretName }}
      key: {{ .Values.kafka.credentials.passwordKey | default "password" }}
{{- else if .Values.config }}
- name: KAFKA_USERNAME
  valueFrom:
    secretKeyRef:
      name: {{ .Values.config.loadGenerationPods.kafka.locustK8sKafkaUser.userName }}
      key: {{ .Values.config.loadGenerationPods.kafka.acl.secret.userKey }}
- name: KAFKA_PASSWORD
  valueFrom:
    secretKeyRef:
      name: {{ .Values.config.loadGenerationPods.kafka.locustK8sKafkaUser.userName }}
      key: {{ .Values.config.loadGenerationPods.kafka.acl.secret.passwordKey }}
{{- end }}
{{- end }}
{{- end }}
