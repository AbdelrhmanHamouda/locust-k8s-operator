package com.locust.controller;

import com.locust.operator.controller.config.SysConfig;
import com.locust.operator.controller.utils.LoadGenHelpers;
import com.locust.operator.controller.utils.resource.manage.ResourceCreationHelpers;
import com.locust.operator.controller.utils.resource.manage.ResourceCreationManager;
import com.locust.operator.controller.utils.resource.manage.ResourceDeletionManager;
import io.fabric8.kubernetes.client.KubernetesClient;
import io.fabric8.kubernetes.client.server.mock.EnableKubernetesMockClient;
import lombok.extern.slf4j.Slf4j;
import lombok.val;
import org.junit.jupiter.api.BeforeAll;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.TestInstance;
import org.junit.runner.RunWith;
import org.mockito.Mock;
import org.mockito.MockitoAnnotations;
import org.mockito.junit.MockitoJUnitRunner;

import static com.locust.controller.utils.TestFixtures.executeWithK8sMockServer;
import static com.locust.controller.utils.TestFixtures.prepareLocustTest;
import static com.locust.controller.utils.TestFixtures.prepareNodeConfig;
import static com.locust.controller.utils.TestFixtures.setupSysconfigMock;
import static com.locust.operator.controller.dto.OperationalMode.MASTER;
import static org.assertj.core.api.AssertionsForClassTypes.assertThat;

@Slf4j
@RunWith(MockitoJUnitRunner.class)
@TestInstance(TestInstance.Lifecycle.PER_CLASS)
@EnableKubernetesMockClient(https = false, crud = true)
public class ResourceDeletionManagerTests {

    @Mock
    private SysConfig sysConfig;
    private LoadGenHelpers loadGenHelpers;
    private ResourceCreationHelpers creationHelper;
    private ResourceCreationManager creationManager;
    private ResourceDeletionManager deletionManager;

    String k8sServerUrl;
    KubernetesClient testClient;

    @BeforeAll
    void setupMethodMock() {

        MockitoAnnotations.openMocks(this);
        loadGenHelpers = new LoadGenHelpers(sysConfig);
        creationHelper = new ResourceCreationHelpers(loadGenHelpers);
        creationManager = new ResourceCreationManager(creationHelper);
        deletionManager = new ResourceDeletionManager(loadGenHelpers);
        setupSysconfigMock(sysConfig);

    }

    @BeforeEach
    void setup() {
        k8sServerUrl = testClient.getMasterUrl().toString();
    }

    @Test
    @DisplayName("Functional: Delete a kubernetes Job")
    void DeleteJobTest() {

        // * Setup
        val namespace = "default";
        val nodeName = "mnt-demo-test-master";
        val resourceName = "mnt.demo-test";
        val nodeConfig = prepareNodeConfig(nodeName, MASTER);
        val locustTest = prepareLocustTest(resourceName);

        // * Act
        executeWithK8sMockServer(k8sServerUrl, () -> creationManager.createJob(nodeConfig, namespace, resourceName));
        executeWithK8sMockServer(k8sServerUrl, () -> deletionManager.deleteJob(locustTest, MASTER));

        // Get All Jobs created by the method
        val jobList = testClient.batch().v1().jobs().inNamespace(namespace).list();
        log.debug("Acquired Job list: {}", jobList);

        // * Assert
        assertThat(jobList.getItems().size()).isEqualTo(0);

    }

    @Test
    @DisplayName("Functional: Check that when Job deletion fails, an empty list is returned.")
    void DeleteJobFailureReturnEmptyListTest() {

        // * Setup
        val resourceName = "mnt.demo-test";
        val locustTest = prepareLocustTest(resourceName);

        // * Act
        val deletedJobStatus = deletionManager.deleteJob(locustTest, MASTER).orElseThrow();

        // * Assert
        assertThat(deletedJobStatus.isEmpty()).isTrue();

    }

    @Test
    @DisplayName("Functional: Delete a kubernetes Service")
    void DeleteServiceTest() {

        // * Setup
        val namespace = "default";
        val nodeName = "act-kafka-test-master";
        val resourceName = "act.kafka-test";
        val nodeConfig = prepareNodeConfig(nodeName, MASTER);
        val locustTest = prepareLocustTest(resourceName);

        // * Act
        executeWithK8sMockServer(k8sServerUrl, () -> creationManager.createMasterService(nodeConfig, namespace));
        executeWithK8sMockServer(k8sServerUrl, () -> deletionManager.deleteService(locustTest, MASTER));

        // Get All Jobs created by the method
        val serviceList = testClient.services().inNamespace(namespace).list();
        log.debug("Acquired Deployment list: {}", serviceList);

        // * Assert
        assertThat(serviceList.getItems().size()).isEqualTo(0);

    }

    @Test
    @DisplayName("Functional: Check that when Service deletion fails, empty list is returned")
    void DeleteServiceFailureReturnEmptyListTest() {

        // * Setup
        val resourceName = "mnt.demo-test";
        val locustTest = prepareLocustTest(resourceName);

        // * Act
        val deletedServiceStatus = deletionManager.deleteService(locustTest, MASTER).orElseThrow();

        // * Assert
        assertThat(deletedServiceStatus.isEmpty()).isTrue();

    }

}
