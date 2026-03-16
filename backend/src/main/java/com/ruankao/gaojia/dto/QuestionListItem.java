package com.ruankao.gaojia.dto;

public record QuestionListItem(
        Long questionId,
        Integer questionIndex,
        String titleHtml,
        String showTypeName,
        String knowledge,
        boolean favorite,
        boolean difficult,
        Integer wrongCount,
        String lastWrongAt,
        String analyzePreview
) {
}
