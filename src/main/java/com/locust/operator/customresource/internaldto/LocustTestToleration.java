package com.locust.operator.customresource.internaldto;

import com.fasterxml.jackson.annotation.JsonInclude;
import com.fasterxml.jackson.annotation.JsonInclude.Include;
import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.io.Serializable;

@Data
@NoArgsConstructor
@AllArgsConstructor
@JsonInclude(Include.NON_NULL)
public class LocustTestToleration implements Serializable {

    private static final long serialVersionUID = 1;

    private String key;
    private String operator;
    private String value;
    private String effect;

}
