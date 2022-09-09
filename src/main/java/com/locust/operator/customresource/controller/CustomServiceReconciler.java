package com.locust.operator.customresource.controller;

import com.locust.operator.customresource.CustomService;
import io.javaoperatorsdk.operator.api.reconciler.Cleaner;
import io.javaoperatorsdk.operator.api.reconciler.Context;
import io.javaoperatorsdk.operator.api.reconciler.ControllerConfiguration;
import io.javaoperatorsdk.operator.api.reconciler.DeleteControl;
import io.javaoperatorsdk.operator.api.reconciler.Reconciler;
import io.javaoperatorsdk.operator.api.reconciler.UpdateControl;
import lombok.extern.slf4j.Slf4j;

@Slf4j
@ControllerConfiguration
public class CustomServiceReconciler implements Reconciler<CustomService>, Cleaner<CustomService> {

    @Override
    public UpdateControl<CustomService> reconcile(CustomService resource, Context<CustomService> context)  {
        log.info("got {}, and {}", resource.getSpec().getLabel(), resource.getSpec().getName());
        return UpdateControl.noUpdate();
    }

    @Override
    public DeleteControl cleanup(CustomService resource, Context<CustomService> context) {
        log.info("deleted {}, and {}", resource.getSpec().getLabel(), resource.getSpec().getName());
        return DeleteControl.defaultDelete();
    }

}
