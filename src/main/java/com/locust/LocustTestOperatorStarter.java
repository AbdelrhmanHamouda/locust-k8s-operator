package com.locust;

import com.locust.operator.controller.LocustTestReconciler;
import io.fabric8.kubernetes.client.KubernetesClient;
import io.fabric8.kubernetes.client.KubernetesClientBuilder;
import io.javaoperatorsdk.operator.Operator;
import io.micronaut.context.event.ApplicationEventListener;
import io.micronaut.runtime.server.event.ServerStartupEvent;
import jakarta.inject.Singleton;
import lombok.extern.slf4j.Slf4j;

@Slf4j
@Singleton
public class LocustTestOperatorStarter implements ApplicationEventListener<ServerStartupEvent> {

    private final LocustTestReconciler reconciler;

    public LocustTestOperatorStarter(LocustTestReconciler reconciler) {
        this.reconciler = reconciler;
    }

    @Override
    public void onApplicationEvent(ServerStartupEvent event) {
        log.info("Starting Kubernetes reconciler!");

        KubernetesClient client = new KubernetesClientBuilder().build();
        Operator operator = new Operator(client);
        operator.register(reconciler);
        operator.start();
    }

}
