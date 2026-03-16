package com.ruankao.gaojia.controller;

import com.ruankao.gaojia.dto.QuestionListItem;
import com.ruankao.gaojia.dto.ReviewSummaryDto;
import com.ruankao.gaojia.service.QuestionService;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.bind.annotation.RestController;

import java.util.List;

@RestController
@RequestMapping("/api/review")
public class ReviewController {

    private final QuestionService questionService;

    public ReviewController(QuestionService questionService) {
        this.questionService = questionService;
    }

    @GetMapping("/wrongs")
    public List<QuestionListItem> wrongs(@RequestParam Long chapterId) {
        return questionService.chapterQuestions(chapterId, false, false, true);
    }

    @GetMapping("/summary")
    public ReviewSummaryDto summary() {
        return questionService.summary();
    }
}
