package com.locust.controller.utils;

import com.locust.operator.controller.config.SysConfig;
import com.locust.operator.controller.dto.LoadGenerationNode;
import com.locust.operator.controller.dto.OperationalMode;
import com.locust.operator.customresource.LocustTest;
import com.locust.operator.customresource.LocustTestSpec;
import io.fabric8.kubernetes.api.model.KubernetesResourceList;
import io.fabric8.kubernetes.api.model.ObjectMetaBuilder;
import lombok.NoArgsConstructor;
import lombok.SneakyThrows;
import lombok.extern.slf4j.Slf4j;

import java.util.HashMap;
import java.util.List;
import java.util.Map;

import static com.github.stefanbirkner.systemlambda.SystemLambda.withEnvironmentVariable;
import static com.locust.operator.controller.dto.OperationalMode.MASTER;
import static com.locust.operator.controller.utils.Constants.CONTAINER_ARGS_SEPARATOR;
import static com.locust.operator.controller.utils.Constants.KAFKA_BOOTSTRAP_SERVERS;
import static com.locust.operator.controller.utils.Constants.KAFKA_PASSWORD;
import static com.locust.operator.controller.utils.Constants.KAFKA_SASL_JAAS_CONFIG;
import static com.locust.operator.controller.utils.Constants.KAFKA_SASL_MECHANISM;
import static com.locust.operator.controller.utils.Constants.KAFKA_SECURITY_ENABLED;
import static com.locust.operator.controller.utils.Constants.KAFKA_SECURITY_PROTOCOL_CONFIG;
import static com.locust.operator.controller.utils.Constants.KAFKA_USERNAME;
import static com.locust.operator.customresource.LocustTest.GROUP;
import static com.locust.operator.customresource.LocustTest.VERSION;
import static lombok.AccessLevel.PRIVATE;
import static org.assertj.core.api.SoftAssertions.assertSoftly;
import static org.mockito.Mockito.when;

@Slf4j
@NoArgsConstructor(access = PRIVATE)
public class TestFixtures {

    public static final List<Integer> DEFAULT_MASTER_PORT_LIST = List.of(5557, 5558, 8089);
    public static final List<Integer> DEFAULT_WORKER_PORT_LIST = List.of(8080);
    public static final Integer MASTER_REPLICA_COUNT = 1;
    public static final String DEFAULT_API_VERSION = GROUP + "/" + VERSION;
    public static final String DEFAULT_NAMESPACE = "default";
    public static final String KIND = "LocustTest";
    public static final String DEFAULT_SEED_COMMAND = "--locustfile src/GQ/src/demo.py";
    public static final int REPLICAS = 50;
    public static final String DEFAULT_TEST_IMAGE = "xlocust:latest";
    public static final int EXPECTED_JOB_COUNT = 1;
    public static final String K8S_SERVER_URL_ENV_VAR = "KUBERNETES_MASTER";
    public static final String MOCK_KAFKA_BOOTSTRAP_VALUE = "localhost:9092";
    public static final boolean MOCK_SECURITY_VALUE = true;
    public static final String MOCK_SECURITY_PROTOCOL_VALUE = "SASL_PLAINTEXT";
    public static final String MOCK_SASL_MECHANISM_VALUE = "SCRAM-SHA-512";
    public static final String MOCK_SASL_JAAS_CONFIG_VALUE = "placeholder";
    public static final String MOCK_USERNAME = "localKafkaUser";
    public static final String MOCK_PASSWORD = "localKafkaPassword";

    public static LocustTest prepareLocustTest(String resourceName) {

        var locustTest = new LocustTest();

        // API version
        locustTest.setApiVersion(DEFAULT_API_VERSION);

        // Kind
        locustTest.setKind(KIND);

        // Metadata
        locustTest.setMetadata(new ObjectMetaBuilder()
            .withName(resourceName)
            .withNamespace(DEFAULT_NAMESPACE)
            .build());

        // Spec
        var spec = new LocustTestSpec();
        spec.setMasterCommandSeed(DEFAULT_SEED_COMMAND);
        spec.setWorkerCommandSeed(DEFAULT_SEED_COMMAND);
        spec.setWorkerReplicas(REPLICAS);
        spec.setImage(DEFAULT_TEST_IMAGE);

        locustTest.setSpec(spec);
        log.debug("Created resource object:\n{}", locustTest);

        return locustTest;

    }

    public static void assertNodeConfig(LocustTest customResource, LoadGenerationNode generatedNodeConfig,
        OperationalMode mode) {

        String expectedConfigName = customResource.getMetadata().getName().replace('.', '-');

        Integer expectedReplicas = mode.equals(MASTER) ? MASTER_REPLICA_COUNT : customResource.getSpec().getWorkerReplicas();

        List<Integer> expectedPortList = mode.equals(MASTER) ? DEFAULT_MASTER_PORT_LIST : DEFAULT_WORKER_PORT_LIST;

        assertSoftly(softly -> {
            softly.assertThat(generatedNodeConfig.getName()).contains(expectedConfigName);
            softly.assertThat(generatedNodeConfig.getOperationalMode()).isEqualTo(mode);
            softly.assertThat(generatedNodeConfig.getPorts()).isEqualTo(expectedPortList);
            softly.assertThat(generatedNodeConfig.getReplicas()).isEqualTo(expectedReplicas);
        });
    }

    public static LoadGenerationNode prepareNodeConfig(String nodeName, OperationalMode mode) {
        var nodeConfig = LoadGenerationNode.builder()
            .name(nodeName)
            .command(List.of(DEFAULT_SEED_COMMAND.split(CONTAINER_ARGS_SEPARATOR)))
            .operationalMode(mode)
            .image(DEFAULT_TEST_IMAGE)
            .replicas(mode.equals(MASTER) ? MASTER_REPLICA_COUNT : REPLICAS)
            .ports(mode.equals(MASTER) ? DEFAULT_MASTER_PORT_LIST : DEFAULT_WORKER_PORT_LIST)
            .build();

        log.debug("Created node configuration: {}", nodeConfig);
        return nodeConfig;
    }

    public static <T extends KubernetesResourceList<?>> void assertK8sResourceCreation(String nodeName, T resourceList) {

        assertSoftly(softly -> {
            softly.assertThat(resourceList.getItems().size()).isEqualTo(EXPECTED_JOB_COUNT);
            softly.assertThat(resourceList.getItems().get(0).getMetadata().getName()).isEqualTo(nodeName);
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

    public static void setupSysconfigMock(SysConfig mockedConfInstance) {

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

    }

}
