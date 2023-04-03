package com.locust.operator.controller.dto;

import com.locust.operator.customresource.internaldto.LocustTestAffinity;
import com.locust.operator.customresource.internaldto.LocustTestToleration;
import lombok.AccessLevel;
import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.util.List;
import java.util.Map;

@Data
@Builder
@AllArgsConstructor
@NoArgsConstructor(access = AccessLevel.NONE)
public class LoadGenerationNode {

    private String name;
    private Map<String, String> labels;
    private Map<String, String> annotations;
    private LocustTestAffinity affinity;
    private List<LocustTestToleration> tolerations;
    private List<String> command;
    private OperationalMode operationalMode;
    private String image;
    private String imagePullPolicy;
    private List<String> imagePullSecrets;
    private Integer replicas;
    private List<Integer> ports;
    private String configMap;

}
