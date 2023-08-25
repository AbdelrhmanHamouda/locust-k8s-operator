package com.locust.operator.controller.config;

import io.micronaut.context.annotation.Property;
import jakarta.inject.Singleton;
import lombok.Getter;
import lombok.ToString;
import org.apache.commons.lang3.math.NumberUtils;

@Getter
@ToString
@Singleton
public class SysConfig {

    // * Kafka
    @Property(name = "config.load-generation-pods.kafka.bootstrap-servers")
    private String kafkaBootstrapServers;
    @Property(name = "config.load-generation-pods.kafka.security.enabled")
    private boolean kafkaSecurityEnabled;
    @Property(name = "config.load-generation-pods.kafka.security.protocol")
    private String kafkaSecurityProtocol;
    @Property(name = "config.load-generation-pods.kafka.security.username")
    private String kafkaUsername;
    @Property(name = "config.load-generation-pods.kafka.security.password")
    private String kafkaUserPassword;
    @Property(name = "config.load-generation-pods.kafka.sasl.mechanism")
    private String kafkaSaslMechanism;
    @ToString.Exclude
    @Property(name = "config.load-generation-pods.kafka.sasl.jaas.config")
    private String kafkaSaslJaasConfig;

    // * Generated job characteristics
    /**
     * We use Object here to prevent automatic conversion from null to 0.
     * <p>
     * See {@link #getTtlSecondsAfterFinished()} for understanding how the
     * value is converted to an integer.
     */
    @Property(name = "config.load-generation-jobs.ttl-seconds-after-finished")
    private Object ttlSecondsAfterFinished;

    // * Generated pod characteristics
    @Property(name = "config.load-generation-pods.resource.cpu-request")
    private String podCpuRequest;
    @Property(name = "config.load-generation-pods.resource.mem-request")
    private String podMemRequest;
    @Property(name = "config.load-generation-pods.resource.ephemeralStorage-request")
    private String podEphemeralStorageRequest;
    @Property(name = "config.load-generation-pods.resource.cpu-limit")
    private String podCpuLimit;
    @Property(name = "config.load-generation-pods.resource.mem-limit")
    private String podMemLimit;
    @Property(name = "config.load-generation-pods.resource.ephemeralStorage-limit")
    private String podEphemeralStorageLimit;

    // * Metrics exporter container characteristics
    @Property(name = "config.load-generation-pods.metricsExporter.image")
    private String metricsExporterImage;
    @Property(name = "config.load-generation-pods.metricsExporter.port")
    private Integer metricsExporterPort;
    @Property(name = "config.load-generation-pods.metricsExporter.pullPolicy")
    private String metricsExporterPullPolicy;
    @Property(name = "config.load-generation-pods.metricsExporter.resource.cpu-request")
    private String metricsExporterCpuRequest;
    @Property(name = "config.load-generation-pods.metricsExporter.resource.mem-request")
    private String metricsExporterMemRequest;
    @Property(name = "config.load-generation-pods.metricsExporter.resource.ephemeralStorage-request")
    private String metricsExporterEphemeralStorageRequest;
    @Property(name = "config.load-generation-pods.metricsExporter.resource.cpu-limit")
    private String metricsExporterCpuLimit;
    @Property(name = "config.load-generation-pods.metricsExporter.resource.mem-limit")
    private String metricsExporterMemLimit;
    @Property(name = "config.load-generation-pods.metricsExporter.resource.ephemeralStorage-limit")
    private String metricsExporterEphemeralStorageLimit;

    @Property(name = "config.load-generation-pods.affinity.enableCrInjection")
    private boolean affinityCrInjectionEnabled;
    @Property(name = "config.load-generation-pods.taintTolerations.enableCrInjection")
    private boolean tolerationsCrInjectionEnabled;

    /**
     * Value configured for setting Kubernetes Jobs' ttlSecondsAfterFinished property.
     * This method will try to convert the value to an integer or fail and report invalid values.
     * {@code null} or empty strings will result in a {@code null} return.
     *
     * @return either {@code null} or an integer value greater than or equal to 0
     */
    public Integer getTtlSecondsAfterFinished() {
        final String stringValue = String.valueOf(this.ttlSecondsAfterFinished);

        if (NumberUtils.isDigits(stringValue)) {
            return Integer.parseInt(stringValue);
        } else if (stringValue.isEmpty()) {
            return null;
        } else {
            throw new IllegalArgumentException(
                String.format(
                    "Invalid value '%s' for property ttl-seconds-after-finished",
                    stringValue
                )
            );
        }
    }
}
