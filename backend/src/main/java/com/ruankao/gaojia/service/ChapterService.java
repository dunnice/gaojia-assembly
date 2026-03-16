package com.ruankao.gaojia.service;

import com.ruankao.gaojia.dto.ChapterTreeNode;
import com.ruankao.gaojia.repository.ChapterRepository;
import org.springframework.stereotype.Service;

import java.util.List;

@Service
public class ChapterService {

    private final ChapterRepository chapterRepository;

    public ChapterService(ChapterRepository chapterRepository) {
        this.chapterRepository = chapterRepository;
    }

    public List<ChapterTreeNode> tree() {
        return chapterRepository.findTree();
    }
}
