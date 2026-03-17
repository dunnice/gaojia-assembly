#!/usr/bin/env python3
"""同步 51CTO 软考高架题库到本地 MySQL。"""

from __future__ import annotations

import argparse
import base64
import json
import re
import sys
import time
import uuid
from dataclasses import dataclass
from datetime import datetime
from pathlib import Path
from typing import Any
from urllib.parse import urljoin

import pymysql
import requests


def now_str() -> str:
    return datetime.now().strftime("%Y-%m-%d %H:%M:%S")


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="同步 51CTO 软考高架题库到 MySQL。章与节结构来自 /napi/chapter/list/sub-{subject} 接口。"
    )
    parser.add_argument(
        "--config",
        default=str(Path(__file__).with_name("config.json")),
        help="配置文件路径，默认读取当前目录 config.json",
    )
    parser.add_argument(
        "--sync-chapters",
        action="store_true",
        help="仅同步章节结构（章+节）及章节与题目的关联关系，不拉取题目详情。等价于 --skip-question-detail",
    )
    parser.add_argument(
        "--skip-question-detail",
        action="store_true",
        help="只同步章节和章节题目列表，不拉取题目详情",
    )
    parser.add_argument(
        "--chapter-id",
        type=int,
        help="只同步指定一级章节ID，不指定则同步全部",
    )
    parser.add_argument(
        "--force",
        action="store_true",
        help="强制重新拉取，忽略本地已存在数据",
    )
    args = parser.parse_args()
    if args.sync_chapters:
        args.skip_question_detail = True
    return args


def load_config(path: str) -> dict[str, Any]:
    config_path = Path(path)
    if not config_path.exists():
        raise FileNotFoundError(
            f"配置文件不存在: {config_path}，请先复制 config.example.json 为 config.json 再填写数据库和 Cookie。"
        )
    return json.loads(config_path.read_text(encoding="utf-8"))


def json_dumps(value: Any) -> str | None:
    if value is None:
        return None
    return json.dumps(value, ensure_ascii=False)


def safe_int(value: Any) -> int | None:
    if value in (None, "", "null", "None"):
        return None
    try:
        return int(value)
    except (TypeError, ValueError):
        return None


def bool_to_tinyint(value: Any) -> int:
    if isinstance(value, bool):
        return 1 if value else 0
    if value in (1, "1", "true", "True"):
        return 1
    return 0


def parse_remote_datetime(value: Any) -> str | None:
    if value in (None, "", "0", 0, "0000-00-00 00:00:00"):
        return None
    if isinstance(value, (int, float)) or (isinstance(value, str) and value.isdigit()):
        ts = int(value)
        if ts <= 0:
            return None
        return datetime.fromtimestamp(ts).strftime("%Y-%m-%d %H:%M:%S")
    if isinstance(value, str):
        try:
            dt = datetime.strptime(value, "%Y-%m-%d %H:%M:%S")
            return dt.strftime("%Y-%m-%d %H:%M:%S")
        except ValueError:
            return None
    return None


def normalize_option_list(raw_option: Any) -> list[str]:
    if raw_option in (None, ""):
        return []
    if isinstance(raw_option, list):
        return [str(item) for item in raw_option]
    if isinstance(raw_option, str):
        try:
            parsed = json.loads(raw_option)
            if isinstance(parsed, list):
                return [str(item) for item in parsed]
        except json.JSONDecodeError:
            return [raw_option]
    return [str(raw_option)]


def normalize_answer_list(raw_answer: Any) -> list[str]:
    if raw_answer in (None, ""):
        return []
    if isinstance(raw_answer, list):
        return [str(item) for item in raw_answer]
    return [str(raw_answer)]


def option_label(index: int) -> str:
    return chr(ord("A") + index - 1)


def _build_chapter_tree(chapter_list: list[dict[str, Any]]) -> list[dict[str, Any]]:
    """
    将 chapter/list 返回的列表构建为章-节树。
    支持嵌套结构（含 child/children）或扁平结构（pid 区分父子）。
    """
    if not chapter_list:
        return []
    if any(item.get("child") or item.get("children") for item in chapter_list):
        return chapter_list
    pid_key = "pid"
    id_key = "id"
    child_map: dict[int, list[dict[str, Any]]] = {}
    tops: list[dict[str, Any]] = []
    for item in chapter_list:
        item_id = safe_int(item.get(id_key))
        pid = safe_int(item.get(pid_key))
        if not item_id:
            continue
        if not pid or pid == 0:
            tops.append(item)
        else:
            child_map.setdefault(pid, []).append(item)
    for children in child_map.values():
        children.sort(key=lambda c: (safe_int(c.get("sort")) or 0, safe_int(c.get("id")) or 0))
    for top in tops:
        top_id = safe_int(top.get(id_key))
        children = child_map.get(top_id, [])
        if children and "child" not in top:
            top["child"] = children
    tops.sort(key=lambda c: (safe_int(c.get("sort")) or 0, safe_int(c.get("id")) or 0))
    return tops


def get_sections_from_chapter(chapter: dict[str, Any]) -> list[dict[str, Any]]:
    """
    从 chapter/list 返回的章节结构中提取小节列表。
    支持 child、children 字段，按 sort、id 排序，去重。
    """
    sections: list[dict[str, Any]] = []
    seen: set[int] = set()
    for key in ("child", "children"):
        for s in chapter.get(key) or []:
            sid = safe_int(s.get("id"))
            if sid and sid not in seen:
                seen.add(sid)
                sections.append(s)
    return sorted(sections, key=lambda c: (safe_int(c.get("sort")) or 0, safe_int(c.get("id")) or 0))


def resolve_section_chapter_id(top_chapter: dict[str, Any], question_index: int) -> int:
    """
    根据题目序号与一级章节下子节的题数，计算题目归属的二级章节 ID。
    子节按 sort 排序，题号 1~n1 归属第 1 节，n1+1~n1+n2 归属第 2 节，以此类推。
    若无子节或题号超出范围，返回一级章节 ID。
    """
    children = get_sections_from_chapter(top_chapter)
    if not children:
        return safe_int(top_chapter.get("id")) or 0

    start = 1
    for child in children:
        n = safe_int(child.get("all_question_num")) or 0
        end = start + n - 1
        if start <= question_index <= end:
            return safe_int(child.get("id")) or 0
        start = end + 1
    return safe_int(top_chapter.get("id")) or 0


# 匹配 <img ... src="..." ...> 或 <img ... src='...' ...>，捕获引号和 URL
_IMG_SRC_RE = re.compile(
    r'<img\s+[^>]*src=(["\'])([^"\']+)\1[^>]*>',
    re.IGNORECASE,
)


def _resolve_image_url(src: str, base_url: str) -> str:
    """将相对或协议相对 URL 转为绝对 URL"""
    if not src or not src.strip():
        return ""
    src = src.strip()
    if src.startswith("//"):
        return "https:" + src
    if src.startswith(("http://", "https://")):
        return src
    return urljoin(base_url.rstrip("/") + "/", src)


def inline_images_in_html(
    html: str,
    base_url: str,
    session: requests.Session,
    timeout: int = 15,
) -> str:
    """
    将 HTML 中的 <img src="http(s)://..."> 拉取并替换为 data:image/xxx;base64,...
    失败时保留原 URL。
    """
    if not html or "<img" not in html.lower():
        return html

    def replace_one(match: re.Match) -> str:
        quote_char = match.group(1)
        src = match.group(2)
        full_url = _resolve_image_url(src, base_url)
        if not full_url:
            return match.group(0)
        try:
            resp = session.get(full_url, timeout=timeout)
            resp.raise_for_status()
            raw = resp.content
        except Exception:
            return match.group(0)
        content_type = resp.headers.get("Content-Type", "image/png").split(";")[0].strip()
        if content_type not in ("image/png", "image/jpeg", "image/jpg", "image/gif", "image/webp"):
            content_type = "image/png"
        b64 = base64.b64encode(raw).decode("ascii")
        data_uri = f"data:{content_type};base64,{b64}"
        return f'<img src={quote_char}{data_uri}{quote_char}'
    return _IMG_SRC_RE.sub(replace_one, html)


@dataclass
class SyncContext:
    batch_no: str
    subject_code: str


class Database:
    def __init__(self, db_config: dict[str, Any]) -> None:
        self.conn = pymysql.connect(
            host=db_config["host"],
            port=int(db_config.get("port", 3306)),
            user=db_config["user"],
            password=db_config["password"],
            database=db_config["database"],
            charset=db_config.get("charset", "utf8mb4"),
            autocommit=False,
            cursorclass=pymysql.cursors.DictCursor,
        )

    def close(self) -> None:
        self.conn.close()

    def execute(self, sql: str, params: tuple[Any, ...] | list[Any]) -> int:
        with self.conn.cursor() as cursor:
            cursor.execute(sql, params)
            return cursor.lastrowid

    def commit(self) -> None:
        self.conn.commit()

    def rollback(self) -> None:
        self.conn.rollback()

    def fetch_one(self, sql: str, params: tuple[Any, ...] | list[Any]) -> dict[str, Any] | None:
        with self.conn.cursor() as cursor:
            cursor.execute(sql, params)
            return cursor.fetchone()

    def fetch_all(self, sql: str, params: tuple[Any, ...] | list[Any]) -> list[dict[str, Any]]:
        with self.conn.cursor() as cursor:
            cursor.execute(sql, params)
            return list(cursor.fetchall())

    def insert_sync_log(
        self,
        *,
        batch_no: str,
        sync_type: str,
        subject_code: str | None,
        chapter_id: int | None,
        question_id: int | None,
        request_url: str,
        request_method: str,
        request_payload: Any,
        response_raw: Any,
        status: str,
        error_message: str = "",
    ) -> int:
        sql = """
        INSERT INTO ag_sync_batch (
            batch_no, sync_type, subject_code, chapter_id, question_id,
            request_url, request_method, request_payload, response_raw,
            status, error_message, started_at, finished_at
        ) VALUES (
            %s, %s, %s, %s, %s,
            %s, %s, CAST(%s AS JSON), CAST(%s AS JSON),
            %s, %s, NOW(), NOW()
        )
        """
        return self.execute(
            sql,
            (
                batch_no,
                sync_type,
                subject_code,
                chapter_id,
                question_id,
                request_url,
                request_method,
                json_dumps(request_payload),
                json_dumps(response_raw),
                status,
                error_message[:1000],
            ),
        )

    def upsert_subject(self, subject_code: str, subject_name: str, raw_json: Any) -> None:
        sql = """
        INSERT INTO ag_subject (subject_code, subject_name, raw_json)
        VALUES (%s, %s, CAST(%s AS JSON))
        ON DUPLICATE KEY UPDATE
            subject_name = VALUES(subject_name),
            raw_json = VALUES(raw_json),
            updated_at = CURRENT_TIMESTAMP
        """
        self.execute(sql, (subject_code, subject_name, json_dumps(raw_json)))

    def upsert_chapter(
        self,
        *,
        chapter: dict[str, Any],
        subject_code: str,
        chapter_level: int,
        last_sync_batch_id: int | None,
    ) -> None:
        sql = """
        INSERT INTO ag_chapter (
            chapter_id, subject_code, parent_chapter_id, chapter_level, chapter_name, sort_no,
            all_question_num, do_question_num, do_subject_num, right_question_num, right_question_rate,
            is_finish, raw_json, last_sync_batch_id
        ) VALUES (
            %s, %s, %s, %s, %s, %s,
            %s, %s, %s, %s, %s,
            %s, CAST(%s AS JSON), %s
        )
        ON DUPLICATE KEY UPDATE
            subject_code = VALUES(subject_code),
            parent_chapter_id = VALUES(parent_chapter_id),
            chapter_level = VALUES(chapter_level),
            chapter_name = VALUES(chapter_name),
            sort_no = VALUES(sort_no),
            all_question_num = VALUES(all_question_num),
            do_question_num = VALUES(do_question_num),
            do_subject_num = VALUES(do_subject_num),
            right_question_num = VALUES(right_question_num),
            right_question_rate = VALUES(right_question_rate),
            is_finish = VALUES(is_finish),
            raw_json = VALUES(raw_json),
            last_sync_batch_id = VALUES(last_sync_batch_id),
            updated_at = CURRENT_TIMESTAMP
        """
        self.execute(
            sql,
            (
                safe_int(chapter.get("id")),
                subject_code,
                safe_int(chapter.get("pid")) or 0,
                chapter_level,
                str(chapter.get("name", "")).strip(),
                safe_int(chapter.get("sort")) or 0,
                safe_int(chapter.get("all_question_num")) or 0,
                safe_int(chapter.get("do_question_num")) or 0,
                safe_int(chapter.get("do_subject_num")) or 0,
                safe_int(chapter.get("right_question_num")) or 0,
                str(chapter.get("right_question_rent", "")),
                safe_int(chapter.get("is_finish")),
                json_dumps(chapter),
                last_sync_batch_id,
            ),
        )

    def upsert_question(
        self,
        *,
        question: dict[str, Any],
        answer_type: Any,
        last_sync_batch_id: int | None,
    ) -> None:
        sql = """
        INSERT INTO ag_question (
            question_id, unique_value, user_id, title_html, question_type, show_type_name,
            answer_type, knowledge, analyze_text, answer_json, material_text, score_rule,
            difficulty_id, first_id, second_id, three_id, parent_id, root_id, new_parent_id,
            sort_no, sort_son, rank_no, orig, is_new, is_delete, is_repeat, error_num,
            from_exam, mark_text, analyze_video_id, creater_uid, creater_name, review_uid,
            review_name, review_status, review_comment, review_time, edit_uid, edit_name,
            edit_time, create_time_remote, updated_at_remote, raw_json, last_sync_batch_id
        ) VALUES (
            %s, %s, %s, %s, %s, %s,
            %s, %s, %s, CAST(%s AS JSON), %s, %s,
            %s, %s, %s, %s, %s, %s, %s,
            %s, %s, %s, %s, %s, %s, %s, %s,
            %s, %s, %s, %s, %s, %s,
            %s, %s, %s, %s, %s, %s,
            %s, %s, %s, CAST(%s AS JSON), %s
        )
        ON DUPLICATE KEY UPDATE
            unique_value = VALUES(unique_value),
            user_id = VALUES(user_id),
            title_html = VALUES(title_html),
            question_type = VALUES(question_type),
            show_type_name = VALUES(show_type_name),
            answer_type = VALUES(answer_type),
            knowledge = VALUES(knowledge),
            analyze_text = VALUES(analyze_text),
            answer_json = VALUES(answer_json),
            material_text = VALUES(material_text),
            score_rule = VALUES(score_rule),
            difficulty_id = VALUES(difficulty_id),
            first_id = VALUES(first_id),
            second_id = VALUES(second_id),
            three_id = VALUES(three_id),
            parent_id = VALUES(parent_id),
            root_id = VALUES(root_id),
            new_parent_id = VALUES(new_parent_id),
            sort_no = VALUES(sort_no),
            sort_son = VALUES(sort_son),
            rank_no = VALUES(rank_no),
            orig = VALUES(orig),
            is_new = VALUES(is_new),
            is_delete = VALUES(is_delete),
            is_repeat = VALUES(is_repeat),
            error_num = VALUES(error_num),
            from_exam = VALUES(from_exam),
            mark_text = VALUES(mark_text),
            analyze_video_id = VALUES(analyze_video_id),
            creater_uid = VALUES(creater_uid),
            creater_name = VALUES(creater_name),
            review_uid = VALUES(review_uid),
            review_name = VALUES(review_name),
            review_status = VALUES(review_status),
            review_comment = VALUES(review_comment),
            review_time = VALUES(review_time),
            edit_uid = VALUES(edit_uid),
            edit_name = VALUES(edit_name),
            edit_time = VALUES(edit_time),
            create_time_remote = VALUES(create_time_remote),
            updated_at_remote = VALUES(updated_at_remote),
            raw_json = VALUES(raw_json),
            last_sync_batch_id = VALUES(last_sync_batch_id),
            updated_at = CURRENT_TIMESTAMP
        """
        self.execute(
            sql,
            (
                safe_int(question.get("id") or question.get("question_id")),
                str(question.get("unique_value", "")),
                safe_int(question.get("user_id")),
                str(question.get("title") or question.get("question_title") or ""),
                str(question.get("type") or question.get("question_type") or ""),
                str(question.get("show_type_name") or ""),
                str(answer_type or question.get("answer_type") or ""),
                str(question.get("knowledge") or ""),
                str(question.get("analyze") or ""),
                json_dumps(normalize_answer_list(question.get("answer"))),
                str(question.get("material_text") or ""),
                str(question.get("score_rule") or ""),
                safe_int(question.get("difficulty_id")),
                safe_int(question.get("first_id")),
                safe_int(question.get("second_id")),
                safe_int(question.get("three_id")),
                safe_int(question.get("parent_id")),
                safe_int(question.get("root_id")),
                safe_int(question.get("new_parent_id")),
                safe_int(question.get("sort")) or 0,
                safe_int(question.get("sort_son")) or 0,
                safe_int(question.get("rank")) or 0,
                bool_to_tinyint(question.get("orig")),
                bool_to_tinyint(question.get("is_new")),
                bool_to_tinyint(question.get("is_delete")),
                bool_to_tinyint(question.get("is_repeat")),
                safe_int(question.get("error_num")) or 0,
                bool_to_tinyint(question.get("from_exam")),
                str(question.get("mark") or ""),
                safe_int(question.get("analyze_video_id")),
                safe_int(question.get("creater_uid")),
                str(question.get("creater_name") or ""),
                safe_int(question.get("review_uid")),
                str(question.get("review_name") or ""),
                safe_int(question.get("review_status")),
                str(question.get("review_comment") or ""),
                parse_remote_datetime(question.get("review_time")),
                safe_int(question.get("edit_uid")),
                str(question.get("edit_name") or ""),
                parse_remote_datetime(question.get("edit_time")),
                parse_remote_datetime(question.get("create_time")),
                parse_remote_datetime(question.get("updated_at")),
                json_dumps(question),
                last_sync_batch_id,
            ),
        )

    def replace_question_options(
        self,
        *,
        question_id: int,
        options: list[str],
        answers: list[str],
    ) -> None:
        delete_sql = "DELETE FROM ag_question_option WHERE question_id = %s"
        self.execute(delete_sql, (question_id,))

        if not options:
            return

        insert_sql = """
        INSERT INTO ag_question_option (
            question_id, option_no, option_label, option_html, is_answer, raw_value
        ) VALUES (%s, %s, %s, %s, %s, %s)
        """
        for idx, option_html in enumerate(options, start=1):
            label = option_label(idx)
            self.execute(
                insert_sql,
                (
                    question_id,
                    idx,
                    label,
                    option_html,
                    1 if label in answers else 0,
                    option_html,
                ),
            )

    def upsert_chapter_question(
        self,
        *,
        subject_code: str,
        chapter_id: int,
        section_chapter_id: int,
        question_item: dict[str, Any],
        chapter_num: int | None,
        section_num: int | None,
        last_sync_batch_id: int | None,
    ) -> None:
        sql = """
        INSERT INTO ag_chapter_question (
            subject_code, chapter_id, section_chapter_id, question_id, question_index, belong_page,
            question_type, answer_type, chapter_num, section_num, raw_json, last_sync_batch_id
        ) VALUES (
            %s, %s, %s, %s, %s, %s,
            %s, %s, %s, %s, CAST(%s AS JSON), %s
        )
        ON DUPLICATE KEY UPDATE
            subject_code = VALUES(subject_code),
            section_chapter_id = VALUES(section_chapter_id),
            question_index = VALUES(question_index),
            belong_page = VALUES(belong_page),
            question_type = VALUES(question_type),
            answer_type = VALUES(answer_type),
            chapter_num = VALUES(chapter_num),
            section_num = VALUES(section_num),
            raw_json = VALUES(raw_json),
            last_sync_batch_id = VALUES(last_sync_batch_id),
            updated_at = CURRENT_TIMESTAMP
        """
        self.execute(
            sql,
            (
                subject_code,
                chapter_id,
                section_chapter_id,
                safe_int(question_item.get("question_id")),
                safe_int(question_item.get("index")) or 0,
                safe_int(question_item.get("belong_page")) or 1,
                str(question_item.get("question_type") or ""),
                str(question_item.get("answer_type") or ""),
                chapter_num,
                section_num,
                json_dumps(question_item),
                last_sync_batch_id,
            ),
        )

    def replace_question_videos(self, question_id: int, videos: list[dict[str, Any]]) -> None:
        self.execute("DELETE FROM ag_question_video WHERE question_id = %s", (question_id,))
        if not videos:
            return
        sql = """
        INSERT INTO ag_question_video (
            question_id, video_id, video_title, video_url, raw_json
        ) VALUES (%s, %s, %s, %s, CAST(%s AS JSON))
        """
        for item in videos:
            self.execute(
                sql,
                (
                    question_id,
                    safe_int(item.get("id") or item.get("video_id")),
                    str(item.get("title") or item.get("video_title") or ""),
                    str(item.get("url") or item.get("video_url") or ""),
                    json_dumps(item),
                ),
            )

    def chapter_question_count(self, chapter_id: int) -> int:
        row = self.fetch_one(
            "SELECT COUNT(*) AS cnt FROM ag_chapter_question WHERE chapter_id = %s",
            (chapter_id,),
        )
        return int(row["cnt"]) if row else 0

    def local_question_ids_by_chapter(self, chapter_id: int) -> list[int]:
        rows = self.fetch_all(
            """
            SELECT question_id
            FROM ag_chapter_question
            WHERE chapter_id = %s
            ORDER BY question_index, question_id
            """,
            (chapter_id,),
        )
        return [int(row["question_id"]) for row in rows]

    def question_detail_exists(self, question_id: int) -> bool:
        row = self.fetch_one(
            """
            SELECT question_id
            FROM ag_question
            WHERE question_id = %s
              AND (
                    COALESCE(unique_value, '') <> ''
                 OR updated_at_remote IS NOT NULL
                 OR COALESCE(analyze_text, '') <> ''
              )
            LIMIT 1
            """,
            (question_id,),
        )
        return row is not None


class RuankaoClient:
    def __init__(self, api_config: dict[str, Any]) -> None:
        self.base_url = api_config["base_url"].rstrip("/")
        self.subject_code = str(api_config["subject_code"])
        self.timeout = int(api_config.get("timeout_seconds", 30))
        self.sleep_seconds = float(api_config.get("sleep_seconds", 0.2))
        self.session = requests.Session()
        self.session.headers.update(api_config.get("headers", {}))

    def _request(
        self,
        method: str,
        path: str,
        *,
        params: dict[str, Any] | None = None,
        json_body: dict[str, Any] | None = None,
    ) -> dict[str, Any]:
        url = f"{self.base_url}{path}"
        response = self.session.request(
            method=method,
            url=url,
            params=params,
            json=json_body,
            timeout=self.timeout,
        )
        response.raise_for_status()
        data = response.json()
        if str(data.get("status")) != "0":
            raise RuntimeError(f"接口返回失败: {url} -> {data}")
        if self.sleep_seconds > 0:
            time.sleep(self.sleep_seconds)
        return data

    def chapter_list(self) -> dict[str, Any]:
        return self._request(
            "POST",
            f"/napi/chapter/list/sub-{self.subject_code}",
            json_body={"subject": self.subject_code},
        )

    def has_progress(self, chapter_id: int) -> dict[str, Any]:
        return self._request(
            "GET",
            "/napi/chapter/has-progress",
            params={"chapter_id": chapter_id},
        )

    def create_assembly(self, chapter_id: int, num: int) -> dict[str, Any]:
        return self._request(
            "POST",
            "/napi/chapter/assembly",
            json_body={"num": num, "chapter_id": str(chapter_id)},
        )

    def chapter_questions(self, assembly_id: int) -> dict[str, Any]:
        return self._request(
            "GET",
            "/napi/chapter/questions",
            params={"assembly_id": assembly_id},
        )

    def section_questions(self, section_id: int) -> dict[str, Any]:
        """按小节 ID 获取该小节下的题目。"""
        return self._request(
            "GET",
            "/napi/chapter/section-questions",
            params={"id": section_id, "user_subject": self.subject_code},
        )

    def get_question(self, question_id: int) -> dict[str, Any]:
        return self._request(
            "GET",
            "/napi/exam/get-question",
            params={"id": question_id},
        )


class SyncService:
    def __init__(self, db: Database, client: RuankaoClient) -> None:
        self.db = db
        self.client = client

    def _inline_images_in_question(
        self,
        question: dict[str, Any],
        options: list[str],
    ) -> tuple[dict[str, Any], list[str]]:
        """将题目 HTML 中的图片 URL 拉取并替换为 base64 data URI"""
        base = self.client.base_url
        session = self.client.session
        timeout = self.client.timeout

        q = dict(question)
        for key in ("title", "question_title", "analyze", "material_text"):
            val = q.get(key)
            if val and isinstance(val, str):
                q[key] = inline_images_in_html(val, base, session, timeout)

        new_options = [
            inline_images_in_html(opt, base, session, timeout)
            for opt in options
        ]
        return q, new_options

    def _call_with_log(
        self,
        *,
        ctx: SyncContext,
        sync_type: str,
        chapter_id: int | None,
        question_id: int | None,
        method: str,
        path: str,
        request_payload: Any,
        func,
    ) -> tuple[dict[str, Any], int]:
        url = f"{self.client.base_url}{path}"
        try:
            data = func()
            log_id = self.db.insert_sync_log(
                batch_no=ctx.batch_no,
                sync_type=sync_type,
                subject_code=ctx.subject_code,
                chapter_id=chapter_id,
                question_id=question_id,
                request_url=url,
                request_method=method,
                request_payload=request_payload,
                response_raw=data,
                status="SUCCESS",
            )
            self.db.commit()
            return data, log_id
        except Exception as exc:
            self.db.insert_sync_log(
                batch_no=ctx.batch_no,
                sync_type=sync_type,
                subject_code=ctx.subject_code,
                chapter_id=chapter_id,
                question_id=question_id,
                request_url=url,
                request_method=method,
                request_payload=request_payload,
                response_raw={"error": str(exc)},
                status="FAIL",
                error_message=str(exc),
            )
            self.db.commit()
            raise

    def sync_all(self, chapter_id_filter: int | None, skip_question_detail: bool, force: bool) -> None:
        ctx = SyncContext(
            batch_no=datetime.now().strftime("%Y%m%d%H%M%S") + "_" + uuid.uuid4().hex[:8],
            subject_code=self.client.subject_code,
        )
        print(f"[{now_str()}] 开始同步，batch_no={ctx.batch_no}, subject={ctx.subject_code}")

        chapter_resp, chapter_log_id = self._call_with_log(
            ctx=ctx,
            sync_type="chapter_tree",
            chapter_id=None,
            question_id=None,
            method="POST",
            path=f"/napi/chapter/list/sub-{ctx.subject_code}",
            request_payload={"subject": ctx.subject_code},
            func=self.client.chapter_list,
        )

        chapter_data = chapter_resp.get("data", {}) or {}
        if isinstance(chapter_data, list):
            chapter_list = chapter_data
            subject_name = f"subject-{ctx.subject_code}"
        else:
            chapter_list = chapter_data.get("list", []) or []
            subject_name = chapter_data.get("title") or f"subject-{ctx.subject_code}"

        self.db.upsert_subject(ctx.subject_code, subject_name, chapter_data)

        top_chapters = _build_chapter_tree(chapter_list)
        for top in top_chapters:
            self.db.upsert_chapter(
                chapter=top,
                subject_code=ctx.subject_code,
                chapter_level=1,
                last_sync_batch_id=chapter_log_id,
            )
            for child in get_sections_from_chapter(top):
                self.db.upsert_chapter(
                    chapter=child,
                    subject_code=ctx.subject_code,
                    chapter_level=2,
                    last_sync_batch_id=chapter_log_id,
                )
        self.db.commit()
        print(f"[{now_str()}] 章节树同步完成，一级章节 {len(top_chapters)} 个")

        if chapter_id_filter is not None:
            top_chapters = [item for item in top_chapters if safe_int(item.get("id")) == chapter_id_filter]
            print(f"[{now_str()}] 已按参数过滤一级章节，剩余 {len(top_chapters)} 个")

        for top in top_chapters:
            self.sync_chapter(ctx, top, skip_question_detail, force)

        print(f"[{now_str()}] 全部同步完成")

    def sync_chapter(self, ctx: SyncContext, top_chapter: dict[str, Any], skip_question_detail: bool, force: bool) -> None:
        chapter_id = safe_int(top_chapter.get("id"))
        if not chapter_id:
            return
        total_num = safe_int(top_chapter.get("all_question_num")) or 0
        sorted_children = get_sections_from_chapter(top_chapter)

        print(f"[{now_str()}] 同步一级章节 {chapter_id} - {top_chapter.get('name')}，题数 {total_num}，小节数 {len(sorted_children)}")

        all_questions: list[dict[str, Any]] = []

        if sorted_children:
            # 按小节获取题目：每个小节调用 section-questions 接口
            for section in sorted_children:
                section_id = safe_int(section.get("id"))
                if not section_id:
                    continue
                section_name = section.get("name", "")
                self._sync_section_questions(ctx, chapter_id, section_id, section_name, all_questions)
        else:
            # 无子节时回退到 assembly 流程
            self._sync_chapter_by_assembly(ctx, top_chapter, chapter_id, total_num, all_questions)

        self.db.commit()
        print(f"[{now_str()}] 章节 {chapter_id} 题目列表已同步，共 {len(all_questions)} 题")

        if skip_question_detail:
            return

        for question_item in all_questions:
            question_id = safe_int(question_item.get("question_id"))
            if not question_id:
                continue
            if not force and self.db.question_detail_exists(question_id):
                print(f"[{now_str()}] 题目详情已存在，跳过 question_id={question_id}")
                continue
            self.sync_question_detail(ctx, chapter_id, question_id)

    def _sync_section_questions(
        self,
        ctx: SyncContext,
        chapter_id: int,
        section_id: int,
        section_name: str,
        out_questions: list[dict[str, Any]],
    ) -> None:
        """按小节 ID 调用 section-questions 接口，获取该小节下的题目。"""
        section_resp, questions_log_id = self._call_with_log(
            ctx=ctx,
            sync_type="section_questions",
            chapter_id=section_id,
            question_id=None,
            method="GET",
            path="/napi/chapter/section-questions",
            request_payload={"id": section_id, "user_subject": ctx.subject_code},
            func=lambda: self.client.section_questions(section_id),
        )

        inner = section_resp.get("data", {}) or {}
        question_data = inner.get("data", {}) or {}
        questions = question_data.get("question", []) or []
        chapter_section_num = question_data.get("chapter_section_num", {}) or {}
        chapter_num = safe_int(chapter_section_num.get("chapter_num"))
        section_num = safe_int(chapter_section_num.get("section_num"))

        for question_item in questions:
            question_id = safe_int(question_item.get("question_id"))
            if not question_id:
                continue

            placeholder = {
                "id": question_id,
                "question_id": question_id,
                "title": question_item.get("question_title", ""),
                "question_type": question_item.get("question_type", ""),
                "show_type_name": question_item.get("show_type_name", ""),
                "answer_type": question_item.get("answer_type", ""),
                "score_rule": question_item.get("score_rule", ""),
                "material_text": question_item.get("material_text", ""),
                "sort_son": question_item.get("sort_son", 0),
                "analyze": question_item.get("analyze", ""),
                "answer": question_item.get("answer", []),
                "option": question_item.get("option", []),
                "new_parent_id": question_item.get("new_parent_id"),
            }
            opts_raw = normalize_option_list(question_item.get("option"))
            placeholder, opts_inlined = self._inline_images_in_question(placeholder, opts_raw)
            self.db.upsert_question(
                question=placeholder,
                answer_type=question_item.get("answer_type"),
                last_sync_batch_id=questions_log_id,
            )
            self.db.replace_question_options(
                question_id=question_id,
                options=opts_inlined,
                answers=normalize_answer_list(question_item.get("answer")),
            )
            self.db.upsert_chapter_question(
                subject_code=ctx.subject_code,
                chapter_id=chapter_id,
                section_chapter_id=section_id,
                question_item=question_item,
                chapter_num=chapter_num,
                section_num=section_num,
                last_sync_batch_id=questions_log_id,
            )
            out_questions.append(question_item)

        print(f"[{now_str()}] 小节 {section_id} - {section_name} 已同步 {len(questions)} 题")

    def _sync_chapter_by_assembly(
        self,
        ctx: SyncContext,
        top_chapter: dict[str, Any],
        chapter_id: int,
        total_num: int,
        out_questions: list[dict[str, Any]],
    ) -> None:
        """无子节时使用 assembly 流程获取整章题目。"""
        progress_resp, _ = self._call_with_log(
            ctx=ctx,
            sync_type="chapter_has_progress",
            chapter_id=chapter_id,
            question_id=None,
            method="GET",
            path="/napi/chapter/has-progress",
            request_payload={"chapter_id": chapter_id},
            func=lambda: self.client.has_progress(chapter_id),
        )
        assembly_id = safe_int(progress_resp.get("data", {}).get("assembly_id"))

        if not assembly_id:
            assembly_resp, _ = self._call_with_log(
                ctx=ctx,
                sync_type="chapter_assembly",
                chapter_id=chapter_id,
                question_id=None,
                method="POST",
                path="/napi/chapter/assembly",
                request_payload={"num": total_num, "chapter_id": str(chapter_id)},
                func=lambda: self.client.create_assembly(chapter_id, total_num),
            )
            assembly_id = safe_int(assembly_resp.get("data", {}).get("assembly_id"))

        if not assembly_id:
            raise RuntimeError(f"章节 {chapter_id} 未获取到 assembly_id")

        questions_resp, questions_log_id = self._call_with_log(
            ctx=ctx,
            sync_type="chapter_questions",
            chapter_id=chapter_id,
            question_id=None,
            method="GET",
            path="/napi/chapter/questions",
            request_payload={"assembly_id": assembly_id},
            func=lambda: self.client.chapter_questions(assembly_id),
        )

        question_data = questions_resp.get("data", {}).get("data", {})
        questions = question_data.get("question", []) or []
        chapter_section_num = question_data.get("chapter_section_num", {}) or {}
        chapter_num = safe_int(chapter_section_num.get("chapter_num"))
        section_num = safe_int(chapter_section_num.get("section_num"))

        for question_item in questions:
            question_id = safe_int(question_item.get("question_id"))
            if not question_id:
                continue

            placeholder = {
                "id": question_id,
                "question_id": question_id,
                "title": question_item.get("question_title", ""),
                "question_type": question_item.get("question_type", ""),
                "show_type_name": question_item.get("show_type_name", ""),
                "answer_type": question_item.get("answer_type", ""),
                "score_rule": question_item.get("score_rule", ""),
                "material_text": question_item.get("material_text", ""),
                "sort_son": question_item.get("sort_son", 0),
                "analyze": question_item.get("analyze", ""),
                "answer": question_item.get("answer", []),
                "option": question_item.get("option", []),
                "new_parent_id": question_item.get("new_parent_id"),
            }
            opts_raw = normalize_option_list(question_item.get("option"))
            placeholder, opts_inlined = self._inline_images_in_question(placeholder, opts_raw)
            self.db.upsert_question(
                question=placeholder,
                answer_type=question_item.get("answer_type"),
                last_sync_batch_id=questions_log_id,
            )
            self.db.replace_question_options(
                question_id=question_id,
                options=opts_inlined,
                answers=normalize_answer_list(question_item.get("answer")),
            )
            question_index = safe_int(question_item.get("index")) or 0
            section_chapter_id = resolve_section_chapter_id(top_chapter, question_index)
            self.db.upsert_chapter_question(
                subject_code=ctx.subject_code,
                chapter_id=chapter_id,
                section_chapter_id=section_chapter_id,
                question_item=question_item,
                chapter_num=chapter_num,
                section_num=section_num,
                last_sync_batch_id=questions_log_id,
            )
            out_questions.append(question_item)

    def sync_question_detail(self, ctx: SyncContext, chapter_id: int, question_id: int) -> None:
        detail_resp, detail_log_id = self._call_with_log(
            ctx=ctx,
            sync_type="question_detail",
            chapter_id=chapter_id,
            question_id=question_id,
            method="GET",
            path="/napi/exam/get-question",
            request_payload={"id": question_id},
            func=lambda: self.client.get_question(question_id),
        )
        detail = detail_resp.get("data", {})
        options_raw = normalize_option_list(detail.get("option"))
        detail, options_inlined = self._inline_images_in_question(detail, options_raw)

        self.db.upsert_question(
            question=detail,
            answer_type=detail.get("answer_type"),
            last_sync_batch_id=detail_log_id,
        )
        self.db.replace_question_options(
            question_id=question_id,
            options=options_inlined,
            answers=normalize_answer_list(detail.get("answer")),
        )
        self.db.replace_question_videos(question_id, detail.get("video_info") or [])
        self.db.commit()

        print(f"[{now_str()}] 题目详情已同步，question_id={question_id}")


def main() -> int:
    args = parse_args()
    config = load_config(args.config)

    db = Database(config["database"])
    client = RuankaoClient(config["api"])
    service = SyncService(db, client)

    try:
        skip_detail = args.skip_question_detail or args.sync_chapters
        service.sync_all(args.chapter_id, skip_detail, args.force)
        return 0
    except KeyboardInterrupt:
        print(f"[{now_str()}] 用户中断", file=sys.stderr)
        db.rollback()
        return 130
    except Exception as exc:
        print(f"[{now_str()}] 同步失败: {exc}", file=sys.stderr)
        db.rollback()
        return 1
    finally:
        db.close()


if __name__ == "__main__":
    raise SystemExit(main())
