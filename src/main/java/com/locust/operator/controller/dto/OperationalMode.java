package com.locust.operator.controller.dto;

import lombok.AllArgsConstructor;
import lombok.Getter;

@AllArgsConstructor
public enum OperationalMode {

    MASTER("master"),
    WORKER("worker");

    @Getter
    public final String mode;
}
