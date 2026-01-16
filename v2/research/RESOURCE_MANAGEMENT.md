# Kubernetes Resource Management in Go

**Research Date:** January 2026  
**Focus:** Building Jobs, Services, and Pods for the Locust Operator

---

## Table of Contents

1. [Resource Building Patterns](#1-resource-building-patterns)
2. [Job Construction](#2-job-construction)
3. [Service Construction](#3-service-construction)
4. [Owner References & Cleanup](#4-owner-references--cleanup)
5. [Label & Annotation Management](#5-label--annotation-management)

---

## 1. Resource Building Patterns

### 1.1 Go Struct Literals vs Java Builders

**Java (Fabric8):**
```java
Job job = new JobBuilder()
    .withMetadata(new ObjectMetaBuilder()
        .withName(name)
        .withNamespace(namespace)
        .build())
    .withSpec(new JobSpecBuilder()
        .withParallelism(replicas)
        .withBackoffLimit(0)
        .build())
    .build();
```

**Go (client-go):**
```go
job := &batchv1.Job{
    ObjectMeta: metav1.ObjectMeta{
        Name:      name,
        Namespace: namespace,
    },
    Spec: batchv1.JobSpec{
        Parallelism:  ptr.To(int32(replicas)),
        BackoffLimit: ptr.To(int32(0)),
    },
}
```

### 1.2 Helper Package for Pointers

```go
// internal/resources/ptr/ptr.go
package ptr

func To[T any](v T) *T {
    return &v
}

func Int32(v int32) *int32 {
    return &v
}

func String(v string) *string {
    return &v
}
```

---

## 2. Job Construction

### 2.1 Master Job Builder

```go
// internal/resources/job.go
package resources

import (
    "fmt"
    "strings"

    batchv1 "k8s.io/api/batch/v1"
    corev1 "k8s.io/api/core/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

    locustv2 "github.com/AbdelrhmanHamouda/locust-k8s-operator/api/v2"
    "github.com/AbdelrhmanHamouda/locust-k8s-operator/internal/config"
)

const (
    MasterPort     = 5557
    MasterBindPort = 5558
    WebUIPort      = 8089
    WorkerPort     = 8080
    MetricsPort    = 9646

    ConfigMapMountPath = "/lotest/src/"
    LibMountPath       = "/opt/locust/lib"
)

func BuildMasterJob(lt *locustv2.LocustTest, cfg *config.OperatorConfig) *batchv1.Job {
    name := NodeName(lt.Name, Master)
    command := BuildMasterCommand(lt.Spec.Master.Command, lt.Spec.Worker.Replicas, &lt.Spec.Master)

    containers := []corev1.Container{
        buildLocustContainer(lt, Master, command, MasterPorts()),
    }

    // Add metrics exporter sidecar if OTel not enabled
    if !isOTelEnabled(lt) {
        containers = append(containers, buildMetricsExporterContainer(cfg))
    }

    // Get scheduling config
    var affinity *corev1.Affinity
    var tolerations []corev1.Toleration
    var nodeSelector map[string]string
    if lt.Spec.Scheduling != nil {
        affinity = lt.Spec.Scheduling.Affinity
        tolerations = lt.Spec.Scheduling.Tolerations
        nodeSelector = lt.Spec.Scheduling.NodeSelector
    }

    return &batchv1.Job{
        ObjectMeta: metav1.ObjectMeta{
            Name:      name,
            Namespace: lt.Namespace,
            Labels:    BuildLabels(lt, Master),
        },
        Spec: batchv1.JobSpec{
            Parallelism:             ptr.To(int32(1)),
            Completions:             ptr.To(int32(1)),
            BackoffLimit:            ptr.To(int32(0)),
            TTLSecondsAfterFinished: cfg.TTLSecondsAfterFinished,
            Template: corev1.PodTemplateSpec{
                ObjectMeta: metav1.ObjectMeta{
                    Labels:      BuildLabels(lt, Master),
                    Annotations: BuildAnnotations(lt, Master, cfg),
                },
                Spec: corev1.PodSpec{
                    RestartPolicy:    corev1.RestartPolicyNever,
                    Containers:       containers,
                    Volumes:          buildVolumes(lt),
                    Affinity:         affinity,
                    Tolerations:      tolerations,
                    NodeSelector:     nodeSelector,
                    ImagePullSecrets: lt.Spec.ImagePullSecrets,
                },
            },
        },
    }
}

func BuildWorkerJob(lt *locustv2.LocustTest, cfg *config.OperatorConfig) *batchv1.Job {
    name := NodeName(lt.Name, Worker)
    masterHost := NodeName(lt.Name, Master)
    command := BuildWorkerCommand(lt.Spec.Worker.Command, masterHost, lt.Spec.Worker.ExtraArgs)

    // Get scheduling config
    var affinity *corev1.Affinity
    var tolerations []corev1.Toleration
    var nodeSelector map[string]string
    if lt.Spec.Scheduling != nil {
        affinity = lt.Spec.Scheduling.Affinity
        tolerations = lt.Spec.Scheduling.Tolerations
        nodeSelector = lt.Spec.Scheduling.NodeSelector
    }

    return &batchv1.Job{
        ObjectMeta: metav1.ObjectMeta{
            Name:      name,
            Namespace: lt.Namespace,
            Labels:    BuildLabels(lt, Worker),
        },
        Spec: batchv1.JobSpec{
            Parallelism:             ptr.To(lt.Spec.Worker.Replicas),
            Completions:             ptr.To(lt.Spec.Worker.Replicas),
            BackoffLimit:            ptr.To(int32(0)),
            TTLSecondsAfterFinished: cfg.TTLSecondsAfterFinished,
            Template: corev1.PodTemplateSpec{
                ObjectMeta: metav1.ObjectMeta{
                    Labels:      BuildLabels(lt, Worker),
                    Annotations: BuildAnnotations(lt, Worker, cfg),
                },
                Spec: corev1.PodSpec{
                    RestartPolicy:    corev1.RestartPolicyNever,
                    Containers:       []corev1.Container{buildLocustContainer(lt, Worker, command, WorkerPorts())},
                    Volumes:          buildVolumes(lt),
                    Affinity:         affinity,
                    Tolerations:      tolerations,
                    NodeSelector:     nodeSelector,
                    ImagePullSecrets: lt.Spec.ImagePullSecrets,
                },
            },
        },
    }
}

func buildLocustContainer(lt *locustv2.LocustTest, mode OperationalMode, command []string, ports []corev1.ContainerPort) corev1.Container {
    container := corev1.Container{
        Name:            "locust",
        Image:           lt.Spec.Image,
        ImagePullPolicy: lt.Spec.ImagePullPolicy,
        Args:            command,
        Ports:           ports,
        VolumeMounts:    buildVolumeMounts(lt, mode),
        Env:             buildEnvVars(lt),
    }

    // Apply resources from grouped config
    if mode == Master {
        container.Resources = lt.Spec.Master.Resources
    } else {
        container.Resources = lt.Spec.Worker.Resources
    }

    return container
}
```

### 2.2 Command Building

```go
// internal/resources/command.go
package resources

import (
    "fmt"
    "strings"

    locustv2 "github.com/AbdelrhmanHamouda/locust-k8s-operator/api/v2"
)

func BuildMasterCommand(command string, workerReplicas int32, master *locustv2.MasterSpec) []string {
    var parts []string

    // Add base command
    parts = append(parts, strings.Fields(command)...)

    // Master mode flags
    parts = append(parts, "--master", fmt.Sprintf("--master-port=%d", MasterPort))
    parts = append(parts, fmt.Sprintf("--expect-workers=%d", workerReplicas))

    // Autostart (default true)
    if master.Autostart == nil || *master.Autostart {
        parts = append(parts, "--autostart")
    }

    // Autoquit (default enabled with 60s timeout)
    if master.Autoquit == nil || master.Autoquit.Enabled {
        timeout := int32(60)
        if master.Autoquit != nil && master.Autoquit.Timeout > 0 {
            timeout = master.Autoquit.Timeout
        }
        parts = append(parts, fmt.Sprintf("--autoquit=%d", timeout))
    }

    // Standard flags
    parts = append(parts, "--enable-rebalancing", "--only-summary")

    // Extra args from master config
    parts = append(parts, master.ExtraArgs...)

    return parts
}

func BuildWorkerCommand(command, masterHost string, extraArgs []string) []string {
    var parts []string

    parts = append(parts, strings.Fields(command)...)
    parts = append(parts,
        "--worker",
        fmt.Sprintf("--master-port=%d", MasterPort),
        fmt.Sprintf("--master-host=%s", masterHost),
    )

    // Extra args from worker config
    parts = append(parts, extraArgs...)

    return parts
}
```

---

## 3. Service Construction

```go
// internal/resources/service.go
package resources

import (
    corev1 "k8s.io/api/core/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/apimachinery/pkg/util/intstr"

    locustv2 "github.com/AbdelrhmanHamouda/locust-k8s-operator/api/v2"
)

func BuildMasterService(lt *locustv2.LocustTest, cfg *config.OperatorConfig) *corev1.Service {
    name := NodeName(lt.Name, Master)

    ports := []corev1.ServicePort{
        {Name: "master", Port: MasterPort, TargetPort: intstr.FromInt(MasterPort), Protocol: corev1.ProtocolTCP},
        {Name: "master-bind", Port: MasterBindPort, TargetPort: intstr.FromInt(MasterBindPort), Protocol: corev1.ProtocolTCP},
        {Name: "web-ui", Port: WebUIPort, TargetPort: intstr.FromInt(WebUIPort), Protocol: corev1.ProtocolTCP},
    }

    // Add metrics port if sidecar is used
    if !isOTelEnabled(lt) {
        ports = append(ports, corev1.ServicePort{
            Name: "metrics", Port: int32(cfg.MetricsExporterPort),
            TargetPort: intstr.FromInt(int(cfg.MetricsExporterPort)), Protocol: corev1.ProtocolTCP,
        })
    }

    return &corev1.Service{
        ObjectMeta: metav1.ObjectMeta{
            Name:      name,
            Namespace: lt.Namespace,
            Labels:    BuildLabels(lt, Master),
        },
        Spec: corev1.ServiceSpec{
            Type: corev1.ServiceTypeClusterIP,
            Selector: map[string]string{
                LabelPodName: name,
            },
            Ports: ports,
        },
    }
}
```

---

## 4. Owner References & Cleanup

### 4.1 Setting Owner References

```go
// internal/controller/locusttest_controller.go
import "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

func (r *LocustTestReconciler) createResources(ctx context.Context, lt *locustv2.LocustTest) (ctrl.Result, error) {
    log := log.FromContext(ctx)

    // Build resources
    masterService := resources.BuildMasterService(lt, r.Config)
    masterJob := resources.BuildMasterJob(lt, r.Config)
    workerJob := resources.BuildWorkerJob(lt, r.Config)

    // Set owner references for automatic garbage collection
    if err := controllerutil.SetControllerReference(lt, masterService, r.Scheme); err != nil {
        return ctrl.Result{}, err
    }
    if err := controllerutil.SetControllerReference(lt, masterJob, r.Scheme); err != nil {
        return ctrl.Result{}, err
    }
    if err := controllerutil.SetControllerReference(lt, workerJob, r.Scheme); err != nil {
        return ctrl.Result{}, err
    }

    // Create resources
    if err := r.Create(ctx, masterService); err != nil && !apierrors.IsAlreadyExists(err) {
        log.Error(err, "Failed to create master Service")
        return ctrl.Result{}, err
    }

    if err := r.Create(ctx, masterJob); err != nil && !apierrors.IsAlreadyExists(err) {
        log.Error(err, "Failed to create master Job")
        return ctrl.Result{}, err
    }

    if err := r.Create(ctx, workerJob); err != nil && !apierrors.IsAlreadyExists(err) {
        log.Error(err, "Failed to create worker Job")
        return ctrl.Result{}, err
    }

    log.Info("All resources created successfully")
    return ctrl.Result{}, nil
}
```

### 4.2 Automatic Cleanup via Owner References

When `LocustTest` CR is deleted, Kubernetes automatically deletes all resources with matching owner references. No manual cleanup needed in most cases.

---

## 5. Label & Annotation Management

```go
// internal/resources/labels.go
package resources

import (
    locustv2 "github.com/AbdelrhmanHamouda/locust-k8s-operator/api/v2"
)

const (
    LabelApp       = "app"
    LabelPodName   = "performance-test-pod-name"
    LabelManagedBy = "managed-by"
    LabelTestName  = "performance-test-name"

    ManagedByValue = "locust-k8s-operator"
)

type OperationalMode string

const (
    Master OperationalMode = "master"
    Worker OperationalMode = "worker"
)

func NodeName(crName string, mode OperationalMode) string {
    name := strings.ReplaceAll(crName, ".", "-")
    return fmt.Sprintf("%s-%s", name, mode)
}

func BuildLabels(lt *locustv2.LocustTest, mode OperationalMode) map[string]string {
    nodeName := NodeName(lt.Name, mode)

    labels := map[string]string{
        LabelApp:       nodeName,
        LabelPodName:   nodeName,
        LabelManagedBy: ManagedByValue,
        LabelTestName:  strings.ReplaceAll(lt.Name, ".", "-"),
    }

    // Merge user-defined labels from grouped config
    var userLabels map[string]string
    if mode == Master {
        userLabels = lt.Spec.Master.Labels
    } else {
        userLabels = lt.Spec.Worker.Labels
    }

    for k, v := range userLabels {
        labels[k] = v
    }

    return labels
}

func BuildAnnotations(lt *locustv2.LocustTest, mode OperationalMode, cfg *config.OperatorConfig) map[string]string {
    annotations := make(map[string]string)

    // Prometheus scrape annotations for master (when OTel not enabled)
    if mode == Master && !isOTelEnabled(lt) {
        annotations["prometheus.io/scrape"] = "true"
        annotations["prometheus.io/path"] = "/metrics"
        annotations["prometheus.io/port"] = fmt.Sprintf("%d", cfg.MetricsExporterPort)
    }

    // Merge user-defined annotations from grouped config
    var userAnnotations map[string]string
    if mode == Master {
        userAnnotations = lt.Spec.Master.Annotations
    } else {
        userAnnotations = lt.Spec.Worker.Annotations
    }

    for k, v := range userAnnotations {
        annotations[k] = v
    }

    return annotations
}
```

---

## References

- [client-go API types](https://pkg.go.dev/k8s.io/api)
- [Kubernetes API Conventions](https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md)
