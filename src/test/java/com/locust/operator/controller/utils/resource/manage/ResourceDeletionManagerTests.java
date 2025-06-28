package com.locust.operator.controller.utils.resource.manage;

import com.locust.operator.controller.config.SysConfig;
import com.locust.operator.controller.utils.LoadGenHelpers;
import io.fabric8.kubernetes.client.KubernetesClient;
import io.fabric8.kubernetes.client.server.mock.EnableKubernetesMockClient;
import lombok.extern.slf4j.Slf4j;
import lombok.val;
import org.junit.jupiter.api.BeforeAll;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.TestInstance;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.Mock;
import org.mockito.MockitoAnnotations;
import org.mockito.junit.jupiter.MockitoExtension;

import java.util.Collections;

import static com.locust.operator.controller.TestFixtures.prepareLocustTest;
import static com.locust.operator.controller.dto.OperationalMode.MASTER;
import static com.locust.operator.controller.utils.TestFixtures.executeWithK8sMockServer;
import static com.locust.operator.controller.utils.TestFixtures.prepareNodeConfig;
import static com.locust.operator.controller.utils.TestFixtures.setupSysconfigMock;
import static org.assertj.core.api.AssertionsForClassTypes.assertThat;

@Slf4j
@ExtendWith(MockitoExtension.class)
@TestInstance(TestInstance.Lifecycle.PER_CLASS)
@EnableKubernetesMockClient(https = false, crud = true)
public class ResourceDeletionManagerTests {

    @Mock
    private SysConfig sysConfig;
    private ResourceCreationManager creationManager;
    private ResourceDeletionManager deletionManager;

    String k8sServerUrl;
    KubernetesClient testClient;

    @BeforeAll
    void setupMethodMock() {

        MockitoAnnotations.openMocks(this);
        var loadGenHelpers = new LoadGenHelpers(sysConfig);
        var creationHelper = new ResourceCreationHelpers(loadGenHelpers);
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
    void deleteJobTest() {

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
    void deleteJobFailureReturnEmptyListTest() {

        // * Setup
        val resourceName = "mnt.demo-test";
        val locustTest = prepareLocustTest(resourceName);

        // * Act
        val deletedJobStatus = deletionManager.deleteJob(locustTest, MASTER).orElse(Collections.emptyList());

        // * Assert
        assertThat(deletedJobStatus.isEmpty()).isTrue();

    }

    @Test
    @DisplayName("Functional: Delete a kubernetes Service")
    void deleteServiceTest() {

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
    void deleteServiceFailureReturnEmptyListTest() {

        // * Setup
        val resourceName = "mnt.demo-test";
        val locustTest = prepareLocustTest(resourceName);

        // * Act
        val deletedServiceStatus = deletionManager.deleteService(locustTest, MASTER).orElse(Collections.emptyList());

        // * Assert
        assertThat(deletedServiceStatus.isEmpty()).isTrue();

    }

}
