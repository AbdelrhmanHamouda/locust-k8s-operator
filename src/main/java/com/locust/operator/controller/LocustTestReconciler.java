package com.locust.operator.controller;

import com.locust.operator.controller.utils.LoadGenHelpers;
import com.locust.operator.controller.utils.resource.manage.ResourceCreationManager;
import com.locust.operator.controller.utils.resource.manage.ResourceDeletionManager;
import com.locust.operator.customresource.LocustTest;
import io.javaoperatorsdk.operator.api.reconciler.Cleaner;
import io.javaoperatorsdk.operator.api.reconciler.Context;
import io.javaoperatorsdk.operator.api.reconciler.ControllerConfiguration;
import io.javaoperatorsdk.operator.api.reconciler.DeleteControl;
import io.javaoperatorsdk.operator.api.reconciler.Reconciler;
import io.javaoperatorsdk.operator.api.reconciler.UpdateControl;
import jakarta.inject.Singleton;
import lombok.extern.slf4j.Slf4j;

import static com.locust.operator.controller.dto.OperationalMode.MASTER;
import static com.locust.operator.controller.dto.OperationalMode.WORKER;

@Slf4j
@Singleton
@ControllerConfiguration
public class LocustTestReconciler implements Reconciler<LocustTest>, Cleaner<LocustTest> {

    private final LoadGenHelpers loadGenHelpers;
    private final ResourceCreationManager creationManager;
    private final ResourceDeletionManager deletionManager;

    public LocustTestReconciler(LoadGenHelpers loadGenHelpers, ResourceCreationManager creationManager,
        ResourceDeletionManager deletionManager) {
        this.loadGenHelpers = loadGenHelpers;
        this.creationManager = creationManager;
        this.deletionManager = deletionManager;
    }

    @Override
    public UpdateControl<LocustTest> reconcile(LocustTest resource, Context<LocustTest> context) {
        // * On update >> NOOP
        if (resource.getMetadata().getGeneration() > 1) {
            // On update will be no op as this use case is not aligned with the use case of this reconciler
            log.info("LocustTest updated: {} in namespace: {}.", resource.getMetadata().getName(), resource.getMetadata().getNamespace());
            log.info("Update operations on {} are not currently supported!", resource.getCRDName());
            return UpdateControl.noUpdate();
        }

        // * On add
        log.info("LocustTest created: '{}'", resource.getMetadata().getName());

        log.debug(
            "Custom resource information: \nNamespace: '{}'  CR name: '{}' \nImage: '{}' \nMaster command: '{}' \nWorker command: '{}' \nWorker replica count:'{}' \nconfigMap:'{}'.",
            resource.getMetadata().getNamespace(),
            resource.getMetadata().getName(),
            resource.getSpec().getImage(),
            resource.getSpec().getMasterCommandSeed(),
            resource.getSpec().getWorkerCommandSeed(),
            resource.getSpec().getWorkerReplicas(),
            resource.getSpec().getConfigMap());

        // * Construct node commands & map to internal dto
        // Generate `master` node object
        var masterNode = loadGenHelpers.generateLoadGenNodeObject(resource, MASTER);
        log.debug("Master node configuration: {}", masterNode);

        // Constructing `worker` node object
        var workerNode = loadGenHelpers.generateLoadGenNodeObject(resource, WORKER);
        log.debug("Worker node configuration: {}", workerNode);

        // * Deploy load generation resource
        // Deploy Master node service
        creationManager.createMasterService(masterNode, resource.getMetadata().getNamespace());

        // Deploy Master job
        creationManager.createJob(masterNode, resource.getMetadata().getNamespace(), resource.getMetadata().getName());

        // Deploy Worker jobs
        creationManager.createJob(workerNode, resource.getMetadata().getNamespace(), resource.getMetadata().getName());

        // TODO update status
        return UpdateControl.noUpdate();
    }

    @Override
    public DeleteControl cleanup(LocustTest resource, Context<LocustTest> context) {

        // * Log custom resource
        log.info("LocustTest deleted: {}", resource.getMetadata().getName());

        log.debug(
            "Deleted in namespace: {}, \nCR with name: {}, and generation: {}, \nimage: {}, \nmaster command: {}, \nworker command: {}, \nreplicas: {} \nconfigMap:'{}'.",
            resource.getMetadata().getNamespace(),
            resource.getMetadata().getName(),
            resource.getMetadata().getGeneration(),
            resource.getSpec().getImage(),
            resource.getSpec().getMasterCommandSeed(),
            resource.getSpec().getWorkerCommandSeed(),
            resource.getSpec().getWorkerReplicas(),
            resource.getSpec().getConfigMap());


        // * Delete load generation resource
        // Delete Master node service
        deletionManager.deleteService(resource, MASTER);

        // Delete Master job
        deletionManager.deleteJob(resource, MASTER);

        // Delete Worker jobs
        deletionManager.deleteJob(resource, WORKER);

        return DeleteControl.defaultDelete();
    }

}
