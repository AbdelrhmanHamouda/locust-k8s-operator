package com.locust.operator.controller.utils.resource.manage;

import com.locust.operator.controller.dto.LoadGenerationNode;
import io.fabric8.kubernetes.api.model.Service;
import io.fabric8.kubernetes.api.model.batch.v1.Job;
import io.fabric8.kubernetes.client.ConfigBuilder;
import io.fabric8.kubernetes.client.KubernetesClient;
import io.fabric8.kubernetes.client.KubernetesClientBuilder;
import jakarta.inject.Singleton;
import lombok.extern.slf4j.Slf4j;

@Slf4j
@Singleton
public class ResourceCreationManager {

    private final ResourceCreationHelpers creationHelper;

    public ResourceCreationManager(ResourceCreationHelpers creationHelper) {
        this.creationHelper = creationHelper;
    }

    public void createJob(LoadGenerationNode nodeConfig, String namespace, String testName) {

        try (KubernetesClient client = new KubernetesClientBuilder().withConfig(new ConfigBuilder().build()).build()) {

            log.info("Creating Job for: {} in namespace: {}", nodeConfig.getName(), namespace);

            Job job = creationHelper.prepareJob(nodeConfig, testName);

            job = client.batch().v1().jobs().inNamespace(namespace).createOrReplace(job);
            log.info("Created job with name: {}", job.getMetadata().getName());
            log.debug("Created job details: {}", job);
        } catch (Exception e) {

            log.error("Exception occurred during Job creation: {}", e.getLocalizedMessage(), e);

        }

    }

    public void createMasterService(LoadGenerationNode nodeConfig, String namespace) {
        try (KubernetesClient client = new KubernetesClientBuilder().withConfig(new ConfigBuilder().build()).build()) {

            log.info("Creating service for: {} in namespace: {}", nodeConfig.getName(), namespace);

            Service service = creationHelper.prepareService(nodeConfig);

            service = client.services().inNamespace(namespace).create(service);
            log.info("Created service with name: {}", service.getMetadata().getName());
            log.debug("Created service {}", service);

        } catch (Exception e) {

            log.error("Exception occurred during service creation: {}", e.getLocalizedMessage(), e);

        }
    }

}
