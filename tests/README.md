This repository uses the existing Python testing framework detected at runtime by CI (pytest preferred).
The newly added tests focus on the following functions:
- normalize_path: Path canonicalization edge cases.
- should_skip_file: Glob pattern evaluation, double-star patterns, invalid patterns, and empty inputs.
- parse_change_scope: Classification of change buckets, including docs-only and large changes.
- build_review_plan: Honoring skip patterns and file count limits, and generating sensible actions.

No new dependencies were introduced.