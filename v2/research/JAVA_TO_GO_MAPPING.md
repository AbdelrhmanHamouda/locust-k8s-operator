# Java to Go Mapping Guide

**Research Date:** January 2026  
**Purpose:** Direct mapping of Java operator patterns to Go equivalents

---

## Table of Contents

1. [Class to Struct Mapping](#1-class-to-struct-mapping)
2. [Dependency Injection](#2-dependency-injection)
3. [Configuration Binding](#3-configuration-binding)
4. [Kubernetes Client Operations](#4-kubernetes-client-operations)
5. [Error Handling](#5-error-handling)
6. [File-by-File Migration Reference](#6-file-by-file-migration-reference)

---

## 1. Class to Struct Mapping

### 1.1 CRD Types

**Java (LocustTestSpec.java):**
```java
@Data
public class LocustTestSpec implements KubernetesResource {
    private Map<String, Map<String, String>> labels;
    private Map<String, Map<String, String>> annotations;
    private LocustTestAffinity affinity;
    private List<LocustTestToleration> tolerations;
    private String masterCommandSeed;
    private String workerCommandSeed;
    private Integer workerReplicas;
    private String configMap;
    private String libConfigMap;
    private String image;
    private String imagePullPolicy;
    private List<String> imagePullSecrets;
}
```

**Go (locusttest_types.go):**
```go
type LocustTestSpec struct {
    Labels            *PodLabels                      `json:"labels,omitempty"`
    Annotations       *PodAnnotations                 `json:"annotations,omitempty"`
    Affinity          *corev1.Affinity                `json:"affinity,omitempty"`
    Tolerations       []corev1.Toleration             `json:"tolerations,omitempty"`
    MasterCommandSeed string                          `json:"masterCommandSeed"`
    WorkerCommandSeed string                          `json:"workerCommandSeed"`
    WorkerReplicas    int32                           `json:"workerReplicas"`
    ConfigMap         string                          `json:"configMap,omitempty"`
    LibConfigMap      string                          `json:"libConfigMap,omitempty"`
    Image             string                          `json:"image"`
    ImagePullPolicy   corev1.PullPolicy               `json:"imagePullPolicy,omitempty"`
    ImagePullSecrets  []corev1.LocalObjectReference   `json:"imagePullSecrets,omitempty"`
}

type PodLabels struct {
    Master map[string]string `json:"master,omitempty"`
    Worker map[string]string `json:"worker,omitempty"`
}

type PodAnnotations struct {
    Master map[string]string `json:"master,omitempty"`
    Worker map[string]string `json:"worker,omitempty"`
}
```

### 1.2 DTOs

**Java (LoadGenerationNode.java):**
```java
@Data
@Builder
public class LoadGenerationNode {
    private String name;
    private Map<String, String> labels;
    private Map<String, String> annotations;
    private LocustTestAffinity affinity;
    private List<LocustTestToleration> tolerations;
    private Integer ttlSecondsAfterFinished;
    private List<String> command;
    private OperationalMode operationalMode;
    private String image;
    private String imagePullPolicy;
    private List<String> imagePullSecrets;
    private Integer replicas;
    private List<Integer> ports;
    private String configMap;
    private String libConfigMap;
}
```

**Go (types.go):**
```go
type LoadGenerationNode struct {
    Name                    string
    Labels                  map[string]string
    Annotations             map[string]string
    Affinity                *corev1.Affinity
    Tolerations             []corev1.Toleration
    TTLSecondsAfterFinished *int32
    Command                 []string
    OperationalMode         OperationalMode
    Image                   string
    ImagePullPolicy         corev1.PullPolicy
    ImagePullSecrets        []corev1.LocalObjectReference
    Replicas                int32
    Ports                   []int32
    ConfigMap               string
    LibConfigMap            string
}

type OperationalMode string

const (
    Master OperationalMode = "master"
    Worker OperationalMode = "worker"
)
```

### 1.3 Enums

**Java:**
```java
public enum OperationalMode {
    MASTER("master"),
    WORKER("worker");
    
    @Getter
    private final String mode;
}
```

**Go:**
```go
type OperationalMode string

const (
    Master OperationalMode = "master"
    Worker OperationalMode = "worker"
)

func (m OperationalMode) String() string {
    return string(m)
}
```

---

## 2. Dependency Injection

### 2.1 Constructor Injection

**Java (Micronaut):**
```java
@Singleton
public class LocustTestReconciler implements Reconciler<LocustTest> {
    private final LoadGenHelpers loadGenHelpers;
    private final ResourceCreationManager creationManager;
    private final ResourceDeletionManager deletionManager;

    public LocustTestReconciler(LoadGenHelpers loadGenHelpers, 
        ResourceCreationManager creationManager,
        ResourceDeletionManager deletionManager) {
        this.loadGenHelpers = loadGenHelpers;
        this.creationManager = creationManager;
        this.deletionManager = deletionManager;
    }
}
```

**Go (Explicit Wiring):**
```go
type LocustTestReconciler struct {
    client.Client
    Scheme   *runtime.Scheme
    Config   *config.OperatorConfig
    Recorder record.EventRecorder
}

// In main.go
func main() {
    mgr, _ := ctrl.NewManager(...)
    
    cfg := config.LoadConfig()
    
    reconciler := &controller.LocustTestReconciler{
        Client:   mgr.GetClient(),
        Scheme:   mgr.GetScheme(),
        Config:   cfg,
        Recorder: mgr.GetEventRecorderFor("locust-controller"),
    }
    
    reconciler.SetupWithManager(mgr)
}
```

---

## 3. Configuration Binding

### 3.1 Property Binding

**Java (SysConfig.java with Micronaut @Property):**
```java
@Singleton
public class SysConfig {
    @Property(name = "config.load-generation-jobs.ttl-seconds-after-finished")
    private Integer ttlSecondsAfterFinished;

    @Property(name = "config.load-generation-pods.resource.cpu-request", defaultValue = "250m")
    private String podCpuRequest;

    @Property(name = "config.load-generation-pods.resource.mem-request", defaultValue = "128Mi")
    private String podMemRequest;

    @Property(name = "config.load-generation-pods.metricsExporter.image")
    private String metricsExporterImage;

    @Property(name = "config.load-generation-pods.affinity.enableCrInjection", defaultValue = "false")
    private boolean affinityCrInjectionEnabled;
}
```

**Go (config.go with environment variables):**
```go
package config

import (
    "os"
    "strconv"
)

type OperatorConfig struct {
    TTLSecondsAfterFinished *int32
    PodCPURequest           string
    PodMemRequest           string
    PodCPULimit             string
    PodMemLimit             string
    MetricsExporterImage    string
    MetricsExporterPort     int32
    AffinityCRInjectionEnabled    bool
    TolerationsCRInjectionEnabled bool
}

func LoadConfig() *OperatorConfig {
    return &OperatorConfig{
        TTLSecondsAfterFinished: getEnvInt32Ptr("JOB_TTL_SECONDS_AFTER_FINISHED"),
        PodCPURequest:           getEnv("POD_CPU_REQUEST", "250m"),
        PodMemRequest:           getEnv("POD_MEM_REQUEST", "128Mi"),
        PodCPULimit:             getEnv("POD_CPU_LIMIT", ""),
        PodMemLimit:             getEnv("POD_MEM_LIMIT", ""),
        MetricsExporterImage:    getEnv("METRICS_EXPORTER_IMAGE", "containersol/locust_exporter:v0.5.0"),
        MetricsExporterPort:     getEnvInt32("METRICS_EXPORTER_PORT", 9646),
        AffinityCRInjectionEnabled:    getEnvBool("ENABLE_AFFINITY_CR_INJECTION", false),
        TolerationsCRInjectionEnabled: getEnvBool("ENABLE_TOLERATIONS_CR_INJECTION", false),
    }
}

func getEnv(key, defaultValue string) string {
    if v := os.Getenv(key); v != "" {
        return v
    }
    return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
    if v := os.Getenv(key); v != "" {
        b, _ := strconv.ParseBool(v)
        return b
    }
    return defaultValue
}

func getEnvInt32(key string, defaultValue int32) int32 {
    if v := os.Getenv(key); v != "" {
        i, _ := strconv.ParseInt(v, 10, 32)
        return int32(i)
    }
    return defaultValue
}

func getEnvInt32Ptr(key string) *int32 {
    if v := os.Getenv(key); v != "" {
        i, _ := strconv.ParseInt(v, 10, 32)
        val := int32(i)
        return &val
    }
    return nil
}
```

---

## 4. Kubernetes Client Operations

### 4.1 Resource Creation

**Java (ResourceCreationManager.java):**
```java
public void createJob(LoadGenerationNode node, String namespace, String crName) {
    log.info("Attempt to create job: {}", node.getName());
    try (KubernetesClient client = new KubernetesClientBuilder().build()) {
        Job job = creationHelpers.generateJob(node, namespace, crName);
        client.batch().v1().jobs()
            .inNamespace(namespace)
            .resource(job)
            .serverSideApply();
        log.info("Job created: {} in namespace: {}", node.getName(), namespace);
    } catch (Exception e) {
        log.error("Exception occurred: {}", e.getLocalizedMessage(), e);
    }
}
```

**Go (controller):**
```go
func (r *LocustTestReconciler) createJob(ctx context.Context, lt *locustv2.LocustTest, job *batchv1.Job) error {
    log := log.FromContext(ctx)
    log.Info("Attempting to create Job", "name", job.Name)

    // Set owner reference
    if err := controllerutil.SetControllerReference(lt, job, r.Scheme); err != nil {
        return fmt.Errorf("failed to set owner reference: %w", err)
    }

    if err := r.Create(ctx, job); err != nil {
        if apierrors.IsAlreadyExists(err) {
            log.V(1).Info("Job already exists", "name", job.Name)
            return nil
        }
        log.Error(err, "Failed to create Job", "name", job.Name)
        return err
    }

    log.Info("Job created", "name", job.Name, "namespace", job.Namespace)
    return nil
}
```

### 4.2 Resource Deletion

**Java (ResourceDeletionManager.java):**
```java
public void deleteJob(LocustTest resource, OperationalMode mode) {
    String namespace = resource.getMetadata().getNamespace();
    String jobName = loadGenHelpers.constructNodeName(resource, mode);

    log.info("Attempting to delete Job: {} in namespace: {}", jobName, namespace);
    try (KubernetesClient client = new KubernetesClientBuilder().build()) {
        client.batch().v1().jobs()
            .inNamespace(namespace)
            .withName(jobName)
            .delete();
        log.info("Job deleted: {}", jobName);
    } catch (Exception e) {
        log.error("Failed to delete job: {}", e.getLocalizedMessage(), e);
    }
}
```

**Go (owner references handle cleanup automatically):**
```go
// With owner references set, deletion is automatic when parent is deleted.
// For explicit deletion:
func (r *LocustTestReconciler) deleteJob(ctx context.Context, namespace, name string) error {
    log := log.FromContext(ctx)
    
    job := &batchv1.Job{
        ObjectMeta: metav1.ObjectMeta{
            Name:      name,
            Namespace: namespace,
        },
    }
    
    if err := r.Delete(ctx, job); err != nil {
        if apierrors.IsNotFound(err) {
            return nil // Already deleted
        }
        log.Error(err, "Failed to delete Job", "name", name)
        return err
    }
    
    log.Info("Job deleted", "name", name)
    return nil
}
```

---

## 5. Error Handling

### 5.1 Try-Catch vs Error Returns

**Java:**
```java
try {
    Job job = client.batch().v1().jobs()
        .inNamespace(namespace)
        .resource(job)
        .serverSideApply();
    log.info("Job created: {}", job.getMetadata().getName());
} catch (KubernetesClientException e) {
    log.error("K8s API error: {}", e.getMessage(), e);
    throw new OperatorException("Failed to create job", e);
}
```

**Go:**
```go
if err := r.Create(ctx, job); err != nil {
    if apierrors.IsAlreadyExists(err) {
        log.V(1).Info("Job already exists", "name", job.Name)
        return nil
    }
    log.Error(err, "Failed to create Job")
    return fmt.Errorf("creating job %s: %w", job.Name, err)
}
log.Info("Job created", "name", job.Name)
```

### 5.2 Optional Handling

**Java:**
```java
Optional.ofNullable(memOverride)
    .filter(s -> !s.isBlank())
    .ifPresent(override -> resourceOverrideMap.put("memory", new Quantity(override)));
```

**Go:**
```go
if memOverride != "" {
    resourceOverrideMap["memory"] = resource.MustParse(memOverride)
}
```

---

## 6. File-by-File Migration Reference

| Java File | Go File | Notes |
|-----------|---------|-------|
| `Application.java` | `cmd/main.go` | Entry point |
| `LocustTestOperatorStarter.java` | `cmd/main.go` | Merged into main |
| `LocustTestReconciler.java` | `internal/controller/locusttest_controller.go` | Core reconciler |
| `SysConfig.java` | `internal/config/config.go` | Configuration |
| `LocustTest.java` | `api/v1/locusttest_types.go` | CRD type |
| `LocustTestSpec.java` | `api/v1/locusttest_types.go` | Merged |
| `LocustTestAffinity.java` | Use `corev1.Affinity` | Standard K8s type |
| `LocustTestToleration.java` | Use `corev1.Toleration` | Standard K8s type |
| `LoadGenerationNode.java` | `internal/resources/types.go` | Internal DTO |
| `MetricsExporterContainer.java` | `internal/resources/container.go` | Container builder |
| `OperationalMode.java` | `internal/resources/types.go` | Const block |
| `Constants.java` | `internal/resources/constants.go` | Constants |
| `LoadGenHelpers.java` | `internal/resources/helpers.go` | Split into builders |
| `ResourceCreationManager.java` | `internal/controller/locusttest_controller.go` | Merged |
| `ResourceCreationHelpers.java` | `internal/resources/job.go`, `service.go` | Split |
| `ResourceDeletionManager.java` | N/A (owner references) | Automatic cleanup |

---

## Quick Reference: Common Conversions

| Java | Go |
|------|-----|
| `@Slf4j` + `log.info()` | `log := log.FromContext(ctx); log.Info()` |
| `@Singleton` | Explicit struct, wired in main |
| `@Property(defaultValue="x")` | `getEnv("KEY", "x")` |
| `Optional.ofNullable(x)` | `if x != nil { ... }` |
| `List.of(a, b, c)` | `[]T{a, b, c}` |
| `Map.of("k", "v")` | `map[string]string{"k": "v"}` |
| `String.format("%s", x)` | `fmt.Sprintf("%s", x)` |
| `stream().map().collect()` | `for range` loop |
| `new JobBuilder()...build()` | `&batchv1.Job{...}` |
| `null` | `nil` |
| `Integer` | `*int32` (nullable) or `int32` |
| `Boolean` | `bool` (zero value is `false`) |
