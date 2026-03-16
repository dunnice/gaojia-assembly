package com.ruankao.gaojia.dto;

import jakarta.validation.constraints.NotEmpty;

import java.util.List;

public record AnswerSubmitRequest(
        Long chapterId,
        @NotEmpty List<String> selectedAnswers,
        Integer durationSeconds
) {
}
