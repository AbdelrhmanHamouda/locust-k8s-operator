package com.locust.operator.customresource;

import io.fabric8.kubernetes.api.model.Namespaced;
import io.fabric8.kubernetes.client.CustomResource;
import io.fabric8.kubernetes.model.annotation.Group;
import io.fabric8.kubernetes.model.annotation.Version;

@Group("tutorial.myfirstoperator")
@Version("v1beta1")
public class CustomService extends CustomResource<CustomServiceSpec, Void> implements Namespaced {

    private CustomServiceSpec spec;

}
