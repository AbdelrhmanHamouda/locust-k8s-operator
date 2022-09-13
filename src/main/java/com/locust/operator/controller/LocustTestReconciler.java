package com.locust.operator.controller;

import com.locust.operator.customresource.LocustTest;
import io.javaoperatorsdk.operator.api.reconciler.Cleaner;
import io.javaoperatorsdk.operator.api.reconciler.Context;
import io.javaoperatorsdk.operator.api.reconciler.ControllerConfiguration;
import io.javaoperatorsdk.operator.api.reconciler.DeleteControl;
import io.javaoperatorsdk.operator.api.reconciler.Reconciler;
import io.javaoperatorsdk.operator.api.reconciler.UpdateControl;
import lombok.extern.slf4j.Slf4j;

@Slf4j
@ControllerConfiguration
public class LocustTestReconciler implements Reconciler<LocustTest>, Cleaner<LocustTest> {

    // TODO:: add noop onUpdate filter
    // TODO:: add on create reaction


    @Override
    public UpdateControl<LocustTest> reconcile(LocustTest resource, Context<LocustTest> context) {
        log.info("got in namespace: {}, \nCR with name: {}, \nimage: {}, \nmaster command: {}, \nworker command: {}, \nreplicas: {},",
            resource.getMetadata().getNamespace(),
            resource.getMetadata().getName(),
            resource.getSpec().getImage(),
            resource.getSpec().getMasterCommandSeed(),
            resource.getSpec().getWorkerCommandSeed(),
            resource.getSpec().getWorkerReplicas());
        return UpdateControl.noUpdate();
    }

    @Override
    public DeleteControl cleanup(LocustTest resource, Context<LocustTest> context) {
        log.info("Deleted in namespace: {}, \nCR with name: {}, \nimage: {}, \nmaster command: {}, \nworker command: {}, \nreplicas: {},",
            resource.getMetadata().getNamespace(),
            resource.getMetadata().getName(),
            resource.getSpec().getImage(),
            resource.getSpec().getMasterCommandSeed(),
            resource.getSpec().getWorkerCommandSeed(),
            resource.getSpec().getWorkerReplicas());
        return DeleteControl.defaultDelete();
    }
}
