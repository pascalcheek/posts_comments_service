-- Таблица постов
CREATE TABLE IF NOT EXISTS posts (
                                     id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title TEXT NOT NULL,
    content TEXT NOT NULL,
    author TEXT NOT NULL,
    allow_comments BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW()
    );

-- Таблица комментариев
CREATE TABLE IF NOT EXISTS comments (
                                        id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    post_id UUID REFERENCES posts(id) ON DELETE CASCADE,
    parent_id UUID REFERENCES comments(id) ON DELETE CASCADE,
    author TEXT NOT NULL,
    text TEXT NOT NULL CHECK (char_length(text) <= 2000),
    created_at TIMESTAMPTZ DEFAULT NOW()
    );

-- Индексы для ускорения выборки
CREATE INDEX IF NOT EXISTS idx_comments_post_id ON comments(post_id);
CREATE INDEX IF NOT EXISTS idx_comments_parent_id ON comments(parent_id);
