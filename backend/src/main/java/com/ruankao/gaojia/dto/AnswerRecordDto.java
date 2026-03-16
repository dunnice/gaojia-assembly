package com.ruankao.gaojia.dto;

import java.util.List;

public record AnswerRecordDto(
        Long id,
        Long chapterId,
        List<String> selectedAnswers,
        boolean correct,
        String answeredAt
) {
}
