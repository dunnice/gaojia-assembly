#!/usr/bin/env python3
"""导出包含章节、子章节、知识点、题目与解析提炼的 OPML。"""

from __future__ import annotations

import html
import json
import re
from collections import Counter, defaultdict
from datetime import datetime
from pathlib import Path
from xml.sax.saxutils import escape

import pymysql


DB_CONFIG = {
    "host": "127.0.0.1",
    "port": 3306,
    "user": "ruankao_user",
    "password": "Rk8!vN3#qL7@xP2$hT9^mZ4&cW6",
    "database": "ruankao_gaojia",
    "charset": "utf8mb4",
    "cursorclass": pymysql.cursors.DictCursor,
}

OUTPUT_FILE = Path("/Users/ice7/Documents/temp/gaojia-assembly/章节知识框架-含题目解析.opml")


def esc(value: str) -> str:
    return escape(value, {'"': "&quot;"})


def clean_text(value: str | None) -> str:
    if not value:
        return ""
    text = html.unescape(value)
    text = re.sub(r"<[^>]+>", " ", text)
    text = html.unescape(text)
    text = re.sub(r"\s+", " ", text).strip()
    return text


def short_text(value: str, limit: int = 72) -> str:
    return value if len(value) <= limit else value[: limit - 3] + "..."


def summarize_analysis(value: str | None) -> str:
    text = clean_text(value)
    if not text:
        return "暂无解析。"
    text = re.sub(r"[A-D][：:][^A-D]*", "", text)
    text = text.replace("【教材依据】", "教材依据：")
    segments = re.split(r"[。！？]", text)
    segments = [seg.strip(" ；;，,") for seg in segments if seg.strip()]
    if not segments:
        return short_text(text, 120)
    summary = "；".join(segments[:2])
    return short_text(summary, 120)


def answer_text(value: str | None) -> str:
    if not value:
        return ""
    try:
        parsed = json.loads(value)
        if isinstance(parsed, list):
            return ",".join(str(item) for item in parsed)
    except json.JSONDecodeError:
        pass
    return value


def main() -> None:
    conn = pymysql.connect(**DB_CONFIG)
    try:
        with conn.cursor() as cursor:
            cursor.execute(
                """
                SELECT chapter_id, parent_chapter_id, chapter_level, chapter_name, sort_no, all_question_num
                FROM ag_chapter
                ORDER BY chapter_level, sort_no, chapter_id
                """
            )
            chapters = cursor.fetchall()

            cursor.execute(
                """
                SELECT
                    cq.chapter_id,
                    cq.question_index,
                    q.question_id,
                    q.title_html,
                    q.analyze_text,
                    q.answer_json,
                    TRIM(q.knowledge) AS knowledge
                FROM ag_chapter_question cq
                JOIN ag_question q ON q.question_id = cq.question_id
                ORDER BY cq.chapter_id, cq.question_index, q.question_id
                """
            )
            question_rows = cursor.fetchall()
    finally:
        conn.close()

    top_chapters = [row for row in chapters if int(row["parent_chapter_id"]) == 0]
    child_map: dict[int, list[dict]] = defaultdict(list)
    for row in chapters:
        parent_id = int(row["parent_chapter_id"])
        if parent_id != 0:
            child_map[parent_id].append(row)

    for children in child_map.values():
        children.sort(key=lambda item: (int(item["sort_no"]), int(item["chapter_id"])))

    chapter_questions: dict[int, list[dict]] = defaultdict(list)
    for row in question_rows:
        chapter_questions[int(row["chapter_id"])].append(row)

    lines: list[str] = [
        '<?xml version="1.0" encoding="UTF-8"?>',
        '<opml version="2.0">',
        "  <head>",
        f"    <title>{esc('软考高架题库知识框架（含题目与解析提炼）')}</title>",
        f"    <dateCreated>{datetime.now().strftime('%a, %d %b %Y %H:%M:%S +0800')}</dateCreated>",
        "    <ownerName>Codex</ownerName>",
        "  </head>",
        "  <body>",
        '    <outline text="软考高架题库知识框架（含题目与解析提炼）">',
    ]

    for top in top_chapters:
        top_id = int(top["chapter_id"])
        top_name = str(top["chapter_name"])
        top_total = int(top["all_question_num"])
        children = child_map.get(top_id, [])
        questions = chapter_questions.get(top_id, [])

        lines.append(
            f'      <outline text="{esc(f"{top_name} [{top_total}题]")}" _chapter_id="{top_id}" _level="1">'
        )

        if not children:
            children = [{
                "chapter_id": top_id,
                "chapter_name": top_name,
                "all_question_num": top_total,
            }]

        index_cursor = 0
        for child_idx, child in enumerate(children):
            child_id = int(child["chapter_id"])
            child_name = str(child["chapter_name"])
            expected_count = int(child["all_question_num"])
            if child_idx == len(children) - 1:
                assigned = questions[index_cursor:]
            else:
                assigned = questions[index_cursor:index_cursor + expected_count]
            index_cursor += expected_count

            lines.append(
                f'        <outline text="{esc(f"{child_name} [目录题量 {expected_count} / 实际归类 {len(assigned)}]")}" _chapter_id="{child_id}" _level="2">'
            )

            knowledge_counter = Counter()
            for item in assigned:
                knowledge = clean_text(item.get("knowledge"))
                if knowledge:
                    knowledge_counter[knowledge] += 1

            lines.append(f'          <outline text="{esc(f"知识点 ({len(knowledge_counter)})")}">')
            if knowledge_counter:
                for knowledge, count in knowledge_counter.most_common():
                    lines.append(
                        f'            <outline text="{esc(f"{knowledge} [{count}题]")}" _knowledge="{esc(knowledge)}" />'
                    )
            else:
                lines.append('            <outline text="无明确 knowledge 字段" />')
            lines.append("          </outline>")

            if knowledge_counter:
                dominant_knowledge, dominant_count = knowledge_counter.most_common(1)[0]
                framework_summary = f"本小节主要围绕“{dominant_knowledge}”展开，当前归类题目 {len(assigned)} 道，其中该知识点出现 {dominant_count} 次。"
            else:
                framework_summary = f"本小节当前归类题目 {len(assigned)} 道，但题目数据未提供明确 knowledge 字段，可结合题干和解析继续细化。"
            lines.append(f'          <outline text="{esc(f"知识框架提炼：{framework_summary}")}" />')

            lines.append(f'          <outline text="{esc(f"题目清单 ({len(assigned)})")}">')
            for item in assigned:
                qid = int(item["question_id"])
                qindex = int(item["question_index"])
                title = short_text(clean_text(item.get("title_html")), 90)
                analyze = summarize_analysis(item.get("analyze_text"))
                knowledge = clean_text(item.get("knowledge")) or "未标注"
                answer = answer_text(item.get("answer_json")) or "未知"

                lines.append(
                    f'            <outline text="{esc(f"Q{qindex}. {title}")}" _question_id="{qid}">'
                )
                lines.append(f'              <outline text="{esc(f"知识点：{knowledge}")}" />')
                lines.append(f'              <outline text="{esc(f"答案：{answer}")}" />')
                lines.append(f'              <outline text="{esc(f"解析提炼：{analyze}")}" />')
                lines.append("            </outline>")
            lines.append("          </outline>")
            lines.append("        </outline>")

        lines.append("      </outline>")

    lines.extend([
        "    </outline>",
        "  </body>",
        "</opml>",
        "",
    ])

    OUTPUT_FILE.write_text("\n".join(lines), encoding="utf-8")
    print(f"已生成: {OUTPUT_FILE}")


if __name__ == "__main__":
    main()
