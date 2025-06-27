package com.locust.operator.controller.utils;

import com.locust.operator.controller.config.SysConfig;
import lombok.val;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.TestInstance;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.junit.jupiter.MockitoExtension;
import org.mockito.junit.jupiter.MockitoSettings;
import org.mockito.quality.Strictness;

import static com.locust.operator.controller.TestFixtures.prepareLocustTest;
import static com.locust.operator.controller.dto.OperationalMode.MASTER;
import static com.locust.operator.controller.dto.OperationalMode.WORKER;
import static com.locust.operator.controller.utils.Constants.DEFAULT_RESOURCE_TARGET;
import static com.locust.operator.controller.utils.TestFixtures.assertNodeConfig;
import static com.locust.operator.controller.utils.TestFixtures.setupSysconfigMock;
import static org.junit.jupiter.api.Assertions.assertFalse;
import static org.mockito.Mockito.mock;
import static org.mockito.Mockito.when;

@ExtendWith(MockitoExtension.class)
@TestInstance(TestInstance.Lifecycle.PER_CLASS)
@MockitoSettings(strictness = Strictness.LENIENT)
public class LoadGenHelpersTests {

    @Test
    @DisplayName("Functional: Master node configuration")
    void masterLoadConfigGeneration() {

        // * Setup
        final SysConfig config = mock(SysConfig.class);
        setupSysconfigMock(config);
        final LoadGenHelpers loadGenHelpers = new LoadGenHelpers(config);
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
        final SysConfig config = mock(SysConfig.class);
        setupSysconfigMock(config);
        final LoadGenHelpers loadGenHelpers = new LoadGenHelpers(config);
        val resourceName = "eq.test";
        val operationalMode = WORKER;
        val locustTest = prepareLocustTest(resourceName);

        // * Act
        val generatedNodeConfig = loadGenHelpers.generateLoadGenNodeObject(locustTest, operationalMode);

        // * Assert
        assertNodeConfig(locustTest, generatedNodeConfig, operationalMode);

    }

    @Test
    @DisplayName("Functional: Unbound CPU limit configuration")
    void unboundCpuLimitConfiguration() {

        // * Setup
        final SysConfig config = mock(SysConfig.class);
        setupSysconfigMock(config);
        when(config.getPodCpuLimit()).thenReturn("");
        final LoadGenHelpers loadGenHelpers = new LoadGenHelpers(config);

        // * Act
        val resourceRequirements = loadGenHelpers.getResourceRequirements(DEFAULT_RESOURCE_TARGET);

        // * Assert
        assertFalse(resourceRequirements.getLimits().containsKey("cpu"), "CPU limit should not be set when the config value is blank");

    }

}
