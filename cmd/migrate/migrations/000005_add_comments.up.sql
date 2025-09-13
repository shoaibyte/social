CREATE TABLE IF NOT EXISTS comments (
    id bigserial PRIMARY KEY,
    post_id bigserial not null,
    user_id bigserial not null,
    content TEXT NOT NULL,
    created_at timestamp(0) with time zone not null default now()
)