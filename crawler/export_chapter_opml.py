#!/usr/bin/env python3
"""从 MySQL 导出章节与知识点的 OPML 文件。"""

from __future__ import annotations

from collections import defaultdict
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

OUTPUT_FILE = Path("/Users/ice7/Documents/temp/gaojia-assembly/章节知识点总览.opml")


def esc(value: str) -> str:
    return escape(value, {'"': "&quot;"})


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
                    TRIM(q.knowledge) AS knowledge,
                    COUNT(*) AS question_count
                FROM ag_chapter_question cq
                JOIN ag_question q ON q.question_id = cq.question_id
                WHERE COALESCE(TRIM(q.knowledge), '') <> ''
                GROUP BY cq.chapter_id, TRIM(q.knowledge)
                ORDER BY cq.chapter_id, COUNT(*) DESC, TRIM(q.knowledge)
                """
            )
            knowledge_rows = cursor.fetchall()
    finally:
        conn.close()

    top_chapters = [row for row in chapters if int(row["parent_chapter_id"]) == 0]
    child_map: dict[int, list[dict]] = defaultdict(list)
    for row in chapters:
        parent_id = int(row["parent_chapter_id"])
        if parent_id != 0:
            child_map[parent_id].append(row)

    knowledge_map: dict[int, list[dict]] = defaultdict(list)
    for row in knowledge_rows:
        knowledge_map[int(row["chapter_id"])].append(row)

    for children in child_map.values():
        children.sort(key=lambda item: (int(item["sort_no"]), int(item["chapter_id"])))

    lines: list[str] = [
        '<?xml version="1.0" encoding="UTF-8"?>',
        '<opml version="2.0">',
        "  <head>",
        f"    <title>{esc('软考高架题库章节与知识点总览')}</title>",
        f"    <dateCreated>{datetime.now().strftime('%a, %d %b %Y %H:%M:%S +0800')}</dateCreated>",
        "    <ownerName>Codex</ownerName>",
        "  </head>",
        "  <body>",
        '    <outline text="软考高架题库章节与知识点总览">',
    ]

    for top in top_chapters:
        chapter_id = int(top["chapter_id"])
        chapter_name = str(top["chapter_name"])
        total_questions = int(top["all_question_num"])
        lines.append(
            f'      <outline text="{esc(f"{chapter_name} [{total_questions}题]")}" _chapter_id="{chapter_id}" _level="1">'
        )

        children = child_map.get(chapter_id, [])
        lines.append(f'        <outline text="{esc(f"子章节 ({len(children)})")}">')
        if children:
            for child in children:
                child_id = int(child["chapter_id"])
                child_name = str(child["chapter_name"])
                child_total = int(child["all_question_num"])
                lines.append(
                    f'          <outline text="{esc(f"{child_name} [{child_total}题]")}" _chapter_id="{child_id}" _level="2" />'
                )
        else:
            lines.append('          <outline text="无子章节" />')
        lines.append("        </outline>")

        knowledges = knowledge_map.get(chapter_id, [])
        lines.append(f'        <outline text="{esc(f"知识点 ({len(knowledges)})")}">')
        if knowledges:
            for item in knowledges:
                knowledge = str(item["knowledge"])
                count = int(item["question_count"])
                lines.append(
                    f'          <outline text="{esc(f"{knowledge} [{count}题]")}" _knowledge="{esc(knowledge)}" />'
                )
        else:
            lines.append('          <outline text="无明确知识点" />')
        lines.append("        </outline>")

        if chapter_id in knowledge_map:
            dominant = knowledge_map[chapter_id][0]
            lines.append(
                f'        <outline text="{esc(f"分析结论：本章题目主要集中在“{dominant["knowledge"]}”，共 {dominant["question_count"]} 题")}" />'
            )
        else:
            lines.append('        <outline text="分析结论：当前章节题目暂无明确 knowledge 字段，建议后续按题干或子章节语义二次归类。" />')

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
