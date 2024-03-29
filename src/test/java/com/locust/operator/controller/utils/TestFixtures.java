package com.locust.operator.controller.utils;

import com.locust.operator.controller.config.SysConfig;
import com.locust.operator.controller.dto.LoadGenerationNode;
import com.locust.operator.controller.dto.MetricsExporterContainer;
import com.locust.operator.controller.dto.OperationalMode;
import com.locust.operator.customresource.LocustTest;
import com.locust.operator.customresource.internaldto.LocustTestAffinity;
import com.locust.operator.customresource.internaldto.LocustTestNodeAffinity;
import com.locust.operator.customresource.internaldto.LocustTestToleration;
import io.fabric8.kubernetes.api.model.KubernetesResourceList;
import io.fabric8.kubernetes.api.model.LocalObjectReference;
import io.fabric8.kubernetes.api.model.NamespaceBuilder;
import io.fabric8.kubernetes.api.model.PodList;
import io.fabric8.kubernetes.api.model.Quantity;
import io.fabric8.kubernetes.api.model.ResourceRequirements;
import io.fabric8.kubernetes.api.model.ServiceList;
import io.fabric8.kubernetes.api.model.batch.v1.JobList;
import io.fabric8.kubernetes.client.KubernetesClient;
import lombok.NoArgsConstructor;
import lombok.SneakyThrows;
import lombok.extern.slf4j.Slf4j;
import lombok.val;

import java.util.Collections;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.stream.Collectors;

import static com.github.stefanbirkner.systemlambda.SystemLambda.withEnvironmentVariable;
import static com.locust.operator.controller.TestFixtures.REPLICAS;
import static com.locust.operator.controller.dto.OperationalMode.MASTER;
import static com.locust.operator.controller.dto.OperatorType.EQUAL;
import static com.locust.operator.controller.utils.Constants.CONTAINER_ARGS_SEPARATOR;
import static com.locust.operator.controller.utils.Constants.EXPORTER_CONTAINER_NAME;
import static com.locust.operator.controller.utils.Constants.KAFKA_BOOTSTRAP_SERVERS;
import static com.locust.operator.controller.utils.Constants.KAFKA_PASSWORD;
import static com.locust.operator.controller.utils.Constants.KAFKA_SASL_JAAS_CONFIG;
import static com.locust.operator.controller.utils.Constants.KAFKA_SASL_MECHANISM;
import static com.locust.operator.controller.utils.Constants.KAFKA_SECURITY_ENABLED;
import static com.locust.operator.controller.utils.Constants.KAFKA_SECURITY_PROTOCOL_CONFIG;
import static com.locust.operator.controller.utils.Constants.KAFKA_USERNAME;
import static lombok.AccessLevel.PRIVATE;
import static org.assertj.core.api.SoftAssertions.assertSoftly;
import static org.mockito.Mockito.when;

@Slf4j
@NoArgsConstructor(access = PRIVATE)
public class TestFixtures {

    public static final List<Integer> DEFAULT_MASTER_PORT_LIST = List.of(5557, 5558, 8089);
    public static final List<Integer> DEFAULT_WORKER_PORT_LIST = List.of(8080);
    public static final Integer MASTER_REPLICA_COUNT = 1;
    public static final String DEFAULT_SEED_COMMAND = "--locustfile src/GQ/src/demo.py";
    public static final String DEFAULT_TEST_IMAGE = "xlocust:latest";
    public static final String DEFAULT_METRICS_IMAGE = "containersol/locust_exporter:v0.5.0";
    public static final int EXPECTED_GENERIC_RESOURCE_COUNT = 1;
    public static final int EXPECTED_SERVICE_RESOURCE_COUNT = 2;
    public static final String K8S_SERVER_URL_ENV_VAR = "KUBERNETES_MASTER";
    public static final String MOCK_KAFKA_BOOTSTRAP_VALUE = "localhost:9092";
    public static final boolean MOCK_SECURITY_VALUE = true;
    public static final boolean MOCK_AFFINITY_INJECTION_VALUE = true;
    public static final boolean MOCK_TOLERATION_INJECTION_VALUE = true;
    public static final String MOCK_SECURITY_PROTOCOL_VALUE = "SASL_PLAINTEXT";
    public static final String MOCK_SASL_MECHANISM_VALUE = "SCRAM-SHA-512";
    public static final String MOCK_SASL_JAAS_CONFIG_VALUE = "placeholder";
    public static final String MOCK_USERNAME = "localKafkaUser";
    public static final String MOCK_PASSWORD = "localKafkaPassword";
    public static final String MOCK_POD_MEM = "1024Mi";
    public static final String MOCK_POD_CPU = "1000m";
    public static final String MOCK_POD_EPHEMERAL_STORAGE = "50M";
    public static final Integer MOCK_POD_PORT = 9646;
    public static final Integer MOCK_TTL_SECONDS_AFTER_FINISHED = 60;
    public static final Map<String, String> DEFAULT_MASTER_LABELS = Map.of("role", "master");
    public static final Map<String, String> DEFAULT_WORKER_LABELS = Map.of("role", "worker");
    public static final Map<String, String> DEFAULT_MASTER_ANNOTATIONS = Map.of("locust.io/role", "master");
    public static final Map<String, String> DEFAULT_WORKER_ANNOTATIONS = new HashMap<>();

    public static void assertNodeConfig(LocustTest customResource, LoadGenerationNode generatedNodeConfig,
        OperationalMode mode) {

        String expectedConfigName = customResource.getMetadata().getName().replace('.', '-');

        Map<String, String> expectedLabels = mode.equals(MASTER) ? DEFAULT_MASTER_LABELS : DEFAULT_WORKER_LABELS;

        Map<String, String> expectedAnnotations = mode.equals(MASTER) ? DEFAULT_MASTER_ANNOTATIONS : DEFAULT_WORKER_ANNOTATIONS;

        Integer expectedReplicas = mode.equals(MASTER) ? MASTER_REPLICA_COUNT : customResource.getSpec().getWorkerReplicas();

        List<Integer> expectedPortList = mode.equals(MASTER) ? DEFAULT_MASTER_PORT_LIST : DEFAULT_WORKER_PORT_LIST;

        assertSoftly(softly -> {
            softly.assertThat(generatedNodeConfig.getName()).contains(expectedConfigName);
            softly.assertThat(generatedNodeConfig.getLabels()).isEqualTo(expectedLabels);
            softly.assertThat(generatedNodeConfig.getAnnotations()).isEqualTo(expectedAnnotations);
            softly.assertThat(generatedNodeConfig.getTtlSecondsAfterFinished()).isEqualTo(MOCK_TTL_SECONDS_AFTER_FINISHED);
            softly.assertThat(generatedNodeConfig.getOperationalMode()).isEqualTo(mode);
            softly.assertThat(generatedNodeConfig.getPorts()).isEqualTo(expectedPortList);
            softly.assertThat(generatedNodeConfig.getReplicas()).isEqualTo(expectedReplicas);
        });
    }

    public static LoadGenerationNode prepareNodeConfig(String nodeName, OperationalMode mode) {
        var nodeConfig = LoadGenerationNode.builder()
            .name(nodeName)
            .labels(mode.equals(MASTER) ? DEFAULT_MASTER_LABELS : DEFAULT_WORKER_LABELS)
            .annotations(mode.equals(MASTER) ? DEFAULT_MASTER_ANNOTATIONS : DEFAULT_WORKER_ANNOTATIONS)
            .command(List.of(DEFAULT_SEED_COMMAND.split(CONTAINER_ARGS_SEPARATOR)))
            .operationalMode(mode)
            .image(DEFAULT_TEST_IMAGE)
            .replicas(mode.equals(MASTER) ? MASTER_REPLICA_COUNT : REPLICAS)
            .ports(mode.equals(MASTER) ? DEFAULT_MASTER_PORT_LIST : DEFAULT_WORKER_PORT_LIST)
            .build();

        log.debug("Created node configuration: {}", nodeConfig);
        return nodeConfig;
    }

    public static LoadGenerationNode prepareNodeConfigWithTtlSecondsAfterFinished(
        String nodeName, OperationalMode mode, Integer ttlSecondsAfterFinished) {
        var nodeConfig = LoadGenerationNode.builder()
            .name(nodeName)
            .labels(mode.equals(MASTER) ? DEFAULT_MASTER_LABELS : DEFAULT_WORKER_LABELS)
            .annotations(mode.equals(MASTER) ? DEFAULT_MASTER_ANNOTATIONS : DEFAULT_WORKER_ANNOTATIONS)
            .ttlSecondsAfterFinished(ttlSecondsAfterFinished)
            .command(List.of(DEFAULT_SEED_COMMAND.split(CONTAINER_ARGS_SEPARATOR)))
            .operationalMode(mode)
            .image(DEFAULT_TEST_IMAGE)
            .replicas(mode.equals(MASTER) ? MASTER_REPLICA_COUNT : REPLICAS)
            .ports(mode.equals(MASTER) ? DEFAULT_MASTER_PORT_LIST : DEFAULT_WORKER_PORT_LIST)
            .build();

        log.debug("Created node configuration: {}", nodeConfig);
        return nodeConfig;
    }

    public static LoadGenerationNode prepareNodeConfigWithNodeAffinity(String nodeName, OperationalMode mode, String affinityKey,
        String affinityValue) {

        // Init instances
        val nodeAffinity = new LocustTestNodeAffinity();
        val affinity = new LocustTestAffinity();
        val nodeConfig = prepareNodeConfig(nodeName, mode);

        // Set affinity
        nodeAffinity.setRequiredDuringSchedulingIgnoredDuringExecution(Map.of(affinityKey, affinityValue));
        affinity.setNodeAffinity(nodeAffinity);

        // Push affinity config to object
        nodeConfig.setAffinity(affinity);
        log.debug("Created node configuration with nodeAffinity: {}", nodeConfig);

        return nodeConfig;

    }

    public static LoadGenerationNode prepareNodeConfigWithTolerations(String nodeName, OperationalMode mode,
        LocustTestToleration toleration) {

        val nodeConfig = prepareNodeConfig(nodeName, mode);
        nodeConfig.setTolerations(Collections.singletonList(toleration));

        return nodeConfig;

    }

    public static LoadGenerationNode prepareNodeConfigWithPullPolicyAndSecrets(
        String nodeName, OperationalMode mode, String pullPolicy, List<String> pullSecrets) {

        val nodeConfig = prepareNodeConfig(nodeName, mode);
        nodeConfig.setImagePullPolicy(pullPolicy);
        nodeConfig.setImagePullSecrets(pullSecrets);

        return nodeConfig;

    }

    public static void assertK8sServiceCreation(String nodeName, ServiceList serviceList) {
        assertK8sResourceCreation(nodeName, serviceList, EXPECTED_SERVICE_RESOURCE_COUNT);
    }

    public static <T extends KubernetesResourceList<?>> void assertK8sResourceCreation(String nodeName, T resourceList) {
        assertK8sResourceCreation(nodeName, resourceList, EXPECTED_GENERIC_RESOURCE_COUNT);
    }

    private static <T extends KubernetesResourceList<?>> void assertK8sResourceCreation(String nodeName, T resourceList,
        int expectedResourceCount) {
        val resourceNamesList = extractNames(resourceList);
        log.debug("Acquired resource list: {}", resourceNamesList);

        assertSoftly(softly -> {
            softly.assertThat(resourceList.getItems().size()).isEqualTo(expectedResourceCount);
            softly.assertThat(resourceNamesList).contains(nodeName);
        });
    }

    private static <T extends KubernetesResourceList<?>> List<String> extractNames(T resourceList) {
        return resourceList.getItems().stream()
            .map(item -> item.getMetadata().getName())
            .collect(Collectors.toList());
    }

    public static void createNamespace(KubernetesClient testClient, String namespace) {

        testClient.namespaces()
            .resource(new NamespaceBuilder()
                .withNewMetadata()
                .withName(namespace)
                .endMetadata()
                .build())
            .serverSideApply();
    }

    public static void assertImagePullData(LoadGenerationNode nodeConfig, PodList podList) {

        podList.getItems().forEach(pod -> {
            final List<String> references = pod.getSpec()
                .getImagePullSecrets()
                .stream()
                .map(LocalObjectReference::getName)
                .toList();

            assertSoftly(softly -> softly.assertThat(references).isEqualTo(nodeConfig.getImagePullSecrets()));

            pod.getSpec()
                .getContainers()
                .forEach(container -> assertSoftly(
                    softly -> softly.assertThat(container.getImagePullPolicy()).isEqualTo(nodeConfig.getImagePullPolicy())));
        });
    }

    public static void assertK8sTtlSecondsAfterFinished(JobList jobList, Integer ttlSecondsAfterFinished) {
        jobList.getItems().forEach(job -> {
            val actualTtlSecondsAfterFinished = job.getSpec().getTtlSecondsAfterFinished();
            assertSoftly(softly -> softly.assertThat(actualTtlSecondsAfterFinished).isEqualTo(ttlSecondsAfterFinished));
        });
    }

    public static void assertK8sNodeAffinity(LoadGenerationNode nodeConfig, JobList jobList, String k8sNodeLabelKey) {

        jobList.getItems().forEach(job -> {
            val nodeSelectorTerms = job.getSpec().getTemplate().getSpec().getAffinity().getNodeAffinity()
                .getRequiredDuringSchedulingIgnoredDuringExecution().getNodeSelectorTerms();

            nodeSelectorTerms.forEach(selectorTerm -> {
                val actualSelectorKey = selectorTerm.getMatchExpressions().get(0).getKey();
                val actualSelectorValue = selectorTerm.getMatchExpressions().get(0).getValues().get(0);
                val desiredSelectorValue = nodeConfig.getAffinity().getNodeAffinity().getRequiredDuringSchedulingIgnoredDuringExecution()
                    .get(k8sNodeLabelKey);

                assertSoftly(softly -> {
                    softly.assertThat(actualSelectorKey).isEqualTo(k8sNodeLabelKey);
                    softly.assertThat(actualSelectorValue).isEqualTo(desiredSelectorValue);
                });
            });

        });

    }

    public static void assertK8sTolerations(JobList jobList, LocustTestToleration expectedToleration) {

        jobList.getItems().forEach(job -> {
            val actualTolerations = job.getSpec().getTemplate().getSpec().getTolerations();

            assertSoftly(softly -> {
                softly.assertThat(actualTolerations.get(0).getKey()).isEqualTo(expectedToleration.getKey());
                softly.assertThat(actualTolerations.get(0).getEffect()).isEqualTo(expectedToleration.getEffect());
                softly.assertThat(actualTolerations.get(0).getOperator()).isEqualTo(expectedToleration.getOperator());

                if (expectedToleration.getOperator().equals(EQUAL.getType())) {
                    softly.assertThat(actualTolerations.get(0).getValue()).isEqualTo(expectedToleration.getValue());
                }
            });

        });

    }

    /**
     * Method to run `runnable` methods while injection the "KUBERNETES_MASTER" in the run environment. This is required as the core methods
     * uses an internally created k8s client that searches for configuration in a specific order. Injecting the environment variable this
     * way allows the internal client to connect to the mock server.
     *
     * @param mockServerUrl Mock server URL
     * @param runnable      Runnable object to run
     */
    @SneakyThrows
    public static void executeWithK8sMockServer(String mockServerUrl, Runnable runnable) {

        withEnvironmentVariable(K8S_SERVER_URL_ENV_VAR, mockServerUrl)
            .execute(runnable::run);

    }

    public static Map<String, String> containerEnvironmentMap() {
        HashMap<String, String> environmentMap = new HashMap<>();

        environmentMap.put(KAFKA_BOOTSTRAP_SERVERS, MOCK_KAFKA_BOOTSTRAP_VALUE);
        environmentMap.put(KAFKA_SECURITY_ENABLED, String.valueOf(MOCK_SECURITY_VALUE));
        environmentMap.put(KAFKA_SECURITY_PROTOCOL_CONFIG, MOCK_SECURITY_PROTOCOL_VALUE);
        environmentMap.put(KAFKA_SASL_MECHANISM, MOCK_SASL_MECHANISM_VALUE);
        environmentMap.put(KAFKA_SASL_JAAS_CONFIG, MOCK_SASL_JAAS_CONFIG_VALUE);
        environmentMap.put(KAFKA_USERNAME, MOCK_USERNAME);
        environmentMap.put(KAFKA_PASSWORD, MOCK_PASSWORD);

        return environmentMap;

    }

    public static MetricsExporterContainer mockMetricsExporterContainer() {

        // Set Resource overrides
        Map<String, Quantity> resourceOverrideMap = new HashMap<>();

        resourceOverrideMap.put("memory", new Quantity(MOCK_POD_MEM));
        resourceOverrideMap.put("cpu", new Quantity(MOCK_POD_CPU));
        resourceOverrideMap.put("ephemeral-storage", new Quantity(MOCK_POD_EPHEMERAL_STORAGE));

        // Construct resource request
        final var mockResourceRequest = new ResourceRequirements();

        mockResourceRequest.setRequests(resourceOverrideMap);
        mockResourceRequest.setLimits(resourceOverrideMap);

        return new MetricsExporterContainer(
            EXPORTER_CONTAINER_NAME,
            "containersol/locust_exporter:v0.5.0",
            "Always",
            9646,
            mockResourceRequest

        );
    }

    public static void setupSysconfigMock(SysConfig mockedConfInstance) {

        // Kafka
        when(mockedConfInstance.getKafkaBootstrapServers())
            .thenReturn(MOCK_KAFKA_BOOTSTRAP_VALUE);
        when(mockedConfInstance.isKafkaSecurityEnabled())
            .thenReturn(MOCK_SECURITY_VALUE);
        when(mockedConfInstance.getKafkaSecurityProtocol())
            .thenReturn(MOCK_SECURITY_PROTOCOL_VALUE);
        when(mockedConfInstance.getKafkaUsername())
            .thenReturn(MOCK_USERNAME);
        when(mockedConfInstance.getKafkaUserPassword())
            .thenReturn(MOCK_PASSWORD);
        when(mockedConfInstance.getKafkaSaslMechanism())
            .thenReturn(MOCK_SASL_MECHANISM_VALUE);
        when(mockedConfInstance.getKafkaSaslJaasConfig())
            .thenReturn(MOCK_SASL_JAAS_CONFIG_VALUE);

        // Resource request :: Load generation node
        when(mockedConfInstance.getPodMemRequest())
            .thenReturn(MOCK_POD_MEM);
        when(mockedConfInstance.getPodCpuRequest())
            .thenReturn(MOCK_POD_CPU);
        when(mockedConfInstance.getPodEphemeralStorageRequest())
            .thenReturn(MOCK_POD_EPHEMERAL_STORAGE);

        // Resource request :: Metrics exporter
        when(mockedConfInstance.getMetricsExporterMemRequest())
            .thenReturn(MOCK_POD_MEM);
        when(mockedConfInstance.getMetricsExporterCpuRequest())
            .thenReturn(MOCK_POD_CPU);
        when(mockedConfInstance.getMetricsExporterEphemeralStorageRequest())
            .thenReturn(MOCK_POD_EPHEMERAL_STORAGE);

        // Port binding :: Metrics exporter
        when(mockedConfInstance.getMetricsExporterPort())
            .thenReturn(MOCK_POD_PORT);

        // Image :: Metrics exporter
        when(mockedConfInstance.getMetricsExporterImage())
            .thenReturn(DEFAULT_METRICS_IMAGE);

        // Job characteristics
        when(mockedConfInstance.getTtlSecondsAfterFinished())
            .thenReturn(MOCK_TTL_SECONDS_AFTER_FINISHED);

        // Resource limit :: Load generation node
        when(mockedConfInstance.getPodMemLimit())
            .thenReturn(MOCK_POD_MEM);
        when(mockedConfInstance.getPodCpuLimit())
            .thenReturn(MOCK_POD_CPU);
        when(mockedConfInstance.getPodEphemeralStorageLimit())
            .thenReturn(MOCK_POD_EPHEMERAL_STORAGE);

        // Resource limit :: Metrics exporter
        when(mockedConfInstance.getMetricsExporterMemLimit())
            .thenReturn(MOCK_POD_MEM);
        when(mockedConfInstance.getMetricsExporterCpuLimit())
            .thenReturn(MOCK_POD_CPU);
        when(mockedConfInstance.getMetricsExporterEphemeralStorageLimit())
            .thenReturn(MOCK_POD_EPHEMERAL_STORAGE);

        // Affinity
        when(mockedConfInstance.isAffinityCrInjectionEnabled())
            .thenReturn(MOCK_AFFINITY_INJECTION_VALUE);

        // Taints Toleration
        when(mockedConfInstance.isTolerationsCrInjectionEnabled())
            .thenReturn(MOCK_TOLERATION_INJECTION_VALUE);
    }

}
