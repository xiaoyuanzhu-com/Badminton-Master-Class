-- data/seed.sql
INSERT INTO categories (id, name, icon, sort_order, parent_id) VALUES
(1, '正手', '🏸', 1, NULL),
(2, '反手', '🏸', 2, NULL),
(3, '杀球', '💥', 3, NULL),
(4, '步法', '👟', 4, NULL),
(5, '发球', '🎯', 5, NULL),
(6, '网前', '🥅', 6, NULL),
(7, '双打', '👥', 7, NULL),
(8, '正手高远球', '🏸', 1, 1),
(9, '正手吊球', '🏸', 2, 1),
(10, '反手高远球', '🏸', 1, 2),
(11, '反手吊球', '🏸', 2, 2);

-- Example content (replace with real curated content)
INSERT INTO contents (title, summary, source_url, source_platform, author_name, category_id, sort_order) VALUES
('正手高远球完整教学', '从握拍到发力，最清晰的正手高远球教程', 'https://www.bilibili.com/video/example1', 'bilibili', '杨晨大神', 8, 1),
('反手高远球三步学会', '反手高远球的核心发力技巧', 'https://www.bilibili.com/video/example2', 'bilibili', '惠程俊', 10, 1);
