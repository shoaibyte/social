# New tests for coderabbit workflow
# ruff: noqa: S101
# Framework: pytest (preferred) falling back to unittest if necessary
import os
import json
import typing as t

import pytest

# Attempt to import the subject under test from common locations.
# Adjust the import path as needed for your repo layout.
try:
    from coderabbit_workflow import (
        build_review_plan,
        parse_change_scope,
        should_skip_file,
        normalize_path,
        ReviewPlan,
        ChangeScope,
    )
except ImportError:
    try:
        from src.coderabbit_workflow import (
            build_review_plan,
            parse_change_scope,
            should_skip_file,
            normalize_path,
            ReviewPlan,
            ChangeScope,
        )
    except ImportError:
        # Minimal fallbacks to allow import errors to surface in dedicated tests
        build_review_plan = None
        parse_change_scope = None
        should_skip_file = None
        normalize_path = None
        ReviewPlan = t.Any
        ChangeScope = t.Any


@pytest.mark.describe("normalize_path")
class TestNormalizePath:
    def test_normalize_basic_relative(self):
        assert normalize_path("./a/b/../c.py") == "a/c.py"

    def test_normalize_collapses_multiple_separators(self):
        assert normalize_path("a//b///c.py") == "a/b/c.py"

    def test_normalize_handles_dot_only(self):
        assert normalize_path(".") == ""

    def test_normalize_preserves_leading_dirs_no_trailing_slash(self):
        assert normalize_path("src/") == "src"

    def test_normalize_windows_backslashes_and_drive_letters(self):
        # Backslashes converted; drive letters stripped or normalized per implementation
        out = normalize_path(r"C:\repo\src\\module\..\utils.py")
        assert out.endswith("src/utils.py")
        assert " " not in out


@pytest.mark.describe("should_skip_file")
class TestShouldSkipFile:
    @pytest.mark.parametrize(
        "path,patterns,expected",
        [
            ("README.md", ["*.md"], True),
            ("docs/guide/intro.md", ["docs/**"], True),
            ("src/app.py", ["tests/**", "*.md"], False),
            ("tests/test_app.py", ["tests/**"], True),
            ("assets/image.png", ["**/*.png"], True),
            ("src/__pycache__/x.pyc", ["**/__pycache__/**", "*.pyc"], True),
        ],
    )
    def test_patterns_glob_and_double_star(self, path, patterns, expected):
        assert should_skip_file(path, patterns) is expected

    def test_empty_patterns_never_skip(self):
        assert should_skip_file("src/app.py", []) is False

    def test_none_patterns_never_skip(self):
        assert should_skip_file("src/app.py", None) is False

    def test_invalid_pattern_is_ignored(self):
        assert should_skip_file("src/app.py", ["[["]) is False


@pytest.mark.describe("parse_change_scope")
class TestParseChangeScope:
    @pytest.mark.parametrize(
        "files,expected",
        [
            (["src/a.py", "src/b.py"], "multiple"),
            (["src/a.py"], "single"),
            ([], "none"),
        ],
    )
    def test_basic_scope_buckets(self, files, expected):
        scope = parse_change_scope(files)
        assert getattr(scope, "bucket", scope) == expected

    def test_recognizes_docs_only_changes(self):
        scope = parse_change_scope(["README.md", "docs/usage.md"])
        assert getattr(scope, "bucket", getattr(scope, "type", "")) in ("docs", "documentation")

    def test_handles_large_file_counts_threshold(self):
        files = [f"src/f{i}.py" for i in range(0, 501)]
        scope = parse_change_scope(files)
        assert getattr(scope, "bucket", getattr(scope, "type", "")) in ("large", "massive", "bulk")


@pytest.mark.describe("build_review_plan")
class TestBuildReviewPlan:
    def test_happy_path_generates_actions(self):
        changed = ["src/app.py", "src/utils/io.py", "tests/test_app.py", "README.md"]
        cfg = {
            "skip": ["tests/**", "**/__pycache__/**"],
            "review_depth": "standard",
            "lint": True,
            "max_files": 500,
        }
        plan = build_review_plan(changed, cfg)
        # Plan should exclude skipped files
        included = getattr(plan, "included_files", [])
        assert "tests/test_app.py" not in included
        assert "src/app.py" in included
        # Includes actionable steps
        steps = getattr(plan, "steps", []) or getattr(plan, "actions", [])
        assert any("lint" in str(s).lower() for s in steps) if cfg["lint"] else True

    def test_respects_max_files_limit(self):
        changed = [f"src/m{i}.py" for i in range(1000)]
        cfg = {"skip": [], "max_files": 50}
        plan = build_review_plan(changed, cfg)
        included = getattr(plan, "included_files", changed)[:60]
        assert len(included) <= 50

    def test_invalid_config_defaults_gracefully(self):
        changed = ["src/app.py"]
        cfg = {"skip": "[[", "review_depth": "unknown", "max_files": -1}
        plan = build_review_plan(changed, cfg)
        assert "src/app" in ",".join(getattr(plan, "included_files", []))

    def test_empty_changes_returns_noop(self):
        plan = build_review_plan([], {"skip": []})
        assert getattr(plan, "included_files", []) == []
        label = getattr(plan, "label", "").lower()
        assert "no changes" in label or "noop" in label

    def test_docs_only_triggers_docs_review_mode(self):
        changed = ["README.md", "docs/guide.md"]
        cfg = {"skip": []}
        plan = build_review_plan(changed, cfg)
        mode = getattr(plan, "mode", getattr(plan, "profile", "")).lower()
        assert "docs" in mode

    def test_nonexistent_functions_are_reported(self):
        # This test ensures that if imports failed above, we surface a clear message.
        if any(x is None for x in (build_review_plan, parse_change_scope, should_skip_file, normalize_path)):
            pytest.skip("Subject under test could not be imported; check module path in tests.")