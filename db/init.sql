CREATE TABLE comments (
                          id SERIAL PRIMARY KEY,            -- Уникальный ID комментария
                          parent_id INTEGER REFERENCES comments(id) ON DELETE CASCADE, -- Ссылка на родителя
                          content TEXT NOT NULL,            -- Текст комментария
                          created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP -- Время создания
);

-- Индекс для быстрого поиска детей конкретного родителя
CREATE INDEX idx_comments_parent_id ON comments(parent_id);

-- Индекс для текстового поиска
CREATE INDEX idx_comments_content ON comments USING gin (to_tsvector('russian', content));
