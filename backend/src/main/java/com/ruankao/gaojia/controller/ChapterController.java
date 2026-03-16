package com.ruankao.gaojia.controller;

import com.ruankao.gaojia.dto.ChapterTreeNode;
import com.ruankao.gaojia.dto.QuestionListItem;
import com.ruankao.gaojia.service.ChapterService;
import com.ruankao.gaojia.service.QuestionService;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.bind.annotation.RestController;

import java.util.List;

@RestController
@RequestMapping("/api/chapters")
public class ChapterController {

    private final ChapterService chapterService;
    private final QuestionService questionService;

    public ChapterController(ChapterService chapterService, QuestionService questionService) {
        this.chapterService = chapterService;
        this.questionService = questionService;
    }

    @GetMapping("/tree")
    public List<ChapterTreeNode> tree() {
        return chapterService.tree();
    }

    @GetMapping("/{chapterId}/questions")
    public List<QuestionListItem> questions(
            @PathVariable Long chapterId,
            @RequestParam(defaultValue = "false") boolean favoriteOnly,
            @RequestParam(defaultValue = "false") boolean difficultOnly,
            @RequestParam(defaultValue = "false") boolean wrongOnly
    ) {
        return questionService.chapterQuestions(chapterId, favoriteOnly, difficultOnly, wrongOnly);
    }

    @GetMapping("/{chapterId}/favorites")
    public List<QuestionListItem> favorites(@PathVariable Long chapterId) {
        return questionService.chapterQuestions(chapterId, true, false, false);
    }
}
