package com.locust.operator.controller;

import com.locust.operator.controller.config.SysConfig;
import com.locust.operator.controller.utils.LoadGenHelpers;
import com.locust.operator.controller.utils.resource.manage.ResourceCreationHelpers;
import com.locust.operator.controller.utils.resource.manage.ResourceCreationManager;
import com.locust.operator.controller.utils.resource.manage.ResourceDeletionManager;
import io.fabric8.kubeapitest.junit.EnableKubeAPIServer;
import io.fabric8.kubeapitest.junit.KubeConfig;
import io.fabric8.kubernetes.client.KubernetesClient;
import lombok.extern.slf4j.Slf4j;
import lombok.val;
import org.junit.jupiter.api.AfterEach;
import org.junit.jupiter.api.BeforeAll;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.TestInstance;
import org.junit.runner.RunWith;
import org.mockito.Mock;
import org.mockito.MockitoAnnotations;
import org.mockito.junit.MockitoJUnitRunner;

import static com.locust.operator.controller.TestFixtures.DEFAULT_NAMESPACE;
import static com.locust.operator.controller.TestFixtures.creatKubernetesClient;
import static com.locust.operator.controller.TestFixtures.deleteLocustTestCrd;
import static com.locust.operator.controller.TestFixtures.prepareLocustTest;
import static com.locust.operator.controller.TestFixtures.setupCustomResourceDefinition;
import static com.locust.operator.controller.utils.TestFixtures.setupSysconfigMock;
import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.SoftAssertions.assertSoftly;

@Slf4j
@RunWith(MockitoJUnitRunner.class)
@TestInstance(TestInstance.Lifecycle.PER_CLASS)
@EnableKubeAPIServer(updateKubeConfigFile = true)
class LocustTestReconcilerTests {

    @Mock
    private SysConfig sysConfig;

    @KubeConfig
    static String configYaml;

    private LocustTestReconciler locustTestReconciler;
    private KubernetesClient k8sTestClient;

    @BeforeAll
    void setupMethodMock() {

        // Mock configuration
        MockitoAnnotations.openMocks(this);
        var loadGenHelpers = new LoadGenHelpers(sysConfig);
        var creationHelper = new ResourceCreationHelpers(loadGenHelpers);
        var creationManager = new ResourceCreationManager(creationHelper);
        var deletionManager = new ResourceDeletionManager(loadGenHelpers);
        locustTestReconciler = new LocustTestReconciler(loadGenHelpers, creationManager, deletionManager);

        // Setup SysConfig mock
        setupSysconfigMock(sysConfig);

        // Setup and deploy the CRD
        k8sTestClient = creatKubernetesClient(configYaml);

    }

    @BeforeEach
    void setup() {
        // Setup and deploy the CRD
        setupCustomResourceDefinition(k8sTestClient);
    }

    @AfterEach
    void tearDown() throws InterruptedException {
        // Clean resources from cluster
        deleteLocustTestCrd(k8sTestClient);
        // Dirty loop until the CRD is deleted to avoid test failures due to the CRD not being deleted
        while (!k8sTestClient.apiextensions().v1().customResourceDefinitions().list().getItems().isEmpty()) {
            Thread.sleep(50);
        }
    }

    @Test
    @DisplayName("Functional: Reconcile - onAdd event")
    void reconcileOnAddEvent() {
        // * Setup
        val expectedJobCount = 2;
        val expectedServiceCount = 2;
        val resourceName = "team.perftest";
        val expectedMasterResourceName = "team-perftest-master"; // Based on the conversion logic
        val expectedWorkerResourceName = "team-perftest-worker"; // Based on the conversion logic
        val locustTest = prepareLocustTest(resourceName);

        // * Act
        // Passing "null" to context is safe as it is not used in the "reconcile()" method
        locustTestReconciler.reconcile(locustTest, null);

        // Get All Jobs created
        val jobList = k8sTestClient.batch().v1().jobs().inNamespace(DEFAULT_NAMESPACE).list();
        log.debug("Acquired Job list: {}", jobList);

        // Get All Services created
        val serviceList = k8sTestClient.services().inNamespace(DEFAULT_NAMESPACE).list();
        log.debug("Acquired Service list: {}", serviceList);

        // * Assert
        assertSoftly(softly -> {
            // Assert master/worker jobs have been created
            softly.assertThat(jobList.getItems().size()).isEqualTo(expectedJobCount);
            softly.assertThat(jobList.getItems().get(0).getMetadata().getName()).isEqualTo(expectedMasterResourceName);
            softly.assertThat(jobList.getItems().get(1).getMetadata().getName()).isEqualTo(expectedWorkerResourceName);

            // Assert master service have been created
            softly.assertThat(serviceList.getItems().size()).isEqualTo(expectedServiceCount);
            // checking for the second item as the first service is the default kubernetes service
            softly.assertThat(serviceList.getItems().get(1).getMetadata().getName()).isEqualTo(expectedMasterResourceName);
        });

    }

    @Test
    @DisplayName("Functional: Reconcile - NOOP to onUpdate event")
    void reconcileOnUpdateEvent() {
        // * Setup
        val resourceName = "team.perftest";
        val resourceGeneration = 1L;
        val workerReplicaCount = 100;
        val locustTest = prepareLocustTest(resourceName, workerReplicaCount, resourceGeneration);

        // Deploy CR
        // Passing "null" to context is safe as it is not used in the "reconcile()" method
        locustTestReconciler.reconcile(locustTest, null);

        // * Act
        // Increase worker count
        val updatedResourceGeneration = 2L;
        val updatedWorkerReplicaCount = 37;
        val updatedLocustTest = prepareLocustTest(resourceName, updatedWorkerReplicaCount, updatedResourceGeneration);

        // Update deployed CR
        // Passing "null" to context is safe as it is not used in the "reconcile()" method
        locustTestReconciler.reconcile(updatedLocustTest, null);

        // Get All Jobs created
        val jobList = k8sTestClient.batch().v1().jobs().inNamespace(DEFAULT_NAMESPACE).list();
        log.debug("Acquired Job list: {}", jobList);

        // * Assert
        // Assert NOOP on update
        assertThat(jobList.getItems().get(1).getSpec().getParallelism()).isEqualTo(workerReplicaCount);

    }

    @Test
    @DisplayName("Functional: Reconcile - cleanup onDelete event")
    void cleanupOnDeleteEvent() {
        // * Setup
        val expectedJobCount = 0;
        val expectedServiceCount = 1; // 1 Because of the default kubernetes service remaining post deletion
        val resourceName = "team.perftest";
        val expectedDefaultServiceName = "kubernetes";
        val locustTest = prepareLocustTest(resourceName);

        // Deploy CR
        // Passing "null" to context is safe as it is not used in the "reconcile()" method
        locustTestReconciler.reconcile(locustTest, null);

        // * Act
        // Delete CR
        // Passing "null" to context is safe as it is not used in the "cleanup()" method
        locustTestReconciler.cleanup(locustTest, null);

        // Get All Jobs created
        val jobList = k8sTestClient.batch().v1().jobs().inNamespace(DEFAULT_NAMESPACE).list();
        log.debug("Acquired Job list: {}", jobList);

        // Get All Services created
        val serviceList = k8sTestClient.services().inNamespace(DEFAULT_NAMESPACE).list();
        log.debug("Acquired Service list: {}", serviceList);

        // * Assert
        assertSoftly(softly -> {
            // Assert master/worker jobs have been deleted
            softly.assertThat(jobList.getItems().size()).isEqualTo(expectedJobCount);

            // Assert master service have been deleted
            softly.assertThat(serviceList.getItems().size()).isEqualTo(expectedServiceCount);
            softly.assertThat(serviceList.getItems().get(0).getMetadata().getName()).isEqualTo(expectedDefaultServiceName);

        });

    }

}
