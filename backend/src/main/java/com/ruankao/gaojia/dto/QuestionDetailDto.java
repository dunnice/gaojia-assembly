package com.ruankao.gaojia.dto;

import java.util.List;

public record QuestionDetailDto(
        Long questionId,
        String titleHtml,
        String showTypeName,
        String knowledge,
        String analyzeText,
        String materialText,
        List<String> answers,
        List<QuestionOptionDto> options,
        boolean favorite,
        boolean difficult,
        String note,
        Integer attemptCount,
        Integer wrongCount,
        String lastWrongAt,
        List<AnswerRecordDto> recentRecords
) {
}
