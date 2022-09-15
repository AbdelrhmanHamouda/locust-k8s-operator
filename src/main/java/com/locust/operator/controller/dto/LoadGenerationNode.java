package com.locust.operator.controller.dto;

import lombok.AccessLevel;
import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.util.List;

@Data
@Builder
@AllArgsConstructor
@NoArgsConstructor(access = AccessLevel.NONE)
public class LoadGenerationNode {

    private String name;
    private String command;
    private OperationalMode operationalMode;
    private String image;
    private Integer replicas;
    private List<Integer> ports;

}
