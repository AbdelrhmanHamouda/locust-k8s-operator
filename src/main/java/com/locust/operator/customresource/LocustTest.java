package com.locust.operator.customresource;

import io.fabric8.kubernetes.api.model.Namespaced;
import io.fabric8.kubernetes.client.CustomResource;
import io.fabric8.kubernetes.model.annotation.Group;
import io.fabric8.kubernetes.model.annotation.Version;

import java.io.Serial;

@Group(LocustTest.GROUP)
@Version(LocustTest.VERSION)
public class LocustTest extends CustomResource<LocustTestSpec, Void> implements Namespaced {

    public static final String GROUP = "locust.io";
    public static final String VERSION = "v1";

    // Used during deserialization to verify that the sender and receiver of a serialized object
    // have loaded classes for that object that are compatible with respect to serialization.
    // Manually setting this avoids the automatic allocation and thus removes the chance of unexpected failure during runtime.
    @Serial
    private static final long serialVersionUID = 1;

}
