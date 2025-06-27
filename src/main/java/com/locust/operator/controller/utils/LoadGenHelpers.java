package com.locust.operator.controller.utils;

import com.locust.operator.controller.config.SysConfig;
import com.locust.operator.controller.dto.LoadGenerationNode;
import com.locust.operator.controller.dto.MetricsExporterContainer;
import com.locust.operator.controller.dto.OperationalMode;
import com.locust.operator.customresource.LocustTest;
import com.locust.operator.customresource.internaldto.LocustTestAffinity;
import com.locust.operator.customresource.internaldto.LocustTestToleration;
import io.fabric8.kubernetes.api.model.Quantity;
import io.fabric8.kubernetes.api.model.ResourceRequirements;
import jakarta.inject.Singleton;
import lombok.extern.slf4j.Slf4j;

import java.util.Collections;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.Optional;

import static com.locust.operator.controller.dto.OperationalMode.MASTER;
import static com.locust.operator.controller.dto.OperationalMode.WORKER;
import static com.locust.operator.controller.utils.Constants.CONTAINER_ARGS_SEPARATOR;
import static com.locust.operator.controller.utils.Constants.DEFAULT_RESOURCE_TARGET;
import static com.locust.operator.controller.utils.Constants.EXPORTER_CONTAINER_NAME;
import static com.locust.operator.controller.utils.Constants.KAFKA_BOOTSTRAP_SERVERS;
import static com.locust.operator.controller.utils.Constants.KAFKA_PASSWORD;
import static com.locust.operator.controller.utils.Constants.KAFKA_SASL_JAAS_CONFIG;
import static com.locust.operator.controller.utils.Constants.KAFKA_SASL_MECHANISM;
import static com.locust.operator.controller.utils.Constants.KAFKA_SECURITY_ENABLED;
import static com.locust.operator.controller.utils.Constants.KAFKA_SECURITY_PROTOCOL_CONFIG;
import static com.locust.operator.controller.utils.Constants.KAFKA_USERNAME;
import static com.locust.operator.controller.utils.Constants.MASTER_CMD_TEMPLATE;
import static com.locust.operator.controller.utils.Constants.MASTER_NODE_PORTS;
import static com.locust.operator.controller.utils.Constants.MASTER_NODE_REPLICA_COUNT;
import static com.locust.operator.controller.utils.Constants.METRICS_EXPORTER_RESOURCE_TARGET;
import static com.locust.operator.controller.utils.Constants.NODE_NAME_TEMPLATE;
import static com.locust.operator.controller.utils.Constants.WORKER_CMD_TEMPLATE;
import static com.locust.operator.controller.utils.Constants.WORKER_NODE_PORT;

@Slf4j
@Singleton
public class LoadGenHelpers {

    private final SysConfig config;

    public LoadGenHelpers(SysConfig config) {
        this.config = config;
    }

    /**
     * Parse an LocustTest resource and convert it a LoadGenerationNode object after: - Constructing the node operational command based on
     * the `mode` parameter - Set the replica count based on the `mode` parameter
     *
     * @param resource Custom resource object
     * @param mode     Operational mode
     * @return Load generation node configuration
     */
    public LoadGenerationNode generateLoadGenNodeObject(LocustTest resource, OperationalMode mode) {

        return new LoadGenerationNode(
            constructNodeName(resource, mode),
            constructNodeLabels(resource, mode),
            constructNodeAnnotations(resource, mode),
            getNodeAffinity(resource),
            getPodToleration(resource),
            getTtlSecondsAfterFinished(),
            constructNodeCommand(resource, mode),
            mode,
            getNodeImage(resource),
            getNodeImagePullPolicy(resource),
            getNodeImagePullSecrets(resource),
            getReplicaCount(resource, mode),
            getNodePorts(resource, mode),
            getConfigMap(resource));

    }

    private Integer getTtlSecondsAfterFinished() {
        return this.config.getTtlSecondsAfterFinished();
    }

    private List<LocustTestToleration> getPodToleration(LocustTest resource) {

        return config.isTolerationsCrInjectionEnabled() ? resource.getSpec().getTolerations() : null;

    }

    public String getConfigMap(LocustTest resource) {

        return resource.getSpec().getConfigMap();

    }

    private String getNodeImage(LocustTest resource) {

        return resource.getSpec().getImage();

    }

    private String getNodeImagePullPolicy(LocustTest resource) {
        return resource.getSpec().getImagePullPolicy();
    }

    private List<String> getNodeImagePullSecrets(LocustTest resource) {
        return resource.getSpec().getImagePullSecrets();
    }

    public LocustTestAffinity getNodeAffinity(LocustTest resource) {

        return config.isAffinityCrInjectionEnabled() ? resource.getSpec().getAffinity() : null;

    }

    public String constructNodeName(LocustTest customResource, OperationalMode mode) {

        return String
            .format(NODE_NAME_TEMPLATE, customResource.getMetadata().getName(), mode.getMode())
            .replace(".", "-");

    }

    /**
     * Constructs the labels to attach to the master and worker pods.
     *
     * @param customResource The custom resource object
     * @param mode           The operational mode
     * @return A non-null, possibly empty map of labels
     */
    public Map<String, String> constructNodeLabels(final LocustTest customResource, final OperationalMode mode) {
        final Map<String, Map<String, String>> labels = Optional.ofNullable(customResource.getSpec().getLabels())
            .orElse(new HashMap<>());
        final Map<String, String> result;
        if (mode.equals(MASTER)) {
            result = labels.getOrDefault(MASTER.getMode(), new HashMap<>());
        } else {
            // Worker
            result = labels.getOrDefault(WORKER.getMode(), new HashMap<>());
        }
        log.debug("Labels attached to {} pod are {}", mode.getMode(), result);
        return result;
    }

    /**
     * Constructs the annotations to attach to the master and worker pods.
     *
     * @param customResource The custom resource object
     * @param mode           The operational mode
     * @return A non-null, possibly empty map of annotations
     */
    public Map<String, String> constructNodeAnnotations(final LocustTest customResource, final OperationalMode mode) {
        final Map<String, Map<String, String>> annotations = Optional.ofNullable(customResource.getSpec().getAnnotations())
            .orElse(new HashMap<>());
        final Map<String, String> result;
        if (mode.equals(MASTER)) {
            result = annotations.getOrDefault(MASTER.getMode(), new HashMap<>());
        } else {
            // Worker
            result = annotations.getOrDefault(WORKER.getMode(), new HashMap<>());
        }
        log.debug("Annotations attached to {} pod are {}", mode.getMode(), result);
        return result;
    }

    /**
     * Construct node command based on mode of operation
     *
     * @param customResource Custom resource object
     * @param mode           Operational mode
     * @return Node command
     */
    private List<String> constructNodeCommand(LocustTest customResource, OperationalMode mode) {

        String cmd;

        if (mode.equals(MASTER)) {
            cmd = String.format(MASTER_CMD_TEMPLATE,
                customResource.getSpec().getMasterCommandSeed(),
                MASTER_NODE_PORTS.get(0),
                customResource.getSpec().getWorkerReplicas());
        } else {
            // worker
            cmd = String.format(WORKER_CMD_TEMPLATE,
                customResource.getSpec().getWorkerCommandSeed(),
                MASTER_NODE_PORTS.get(0),
                constructNodeName(customResource, MASTER)
            );
        }

        log.debug("Constructed command: {}", cmd);
        // Split the command on <\s> to match expected container args
        return List.of(cmd.split(CONTAINER_ARGS_SEPARATOR));
    }

    /**
     * Get Replica count based on mode of operation
     *
     * @param customResource Custom resource object
     * @param mode           Operational mode
     * @return Replica count
     */
    private int getReplicaCount(LocustTest customResource, OperationalMode mode) {

        Integer replicaCount;

        if (mode.equals(MASTER)) {
            replicaCount = MASTER_NODE_REPLICA_COUNT;
        } else {
            replicaCount = customResource.getSpec().getWorkerReplicas();
        }

        log.debug("Replica count for node: {}, with mode: {}, is: {}", customResource.getMetadata().getName(), mode, replicaCount);
        return replicaCount;

    }

    private List<Integer> getNodePorts(LocustTest customResource, OperationalMode mode) {

        List<Integer> ports;

        if (mode.equals(MASTER)) {
            ports = MASTER_NODE_PORTS;
        } else {
            ports = Collections.singletonList(WORKER_NODE_PORT);
        }

        log.debug("Ports list for node: {}, with mode: {}, is: {}", customResource.getMetadata().getName(), mode, ports);
        return ports;

    }

    public Map<String, String> generateContainerEnvironmentMap() {
        HashMap<String, String> environmentMap = new HashMap<>();

        environmentMap.put(KAFKA_BOOTSTRAP_SERVERS, config.getKafkaBootstrapServers());
        environmentMap.put(KAFKA_SECURITY_ENABLED, String.valueOf(config.isKafkaSecurityEnabled()));
        environmentMap.put(KAFKA_SECURITY_PROTOCOL_CONFIG, config.getKafkaSecurityProtocol());
        environmentMap.put(KAFKA_SASL_MECHANISM, config.getKafkaSaslMechanism());
        environmentMap.put(KAFKA_SASL_JAAS_CONFIG, config.getKafkaSaslJaasConfig());
        environmentMap.put(KAFKA_USERNAME, config.getKafkaUsername());
        environmentMap.put(KAFKA_PASSWORD, config.getKafkaUserPassword());

        return environmentMap;
    }

    /**
     * Constructs a MetricsExporterContainer using the configuration settings and resource requirements.
     *
     * @return A MetricsExporterContainer instance configured with the specified settings and resource requirements.
     */
    public MetricsExporterContainer constructMetricsExporterContainer() {
        return new MetricsExporterContainer(
            EXPORTER_CONTAINER_NAME,
            config.getMetricsExporterImage(),
            config.getMetricsExporterPullPolicy(),
            config.getMetricsExporterPort(),
            this.getResourceRequirements(METRICS_EXPORTER_RESOURCE_TARGET)

        );
    }

    /**
     * Get resource request and limit for containers
     *
     * @return resource requirements
     */
    public ResourceRequirements getResourceRequirements(String target) {

        Map<String, Quantity> resourceRequests;
        Map<String, Quantity> resourceLimits;

        // Default target
        if (target.equals(DEFAULT_RESOURCE_TARGET)) {

            resourceRequests = this.getResourceRequests();
            resourceLimits = this.getResourceLimits();

            // If not default target, then the assumed target is a "Metrics Exporter" container!
            // + No need for "else if" in order to avoid unneeded checks and increased complexity
            // + in a future implementation if another "target" is introduced,
            // + the method should be updated and this comment removed.
        } else {

            resourceRequests = this.getMetricsExporterResourceRequests();
            resourceLimits = this.getMetricsExporterResourceLimits();

        }

        final var resourceRequest = new ResourceRequirements();

        // Add memory and cpu resource requests
        resourceRequest.setRequests(resourceRequests);

        // Add memory and cpu resource limits
        resourceRequest.setLimits(resourceLimits);

        return resourceRequest;

    }

    /**
     * Get requested resources based on configuration (defaults or HELM).
     *
     * @return the resources request to use
     */
    private Map<String, Quantity> getResourceRequests() {
        String memOverride = config.getPodMemRequest();
        String cpuOverride = config.getPodCpuRequest();
        String ephemeralOverride = config.getPodEphemeralStorageRequest();

        log.debug("Using resource requests - cpu: {}, mem: {}, ephemeral: {}", cpuOverride, memOverride, ephemeralOverride);

        return generateResourceOverrideMap(memOverride, cpuOverride, ephemeralOverride);
    }

    /**
     * Get resource limits based on configuration (defaults or HELM).
     *
     * @return the resource limits to use
     */
    private Map<String, Quantity> getResourceLimits() {
        String memOverride = config.getPodMemLimit();
        String cpuOverride = config.getPodCpuLimit();
        String ephemeralOverride = config.getPodEphemeralStorageLimit();

        log.debug("Using resource limits - cpu: {}, mem: {}, ephemeral: {}", cpuOverride, memOverride, ephemeralOverride);

        return generateResourceOverrideMap(memOverride, cpuOverride, ephemeralOverride);
    }

    /**
     * Get resources request for Metrics Exporter container.
     *
     * @return the resource requests to use
     */
    private Map<String, Quantity> getMetricsExporterResourceRequests() {
        String memOverride = config.getMetricsExporterMemRequest();
        String cpuOverride = config.getMetricsExporterCpuRequest();
        String ephemeralOverride = config.getMetricsExporterEphemeralStorageRequest();

        log.debug("Using resource requests for metrics exporter - cpu: {}, mem: {}, ephemeral: {}", cpuOverride, memOverride,
            ephemeralOverride);

        return generateResourceOverrideMap(memOverride, cpuOverride, ephemeralOverride);
    }

    /**
     * Get resource limits for Metrics Exporter container.
     *
     * @return the resource requests to use
     */
    private Map<String, Quantity> getMetricsExporterResourceLimits() {
        String memOverride = config.getMetricsExporterMemLimit();
        String cpuOverride = config.getMetricsExporterCpuLimit();
        String ephemeralOverride = config.getMetricsExporterEphemeralStorageLimit();

        log.debug("Using resource limits - cpu: {}, mem: {}, ephemeral: {}", cpuOverride, memOverride, ephemeralOverride);

        return generateResourceOverrideMap(memOverride, cpuOverride, ephemeralOverride);
    }

    /**
     * Generates a resource override map based on the provided memory, CPU, and ephemeral storage overrides.
     *
     * @param memOverride       The memory override value to be used for the "memory" resource.
     * @param cpuOverride       The CPU override value to be used for the "cpu" resource.
     * @param ephemeralOverride The ephemeral storage override value to be used for the "ephemeral-storage" resource. This value will be
     *                          applied only if the Kubernetes version supports "ephemeral-storage" requests.
     * @return A Map containing resource overrides for memory, CPU, and ephemeral storage.
     */
    private Map<String, Quantity> generateResourceOverrideMap(String memOverride, String cpuOverride, String ephemeralOverride) {
        Map<String, Quantity> resourceOverrideMap = new HashMap<>();

        Optional.ofNullable(memOverride)
            .filter(s -> !s.isBlank())
            .ifPresent(override -> resourceOverrideMap.put("memory", new Quantity(override)));

        Optional.ofNullable(cpuOverride)
            .filter(s -> !s.isBlank())
            .ifPresent(override -> resourceOverrideMap.put("cpu", new Quantity(override)));

        Optional.ofNullable(ephemeralOverride)
            .filter(s -> !s.isBlank())
            .ifPresent(override -> resourceOverrideMap.put("ephemeral-storage", new Quantity(ephemeralOverride)));

        return resourceOverrideMap;
    }

}
