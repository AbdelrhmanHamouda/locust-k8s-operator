package com.locust.operator.controller.dto;

import io.fabric8.kubernetes.api.model.ResourceRequirements;
import lombok.AccessLevel;
import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

@Data
@Builder
@AllArgsConstructor
@NoArgsConstructor(access = AccessLevel.NONE)
public class MetricsExporterContainer {

    private String containerName;
    private String containerImage;
    private String pullPolicy;
    private int exporterPort;
    private ResourceRequirements resourceRequirements;

}
