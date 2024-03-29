package com.locust.operator.controller;

import com.locust.operator.customresource.LocustTest;
import com.locust.operator.customresource.LocustTestSpec;
import io.fabric8.kubernetes.api.model.ObjectMetaBuilder;
import io.fabric8.kubernetes.api.model.StatusDetails;
import io.fabric8.kubernetes.api.model.apiextensions.v1.CustomResourceDefinition;
import io.fabric8.kubernetes.client.Config;
import io.fabric8.kubernetes.client.KubernetesClient;
import io.fabric8.kubernetes.client.KubernetesClientBuilder;
import lombok.NoArgsConstructor;
import lombok.SneakyThrows;
import lombok.extern.slf4j.Slf4j;
import lombok.val;

import java.io.ByteArrayInputStream;
import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.util.Collections;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

import static com.locust.operator.controller.dto.OperationalMode.MASTER;
import static com.locust.operator.controller.dto.OperationalMode.WORKER;
import static com.locust.operator.customresource.LocustTest.GROUP;
import static com.locust.operator.customresource.LocustTest.VERSION;
import static lombok.AccessLevel.PRIVATE;

import static java.nio.charset.StandardCharsets.UTF_8;

@Slf4j
@NoArgsConstructor(access = PRIVATE)
public class TestFixtures {

    public static final String CRD_FILE_PATH = "charts/locust-k8s-operator/crds/locust-test-crd.yaml";
    public static final String DEFAULT_API_VERSION = GROUP + "/" + VERSION;
    public static final String KIND = "LocustTest";
    public static final String DEFAULT_SEED_COMMAND = "--locustfile src/demo.py";
    public static final String DEFAULT_TEST_IMAGE = "xlocust:latest";
    public static final String DEFAULT_IMAGE_PULL_POLICY = "IfNotPresent";
    public static final List<String> DEFAULT_IMAGE_PULL_SECRETS = Collections.emptyList();
    public static final String DEFAULT_TEST_CONFIGMAP = "demo-test-configmap";
    public static final String DEFAULT_NAMESPACE = "default";
    public static final int REPLICAS = 50;
    public static final long DEFAULT_CR_GENERATION = 1L;
    public static final Map<String, String> DEFAULT_MASTER_LABELS = Map.of("role", "master");
    public static final Map<String, String> DEFAULT_WORKER_LABELS = Map.of("role", "worker");
    public static final Map<String, String> DEFAULT_MASTER_ANNOTATIONS = Map.of("locust.io/role", "master");
    public static final Map<String, String> DEFAULT_WORKER_ANNOTATIONS = new HashMap<>();

    @SneakyThrows
    public static CustomResourceDefinition prepareCustomResourceDefinition(KubernetesClient k8sClient) {

        return loadCrdFile(Paths.get(CRD_FILE_PATH), k8sClient);
    }

    private static CustomResourceDefinition loadCrdFile(Path path, KubernetesClient k8sClient) throws IOException {

        // Purge HELM specific lines from CRD file
        ByteArrayInputStream inputStream = removeHelmSpecialLines(path);

        // Load CRD
        return k8sClient.apiextensions().v1()
            .customResourceDefinitions()
            .load(inputStream)
            .item();
    }

    /**
     * Removes HELM condition that is not supported when loading the CRD for the component tests.
     *
     * @param path Path to the CRD file
     * @return Processed file content
     */
    private static ByteArrayInputStream removeHelmSpecialLines(Path path) throws IOException {

        StringBuilder stringBuilder = new StringBuilder();
        Files.lines(path).filter(line -> !line.startsWith("{{"))
            .map(line -> line + "\n")
            .forEach(stringBuilder::append);
        return new ByteArrayInputStream(stringBuilder.toString().getBytes(UTF_8));
    }

    public static CustomResourceDefinition createCrd(CustomResourceDefinition crd, KubernetesClient k8sClient) {
        return k8sClient.apiextensions().v1().customResourceDefinitions().resource(crd).create();
    }

    public static List<StatusDetails> deleteLocustTestCrd(KubernetesClient k8sClient) {

        log.debug("Deleting LocustTest CRD instances");

        val crdClient = k8sClient.apiextensions().v1().customResourceDefinitions().withName("locusttests.locust.io");
        return crdClient.delete();
    }

    public static LocustTest prepareLocustTest(String resourceName) {

        return prepareLocustTest(resourceName, REPLICAS, DEFAULT_CR_GENERATION);

    }

    public static LocustTest prepareLocustTest(String resourceName, Integer replicas, Long generation) {

        var locustTest = new LocustTest();

        // API version
        locustTest.setApiVersion(DEFAULT_API_VERSION);

        // Kind
        locustTest.setKind(KIND);

        // Metadata
        locustTest.setMetadata(new ObjectMetaBuilder()
            .withName(resourceName)
            .withNamespace(DEFAULT_NAMESPACE)
            .withGeneration(generation)
            .build());

        // Spec
        var spec = new LocustTestSpec();
        spec.setMasterCommandSeed(DEFAULT_SEED_COMMAND);
        spec.setWorkerCommandSeed(DEFAULT_SEED_COMMAND);
        spec.setConfigMap(DEFAULT_TEST_CONFIGMAP);
        spec.setImage(DEFAULT_TEST_IMAGE);
        spec.setImagePullPolicy(DEFAULT_IMAGE_PULL_POLICY);
        spec.setImagePullSecrets(DEFAULT_IMAGE_PULL_SECRETS);
        spec.setWorkerReplicas(replicas);

        var labels = new HashMap<String, Map<String, String>>();
        labels.put(MASTER.getMode(), DEFAULT_MASTER_LABELS);
        labels.put(WORKER.getMode(), DEFAULT_WORKER_LABELS);
        spec.setLabels(labels);

        var annotations = new HashMap<String, Map<String, String>>();
        annotations.put(MASTER.getMode(), DEFAULT_MASTER_ANNOTATIONS);
        annotations.put(WORKER.getMode(), DEFAULT_WORKER_ANNOTATIONS);
        spec.setAnnotations(annotations);

        locustTest.setSpec(spec);
        log.debug("Created resource object:\n{}", locustTest);

        return locustTest;

    }

    /**
     * Creates  a new instance of KubernetesClient using the provided YAML configuration.
     * <p>
     * This method uses the KubernetesClientBuilder to create a new KubernetesClient. The builder is configured with a Config object, which
     * is created from the provided YAML configuration using the Config.fromKubeconfig method.
     *
     * @param configYaml A string representing the Kubernetes configuration in YAML format.
     * @return A new instance of KubernetesClient configured according to the provided YAML configuration.
     */
    public static KubernetesClient creatKubernetesClient(String configYaml) {
        // Instantiate a KubernetesClientBuilder, configure it with the provided YAML configuration
        return new KubernetesClientBuilder().
            withConfig(Config.fromKubeconfig(configYaml))
            .build();
    }

    /**
     * Prepares and creates a Custom Resource Definition (CRD) in the Kubernetes cluster associated with the provided client.
     * <p>
     * This method first prepares a CRD using the prepareCustomResourceDefinition method. It then creates that CRD in the Kubernetes cluster
     * using the createCrd method. Both of these methods use the provided KubernetesClient to interact with the Kubernetes API.
     * After the CRD is created, it logs the details of the created CRD and returns it.
     *
     * @param testClient The KubernetesClient to use when interacting with the Kubernetes server API.
     * @return The created CustomResourceDefinition.
     */
    public static CustomResourceDefinition setupCustomResourceDefinition(KubernetesClient testClient) {
        // Prepare and create the Custom Resource Definition
        val expectedCrd = prepareCustomResourceDefinition(testClient);

        // Create the Custom Resource Definition
        val crd = createCrd(expectedCrd, testClient);

        // Log and return the created CRD
        log.debug("Created CRD details: {}", crd);
        return crd;
    }

}
