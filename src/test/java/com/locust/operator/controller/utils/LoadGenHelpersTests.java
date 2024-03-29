package com.locust.operator.controller.utils;

import com.locust.operator.controller.config.SysConfig;
import lombok.val;
import org.junit.jupiter.api.BeforeAll;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.TestInstance;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.Mock;
import org.mockito.MockitoAnnotations;
import org.mockito.junit.jupiter.MockitoExtension;

import static com.locust.operator.controller.TestFixtures.prepareLocustTest;
import static com.locust.operator.controller.dto.OperationalMode.MASTER;
import static com.locust.operator.controller.dto.OperationalMode.WORKER;
import static com.locust.operator.controller.utils.TestFixtures.assertNodeConfig;
import static com.locust.operator.controller.utils.TestFixtures.setupSysconfigMock;

@ExtendWith(MockitoExtension.class)
@TestInstance(TestInstance.Lifecycle.PER_CLASS)
public class LoadGenHelpersTests {

    @Mock
    private SysConfig config;
    private LoadGenHelpers loadGenHelpers;

    @BeforeAll
    void setupMock() {
        MockitoAnnotations.openMocks(this);
        loadGenHelpers = new LoadGenHelpers(config);
        setupSysconfigMock(config);
    }

    @Test
    @DisplayName("Functional: Master node configuration")
    void masterLoadConfigGeneration() {

        // * Setup
        // QE -> Quality Engineering team
        val resourceName = "qe.performanceTest";
        val operationalMode = MASTER;
        val locustTest = prepareLocustTest(resourceName);

        // * Act
        val generatedNodeConfig = loadGenHelpers.generateLoadGenNodeObject(locustTest, operationalMode);

        // * Assert
        assertNodeConfig(locustTest, generatedNodeConfig, operationalMode);

    }

    @Test
    @DisplayName("Functional: Worker node configuration")
    void workerLoadConfigGeneration() {

        // * Setup
        val resourceName = "eq.test";
        val operationalMode = WORKER;
        val locustTest = prepareLocustTest(resourceName);

        // * Act
         val generatedNodeConfig = loadGenHelpers.generateLoadGenNodeObject(locustTest, operationalMode);

        // * Assert
        assertNodeConfig(locustTest, generatedNodeConfig, operationalMode);

    }

}
