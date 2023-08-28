package com.locust.operator.controller.utils;

import lombok.NoArgsConstructor;

import java.util.List;

import static lombok.AccessLevel.PRIVATE;

@NoArgsConstructor(access = PRIVATE)
public class Constants {

    public static final String NODE_NAME_TEMPLATE = "%s-%s";

    // Master node constants
    public static final int MASTER_NODE_REPLICA_COUNT = 1;
    public static final int DEFAULT_WEB_UI_PORT = 8089;
    // 8089 -> Web interface
    // 5557, 5558 -> Node communication
    public static final List<Integer> MASTER_NODE_PORTS = List.of(5557, 5558, DEFAULT_WEB_UI_PORT);

    public static final Integer WORKER_NODE_PORT = 8080;
    // Master node command template: %s -> Team test configuration
    public static final String MASTER_CMD_TEMPLATE = "%s "
        // Declare `master` operation mode & availability port
        + "--master --master-port=%d "
        // Number of workers to wait for before starting the test
        + "--expect-workers=%d "
        // Auto start the test while keeping the UI available
        + "--autostart --autoquit 60 "
        // Allow to automatically rebalance users if new workers are added or removed during a test run.
        + "--enable-rebalancing "
        // Log only the summary
        + "--only-summary ";

    // Worker node constants
    // When used, output will be: "<team test config>  --worker --master-port=<port> --master-host=<host cluster url>"
    public static final String WORKER_CMD_TEMPLATE = "%s --worker --master-port=%d --master-host=%s";

    // Generic k8s constants
    public static final String APP_DEFAULT_LABEL = "performance-test-name";
    public static final String SERVICE_SELECTOR_LABEL = "performance-test-pod-name";
    public static final String MANAGED_BY_LABEL_KEY = "managed-by";
    public static final String MANAGED_BY_LABEL_VALUE = "locust-k8s-operator";

    // Environment variables names
    public static final String KAFKA_BOOTSTRAP_SERVERS = "KAFKA_BOOTSTRAP_SERVERS";
    public static final String KAFKA_SECURITY_ENABLED = "KAFKA_SECURITY_ENABLED";
    public static final String KAFKA_SECURITY_PROTOCOL_CONFIG = "KAFKA_SECURITY_PROTOCOL_CONFIG";
    public static final String KAFKA_SASL_MECHANISM = "KAFKA_SASL_MECHANISM";
    public static final String KAFKA_SASL_JAAS_CONFIG = "KAFKA_SASL_JAAS_CONFIG";
    public static final String KAFKA_USERNAME = "KAFKA_USERNAME";
    public static final String KAFKA_PASSWORD = "KAFKA_PASSWORD";

    // Service constants
    public static final String PORT_DEFAULT_NAME = "port";
    public static final String TCP_PROTOCOL = "TCP";
    public static final String METRICS_PORT_NAME = "prometheus-metrics";

    // Job constants
    public static final String DEFAULT_RESTART_POLICY = "Never";
    public static final int BACKOFF_LIMIT = 0;
    public static final String DEFAULT_MOUNT_PATH = "/lotest/src/";
    public static final String CONTAINER_ARGS_SEPARATOR = " ";

    // Node Affinity constants
    public static final String DEFAULT_NODE_MATCH_EXPRESSION_OPERATOR = "In";

    // Metrics
    public static final String PROMETHEUS_IO_SCRAPE = "prometheus.io/scrape";
    public static final String PROMETHEUS_IO_PATH = "prometheus.io/path";
    public static final String PROMETHEUS_IO_PORT = "prometheus.io/port";
    public static final String PROMETHEUS_IO_ENDPOINT = "/metrics";

    // Metrics container
    public static final String EXPORTER_CONTAINER_NAME = "locust-metrics-exporter";

    public static final String EXPORTER_URI_ENV_VAR = "LOCUST_EXPORTER_URI";
    // localhost is used because the exporter container is in the same pod as the master container.
    // This means that they share the same network
    public static final String EXPORTER_URI_ENV_VAR_VALUE = String.format("http://localhost:%s", DEFAULT_WEB_UI_PORT);

    public static final String EXPORTER_PORT_ENV_VAR = "LOCUST_EXPORTER_WEB_LISTEN_ADDRESS";

    public static final String DEFAULT_RESOURCE_TARGET = "defaultTarget";
    public static final String METRICS_EXPORTER_RESOURCE_TARGET = "metricsExporter";

}
