package com.locust.operator.controller.utils.resource.manage;

import com.locust.operator.controller.dto.LoadGenerationNode;
import com.locust.operator.controller.utils.LoadGenHelpers;
import io.fabric8.kubernetes.api.model.Container;
import io.fabric8.kubernetes.api.model.ContainerBuilder;
import io.fabric8.kubernetes.api.model.ContainerPort;
import io.fabric8.kubernetes.api.model.ContainerPortBuilder;
import io.fabric8.kubernetes.api.model.EnvVar;
import io.fabric8.kubernetes.api.model.EnvVarBuilder;
import io.fabric8.kubernetes.api.model.ObjectMeta;
import io.fabric8.kubernetes.api.model.ObjectMetaBuilder;
import io.fabric8.kubernetes.api.model.PodSpec;
import io.fabric8.kubernetes.api.model.PodSpecBuilder;
import io.fabric8.kubernetes.api.model.PodTemplateSpec;
import io.fabric8.kubernetes.api.model.PodTemplateSpecBuilder;
import io.fabric8.kubernetes.api.model.Service;
import io.fabric8.kubernetes.api.model.ServiceBuilder;
import io.fabric8.kubernetes.api.model.batch.v1.Job;
import io.fabric8.kubernetes.api.model.batch.v1.JobBuilder;
import io.fabric8.kubernetes.api.model.batch.v1.JobSpec;
import io.fabric8.kubernetes.api.model.batch.v1.JobSpecBuilder;
import jakarta.inject.Singleton;
import lombok.extern.slf4j.Slf4j;
import lombok.val;

import java.util.ArrayList;
import java.util.Collections;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.stream.Collectors;

import static com.locust.operator.controller.dto.OperationalMode.MASTER;
import static com.locust.operator.controller.utils.Constants.APP_DEFAULT_LABEL;
import static com.locust.operator.controller.utils.Constants.BACKOFF_LIMIT;
import static com.locust.operator.controller.utils.Constants.DEFAULT_RESTART_POLICY;
import static com.locust.operator.controller.utils.Constants.DEFAULT_WEB_UI_PORT;
import static com.locust.operator.controller.utils.Constants.EXPORTER_CONTAINER_NAME;
import static com.locust.operator.controller.utils.Constants.EXPORTER_IMAGE;
import static com.locust.operator.controller.utils.Constants.EXPORTER_PORT_ENV_VAR;
import static com.locust.operator.controller.utils.Constants.EXPORTER_PORT_ENV_VAR_VALUE;
import static com.locust.operator.controller.utils.Constants.EXPORTER_URI_ENV_VAR;
import static com.locust.operator.controller.utils.Constants.EXPORTER_URI_ENV_VAR_VALUE;
import static com.locust.operator.controller.utils.Constants.LOCUST_COMMAND_ENV_VAR;
import static com.locust.operator.controller.utils.Constants.LOCUST_EXPORTER_PORT;
import static com.locust.operator.controller.utils.Constants.METRICS_PORT_NAME;
import static com.locust.operator.controller.utils.Constants.PORT_DEFAULT_NAME;
import static com.locust.operator.controller.utils.Constants.PROMETHEUS_IO_ENDPOINT;
import static com.locust.operator.controller.utils.Constants.PROMETHEUS_IO_PATH;
import static com.locust.operator.controller.utils.Constants.PROMETHEUS_IO_PORT;
import static com.locust.operator.controller.utils.Constants.PROMETHEUS_IO_SCRAPE;
import static com.locust.operator.controller.utils.Constants.SERVICE_SELECTOR_LABEL;
import static com.locust.operator.controller.utils.Constants.TCP_PROTOCOL;

@Slf4j
@Singleton
public class ResourceCreationHelpers {

    private final LoadGenHelpers loadGenHelpers;

    public ResourceCreationHelpers(LoadGenHelpers loadGenHelpers) {
        this.loadGenHelpers = loadGenHelpers;
    }

    /**
     * Prepare a Kubernetes Job.
     * <p>
     * Reference: <a href="https://kubernetes.io/docs/concepts/workloads/controllers/job/">Kubernetes Job Docs</a>
     *
     * @param nodeConfig Load generation configuration
     * @return Job
     */
    protected Job prepareJob(LoadGenerationNode nodeConfig, String testName) {
        return new JobBuilder()
            .withMetadata(prepareJobMetadata(nodeConfig))
            .withSpec(prepareJobSpec(nodeConfig, testName))
            .build();
    }

    /**
     * Prepare Kubernetes 'Job > Metadata'.
     * <p>
     * Reference: <a href="https://kubernetes.io/docs/concepts/workloads/controllers/job/">Kubernetes Job Docs</a>
     *
     * @param nodeConfig Load generation configuration
     * @return ObjectMeta
     */
    private ObjectMeta prepareJobMetadata(LoadGenerationNode nodeConfig) {

        // * Metadata
        ObjectMeta jobMeta = new ObjectMetaBuilder()
            .withName(nodeConfig.getName())
            .build();

        log.debug("Prepared Kubernetes 'Job > Metadata': {}", jobMeta);

        return jobMeta;

    }

    /**
     * Prepare Kubernetes 'Job > Spec'.
     * <p>
     * Reference: <a href="https://kubernetes.io/docs/concepts/workloads/controllers/job/">Kubernetes Job Docs</a>
     *
     * @param nodeConfig Load generation configuration
     * @return JobSpec
     */
    private JobSpec prepareJobSpec(LoadGenerationNode nodeConfig, String testName) {

        // * Job Spec configuration
        JobSpec jobSpec = new JobSpecBuilder()

            // Pods count
            // Setting the `Parallelism` attribute will result in k8s deploying pods to match the requested value
            // effectively enabling control over the deployed pod count.
            .withParallelism(nodeConfig.getReplicas())

            // Backoff limit
            .withBackoffLimit(BACKOFF_LIMIT)

            // Template
            .withTemplate(prepareSpecTemplate(nodeConfig, testName))

            .build();

        log.debug("Prepared Kubernetes 'Job > Spec': {}", jobSpec);

        return jobSpec;

    }

    /**
     * Prepare Kubernetes 'Job > Spec > Template'.
     * <p>
     * Reference: <a href="https://kubernetes.io/docs/concepts/workloads/controllers/job/">Kubernetes Job Docs</a>
     *
     * @param nodeConfig Load generation configuration
     * @return PodTemplateSpec
     */
    private PodTemplateSpec prepareSpecTemplate(LoadGenerationNode nodeConfig, String testName) {

        PodTemplateSpec specTemplate = new PodTemplateSpecBuilder()
            .withMetadata(prepareTemplateMetadata(nodeConfig.getName(), testName))
            .withSpec(prepareTemplateSpec(nodeConfig))
            .build();

        log.debug("Prepared Kubernetes 'Job > Spec > Template': {}", specTemplate);

        return specTemplate;

    }

    /**
     * Prepare Kubernetes 'Job > Spec > Template > Metadata'.
     * <p>
     * Reference: <a href="https://kubernetes.io/docs/concepts/workloads/controllers/job/">Kubernetes Job Docs</a>
     *
     * @param testName Performance Test name
     * @return PodTemplateSpec
     */
    private ObjectMeta prepareTemplateMetadata(String nodeName, String testName) {

        ObjectMeta templateMeta = new ObjectMetaBuilder()
            // Labels
            .addToLabels(APP_DEFAULT_LABEL, testName)
            .addToLabels(SERVICE_SELECTOR_LABEL, nodeName)

            // Annotations
            // Enable Prometheus endpoint discovery by Prometheus server
            .addToAnnotations(PROMETHEUS_IO_SCRAPE, "true")
            .addToAnnotations(PROMETHEUS_IO_PATH, PROMETHEUS_IO_ENDPOINT)
            .addToAnnotations(PROMETHEUS_IO_PORT, String.valueOf(LOCUST_EXPORTER_PORT))

            .build();

        log.debug("Prepared Kubernetes 'Job > Spec > Template > Metadata': {}", templateMeta);

        return templateMeta;

    }

    /**
     * Prepare Kubernetes 'Job > Spec > Template > Spec'.
     * <p>
     * Reference: <a href="https://kubernetes.io/docs/concepts/workloads/controllers/job/">Kubernetes Job Docs</a>
     *
     * @param nodeConfig Load generation configuration
     * @return PodTemplateSpec
     */
    private PodSpec prepareTemplateSpec(LoadGenerationNode nodeConfig) {

        PodSpec templateSpec = new PodSpecBuilder()

            // Containers
            .withContainers(prepareContainerList(nodeConfig))
            .withRestartPolicy(DEFAULT_RESTART_POLICY)
            .build();

        log.debug("Prepared Kubernetes 'Job > Spec > Template > Spec': {}", templateSpec);

        return templateSpec;

    }

    private List<Container> prepareContainerList(LoadGenerationNode nodeConfig) {

        List<Container> constantsList = new ArrayList<>();

        // Load generation container
        constantsList.add(prepareLoadGenContainer(nodeConfig));

        // Inject metrics container only if `master`
        if (nodeConfig.getOperationalMode().equals(MASTER)) {
            constantsList.add(prepareMetricsExporterContainer(nodeConfig));
        }

        return constantsList;

    }

    /**
     * Prepare locust prometheus metrics exporter container.
     * <p>
     * Reference: <a href="https://github.com/ContainerSolutions/locust_exporter">locust exporter docs</a>
     *
     * @param nodeConfig Load generation configuration
     * @return Container
     */
    private Container prepareMetricsExporterContainer(LoadGenerationNode nodeConfig) {

        HashMap<String, String> envMap = new HashMap<>();

        envMap.put(EXPORTER_URI_ENV_VAR, EXPORTER_URI_ENV_VAR_VALUE);
        envMap.put(EXPORTER_PORT_ENV_VAR, EXPORTER_PORT_ENV_VAR_VALUE);

        Container container = new ContainerBuilder()

            // Name
            .withName(EXPORTER_CONTAINER_NAME)

            // Image
            .withImage(EXPORTER_IMAGE)

            // Ports
            .withPorts(new ContainerPortBuilder().withContainerPort(LOCUST_EXPORTER_PORT).build())

            // Environment
            .withEnv(prepareContainerEnvironmentVariables(envMap, nodeConfig))

            .build();

        log.debug("Prepared Kubernetes metrics exporter container: {}", container);

        return container;
    }

    /**
     * Prepare a load generation container.
     * <p>
     * Reference: <a href="https://kubernetes.io/docs/concepts/containers/">Kubernetes containers Docs</a>
     *
     * @param nodeConfig Load generation configuration
     * @return Container
     */
    private Container prepareLoadGenContainer(LoadGenerationNode nodeConfig) {
        Container container = new ContainerBuilder()

            // Name
            .withName(nodeConfig.getName())

            // Resource config
            .withResources(loadGenHelpers.getResourceRequirements())

            // Image
            .withImage(nodeConfig.getImage())

            // Ports
            .withPorts(prepareContainerPorts(nodeConfig.getPorts()))

            // Environment
            .withEnv(prepareContainerEnvironmentVariables(loadGenHelpers.generateContainerEnvironmentMap(), nodeConfig))

            .build();

        log.debug("Prepared Kubernetes load generator container: {}", container);

        return container;
    }

    /**
     * Prepare container Environment variable.
     * <p>
     * Reference: <a href="https://kubernetes.io/docs/concepts/containers/">Kubernetes containers Docs</a>
     *
     * @param envMap Environment variable map
     * @return ContainerPort
     */
    private List<EnvVar> prepareContainerEnvironmentVariables(Map<String, String> envMap, LoadGenerationNode nodeConfig) {

        // Update LOCUST_COMMAND with runtime config
        envMap.replace(LOCUST_COMMAND_ENV_VAR, nodeConfig.getCommand());

        List<EnvVar> containerEnvVars = envMap
            .entrySet()
            .stream()
            .map(entry -> new EnvVarBuilder()
                .withName(entry.getKey())
                .withValue(entry.getValue())
                .build())
            .collect(Collectors.toList());

        log.debug("Prepared container environment variable list: {}", containerEnvVars);

        return containerEnvVars;

    }

    /**
     * Prepare container ports.
     * <p>
     * Reference: <a href="https://kubernetes.io/docs/concepts/containers/">Kubernetes containers Docs</a>
     *
     * @param portsList Container port list
     * @return ContainerPort
     */
    private List<ContainerPort> prepareContainerPorts(List<Integer> portsList) {

        List<ContainerPort> containerPortList = portsList
            .stream()
            .map(port -> new ContainerPortBuilder().withContainerPort(port).build())
            .collect(Collectors.toList());

        log.debug("Prepared container ports list: {}", containerPortList);

        return containerPortList;

    }

    protected Service prepareService(LoadGenerationNode nodeConfig) {

        // Initial service configuration
        var serviceConfig = new ServiceBuilder()

            // Metadata
            .withNewMetadata()
            .withName(nodeConfig.getName())
            .endMetadata()

            // Spec
            .withNewSpec()
            .withSelector(Collections.singletonMap(SERVICE_SELECTOR_LABEL, nodeConfig.getName()));

        // Map ports
        nodeConfig.getPorts()
            .stream()
            .filter(port -> !port.equals(DEFAULT_WEB_UI_PORT))
            .forEach(port -> {

                val portName = PORT_DEFAULT_NAME + port;

                serviceConfig
                    .addNewPort()
                    .withName(portName)
                    .withProtocol(TCP_PROTOCOL)
                    .withPort(port)
                    .endPort();
            });

        // Metrics port
        serviceConfig
            .addNewPort()
            .withName(METRICS_PORT_NAME)
            .withProtocol(TCP_PROTOCOL)
            .withPort(LOCUST_EXPORTER_PORT)
            .endPort();

        // Finalize building the service object
        var service = serviceConfig.endSpec();

        return service.build();

    }

}
