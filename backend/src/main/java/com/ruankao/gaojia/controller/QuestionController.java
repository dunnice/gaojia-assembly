package com.ruankao.gaojia.controller;

import com.ruankao.gaojia.dto.AnswerSubmitRequest;
import com.ruankao.gaojia.dto.AnswerSubmitResponse;
import com.ruankao.gaojia.dto.QuestionDetailDto;
import com.ruankao.gaojia.dto.QuestionStatusRequest;
import com.ruankao.gaojia.service.QuestionService;
import jakarta.validation.Valid;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

@RestController
@RequestMapping("/api/questions")
public class QuestionController {

    private final QuestionService questionService;

    public QuestionController(QuestionService questionService) {
        this.questionService = questionService;
    }

    @GetMapping("/{questionId}")
    public QuestionDetailDto detail(@PathVariable Long questionId) {
        return questionService.questionDetail(questionId);
    }

    @PostMapping("/{questionId}/status")
    public void updateStatus(@PathVariable Long questionId, @RequestBody QuestionStatusRequest request) {
        questionService.updateStatus(questionId, request);
    }

    @PostMapping("/{questionId}/answer")
    public AnswerSubmitResponse submitAnswer(@PathVariable Long questionId, @Valid @RequestBody AnswerSubmitRequest request) {
        return questionService.submitAnswer(questionId, request);
    }
}
