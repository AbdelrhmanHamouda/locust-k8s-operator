package com.locust.operator.controller.config;

import io.micronaut.context.annotation.Property;
import jakarta.inject.Singleton;
import lombok.Getter;
import lombok.ToString;

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

    // * Generated pod resources
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

}
