package com.locust.operator.controller.dto;

import lombok.AllArgsConstructor;
import lombok.Getter;

@AllArgsConstructor
public enum OperatorType {

    EXISTS("Exists"),
    EQUAL("Equal");

    @Getter
    public final String type;

}
