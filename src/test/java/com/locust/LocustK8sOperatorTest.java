
package com.locust;

import io.micronaut.runtime.EmbeddedApplication;
import io.micronaut.test.extensions.junit5.annotation.MicronautTest;
import lombok.extern.slf4j.Slf4j;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.Assertions;

import jakarta.inject.Inject;

@Slf4j
@MicronautTest
class LocustK8sOperatorTest {

    @Inject
    EmbeddedApplication<?> application;

    @Test
    void testItWorks() {
        log.debug("my debug message!");
        Assertions.assertTrue(application.isRunning());
    }

}
