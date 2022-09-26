package com.locust.operator.controller.utils;

import com.locust.operator.controller.config.SysConfig;
import com.locust.operator.controller.dto.LoadGenerationNode;
import com.locust.operator.controller.dto.OperationalMode;
import com.locust.operator.customresource.LocustTest;
import io.fabric8.kubernetes.api.model.Quantity;
import io.fabric8.kubernetes.api.model.ResourceRequirements;
import jakarta.inject.Singleton;
import lombok.extern.slf4j.Slf4j;

import javax.validation.constraints.NotNull;

import java.util.Collections;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

import static com.locust.operator.controller.dto.OperationalMode.MASTER;
import static com.locust.operator.controller.utils.Constants.CONTAINER_ARGS_SEPARATOR;
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
            constructNodeCommand(resource, mode),
            mode,
            resource.getSpec().getImage(),
            getReplicaCount(resource, mode),
            getNodePorts(resource, mode),
            resource.getSpec().getConfigMap());

    }

    public String constructNodeName(LocustTest customResource, OperationalMode mode) {

        return String
            .format(NODE_NAME_TEMPLATE, customResource.getMetadata().getName(), mode.getMode())
            .replace(".", "-");

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
     * Get resource request and limit for pods
     *
     * @return resource requirements
     */
    public ResourceRequirements getResourceRequirements() {

        final var resourceRequest = new ResourceRequirements();

        // Add memory and cpu resource requests
        resourceRequest.setRequests(this.getResourceRequests());

        // Add memory and cpu resource limits
        resourceRequest.setLimits(this.getResourceLimits());

        return resourceRequest;

    }

    /*
     * Resource requests are guaranteed by the Kubernetesruntime.
     *
     * @return the resource requests to use
     */
    private Map<String, Quantity> getResourceRequests() {
        String memOverride = config.getPodMemRequest();
        String cpuOverride = config.getPodCpuRequest();
        String ephemeralOverride = config.getPodEphemeralStorageRequest();

        log.debug("Using resource requests - cpu: {}, mem: {}, ephemeral: {}", cpuOverride, memOverride, ephemeralOverride);

        return generateResourceOverrideMap(memOverride, cpuOverride, ephemeralOverride);
    }

    /*
     * Resource limits for pods.
     *
     * @return the resource requests to use
     */
    private Map<String, Quantity> getResourceLimits() {
        String memOverride = config.getPodMemLimit();
        String cpuOverride = config.getPodCpuLimit();
        String ephemeralOverride = config.getPodEphemeralStorageLimit();

        log.debug("Using resource limits - cpu: {}, mem: {}, ephemeral: {}", cpuOverride, memOverride, ephemeralOverride);

        return generateResourceOverrideMap(memOverride, cpuOverride, ephemeralOverride);
    }

    @NotNull
    private Map<String, Quantity> generateResourceOverrideMap(String memOverride, String cpuOverride, String ephemeralOverride) {
        Map<String, Quantity> resourceOverrideMap = new HashMap<>();

        resourceOverrideMap.put("memory", new Quantity(memOverride));

        resourceOverrideMap.put("cpu", new Quantity(cpuOverride));

        // This conditional is to enable the operator to run on older kubernetes versions that doesn't support "ephemeral-storage" requests
        if (!config.getPodEphemeralStorageRequest().isBlank()) {
            resourceOverrideMap.put("ephemeral-storage", new Quantity(ephemeralOverride));

        }

        return resourceOverrideMap;
    }

}
