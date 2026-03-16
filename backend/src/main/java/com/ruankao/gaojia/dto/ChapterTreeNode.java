package com.ruankao.gaojia.dto;

import java.util.ArrayList;
import java.util.List;

public class ChapterTreeNode {

    private Long chapterId;
    private String chapterName;
    private Integer chapterLevel;
    private Integer sortNo;
    private Integer allQuestionNum;
    private List<ChapterTreeNode> children = new ArrayList<>();

    public ChapterTreeNode() {
    }

    public ChapterTreeNode(Long chapterId, String chapterName, Integer chapterLevel, Integer sortNo, Integer allQuestionNum) {
        this.chapterId = chapterId;
        this.chapterName = chapterName;
        this.chapterLevel = chapterLevel;
        this.sortNo = sortNo;
        this.allQuestionNum = allQuestionNum;
    }

    public Long getChapterId() {
        return chapterId;
    }

    public String getChapterName() {
        return chapterName;
    }

    public Integer getChapterLevel() {
        return chapterLevel;
    }

    public Integer getSortNo() {
        return sortNo;
    }

    public Integer getAllQuestionNum() {
        return allQuestionNum;
    }

    public List<ChapterTreeNode> getChildren() {
        return children;
    }
}
