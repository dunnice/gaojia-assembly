package com.ruankao.gaojia.service;

import com.ruankao.gaojia.dto.AnswerSubmitRequest;
import com.ruankao.gaojia.dto.AnswerSubmitResponse;
import com.ruankao.gaojia.dto.QuestionDetailDto;
import com.ruankao.gaojia.dto.QuestionListItem;
import com.ruankao.gaojia.dto.QuestionStatusRequest;
import com.ruankao.gaojia.dto.ReviewSummaryDto;
import com.ruankao.gaojia.repository.QuestionRepository;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Service;

import java.time.LocalDateTime;
import java.time.format.DateTimeFormatter;
import java.util.List;

@Service
public class QuestionService {

    private static final DateTimeFormatter DATE_TIME_FORMATTER = DateTimeFormatter.ofPattern("yyyy-MM-dd HH:mm:ss");

    private final QuestionRepository questionRepository;
    private final Long defaultUserId;

    public QuestionService(QuestionRepository questionRepository, @Value("${app.default-user-id:1}") Long defaultUserId) {
        this.questionRepository = questionRepository;
        this.defaultUserId = defaultUserId;
    }

    public List<QuestionListItem> chapterQuestions(Long chapterId, boolean favoriteOnly, boolean difficultOnly, boolean wrongOnly) {
        return questionRepository.findQuestions(defaultUserId, chapterId, favoriteOnly, difficultOnly, wrongOnly);
    }

    public QuestionDetailDto questionDetail(Long questionId) {
        return questionRepository.findQuestionDetail(defaultUserId, questionId);
    }

    public void updateStatus(Long questionId, QuestionStatusRequest request) {
        QuestionDetailDto current = questionDetail(questionId);
        questionRepository.upsertQuestionStatus(
                defaultUserId,
                questionId,
                request.favorite() == null ? current.favorite() : request.favorite(),
                request.difficult() == null ? current.difficult() : request.difficult(),
                request.note() == null ? current.note() : request.note()
        );
    }

    public AnswerSubmitResponse submitAnswer(Long questionId, AnswerSubmitRequest request) {
        List<String> correctAnswers = questionRepository.findCorrectAnswers(questionId);
        boolean correct = questionRepository.isCorrect(request.selectedAnswers(), correctAnswers);
        questionRepository.insertAnswerRecord(
                defaultUserId,
                request.chapterId(),
                questionId,
                request.selectedAnswers(),
                correctAnswers,
                correct,
                request.durationSeconds()
        );
        return new AnswerSubmitResponse(
                correct,
                correctAnswers,
                questionRepository.countAttempts(defaultUserId, questionId),
                questionRepository.countWrongs(defaultUserId, questionId),
                LocalDateTime.now().format(DATE_TIME_FORMATTER)
        );
    }

    public ReviewSummaryDto summary() {
        return new ReviewSummaryDto(
                questionRepository.totalQuestions(),
                questionRepository.favoriteQuestions(defaultUserId),
                questionRepository.difficultQuestions(defaultUserId),
                questionRepository.wrongQuestions(defaultUserId)
        );
    }
}
