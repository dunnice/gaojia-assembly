package com.ruankao.gaojia.dto;

public record QuestionStatusRequest(
        Boolean favorite,
        Boolean difficult,
        String note
) {
}
