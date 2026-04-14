-- data/seed.sql

-- ============================================================
-- Top-level categories
-- ============================================================
INSERT INTO categories (id, name, icon, sort_order, parent_id) VALUES
(1,  '基本功',     '🏸', 1, NULL),
(2,  '进攻技术',   '💥', 2, NULL),
(3,  '防守技术',   '🛡', 3, NULL),
(4,  '网前技术',   '🥅', 4, NULL),
(5,  '双打配合',   '👥', 5, NULL),
(6,  '体能训练',   '💪', 6, NULL);

-- ============================================================
-- Subcategories
-- ============================================================

-- 基本功
INSERT INTO categories (id, name, icon, sort_order, parent_id) VALUES
(10, '握拍',       '✋', 1, 1),
(11, '步法',       '👟', 2, 1),
(12, '发球',       '🎯', 3, 1),
(13, '高远球',     '🏸', 4, 1);

-- 进攻技术
INSERT INTO categories (id, name, icon, sort_order, parent_id) VALUES
(20, '杀球',       '💥', 1, 2),
(21, '吊球',       '🪶', 2, 2),
(22, '劈吊',       '⚡', 3, 2);

-- 防守技术
INSERT INTO categories (id, name, icon, sort_order, parent_id) VALUES
(30, '接杀',       '🛡', 1, 3),
(31, '挡网',       '🥅', 2, 3),
(32, '抽挡',       '🔄', 3, 3);

-- 网前技术
INSERT INTO categories (id, name, icon, sort_order, parent_id) VALUES
(40, '搓球',       '🌀', 1, 4),
(41, '勾球',       '↩', 2, 4),
(42, '推球',       '➡', 3, 4),
(43, '扑球',       '⬇', 4, 4);

-- 双打配合
INSERT INTO categories (id, name, icon, sort_order, parent_id) VALUES
(50, '站位',       '📍', 1, 5),
(51, '轮转',       '🔄', 2, 5),
(52, '发接发',     '🎯', 3, 5);

-- 体能训练
INSERT INTO categories (id, name, icon, sort_order, parent_id) VALUES
(60, '力量训练',   '🏋', 1, 6),
(61, '速度与敏捷', '⚡', 2, 6),
(62, '柔韧性',     '🧘', 3, 6);

-- ============================================================
-- Content entries (20 tutorials across categories)
-- ============================================================

INSERT INTO contents (title, summary, source_url, source_platform, author_name, category_id, sort_order) VALUES

-- 握拍 (10)
('正手握拍与反手握拍详解',
 '最基础也最重要的握拍教学，正确握拍是一切技术的起点',
 'https://www.bilibili.com/video/BV1xK4y1E7Qp',
 'bilibili', '杨晨大神', 10, 1),

('握拍常见错误纠正',
 '90%初学者都会犯的握拍错误，看看你中了几个',
 'https://www.xiaohongshu.com/explore/6501a2e3000000001c00ef12',
 'xiaohongshu', '李宇轩羽毛球', 10, 2),

-- 步法 (11)
('羽毛球全场步法教学',
 '前场、中场、后场六个方向步法完整讲解，含慢动作演示',
 'https://www.bilibili.com/video/BV1rT4y1c7bN',
 'bilibili', '惠程俊', 11, 1),

('米字步法专项训练',
 '每天15分钟米字步法练习，快速提升移动能力',
 'https://www.douyin.com/video/7234567890123456789',
 'douyin', '杨晨大神', 11, 2),

-- 发球 (12)
('反手发小球技巧',
 '双打中最常用的发球方式，教你发出又低又短的球',
 'https://www.bilibili.com/video/BV1Wp4y1D7Xa',
 'bilibili', '惠程俊', 12, 1),

('正手发高远球教学',
 '单打必备发球技术，又高又远压到底线',
 'https://www.youtube.com/watch?v=dQw4w9WgXcQ',
 'youtube', '李宇轩羽毛球', 12, 2),

-- 高远球 (13)
('正手高远球完整教学',
 '从握拍到发力，最清晰的正手高远球教程，适合初学者反复观看',
 'https://www.bilibili.com/video/BV1sK411n7Hm',
 'bilibili', '杨晨大神', 13, 1),

('反手高远球三步学会',
 '反手高远球的核心发力技巧，手指发力是关键',
 'https://www.bilibili.com/video/BV1d34y1R7wZ',
 'bilibili', '惠程俊', 13, 2),

-- 杀球 (20)
('正手杀球教学：从引拍到收拍',
 '杀球不是靠蛮力，正确的鞭打动作让你杀球又快又省力',
 'https://www.bilibili.com/video/BV1gZ4y1H7kL',
 'bilibili', '杨晨大神', 20, 1),

('跳杀技术详解',
 '进阶杀球技术，起跳时机和空中姿态是关键',
 'https://www.youtube.com/watch?v=abc123jump',
 'youtube', 'Badminton Insight', 20, 2),

-- 吊球 (21)
('正手吊球与劈吊组合',
 '后场得分利器，学会吊球让你的进攻更有层次感',
 'https://www.bilibili.com/video/BV1mq4y1w7Tc',
 'bilibili', '李宇轩羽毛球', 21, 1),

-- 接杀 (30)
('接杀球防守技巧',
 '面对对手杀球不再慌张，正确的接杀姿势和拍面角度',
 'https://www.bilibili.com/video/BV1Nq4y1K7eP',
 'bilibili', '惠程俊', 30, 1),

('双打接杀挡网训练',
 '双打防守中最实用的接杀技术，化被动为主动',
 'https://www.douyin.com/video/7345678901234567890',
 'douyin', '杨晨大神', 30, 2),

-- 搓球 (40)
('网前搓球技术教学',
 '搓球是网前最基本的技术，手指的细腻控制决定质量',
 'https://www.bilibili.com/video/BV1Fq4y1P7nR',
 'bilibili', '惠程俊', 40, 1),

-- 勾球 (41)
('网前勾对角技巧',
 '出其不意的网前得分手段，关键在于手腕的隐蔽性',
 'https://www.xiaohongshu.com/explore/6602b3f4000000001d01af23',
 'xiaohongshu', '李宇轩羽毛球', 41, 1),

-- 推球 (42)
('网前推球突击教学',
 '抓住对手网前回球质量不高的机会，快速推压后场',
 'https://www.douyin.com/video/7456789012345678901',
 'douyin', '杨晨大神', 42, 1),

-- 站位 (50)
('双打基本站位讲解',
 '前后站位与左右站位的切换时机，双打入门必看',
 'https://www.bilibili.com/video/BV1Aq4y1G7sT',
 'bilibili', '杨晨大神', 50, 1),

-- 轮转 (51)
('双打轮转配合详解',
 '进攻转防守、防守转进攻的轮转要领，打好双打的核心',
 'https://www.youtube.com/watch?v=rotation123',
 'youtube', 'Badminton Insight', 51, 1),

-- 力量训练 (60)
('羽毛球专项力量训练',
 '针对羽毛球运动的手腕、核心、腿部力量训练计划',
 'https://www.bilibili.com/video/BV1Zq4y1L7uV',
 'bilibili', '李宇轩羽毛球', 60, 1),

-- 速度与敏捷 (61)
('羽毛球脚步敏捷性训练',
 '绳梯训练、反应训练等提升场上移动速度的方法',
 'https://www.xiaohongshu.com/explore/6703c4a5000000001e02bf34',
 'xiaohongshu', '惠程俊', 61, 1);
