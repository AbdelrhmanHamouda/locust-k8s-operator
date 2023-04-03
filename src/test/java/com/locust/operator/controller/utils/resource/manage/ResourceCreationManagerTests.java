package com.locust.operator.controller.utils.resource.manage;

import com.locust.operator.controller.utils.LoadGenHelpers;
import com.locust.operator.customresource.internaldto.LocustTestToleration;
import io.fabric8.kubernetes.client.KubernetesClient;
import io.fabric8.kubernetes.client.server.mock.EnableKubernetesMockClient;
import lombok.extern.slf4j.Slf4j;
import lombok.val;
import org.junit.jupiter.api.BeforeAll;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.TestInstance;
import org.mockito.Mock;
import org.mockito.MockitoAnnotations;

import static com.locust.operator.controller.dto.OperationalMode.MASTER;
import static com.locust.operator.controller.utils.TestFixtures.assertK8sNodeAffinity;
import static com.locust.operator.controller.utils.TestFixtures.assertK8sResourceCreation;
import static com.locust.operator.controller.utils.TestFixtures.assertK8sTolerations;
import static com.locust.operator.controller.utils.TestFixtures.assertK8sTtlSecondsAfterFinished;
import static com.locust.operator.controller.utils.TestFixtures.containerEnvironmentMap;
import static com.locust.operator.controller.utils.TestFixtures.executeWithK8sMockServer;
import static com.locust.operator.controller.utils.TestFixtures.prepareNodeConfig;
import static com.locust.operator.controller.utils.TestFixtures.prepareNodeConfigWithNodeAffinity;
import static com.locust.operator.controller.utils.TestFixtures.prepareNodeConfigWithTolerations;
import static com.locust.operator.controller.utils.TestFixtures.prepareNodeConfigWithTtlSecondsAfterFinished;
import static org.mockito.Mockito.when;

@Slf4j
@TestInstance(TestInstance.Lifecycle.PER_CLASS)
@EnableKubernetesMockClient(https = false, crud = true)
public class ResourceCreationManagerTests {

    @Mock
    private LoadGenHelpers loadGenHelpers;
    private ResourceCreationManager CreationManager;

    String k8sServerUrl;
    KubernetesClient testClient;

    @BeforeAll
    void setupMethodMock() {

        MockitoAnnotations.openMocks(this);
        var creationHelper = new ResourceCreationHelpers(loadGenHelpers);
        CreationManager = new ResourceCreationManager(creationHelper);
        when(loadGenHelpers.generateContainerEnvironmentMap())
            .thenReturn(containerEnvironmentMap());

    }

    @BeforeEach
    void setup() {
        k8sServerUrl = testClient.getMasterUrl().toString();
    }

    @Test
    @DisplayName("Functional: Create a kubernetes Job")
    void createJobTest() {

        // * Setup
        val namespace = "default";
        val nodeName = "mnt-demo-test";
        val resourceName = "mnt.demo-test";
        val nodeConfig = prepareNodeConfig(nodeName, MASTER);

        // * Act
        executeWithK8sMockServer(k8sServerUrl, () -> CreationManager.createJob(nodeConfig, namespace, resourceName));

        // Get All Jobs created by the method
        val jobList = testClient.batch().v1().jobs().inNamespace(namespace).list();
        log.debug("Acquired Job list: {}", jobList);

        // * Assert
        assertK8sResourceCreation(nodeName, jobList);

    }

    @Test
    @DisplayName("Functional: Create a kubernetes Service")
    void createServiceTest() {

        // * Setup
        val namespace = "default";
        val nodeName = "act-kafka-test";
        val nodeConfig = prepareNodeConfig(nodeName, MASTER);

        // * Act
        executeWithK8sMockServer(k8sServerUrl, () -> CreationManager.createMasterService(nodeConfig, namespace));

        // Get All Services created by the method
        val serviceList = testClient.services().inNamespace(namespace).list();
        log.debug("Acquired Service list: {}", serviceList);

        // * Assert
        assertK8sResourceCreation(nodeName, serviceList);

    }

    @Test
    @DisplayName("Functional: Create a kubernetes Job with Default TTL Seconds After Finished")
    void createJobWithDefaultTtlSecondsAfterFinishedTest() {

        // * Setup
        val namespace = "ttl-ns";
        val nodeName = "ttl-demo-test";
        val resourceName = "ttl.demo-test";
        final Integer defaultTtlSecondsAfterFinished = null;
        val nodeConfig = prepareNodeConfigWithTtlSecondsAfterFinished(nodeName, MASTER, defaultTtlSecondsAfterFinished);

        // * Act
        executeWithK8sMockServer(k8sServerUrl, () -> CreationManager.createJob(nodeConfig, namespace, resourceName));

        // Get All Jobs created by the method
        val jobList = testClient.batch().v1().jobs().inNamespace(namespace).list();
        log.debug("Acquired Job list: {}", jobList);

        // * Assert
        assertK8sTtlSecondsAfterFinished(jobList, defaultTtlSecondsAfterFinished);

    }

    @Test
    @DisplayName("Functional: Create a kubernetes Job with TTL Seconds After Finished")
    void createJobWithTtlSecondsAfterFinishedTest() {

        // * Setup
        val namespace = "ttl-ns";
        val nodeName = "ttl-demo-test";
        val resourceName = "ttl.demo-test";
        val ttlSecondsAfterFinished = Integer.valueOf(120);
        val nodeConfig = prepareNodeConfigWithTtlSecondsAfterFinished(nodeName, MASTER, ttlSecondsAfterFinished);

        // * Act
        executeWithK8sMockServer(k8sServerUrl, () -> CreationManager.createJob(nodeConfig, namespace, resourceName));

        // Get All Jobs created by the method
        val jobList = testClient.batch().v1().jobs().inNamespace(namespace).list();
        log.debug("Acquired Job list: {}", jobList);

        // * Assert
        assertK8sTtlSecondsAfterFinished(jobList, ttlSecondsAfterFinished);

    }

    @Test
    @DisplayName("Functional: Create a kubernetes Job with Node Affinity")
    void createJobWithNodeAffinityTest() {

        // * Setup
        val namespace = "node-affinity";
        val nodeName = "locust-demo-test";
        val resourceName = "locust.demo-test";
        val k8sNodeLabelKey = "organisation.com/nodeLabel";
        val k8sNodeLabelValue = "performance-nodes";
        val nodeConfig = prepareNodeConfigWithNodeAffinity(nodeName, MASTER, k8sNodeLabelKey, k8sNodeLabelValue);

        // * Act
        executeWithK8sMockServer(k8sServerUrl, () -> CreationManager.createJob(nodeConfig, namespace, resourceName));

        // Get All Jobs created by the method
        val jobList = testClient.batch().v1().jobs().inNamespace(namespace).list();
        log.debug("Acquired Job list: {}", jobList);

        // * Assert
        assertK8sResourceCreation(nodeName, jobList);
        assertK8sNodeAffinity(nodeConfig, jobList, k8sNodeLabelKey);

    }

    @Test
    @DisplayName("Functional: Create a kubernetes Job with Tolerations and Toleration Operator set to Equal")
    void createJobWithTolerationsAndOperatorEqualTest() {
        // * Setup
        val namespace = "node-affinity";
        val nodeName = "locust-demo-test";
        val resourceName = "locust.demo-test";

        // Toleration
        val tolerationKey = "taintA";
        val tolerationEffect = "NoSchedule";
        val tolerationEqualOperator = "Equal";
        val tolerationValue = "dedicatedToPerformance";

        val toleration = new LocustTestToleration(tolerationKey, tolerationEqualOperator, tolerationValue, tolerationEffect);
        val nodeConfig = prepareNodeConfigWithTolerations(nodeName, MASTER, toleration);

        // * Act
        executeWithK8sMockServer(k8sServerUrl, () -> CreationManager.createJob(nodeConfig, namespace, resourceName));

        // Get All Jobs created by the method
        val jobList = testClient.batch().v1().jobs().inNamespace(namespace).list();
        log.debug("Acquired Job list: {}", jobList);

        // * Assert
        assertK8sResourceCreation(nodeName, jobList);
        assertK8sTolerations(jobList, toleration);

    }

    @Test
    @DisplayName("Functional: Create a kubernetes Job with Tolerations and Toleration Operator set to Exists")
    void createJobWithTolerationsAndOperatorExistsTest() {
        // * Setup
        val namespace = "node-affinity";
        val nodeName = "locust-demo-test";
        val resourceName = "locust.demo-test";

        // Toleration
        val tolerationKey = "taintA";
        val tolerationEffect = "NoSchedule";
        val tolerationEqualOperator = "Exists";

        val toleration = new LocustTestToleration(tolerationKey, tolerationEqualOperator, null, tolerationEffect);
        val nodeConfig = prepareNodeConfigWithTolerations(nodeName, MASTER, toleration);

        // * Act
        executeWithK8sMockServer(k8sServerUrl, () -> CreationManager.createJob(nodeConfig, namespace, resourceName));

        // Get All Jobs created by the method
        val jobList = testClient.batch().v1().jobs().inNamespace(namespace).list();
        log.debug("Acquired Job list: {}", jobList);

        // * Assert
        assertK8sResourceCreation(nodeName, jobList);
        assertK8sTolerations(jobList, toleration);

    }
}
