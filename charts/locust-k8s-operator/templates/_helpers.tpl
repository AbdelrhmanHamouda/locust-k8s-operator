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
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
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

Priority order (checks leaf values, not just parent keys):
  1. New path value exists → use it
  2. Old path value exists → use it (backward compat)
  3. Neither → use hardcoded default

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
{{- if and .Values.locustPods .Values.locustPods.resources .Values.locustPods.resources.requests .Values.locustPods.resources.requests.cpu }}
{{- .Values.locustPods.resources.requests.cpu }}
{{- else if and .Values.config .Values.config.loadGenerationPods .Values.config.loadGenerationPods.resource .Values.config.loadGenerationPods.resource.cpuRequest }}
{{- .Values.config.loadGenerationPods.resource.cpuRequest }}
{{- else }}
{{- "250m" }}
{{- end }}
{{- end }}

{{/*
Pod Memory Request - new path with fallback to old path
*/}}
{{- define "locust.podMemRequest" -}}
{{- if and .Values.locustPods .Values.locustPods.resources .Values.locustPods.resources.requests .Values.locustPods.resources.requests.memory }}
{{- .Values.locustPods.resources.requests.memory }}
{{- else if and .Values.config .Values.config.loadGenerationPods .Values.config.loadGenerationPods.resource .Values.config.loadGenerationPods.resource.memRequest }}
{{- .Values.config.loadGenerationPods.resource.memRequest }}
{{- else }}
{{- "128Mi" }}
{{- end }}
{{- end }}

{{/*
Pod Ephemeral Storage Request - new path with fallback to old path
*/}}
{{- define "locust.podEphemeralRequest" -}}
{{- if and .Values.locustPods .Values.locustPods.resources .Values.locustPods.resources.requests .Values.locustPods.resources.requests.ephemeralStorage }}
{{- .Values.locustPods.resources.requests.ephemeralStorage }}
{{- else if and .Values.config .Values.config.loadGenerationPods .Values.config.loadGenerationPods.resource .Values.config.loadGenerationPods.resource.ephemeralRequest }}
{{- .Values.config.loadGenerationPods.resource.ephemeralRequest }}
{{- else }}
{{- "30M" }}
{{- end }}
{{- end }}

{{/*
Pod CPU Limit - new path with fallback to old path
*/}}
{{- define "locust.podCpuLimit" -}}
{{- if and .Values.locustPods .Values.locustPods.resources .Values.locustPods.resources.limits .Values.locustPods.resources.limits.cpu }}
{{- .Values.locustPods.resources.limits.cpu }}
{{- else if and .Values.config .Values.config.loadGenerationPods .Values.config.loadGenerationPods.resource .Values.config.loadGenerationPods.resource.cpuLimit }}
{{- .Values.config.loadGenerationPods.resource.cpuLimit }}
{{- else }}
{{- "1000m" }}
{{- end }}
{{- end }}

{{/*
Pod Memory Limit - new path with fallback to old path
*/}}
{{- define "locust.podMemLimit" -}}
{{- if and .Values.locustPods .Values.locustPods.resources .Values.locustPods.resources.limits .Values.locustPods.resources.limits.memory }}
{{- .Values.locustPods.resources.limits.memory }}
{{- else if and .Values.config .Values.config.loadGenerationPods .Values.config.loadGenerationPods.resource .Values.config.loadGenerationPods.resource.memLimit }}
{{- .Values.config.loadGenerationPods.resource.memLimit }}
{{- else }}
{{- "1024Mi" }}
{{- end }}
{{- end }}

{{/*
Pod Ephemeral Storage Limit - new path with fallback to old path
*/}}
{{- define "locust.podEphemeralLimit" -}}
{{- if and .Values.locustPods .Values.locustPods.resources .Values.locustPods.resources.limits .Values.locustPods.resources.limits.ephemeralStorage }}
{{- .Values.locustPods.resources.limits.ephemeralStorage }}
{{- else if and .Values.config .Values.config.loadGenerationPods .Values.config.loadGenerationPods.resource .Values.config.loadGenerationPods.resource.ephemeralLimit }}
{{- .Values.config.loadGenerationPods.resource.ephemeralLimit }}
{{- else }}
{{- "50M" }}
{{- end }}
{{- end }}

{{/*
Affinity Injection - new path with fallback to old path
*/}}
{{- define "locust.affinityInjection" -}}
{{- if and .Values.locustPods (hasKey .Values.locustPods "affinityInjection") }}
{{- .Values.locustPods.affinityInjection }}
{{- else if and .Values.config .Values.config.loadGenerationPods .Values.config.loadGenerationPods.affinity (hasKey .Values.config.loadGenerationPods.affinity "enableCrInjection") }}
{{- .Values.config.loadGenerationPods.affinity.enableCrInjection }}
{{- else }}
{{- true }}
{{- end }}
{{- end }}

{{/*
Tolerations Injection - new path with fallback to old path
*/}}
{{- define "locust.tolerationsInjection" -}}
{{- if and .Values.locustPods (hasKey .Values.locustPods "tolerationsInjection") }}
{{- .Values.locustPods.tolerationsInjection }}
{{- else if and .Values.config .Values.config.loadGenerationPods .Values.config.loadGenerationPods.taintTolerations (hasKey .Values.config.loadGenerationPods.taintTolerations "enableCrInjection") }}
{{- .Values.config.loadGenerationPods.taintTolerations.enableCrInjection }}
{{- else }}
{{- true }}
{{- end }}
{{- end }}

{{/*
TTL Seconds After Finished - new path with fallback to old path
*/}}
{{- define "locust.ttlSecondsAfterFinished" -}}
{{- if and .Values.locustPods .Values.locustPods.ttlSecondsAfterFinished }}
{{- .Values.locustPods.ttlSecondsAfterFinished }}
{{- else if and .Values.config .Values.config.loadGenerationJobs .Values.config.loadGenerationJobs.ttlSecondsAfterFinished }}
{{- .Values.config.loadGenerationJobs.ttlSecondsAfterFinished }}
{{- else }}
{{- "" }}
{{- end }}
{{- end }}

{{/*
Metrics Exporter Image - new path with fallback to old path
*/}}
{{- define "locust.metricsExporterImage" -}}
{{- if and .Values.locustPods .Values.locustPods.metricsExporter .Values.locustPods.metricsExporter.image }}
{{- .Values.locustPods.metricsExporter.image }}
{{- else if and .Values.config .Values.config.loadGenerationPods .Values.config.loadGenerationPods.metricsExporter .Values.config.loadGenerationPods.metricsExporter.image }}
{{- .Values.config.loadGenerationPods.metricsExporter.image }}
{{- else }}
{{- "containersol/locust_exporter:v0.5.0" }}
{{- end }}
{{- end }}

{{/*
Metrics Exporter Port - new path with fallback to old path
*/}}
{{- define "locust.metricsExporterPort" -}}
{{- if and .Values.locustPods .Values.locustPods.metricsExporter .Values.locustPods.metricsExporter.port }}
{{- .Values.locustPods.metricsExporter.port }}
{{- else if and .Values.config .Values.config.loadGenerationPods .Values.config.loadGenerationPods.metricsExporter .Values.config.loadGenerationPods.metricsExporter.port }}
{{- .Values.config.loadGenerationPods.metricsExporter.port }}
{{- else }}
{{- 9646 }}
{{- end }}
{{- end }}

{{/*
Metrics Exporter Pull Policy - new path with fallback to old path
*/}}
{{- define "locust.metricsExporterPullPolicy" -}}
{{- if and .Values.locustPods .Values.locustPods.metricsExporter .Values.locustPods.metricsExporter.pullPolicy }}
{{- .Values.locustPods.metricsExporter.pullPolicy }}
{{- else if and .Values.config .Values.config.loadGenerationPods .Values.config.loadGenerationPods.metricsExporter .Values.config.loadGenerationPods.metricsExporter.pullPolicy }}
{{- .Values.config.loadGenerationPods.metricsExporter.pullPolicy }}
{{- else }}
{{- "IfNotPresent" }}
{{- end }}
{{- end }}

{{/*
Metrics Exporter CPU Request
*/}}
{{- define "locust.metricsExporterCpuRequest" -}}
{{- if and .Values.locustPods .Values.locustPods.metricsExporter .Values.locustPods.metricsExporter.resources .Values.locustPods.metricsExporter.resources.requests .Values.locustPods.metricsExporter.resources.requests.cpu }}
{{- .Values.locustPods.metricsExporter.resources.requests.cpu }}
{{- else if and .Values.config .Values.config.loadGenerationPods .Values.config.loadGenerationPods.metricsExporter .Values.config.loadGenerationPods.metricsExporter.resource .Values.config.loadGenerationPods.metricsExporter.resource.cpuRequest }}
{{- .Values.config.loadGenerationPods.metricsExporter.resource.cpuRequest }}
{{- else }}
{{- "100m" }}
{{- end }}
{{- end }}

{{/*
Metrics Exporter Memory Request
*/}}
{{- define "locust.metricsExporterMemRequest" -}}
{{- if and .Values.locustPods .Values.locustPods.metricsExporter .Values.locustPods.metricsExporter.resources .Values.locustPods.metricsExporter.resources.requests .Values.locustPods.metricsExporter.resources.requests.memory }}
{{- .Values.locustPods.metricsExporter.resources.requests.memory }}
{{- else if and .Values.config .Values.config.loadGenerationPods .Values.config.loadGenerationPods.metricsExporter .Values.config.loadGenerationPods.metricsExporter.resource .Values.config.loadGenerationPods.metricsExporter.resource.memRequest }}
{{- .Values.config.loadGenerationPods.metricsExporter.resource.memRequest }}
{{- else }}
{{- "64Mi" }}
{{- end }}
{{- end }}

{{/*
Metrics Exporter Ephemeral Request
*/}}
{{- define "locust.metricsExporterEphemeralRequest" -}}
{{- if and .Values.locustPods .Values.locustPods.metricsExporter .Values.locustPods.metricsExporter.resources .Values.locustPods.metricsExporter.resources.requests .Values.locustPods.metricsExporter.resources.requests.ephemeralStorage }}
{{- .Values.locustPods.metricsExporter.resources.requests.ephemeralStorage }}
{{- else if and .Values.config .Values.config.loadGenerationPods .Values.config.loadGenerationPods.metricsExporter .Values.config.loadGenerationPods.metricsExporter.resource .Values.config.loadGenerationPods.metricsExporter.resource.ephemeralRequest }}
{{- .Values.config.loadGenerationPods.metricsExporter.resource.ephemeralRequest }}
{{- else }}
{{- "30M" }}
{{- end }}
{{- end }}

{{/*
Metrics Exporter CPU Limit
*/}}
{{- define "locust.metricsExporterCpuLimit" -}}
{{- if and .Values.locustPods .Values.locustPods.metricsExporter .Values.locustPods.metricsExporter.resources .Values.locustPods.metricsExporter.resources.limits .Values.locustPods.metricsExporter.resources.limits.cpu }}
{{- .Values.locustPods.metricsExporter.resources.limits.cpu }}
{{- else if and .Values.config .Values.config.loadGenerationPods .Values.config.loadGenerationPods.metricsExporter .Values.config.loadGenerationPods.metricsExporter.resource .Values.config.loadGenerationPods.metricsExporter.resource.cpuLimit }}
{{- .Values.config.loadGenerationPods.metricsExporter.resource.cpuLimit }}
{{- else }}
{{- "250m" }}
{{- end }}
{{- end }}

{{/*
Metrics Exporter Memory Limit
*/}}
{{- define "locust.metricsExporterMemLimit" -}}
{{- if and .Values.locustPods .Values.locustPods.metricsExporter .Values.locustPods.metricsExporter.resources .Values.locustPods.metricsExporter.resources.limits .Values.locustPods.metricsExporter.resources.limits.memory }}
{{- .Values.locustPods.metricsExporter.resources.limits.memory }}
{{- else if and .Values.config .Values.config.loadGenerationPods .Values.config.loadGenerationPods.metricsExporter .Values.config.loadGenerationPods.metricsExporter.resource .Values.config.loadGenerationPods.metricsExporter.resource.memLimit }}
{{- .Values.config.loadGenerationPods.metricsExporter.resource.memLimit }}
{{- else }}
{{- "128Mi" }}
{{- end }}
{{- end }}

{{/*
Metrics Exporter Ephemeral Limit
*/}}
{{- define "locust.metricsExporterEphemeralLimit" -}}
{{- if and .Values.locustPods .Values.locustPods.metricsExporter .Values.locustPods.metricsExporter.resources .Values.locustPods.metricsExporter.resources.limits .Values.locustPods.metricsExporter.resources.limits.ephemeralStorage }}
{{- .Values.locustPods.metricsExporter.resources.limits.ephemeralStorage }}
{{- else if and .Values.config .Values.config.loadGenerationPods .Values.config.loadGenerationPods.metricsExporter .Values.config.loadGenerationPods.metricsExporter.resource .Values.config.loadGenerationPods.metricsExporter.resource.ephemeralLimit }}
{{- .Values.config.loadGenerationPods.metricsExporter.resource.ephemeralLimit }}
{{- else }}
{{- "50M" }}
{{- end }}
{{- end }}

{{/*
Kafka Bootstrap Servers - new path with fallback to old path
*/}}
{{- define "locust.kafkaBootstrapServers" -}}
{{- if and .Values.kafka .Values.kafka.bootstrapServers }}
{{- .Values.kafka.bootstrapServers }}
{{- else if and .Values.config .Values.config.loadGenerationPods .Values.config.loadGenerationPods.kafka .Values.config.loadGenerationPods.kafka.bootstrapServers }}
{{- .Values.config.loadGenerationPods.kafka.bootstrapServers }}
{{- else }}
{{- "localhost:9092" }}
{{- end }}
{{- end }}

{{/*
Kafka Security Enabled - new path with fallback to old path
*/}}
{{- define "locust.kafkaSecurityEnabled" -}}
{{- if and .Values.kafka .Values.kafka.security (hasKey .Values.kafka.security "enabled") }}
{{- .Values.kafka.security.enabled }}
{{- else if and .Values.config .Values.config.loadGenerationPods .Values.config.loadGenerationPods.kafka .Values.config.loadGenerationPods.kafka.acl (hasKey .Values.config.loadGenerationPods.kafka.acl "enabled") }}
{{- .Values.config.loadGenerationPods.kafka.acl.enabled }}
{{- else }}
{{- false }}
{{- end }}
{{- end }}

{{/*
=============================================================================
SECTION 2.5: Role-Specific Resource Helpers
=============================================================================
These helpers return role-specific resources for master and worker pods.
If masterResources/workerResources are set, they override the unified resources.
If empty, the helper returns empty string, meaning "use unified resources".

This implements a three-level precedence:
  1. CR-level resources (highest precedence, in Go code)
  2. Helm role-specific resources (these helpers)
  3. Helm unified resources (fallback, in Section 2 helpers above)
=============================================================================
*/}}

{{/*
Master CPU Request - role-specific override
*/}}
{{- define "locust.masterCpuRequest" -}}
{{- if and .Values.locustPods .Values.locustPods.masterResources .Values.locustPods.masterResources.requests }}
{{- .Values.locustPods.masterResources.requests.cpu | default "" }}
{{- else }}
{{- "" }}
{{- end }}
{{- end }}

{{/*
Master Memory Request - role-specific override
*/}}
{{- define "locust.masterMemRequest" -}}
{{- if and .Values.locustPods .Values.locustPods.masterResources .Values.locustPods.masterResources.requests }}
{{- .Values.locustPods.masterResources.requests.memory | default "" }}
{{- else }}
{{- "" }}
{{- end }}
{{- end }}

{{/*
Master Ephemeral Storage Request - role-specific override
*/}}
{{- define "locust.masterEphemeralRequest" -}}
{{- if and .Values.locustPods .Values.locustPods.masterResources .Values.locustPods.masterResources.requests }}
{{- .Values.locustPods.masterResources.requests.ephemeralStorage | default "" }}
{{- else }}
{{- "" }}
{{- end }}
{{- end }}

{{/*
Master CPU Limit - role-specific override
*/}}
{{- define "locust.masterCpuLimit" -}}
{{- if and .Values.locustPods .Values.locustPods.masterResources .Values.locustPods.masterResources.limits }}
{{- .Values.locustPods.masterResources.limits.cpu | default "" }}
{{- else }}
{{- "" }}
{{- end }}
{{- end }}

{{/*
Master Memory Limit - role-specific override
*/}}
{{- define "locust.masterMemLimit" -}}
{{- if and .Values.locustPods .Values.locustPods.masterResources .Values.locustPods.masterResources.limits }}
{{- .Values.locustPods.masterResources.limits.memory | default "" }}
{{- else }}
{{- "" }}
{{- end }}
{{- end }}

{{/*
Master Ephemeral Storage Limit - role-specific override
*/}}
{{- define "locust.masterEphemeralLimit" -}}
{{- if and .Values.locustPods .Values.locustPods.masterResources .Values.locustPods.masterResources.limits }}
{{- .Values.locustPods.masterResources.limits.ephemeralStorage | default "" }}
{{- else }}
{{- "" }}
{{- end }}
{{- end }}

{{/*
Worker CPU Request - role-specific override
*/}}
{{- define "locust.workerCpuRequest" -}}
{{- if and .Values.locustPods .Values.locustPods.workerResources .Values.locustPods.workerResources.requests }}
{{- .Values.locustPods.workerResources.requests.cpu | default "" }}
{{- else }}
{{- "" }}
{{- end }}
{{- end }}

{{/*
Worker Memory Request - role-specific override
*/}}
{{- define "locust.workerMemRequest" -}}
{{- if and .Values.locustPods .Values.locustPods.workerResources .Values.locustPods.workerResources.requests }}
{{- .Values.locustPods.workerResources.requests.memory | default "" }}
{{- else }}
{{- "" }}
{{- end }}
{{- end }}

{{/*
Worker Ephemeral Storage Request - role-specific override
*/}}
{{- define "locust.workerEphemeralRequest" -}}
{{- if and .Values.locustPods .Values.locustPods.workerResources .Values.locustPods.workerResources.requests }}
{{- .Values.locustPods.workerResources.requests.ephemeralStorage | default "" }}
{{- else }}
{{- "" }}
{{- end }}
{{- end }}

{{/*
Worker CPU Limit - role-specific override
*/}}
{{- define "locust.workerCpuLimit" -}}
{{- if and .Values.locustPods .Values.locustPods.workerResources .Values.locustPods.workerResources.limits }}
{{- .Values.locustPods.workerResources.limits.cpu | default "" }}
{{- else }}
{{- "" }}
{{- end }}
{{- end }}

{{/*
Worker Memory Limit - role-specific override
*/}}
{{- define "locust.workerMemLimit" -}}
{{- if and .Values.locustPods .Values.locustPods.workerResources .Values.locustPods.workerResources.limits }}
{{- .Values.locustPods.workerResources.limits.memory | default "" }}
{{- else }}
{{- "" }}
{{- end }}
{{- end }}

{{/*
Worker Ephemeral Storage Limit - role-specific override
*/}}
{{- define "locust.workerEphemeralLimit" -}}
{{- if and .Values.locustPods .Values.locustPods.workerResources .Values.locustPods.workerResources.limits }}
{{- .Values.locustPods.workerResources.limits.ephemeralStorage | default "" }}
{{- else }}
{{- "" }}
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
# Role-specific resources for master Locust containers
# Empty values mean "use unified resources" (backward compatible)
{{- $masterCpuReq := include "locust.masterCpuRequest" . }}
{{- if $masterCpuReq }}
- name: MASTER_POD_CPU_REQUEST
  value: {{ $masterCpuReq | quote }}
{{- end }}
{{- $masterMemReq := include "locust.masterMemRequest" . }}
{{- if $masterMemReq }}
- name: MASTER_POD_MEM_REQUEST
  value: {{ $masterMemReq | quote }}
{{- end }}
{{- $masterEphemeralReq := include "locust.masterEphemeralRequest" . }}
{{- if $masterEphemeralReq }}
- name: MASTER_POD_EPHEMERAL_REQUEST
  value: {{ $masterEphemeralReq | quote }}
{{- end }}
{{- $masterCpuLim := include "locust.masterCpuLimit" . }}
{{- if $masterCpuLim }}
- name: MASTER_POD_CPU_LIMIT
  value: {{ $masterCpuLim | quote }}
{{- end }}
{{- $masterMemLim := include "locust.masterMemLimit" . }}
{{- if $masterMemLim }}
- name: MASTER_POD_MEM_LIMIT
  value: {{ $masterMemLim | quote }}
{{- end }}
{{- $masterEphemeralLim := include "locust.masterEphemeralLimit" . }}
{{- if $masterEphemeralLim }}
- name: MASTER_POD_EPHEMERAL_LIMIT
  value: {{ $masterEphemeralLim | quote }}
{{- end }}
# Role-specific resources for worker Locust containers
# Empty values mean "use unified resources" (backward compatible)
{{- $workerCpuReq := include "locust.workerCpuRequest" . }}
{{- if $workerCpuReq }}
- name: WORKER_POD_CPU_REQUEST
  value: {{ $workerCpuReq | quote }}
{{- end }}
{{- $workerMemReq := include "locust.workerMemRequest" . }}
{{- if $workerMemReq }}
- name: WORKER_POD_MEM_REQUEST
  value: {{ $workerMemReq | quote }}
{{- end }}
{{- $workerEphemeralReq := include "locust.workerEphemeralRequest" . }}
{{- if $workerEphemeralReq }}
- name: WORKER_POD_EPHEMERAL_REQUEST
  value: {{ $workerEphemeralReq | quote }}
{{- end }}
{{- $workerCpuLim := include "locust.workerCpuLimit" . }}
{{- if $workerCpuLim }}
- name: WORKER_POD_CPU_LIMIT
  value: {{ $workerCpuLim | quote }}
{{- end }}
{{- $workerMemLim := include "locust.workerMemLimit" . }}
{{- if $workerMemLim }}
- name: WORKER_POD_MEM_LIMIT
  value: {{ $workerMemLim | quote }}
{{- end }}
{{- $workerEphemeralLim := include "locust.workerEphemeralLimit" . }}
{{- if $workerEphemeralLim }}
- name: WORKER_POD_EPHEMERAL_LIMIT
  value: {{ $workerEphemeralLim | quote }}
{{- end }}
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
{{- if .Values.kafka.enabled }}
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
{{- end }}
