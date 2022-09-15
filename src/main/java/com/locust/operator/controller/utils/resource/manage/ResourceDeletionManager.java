package com.locust.operator.controller.utils.resource.manage;

import com.locust.operator.controller.dto.OperationalMode;
import com.locust.operator.controller.utils.LoadGenHelpers;
import com.locust.operator.customresource.LocustTest;
import io.fabric8.kubernetes.api.model.StatusDetails;
import io.fabric8.kubernetes.client.ConfigBuilder;
import io.fabric8.kubernetes.client.KubernetesClient;
import io.fabric8.kubernetes.client.KubernetesClientBuilder;
import jakarta.inject.Singleton;
import lombok.extern.slf4j.Slf4j;
import lombok.val;

import java.util.List;
import java.util.Optional;

@Slf4j
@Singleton
public class ResourceDeletionManager {

    private final LoadGenHelpers loadGenHelpers;

    public ResourceDeletionManager(LoadGenHelpers loadGenHelpers) {
        this.loadGenHelpers = loadGenHelpers;
    }

    public Optional<List<StatusDetails>> deleteJob(LocustTest crdInstance, OperationalMode mode) {

        try (KubernetesClient client = new KubernetesClientBuilder().withConfig(new ConfigBuilder().build()).build()) {

            val namespace = crdInstance.getMetadata().getNamespace();
            val resourceName = loadGenHelpers.constructNodeName(crdInstance, mode);

            log.info("Deleting Job for: {} in namespace: {}", crdInstance.getMetadata().getName(), namespace);
            return Optional.ofNullable(client.batch().v1().jobs().inNamespace(namespace).withName(resourceName).delete());

        } catch (Exception e) {

            log.error("Exception occurred during Job deletion: {}", e.getLocalizedMessage(), e);
            return Optional.empty();

        }

    }

    public Optional<List<StatusDetails>> deleteService(LocustTest crdInstance, OperationalMode mode) {

        try (KubernetesClient client = new KubernetesClientBuilder().withConfig(new ConfigBuilder().build()).build()) {

            val namespace = crdInstance.getMetadata().getNamespace();
            val resourceName = loadGenHelpers.constructNodeName(crdInstance, mode);

            log.info("Deleting Service for: {} in namespace: {}", crdInstance.getMetadata().getName(), namespace);
            return Optional.ofNullable(client.services().inNamespace(namespace).withName(resourceName).delete());

        } catch (Exception e) {

            log.error("Exception occurred during Service deletion: {}", e.getLocalizedMessage(), e);
            return Optional.empty();

        }

    }

}
