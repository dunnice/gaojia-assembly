package com.ruankao.gaojia.dto;

public record ReviewSummaryDto(
        Integer totalQuestions,
        Integer favoriteQuestions,
        Integer difficultQuestions,
        Integer wrongQuestions
) {
}
