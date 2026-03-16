package com.ruankao.gaojia.dto;

import java.util.List;

public record AnswerSubmitResponse(
        boolean correct,
        List<String> correctAnswers,
        Integer totalAttempts,
        Integer wrongCount,
        String answeredAt
) {
}
