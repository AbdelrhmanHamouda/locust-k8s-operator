package com.locust;

import com.locust.operator.controller.LocustTestReconciler;
import io.fabric8.kubernetes.client.KubernetesClient;
import io.fabric8.kubernetes.client.server.mock.EnableKubernetesMockClient;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.TestInstance;
import org.mockito.Mock;
import org.mockito.MockitoAnnotations;

import static com.locust.operator.controller.TestFixtures.executeWithK8sMockServer;

@TestInstance(TestInstance.Lifecycle.PER_CLASS)
@EnableKubernetesMockClient(https = false, crud = true)
public class OperatorStarterTests {

    String k8sServerUrl;
    KubernetesClient testClient;

    @Mock
    private LocustTestReconciler reconciler;

    @BeforeEach
    void setup() {
        MockitoAnnotations.openMocks(this);
        k8sServerUrl = testClient.getMasterUrl().toString();
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
        executeWithK8sMockServer(k8sServerUrl, () -> operatorStarter.onApplicationEvent(null));

        // * Assert
        // This test doesn't need an explicit assertion statement since the onApplicationEvent() logic
        // will throw an exception if it doesn't manage to start the Operator.

    }

}
