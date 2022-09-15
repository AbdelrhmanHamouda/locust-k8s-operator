package com.locust.operator.controller.utils;

import lombok.NoArgsConstructor;

import java.time.LocalDateTime;
import java.time.format.DateTimeFormatter;

import static lombok.AccessLevel.PRIVATE;

@NoArgsConstructor(access = PRIVATE)
public class DateUtils {

    public static final String DEFAULT_PATTERN = "yyyy-MM-dd_hh:mm";

    /**
     * Get the actual date in the Year-Month-Day-Hour-Minute-Second format Date pattern: "yyyy-MM-dd_hh:mm"
     *
     * @return String with the date representation
     */
    public static String getDate() {

        return getDate(DEFAULT_PATTERN);
    }

    public static String getDate(String pattern) {

        DateTimeFormatter formatter = DateTimeFormatter.ofPattern(pattern);
        return formatter.format(LocalDateTime.now());
    }

}
