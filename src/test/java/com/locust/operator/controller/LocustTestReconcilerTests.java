package com.locust.operator.controller;

import com.locust.operator.controller.config.SysConfig;
import com.locust.operator.controller.utils.LoadGenHelpers;
import com.locust.operator.controller.utils.resource.manage.ResourceCreationHelpers;
import com.locust.operator.controller.utils.resource.manage.ResourceCreationManager;
import com.locust.operator.controller.utils.resource.manage.ResourceDeletionManager;
import io.fabric8.kubernetes.client.KubernetesClient;
import io.fabric8.kubernetes.client.server.mock.EnableKubernetesMockClient;
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
import static com.locust.operator.controller.TestFixtures.createCrd;
import static com.locust.operator.controller.TestFixtures.deleteLocustTestCrd;
import static com.locust.operator.controller.TestFixtures.prepareCustomResourceDefinition;
import static com.locust.operator.controller.TestFixtures.prepareLocustTest;
import static com.locust.operator.controller.utils.TestFixtures.executeWithK8sMockServer;
import static com.locust.operator.controller.utils.TestFixtures.setupSysconfigMock;
import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.SoftAssertions.assertSoftly;

@Slf4j
@RunWith(MockitoJUnitRunner.class)
@TestInstance(TestInstance.Lifecycle.PER_CLASS)
@EnableKubernetesMockClient(https = false, crud = true)
class LocustTestReconcilerTests {

    @Mock
    private SysConfig sysConfig;
    private LocustTestReconciler locustTestReconciler;

    String k8sServerUrl;
    KubernetesClient k8sTestClient;

    @BeforeAll
    void setupMethodMock() {

        MockitoAnnotations.openMocks(this);
        var loadGenHelpers = new LoadGenHelpers(sysConfig);
        var creationHelper = new ResourceCreationHelpers(loadGenHelpers);
        var creationManager = new ResourceCreationManager(creationHelper);
        var deletionManager = new ResourceDeletionManager(loadGenHelpers);
        locustTestReconciler = new LocustTestReconciler(loadGenHelpers, creationManager, deletionManager);
        setupSysconfigMock(sysConfig);

    }

    @BeforeEach
    void setUp() {
        k8sServerUrl = k8sTestClient.getMasterUrl().toString();
        val expectedCrd = prepareCustomResourceDefinition(k8sTestClient);
        val crd = createCrd(expectedCrd, k8sTestClient);
        log.debug("Created CRD details: {}", crd);
    }

    @AfterEach
    void tearDown() {
        // Clean resources from cluster
        deleteLocustTestCrd(k8sTestClient);
    }

    @Test
    @DisplayName("Functional: Reconcile - onAdd event")
    void reconcileOnAddEvent() {
        // * Setup
        val expectedJobCount = 2;
        val expectedServiceCount = 1;
        val resourceName = "team.perftest";
        val expectedMasterResourceName = "team-perftest-master"; // Based on the conversion logic
        val expectedWorkerResourceName = "team-perftest-worker"; // Based on the conversion logic
        val locustTest = prepareLocustTest(resourceName);

        // * Act
        // Passing "null" to context is safe as it is not used in the "reconcile()" method
        executeWithK8sMockServer(k8sServerUrl, () -> locustTestReconciler.reconcile(locustTest, null));

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
            softly.assertThat(serviceList.getItems().get(0).getMetadata().getName()).isEqualTo(expectedMasterResourceName);
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
        executeWithK8sMockServer(k8sServerUrl, () -> locustTestReconciler.reconcile(locustTest, null));

        // * Act
        // Increase worker count
        val updatedResourceGeneration = 2L;
        val updatedWorkerReplicaCount = 37;
        val updatedLocustTest = prepareLocustTest(resourceName, updatedWorkerReplicaCount, updatedResourceGeneration);

        // Update deployed CR
        // Passing "null" to context is safe as it is not used in the "reconcile()" method
        executeWithK8sMockServer(k8sServerUrl, () -> locustTestReconciler.reconcile(updatedLocustTest, null));

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
        val expectedResourceCount = 0;
        val resourceName = "team.perftest";
        val locustTest = prepareLocustTest(resourceName);

        // Deploy CR
        // Passing "null" to context is safe as it is not used in the "reconcile()" method
        executeWithK8sMockServer(k8sServerUrl, () -> locustTestReconciler.reconcile(locustTest, null));

        // * Act
        // Delete CR
        // Passing "null" to context is safe as it is not used in the "cleanup()" method
        executeWithK8sMockServer(k8sServerUrl, () -> locustTestReconciler.cleanup(locustTest, null));

        // Get All Jobs created
        val jobList = k8sTestClient.batch().v1().jobs().inNamespace(DEFAULT_NAMESPACE).list();
        log.debug("Acquired Job list: {}", jobList);

        // Get All Services created
        val serviceList = k8sTestClient.services().inNamespace(DEFAULT_NAMESPACE).list();
        log.debug("Acquired Service list: {}", serviceList);

        // * Assert
        assertSoftly(softly -> {
            // Assert master/worker jobs have been deleted
            softly.assertThat(jobList.getItems().size()).isEqualTo(expectedResourceCount);

            // Assert master service have been deleted
            softly.assertThat(serviceList.getItems().size()).isEqualTo(expectedResourceCount);

        });

    }

}
