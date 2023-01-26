package com.locust.operator.customresource.internaldto;

import com.fasterxml.jackson.annotation.JsonInclude;
import com.fasterxml.jackson.annotation.JsonInclude.Include;
import com.fasterxml.jackson.databind.JsonDeserializer;
import com.fasterxml.jackson.databind.annotation.JsonDeserialize;
import lombok.Data;

import java.io.Serializable;

@JsonDeserialize(using = JsonDeserializer.None.class)
@JsonInclude(Include.NON_NULL)
@Data
public class LocustTestAffinity implements Serializable {

    private static final long serialVersionUID = 1;
    private LocustTestNodeAffinity nodeAffinity;

}
