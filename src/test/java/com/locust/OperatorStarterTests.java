package com.locust;

import com.locust.operator.controller.LocustTestReconciler;
import io.javaoperatorsdk.jenvtest.junit.EnableKubeAPIServer;
import io.javaoperatorsdk.jenvtest.junit.KubeConfig;
import org.junit.jupiter.api.BeforeAll;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.TestInstance;
import org.mockito.Mock;
import org.mockito.MockitoAnnotations;

import static com.locust.operator.controller.TestFixtures.setupCustomResourceDefinition;
import static com.locust.operator.controller.TestFixtures.creatKubernetesClient;

@TestInstance(TestInstance.Lifecycle.PER_CLASS)
@EnableKubeAPIServer(updateKubeConfigFile = true)
public class OperatorStarterTests {

    @KubeConfig
    static String configYaml;

    @Mock
    private LocustTestReconciler reconciler;

    @BeforeAll
    void setup() {
        MockitoAnnotations.openMocks(this);
        setupCustomResourceDefinition(creatKubernetesClient(configYaml));
    }

    /**
     * Ideally this test should be replaced with a combination of
     * {@link io.micronaut.test.extensions.junit5.annotation.MicronautTest @MicronautTest} and a check that "Application.isRunning()"
     * returns true.
     * <p>
     * The only reason this test was designed this way is that I am not able to find how to inject `k8sServerUrl` variable into the test
     * environment before @MicronautTest boots up the Application. If the application trys to boot without `k8sServerUrl` being injected, it
     * will throw an Exception.
     */
    @Test
    @DisplayName("Functional: Check operator startup core")
    void operatorStarterCore() {

        // * Setup
        var operatorStarter = new LocustTestOperatorStarter(reconciler);

        // * Act & Assert
        // Passing "null" to onApplicationEvent(ServerStartupEvent event) is safe since the event is not used by the "operatorStarter" logic.
        //executeWithK8sMockServer(k8sServerUrl, () -> operatorStarter.onApplicationEvent(null));
        operatorStarter.onApplicationEvent(null);

        // * Assert
        // This test doesn't need an explicit assertion statement since the onApplicationEvent() logic
        // will throw an exception if it doesn't manage to start the Operator.

    }

}
