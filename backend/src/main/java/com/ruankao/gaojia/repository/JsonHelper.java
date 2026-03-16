package com.ruankao.gaojia.repository;

import com.fasterxml.jackson.core.type.TypeReference;
import com.fasterxml.jackson.databind.ObjectMapper;

import java.util.Collections;
import java.util.List;
import java.util.Objects;

public final class JsonHelper {

    private static final ObjectMapper OBJECT_MAPPER = new ObjectMapper();

    private JsonHelper() {
    }

    public static List<String> toStringList(Object raw) {
        if (raw == null) {
            return Collections.emptyList();
        }
        try {
            if (raw instanceof String str && !str.isBlank()) {
                return OBJECT_MAPPER.readValue(str, new TypeReference<>() {
                });
            }
            return OBJECT_MAPPER.convertValue(raw, new TypeReference<>() {
            });
        } catch (Exception ex) {
            return Collections.emptyList();
        }
    }

    public static String toJson(List<String> values) {
        try {
            return OBJECT_MAPPER.writeValueAsString(Objects.requireNonNullElse(values, Collections.<String>emptyList()));
        } catch (Exception ex) {
            return "[]";
        }
    }
}
