package com.locust.operator.customresource;

import com.fasterxml.jackson.annotation.JsonInclude;
import com.fasterxml.jackson.annotation.JsonInclude.Include;
import com.fasterxml.jackson.databind.JsonDeserializer;
import com.fasterxml.jackson.databind.annotation.JsonDeserialize;
import com.locust.operator.customresource.internaldto.LocustTestAffinity;
import io.fabric8.kubernetes.api.model.KubernetesResource;
import lombok.Data;

import java.io.Serial;
import java.util.Map;

// This @JsonDeserialize overrides the deserializer used in KubernetesResource,
// in order to be able to deserialize correctly the fields in the 'spec' field of the json
@JsonDeserialize(using = JsonDeserializer.None.class)
@JsonInclude(Include.NON_NULL)
@Data
public class LocustTestSpec implements KubernetesResource {

    // Used during deserialization to verify that the sender and receiver of a serialized object
    // have loaded classes for that object that are compatible with respect to serialization.
    // Manually setting this avoids the automatic allocation and thus removes the chance of unexpected failure during runtime.
    @Serial
    private static final long serialVersionUID = 1;

    private Map<String, Map<String, String>> labels;
    private Map<String, Map<String, String>> annotations;
    private LocustTestAffinity affinity;
    private String masterCommandSeed;
    private String workerCommandSeed;
    private Integer workerReplicas;
    private String configMap;
    private String image;

}
