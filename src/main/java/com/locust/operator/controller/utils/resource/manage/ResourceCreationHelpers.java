package com.locust.operator.controller.utils.resource.manage;

import com.locust.operator.controller.dto.LoadGenerationNode;
import com.locust.operator.controller.dto.MetricsExporterContainer;
import com.locust.operator.controller.utils.LoadGenHelpers;
import io.fabric8.kubernetes.api.model.Affinity;
import io.fabric8.kubernetes.api.model.AffinityBuilder;
import io.fabric8.kubernetes.api.model.ConfigMapVolumeSource;
import io.fabric8.kubernetes.api.model.ConfigMapVolumeSourceBuilder;
import io.fabric8.kubernetes.api.model.Container;
import io.fabric8.kubernetes.api.model.ContainerBuilder;
import io.fabric8.kubernetes.api.model.ContainerPort;
import io.fabric8.kubernetes.api.model.ContainerPortBuilder;
import io.fabric8.kubernetes.api.model.EnvVar;
import io.fabric8.kubernetes.api.model.EnvVarBuilder;
import io.fabric8.kubernetes.api.model.LocalObjectReference;
import io.fabric8.kubernetes.api.model.LocalObjectReferenceBuilder;
import io.fabric8.kubernetes.api.model.NodeAffinity;
import io.fabric8.kubernetes.api.model.NodeAffinityBuilder;
import io.fabric8.kubernetes.api.model.NodeSelector;
import io.fabric8.kubernetes.api.model.NodeSelectorBuilder;
import io.fabric8.kubernetes.api.model.NodeSelectorRequirement;
import io.fabric8.kubernetes.api.model.NodeSelectorRequirementBuilder;
import io.fabric8.kubernetes.api.model.NodeSelectorTermBuilder;
import io.fabric8.kubernetes.api.model.ObjectMeta;
import io.fabric8.kubernetes.api.model.ObjectMetaBuilder;
import io.fabric8.kubernetes.api.model.PodSpec;
import io.fabric8.kubernetes.api.model.PodSpecBuilder;
import io.fabric8.kubernetes.api.model.PodTemplateSpec;
import io.fabric8.kubernetes.api.model.PodTemplateSpecBuilder;
import io.fabric8.kubernetes.api.model.Service;
import io.fabric8.kubernetes.api.model.ServiceBuilder;
import io.fabric8.kubernetes.api.model.Toleration;
import io.fabric8.kubernetes.api.model.TolerationBuilder;
import io.fabric8.kubernetes.api.model.Volume;
import io.fabric8.kubernetes.api.model.VolumeBuilder;
import io.fabric8.kubernetes.api.model.VolumeMount;
import io.fabric8.kubernetes.api.model.VolumeMountBuilder;
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
import java.util.Optional;
import java.util.stream.Collectors;

import static com.locust.operator.controller.dto.OperationalMode.MASTER;
import static com.locust.operator.controller.dto.OperatorType.EQUAL;
import static com.locust.operator.controller.utils.Constants.APP_DEFAULT_LABEL;
import static com.locust.operator.controller.utils.Constants.BACKOFF_LIMIT;
import static com.locust.operator.controller.utils.Constants.DEFAULT_MOUNT_PATH;
import static com.locust.operator.controller.utils.Constants.LIB_MOUNT_PATH;
import static com.locust.operator.controller.utils.Constants.DEFAULT_NODE_MATCH_EXPRESSION_OPERATOR;
import static com.locust.operator.controller.utils.Constants.DEFAULT_RESOURCE_TARGET;
import static com.locust.operator.controller.utils.Constants.DEFAULT_RESTART_POLICY;
import static com.locust.operator.controller.utils.Constants.DEFAULT_WEB_UI_PORT;
import static com.locust.operator.controller.utils.Constants.EXPORTER_PORT_ENV_VAR;
import static com.locust.operator.controller.utils.Constants.EXPORTER_URI_ENV_VAR;
import static com.locust.operator.controller.utils.Constants.EXPORTER_URI_ENV_VAR_VALUE;
import static com.locust.operator.controller.utils.Constants.MANAGED_BY_LABEL_KEY;
import static com.locust.operator.controller.utils.Constants.MANAGED_BY_LABEL_VALUE;
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
            .withTtlSecondsAfterFinished(nodeConfig.getTtlSecondsAfterFinished())

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
            .withMetadata(prepareTemplateMetadata(nodeConfig, testName))
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
     * @param nodeConfig The node configuration object.
     * @param testName   Test name.
     * @return PodTemplateSpec.
     */
    private ObjectMeta prepareTemplateMetadata(LoadGenerationNode nodeConfig, String testName) {

        ObjectMeta templateMeta = new ObjectMetaBuilder()
            // Labels
            .addToLabels(APP_DEFAULT_LABEL, testName)
            .addToLabels(SERVICE_SELECTOR_LABEL, nodeConfig.getName())
            .addToLabels(MANAGED_BY_LABEL_KEY, MANAGED_BY_LABEL_VALUE)
            .addToLabels(nodeConfig.getLabels())

            // Annotations
            // Enable Prometheus endpoint discovery by Prometheus server
            .addToAnnotations(PROMETHEUS_IO_SCRAPE, "true")
            .addToAnnotations(PROMETHEUS_IO_PATH, PROMETHEUS_IO_ENDPOINT)
            .addToAnnotations(PROMETHEUS_IO_PORT, String.valueOf(loadGenHelpers.constructMetricsExporterContainer().getExporterPort()))
            .addToAnnotations(nodeConfig.getAnnotations())

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
            // images
            .withImagePullSecrets(prepareImagePullSecrets(nodeConfig))

            // Containers
            .withContainers(prepareContainerList(nodeConfig))
            .withVolumes(prepareVolumesList(nodeConfig))
            .withAffinity(prepareAffinity(nodeConfig))
            .withTolerations(prepareTolerations(nodeConfig))
            .withRestartPolicy(DEFAULT_RESTART_POLICY)
            .build();

        log.debug("Prepared Kubernetes 'Job > Spec > Template > Spec': {}", templateSpec);

        return templateSpec;

    }

    private List<LocalObjectReference> prepareImagePullSecrets(LoadGenerationNode nodeConfig) {
        final List<LocalObjectReference> references = new ArrayList<>();

        if (nodeConfig.getImagePullSecrets() != null) {
            references.addAll(
                nodeConfig.getImagePullSecrets()
                    .stream()
                    .map(secretName -> new LocalObjectReferenceBuilder().withName(secretName).build())
                    .toList()
            );
        }

        log.debug("Prepared image pull secrets: {}", references);

        return references;
    }

    private List<Volume> prepareVolumesList(LoadGenerationNode nodeConfig) {

        List<Volume> volumeList = new ArrayList<>();

        if (nodeConfig.getConfigMap() != null) {
            volumeList.add(prepareVolume(nodeConfig));
        }

        if (nodeConfig.getLibConfigMap() != null) {
            volumeList.add(prepareLibVolume(nodeConfig));
        }

        return volumeList;

    }

    private Affinity prepareAffinity(LoadGenerationNode nodeConfig) {

        // Construct Affinity
        var affinityBuilder = new AffinityBuilder();

        //! Note for future feature extensions:
        //! When adding support for more "Affinity" options, the evaluation inside the `if` condition is to be split into several checks.
        if (nodeConfig.getAffinity() != null && nodeConfig.getAffinity().getNodeAffinity() != null) {
            affinityBuilder.withNodeAffinity(prepareNodeAffinity(nodeConfig));
        }

        var affinity = affinityBuilder.build();
        log.debug("Prepared pod affinity: '{}'", affinity);

        return affinity;

    }

    private NodeAffinity prepareNodeAffinity(LoadGenerationNode nodeConfig) {

        var nodeAffinityBuilder = new NodeAffinityBuilder();

        // Prepare Required during scheduling node selector
        var requiredDuringSchedulingNodeSelector = prepareRequiredDuringSchedulingNodeSelector(nodeConfig);

        nodeAffinityBuilder.withRequiredDuringSchedulingIgnoredDuringExecution(requiredDuringSchedulingNodeSelector);

        return nodeAffinityBuilder.build();

    }

    private NodeSelector prepareRequiredDuringSchedulingNodeSelector(LoadGenerationNode nodeConfig) {

        // Required during scheduling
        List<NodeSelectorRequirement> matchExpressions = new ArrayList<>();

        final var requiredDuringScheduling = Optional.ofNullable(
            nodeConfig.getAffinity().getNodeAffinity().getRequiredDuringSchedulingIgnoredDuringExecution()).orElse(new HashMap<>());

        requiredDuringScheduling.forEach((requiredAffinityKey, requiredAffinityValue) -> matchExpressions
            .add(new NodeSelectorRequirementBuilder().withKey(requiredAffinityKey).withOperator(DEFAULT_NODE_MATCH_EXPRESSION_OPERATOR)
                .withValues(requiredAffinityValue).build()));

        var nodeSelectorTerms = new NodeSelectorTermBuilder().withMatchExpressions(matchExpressions).build();

        return new NodeSelectorBuilder().withNodeSelectorTerms(nodeSelectorTerms).build();
    }

    private List<Toleration> prepareTolerations(LoadGenerationNode nodeConfig) {

        List<Toleration> tolerations = new ArrayList<>();

        if (nodeConfig.getTolerations() != null) {

            // For each configured node toleration from the Custom Resource, build a toleration object and add it to list
            nodeConfig.getTolerations().forEach(nodeToleration -> {
                var tolerationBuilder = new TolerationBuilder();
                tolerationBuilder
                    .withKey(nodeToleration.getKey())
                    .withOperator(nodeToleration.getOperator())
                    .withEffect(nodeToleration.getEffect());

                if (nodeToleration.getOperator().equals(EQUAL.getType())) {
                    tolerationBuilder.withValue(nodeToleration.getValue());
                }

                tolerations.add(tolerationBuilder.build());
            });
        }

        log.debug("Prepared pod tolerations: '{}'", tolerations);
        return tolerations;

    }

    private static Volume prepareVolume(LoadGenerationNode nodeConfig) {
        return new VolumeBuilder()
            .withName(nodeConfig.getName())
            .withConfigMap(prepareConfigMapSource(nodeConfig))
            .build();
    }

    private static Volume prepareLibVolume(LoadGenerationNode nodeConfig) {
        return new VolumeBuilder()
            .withName("lib")
            .withConfigMap(prepareLibConfigMapSource(nodeConfig))
            .build();
    }

    private static ConfigMapVolumeSource prepareConfigMapSource(LoadGenerationNode nodeConfig) {
        return new ConfigMapVolumeSourceBuilder()
            .withName(nodeConfig.getConfigMap())
            .build();
    }

    private static ConfigMapVolumeSource prepareLibConfigMapSource(LoadGenerationNode nodeConfig) {
        return new ConfigMapVolumeSourceBuilder()
            .withName(nodeConfig.getLibConfigMap())
            .build();
    }

    private List<Container> prepareContainerList(LoadGenerationNode nodeConfig) {

        List<Container> constantsList = new ArrayList<>();

        // Load generation container
        constantsList.add(prepareLoadGenContainer(nodeConfig));

        // Inject metrics container only if `master`
        if (nodeConfig.getOperationalMode().equals(MASTER)) {
            constantsList.add(prepareMetricsExporterContainer(loadGenHelpers.constructMetricsExporterContainer()));
        }

        return constantsList;

    }

    /**
     * Prepare locust prometheus metrics exporter container.
     * <p>
     * Reference for default exporter: <a href="https://github.com/ContainerSolutions/locust_exporter">locust exporter docs</a>
     *
     * @param exporterContainer The metrics exporter container
     * @return Container
     */
    private Container prepareMetricsExporterContainer(final MetricsExporterContainer exporterContainer) {

        HashMap<String, String> envMap = new HashMap<>();

        envMap.put(EXPORTER_URI_ENV_VAR, EXPORTER_URI_ENV_VAR_VALUE);
        envMap.put(EXPORTER_PORT_ENV_VAR, String.format(":%s", exporterContainer.getExporterPort()));

        Container container = new ContainerBuilder()

            // Name
            .withName(exporterContainer.getContainerName())

            // Image
            .withImage(exporterContainer.getContainerImage())
            .withImagePullPolicy(exporterContainer.getPullPolicy())

            // Resources
            .withResources(exporterContainer.getResourceRequirements())

            // Ports
            .withPorts(new ContainerPortBuilder().withContainerPort(exporterContainer.getExporterPort()).build())

            // Environment
            .withEnv(prepareContainerEnvironmentVariables(envMap))

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
            .withResources(loadGenHelpers.getResourceRequirements(DEFAULT_RESOURCE_TARGET))

            // Image
            .withImage(nodeConfig.getImage())
            .withImagePullPolicy(nodeConfig.getImagePullPolicy())

            // Ports
            .withPorts(prepareContainerPorts(nodeConfig.getPorts()))

            // Environment
            .withEnv(prepareContainerEnvironmentVariables(loadGenHelpers.generateContainerEnvironmentMap()))

            // Container command
            .withArgs(nodeConfig.getCommand())

            // Mount configMap as volume
            .withVolumeMounts(prepareVolumeMounts(nodeConfig))

            .build();

        log.debug("Prepared Kubernetes load generator container: {}", container);

        return container;
    }

    private List<VolumeMount> prepareVolumeMounts(LoadGenerationNode nodeConfig) {

        List<VolumeMount> mounts = new ArrayList<>();

        if (nodeConfig.getConfigMap() != null) {
            // Prepare main configMap mount
            mounts.add(new VolumeMountBuilder()
                .withName(nodeConfig.getName())
                .withMountPath(DEFAULT_MOUNT_PATH)
                .withReadOnly(false)
                .build());
        }

        if (nodeConfig.getLibConfigMap() != null) {
            // Prepare lib configMap mount
            mounts.add(new VolumeMountBuilder()
                .withName("lib")
                .withMountPath(LIB_MOUNT_PATH)
                .withReadOnly(false)
                .build());
        }

        return mounts;

    }

    /**
     * Prepare container Environment variable.
     * <p>
     * Reference: <a href="https://kubernetes.io/docs/concepts/containers/">Kubernetes containers Docs</a>
     *
     * @param envMap Environment variable map
     * @return ContainerPort
     */
    private List<EnvVar> prepareContainerEnvironmentVariables(Map<String, String> envMap) {

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
            .withPort(loadGenHelpers.constructMetricsExporterContainer().getExporterPort())
            .endPort();

        // Finalize building the service object
        var service = serviceConfig.endSpec();

        return service.build();

    }

}
