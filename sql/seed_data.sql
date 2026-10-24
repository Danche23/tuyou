-- ==========================================
-- tuyou 测试数据 SQL（Navicat 中直接执行）
-- 生成时间: 2026-07-25
-- 密码统一: 123456（bcrypt 加密）
--
-- ⚠️ 仅供本地开发 / 测试使用
--    本脚本开头会 TRUNCATE 全部业务表，请勿在已有数据的环境
--    （尤其是生产环境）执行。
--    所有手机号为 138000000xx 测试号段，非真实号码。
-- ==========================================

SET FOREIGN_KEY_CHECKS = 0;

-- ==========================================
-- 0. 清空所有表（与外键无关，直接 TRUNCATE）
-- ==========================================
TRUNCATE TABLE comment_like;
TRUNCATE TABLE comment;
TRUNCATE TABLE follow;
TRUNCATE TABLE msg_user_record;
TRUNCATE TABLE msg_info;
TRUNCATE TABLE msg_template;
TRUNCATE TABLE activity_join;
TRUNCATE TABLE activity;
TRUNCATE TABLE strategy_collect;
TRUNCATE TABLE strategy_like;
TRUNCATE TABLE strategy_tag;
TRUNCATE TABLE strategy_image;
TRUNCATE TABLE strategy;
TRUNCATE TABLE note_collect;
TRUNCATE TABLE note_like;
TRUNCATE TABLE note_tag;
TRUNCATE TABLE note_image;
TRUNCATE TABLE note;
TRUNCATE TABLE tag;
TRUNCATE TABLE sys_role_perm;
TRUNCATE TABLE sys_role_menu;
TRUNCATE TABLE sys_permission;
TRUNCATE TABLE sys_menu;
TRUNCATE TABLE sys_admin;
TRUNCATE TABLE sys_role;
TRUNCATE TABLE member;

-- ==========================================
-- 1. member（5 人：2 领队 + 3 普通会员）
-- 密码: 123456  →  bcrypt hash
-- ==========================================
INSERT INTO member (id, create_time, update_time, delete_time, avatar, nickname, phone, password, bio, bg_img, member_type, status, last_login_time, follow_count, favorite_count, publish_count, fans_count, leading_count)
VALUES
(1, NOW(), NOW(), NULL, 'https://picsum.photos/200/200?random=1', '山鹰队长',   '13800000001', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', '十年户外领队经验，带你走遍大好河山',        '', 0, 1, NOW(), 128, 45, 32, 256, 18),
(2, NOW(), NOW(), NULL, 'https://picsum.photos/200/200?random=2', '行走的骆驼', '13800000002', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', '热爱徒步和露营',                            '', 1, 1, NOW(), 56,  30, 15, 89,  NULL),
(3, NOW(), NOW(), NULL, 'https://picsum.photos/200/200?random=3', '背包客小张', '13800000003', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', '世界那么大，我想去看看',                      '', 1, 1, NOW(), 34,  22, 8,  42,  NULL),
(4, NOW(), NOW(), NULL, 'https://picsum.photos/200/200?random=4', '极星领队',   '13800000004', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', '专业高山向导，登顶过四座5000米雪山',          '', 0, 1, NOW(), 200, 60, 45, 512, 35),
(5, NOW(), NOW(), NULL, 'https://picsum.photos/200/200?random=5', '小鱼爱旅行', '13800000005', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', '分享旅途中的美好瞬间',                        '', 1, 1, NOW(), 18,  15, 6,  25,  NULL);

-- ==========================================
-- 2. sys_role（3 个角色）
-- ==========================================
INSERT INTO sys_role (id, create_time, update_time, delete_time, role_name, role_code, description, status)
VALUES
(1, NOW(), NOW(), NULL, '超级管理员', 'super_admin', '拥有系统全部权限', 1),
(2, NOW(), NOW(), NULL, '内容管理员', 'content_admin', '管理笔记、攻略、活动等内容', 1),
(3, NOW(), NOW(), NULL, '普通管理员', 'normal_admin', '查看数据，无修改权限', 1);

-- ==========================================
-- 3. sys_admin（1 个后台管理员）  密码: admin123
-- ==========================================
INSERT INTO sys_admin (id, create_time, update_time, delete_time, username, password, admin_name, role_id, avatar, status, last_login_time)
VALUES
(1, NOW(), NOW(), NULL, 'admin', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', '系统管理员', 1, '', 1, NOW());

-- ==========================================
-- 4. sys_menu（7 条：一级目录 + 二级菜单）
-- ==========================================
INSERT INTO sys_menu (id, create_time, update_time, delete_time, menu_name, parent_id, menu_type, route_path, sort, status)
VALUES
(1, NOW(), NOW(), NULL, '用户管理', 0, 1, '/user',       1, 1),
(2, NOW(), NOW(), NULL, '会员列表', 1, 2, '/user/member', 1, 1),
(3, NOW(), NOW(), NULL, '领队管理', 1, 2, '/user/leader', 2, 1),
(4, NOW(), NOW(), NULL, '内容管理', 0, 1, '/content',      2, 1),
(5, NOW(), NOW(), NULL, '笔记列表', 4, 2, '/content/note',     1, 1),
(6, NOW(), NOW(), NULL, '攻略管理', 4, 2, '/content/strategy', 2, 1),
(7, NOW(), NOW(), NULL, '活动管理', 4, 2, '/content/activity', 3, 1);

-- ==========================================
-- 5. sys_permission（9 条权限）
-- ==========================================
INSERT INTO sys_permission (id, create_time, update_time, delete_time, perm_name, module, perm_key, request_method, api_path, status)
VALUES
(1, NOW(), NOW(), NULL, '查询会员',     'member',   'member:list',   'GET',    '/api/v1/admin/members',       1),
(2, NOW(), NOW(), NULL, '编辑会员',     'member',   'member:edit',   'PUT',    '/api/v1/admin/members/:id',   1),
(3, NOW(), NOW(), NULL, '删除会员',     'member',   'member:delete', 'DELETE', '/api/v1/admin/members/:id',   1),
(4, NOW(), NOW(), NULL, '查询笔记',     'note',     'note:list',     'GET',    '/api/v1/admin/notes',         1),
(5, NOW(), NOW(), NULL, '编辑笔记',     'note',     'note:edit',     'PUT',    '/api/v1/admin/notes/:id',     1),
(6, NOW(), NOW(), NULL, '删除笔记',     'note',     'note:delete',   'DELETE', '/api/v1/admin/notes/:id',     1),
(7, NOW(), NOW(), NULL, '查询攻略',     'strategy', 'strategy:list', 'GET',    '/api/v1/admin/strategies',    1),
(8, NOW(), NOW(), NULL, '查询活动',     'activity', 'activity:list', 'GET',    '/api/v1/admin/activities',    1),
(9, NOW(), NOW(), NULL, '编辑活动',     'activity', 'activity:edit', 'PUT',    '/api/v1/admin/activities/:id',1);

-- ==========================================
-- 6. sys_role_menu（超级管理员 → 全部菜单）
-- ==========================================
INSERT INTO sys_role_menu (id, create_time, update_time, delete_time, role_id, menu_id)
VALUES
(1, NOW(), NOW(), NULL, 1, 1), (2, NOW(), NOW(), NULL, 1, 2), (3, NOW(), NOW(), NULL, 1, 3),
(4, NOW(), NOW(), NULL, 1, 4), (5, NOW(), NOW(), NULL, 1, 5), (6, NOW(), NOW(), NULL, 1, 6),
(7, NOW(), NOW(), NULL, 1, 7);

-- ==========================================
-- 7. sys_role_perm（超级管理员 → 全部权限）
-- ==========================================
INSERT INTO sys_role_perm (id, create_time, update_time, delete_time, role_id, perm_id)
VALUES
(1, NOW(), NOW(), NULL, 1, 1), (2, NOW(), NOW(), NULL, 1, 2), (3, NOW(), NOW(), NULL, 1, 3),
(4, NOW(), NOW(), NULL, 1, 4), (5, NOW(), NOW(), NULL, 1, 5), (6, NOW(), NOW(), NULL, 1, 6),
(7, NOW(), NOW(), NULL, 1, 7), (8, NOW(), NOW(), NULL, 1, 8), (9, NOW(), NOW(), NULL, 1, 9);

-- ==========================================
-- 8. tag（8 个通用标签）
-- ==========================================
INSERT INTO tag (id, create_time, update_time, delete_time, tag_name, status)
VALUES
(1, NOW(), NOW(), NULL, '登山', 1),
(2, NOW(), NOW(), NULL, '露营', 1),
(3, NOW(), NOW(), NULL, '徒步', 1),
(4, NOW(), NOW(), NULL, '骑行', 1),
(5, NOW(), NOW(), NULL, '摄影', 1),
(6, NOW(), NOW(), NULL, '美食', 1),
(7, NOW(), NOW(), NULL, '自驾', 1),
(8, NOW(), NOW(), NULL, '亲子', 1);

-- ==========================================
-- 9. note（5 篇游记笔记）
-- ==========================================
INSERT INTO note (id, create_time, update_time, delete_time, member_id, title, video_url, content, view_count, like_count, collect_count, comment_count, status, publish_time, is_slide, category, heat)
VALUES
(1, NOW(), NOW(), NULL, 1, '四姑娘山大峰登顶记：三天两夜的雪山之旅',
    '',
    '## Day1 成都出发\n凌晨5点从成都出发，经映秀、卧龙，翻越巴朗山垭口...\n\n## Day2 冲顶日\n凌晨2点起床，3点出发。头灯在黑暗中连成一条光带...\n\n## Day3 下撤\n站在5025米的顶峰，云海在脚下翻涌，所有的疲惫都值了。',
    1520, 86, 32, 12, 1, '2026-07-10 08:30:00', 1, 1, 95),
(2, NOW(), NOW(), NULL, 2, '牛背山露营全攻略：云海日出星空一次集齐',
    'https://example.com/videos/niubei.mp4',
    '牛背山绝对是川西最值得去的露营地之一。\n\n**交通**：成都出发走雅康高速，约4.5小时到冷碛镇，再换乘当地越野车上山。\n\n**最佳时间**：4-6月、10-11月是观赏云海的最佳季节。\n\n**装备清单**：帐篷（三季帐即可）、-5℃睡袋、防潮垫、头灯、炉头...',
    2340, 128, 56, 18, 1, '2026-07-15 14:20:00', 1, 2, 88),
(3, NOW(), NOW(), NULL, 3, '川西小环线骑行日记：挑战与美景并存',
    '',
    '从康定出发，翻越折多山、高尔寺山、卡子拉山三座4000米以上垭口。\n\nDay1 康定→折多塘 15km 爬升800m\nDay2 折多塘→新都桥 45km 翻折多山垭口4298m\nDay3 新都桥→相克宗 85km 翻高尔寺山4412m\n\n全程高反是最大挑战，但318国道沿途的风景让一切变得值得。',
    980, 65, 28, 8, 1, '2026-07-18 10:00:00', 0, 3, 72),
(4, NOW(), NOW(), NULL, 5, '杭州西湖徒步：城市中的诗意山水',
    '',
    '周末来一场说走就走的西湖徒步。\n\n**路线**：断桥→白堤→孤山→苏堤→花港观鱼→雷峰塔，全程约8km，3小时。\n\n三月烟雨中的西湖最美，柳树刚发芽，桃花也开了，一路走一路拍照。湖边的茶馆值得一坐，龙井配西湖，绝了。',
    620, 38, 15, 5, 1, '2026-07-20 16:45:00', 0, 4, 55),
(5, NOW(), NOW(), NULL, 1, '贡嘎西南坡徒步：近距离仰望蜀山之王',
    'https://example.com/videos/gongga.mp4',
    '贡嘎，海拔7556米，蜀山之王。西南坡是最经典的接近路线。\n\n**行程**（5天）：\nD1 成都→康定→老榆林\nD2 老榆林→格西草原→两岔河\nD3 两岔河→日乌且垭口→莫溪沟\nD4 莫溪沟→贡嘎寺\nD5 贡嘎寺→子梅垭口→康定\n\n子梅垭口是看贡嘎日照金山的最佳位置，一定要赶在日落前到达！',
    3180, 210, 89, 24, 1, '2026-07-22 09:15:00', 1, 1, 98);

-- ==========================================
-- 10. note_image（每篇笔记 2 张图）
-- ==========================================
INSERT INTO note_image (id, create_time, update_time, delete_time, note_id, url, sort)
VALUES
(1,  NOW(), NOW(), NULL, 1, 'https://picsum.photos/800/600?random=10', 0),
(2,  NOW(), NOW(), NULL, 1, 'https://picsum.photos/800/600?random=11', 1),
(3,  NOW(), NOW(), NULL, 2, 'https://picsum.photos/800/600?random=12', 0),
(4,  NOW(), NOW(), NULL, 2, 'https://picsum.photos/800/600?random=13', 1),
(5,  NOW(), NOW(), NULL, 3, 'https://picsum.photos/800/600?random=14', 0),
(6,  NOW(), NOW(), NULL, 3, 'https://picsum.photos/800/600?random=15', 1),
(7,  NOW(), NOW(), NULL, 4, 'https://picsum.photos/800/600?random=16', 0),
(8,  NOW(), NOW(), NULL, 4, 'https://picsum.photos/800/600?random=17', 1),
(9,  NOW(), NOW(), NULL, 5, 'https://picsum.photos/800/600?random=18', 0),
(10, NOW(), NOW(), NULL, 5, 'https://picsum.photos/800/600?random=19', 1);

-- ==========================================
-- 11. note_tag（每篇笔记 2 个标签）
-- ==========================================
INSERT INTO note_tag (id, create_time, update_time, delete_time, note_id, tag_id)
VALUES
(1, NOW(), NOW(), NULL, 1, 1), (2, NOW(), NOW(), NULL, 1, 5),
(3, NOW(), NOW(), NULL, 2, 2), (4, NOW(), NOW(), NULL, 2, 5),
(5, NOW(), NOW(), NULL, 3, 4), (6, NOW(), NOW(), NULL, 3, 3),
(7, NOW(), NOW(), NULL, 4, 3), (8, NOW(), NOW(), NULL, 4, 5),
(9, NOW(), NOW(), NULL, 5, 1), (10, NOW(), NOW(), NULL, 5, 3);

-- ==========================================
-- 12. note_like（8 条点赞）
-- ==========================================
INSERT INTO note_like (id, create_time, update_time, delete_time, note_id, member_id)
VALUES
(1, NOW(), NOW(), NULL, 1, 2), (2, NOW(), NOW(), NULL, 1, 3),
(3, NOW(), NOW(), NULL, 2, 1), (4, NOW(), NOW(), NULL, 2, 3),
(5, NOW(), NOW(), NULL, 3, 1), (6, NOW(), NOW(), NULL, 3, 5),
(7, NOW(), NOW(), NULL, 4, 1), (8, NOW(), NOW(), NULL, 5, 2);

-- ==========================================
-- 13. note_collect（5 条收藏）
-- ==========================================
INSERT INTO note_collect (id, create_time, update_time, delete_time, note_id, member_id)
VALUES
(1, NOW(), NOW(), NULL, 1, 3), (2, NOW(), NOW(), NULL, 2, 2),
(3, NOW(), NOW(), NULL, 2, 5), (4, NOW(), NOW(), NULL, 5, 3),
(5, NOW(), NOW(), NULL, 5, 2);

-- ==========================================
-- 14. strategy（3 篇攻略）
-- ==========================================
INSERT INTO strategy (id, create_time, update_time, delete_time, member_id, title, video_url, content, view_count, like_count, collect_count, comment_count, status, publish_time, is_slide, category, heat, distance, duration, difficulty, transport)
VALUES
(1, NOW(), NOW(), NULL, 4, '川藏线自驾终极攻略：318国道15天全记录',
    '',
    '## 总览\n川藏南线（318国道）是中国最美的景观大道，全长约2140公里。\n\n## 详细行程\n| 天数 | 路线 | 里程 | 亮点 |\n|------|------|------|------|\n| D1 | 成都→雅安→康定 | 290km | 二郎山隧道 |\n| D2 | 康定→折多山→新都桥 | 75km | 摄影天堂 |\n...\n\n## 注意事项\n1. 提前办理边防证\n2. 备好高反药品\n3. 每天检查车辆刹车和轮胎',
    5600, 320, 145, 35, 1, '2026-06-01 10:00:00', 1, 1, 99, '2100公里', '15天', 3, 3),
(2, NOW(), NOW(), NULL, 1, '雨崩徒步全攻略：梅里雪山脚下的世外桃源',
    '',
    '雨崩村，藏在梅里雪山脚下的小村庄，不通公路，只能徒步或骑骡子进入。\n\n## 路线\n- **进村**：西当→南宗垭口→上雨崩，约6小时\n- **冰湖线**：上雨崩→笑农大本营→冰湖，往返8小时\n- **神瀑线**：下雨崩→神瀑，往返5小时\n\n## 最佳季节\n5-6月：杜鹃花开\n10-11月：秋色最美，看到日照金山概率最高\n\n## 住宿\n上下雨崩村都有客栈，条件简朴但干净，旺季需提前一周预订。',
    4200, 256, 98, 28, 1, '2026-05-20 14:00:00', 1, 2, 92, '45公里', '4天', 2, 1),
(3, NOW(), NOW(), NULL, 2, '周末京郊骑行路线推荐：妙峰山经典爬坡线',
    'https://example.com/videos/miaofeng.mp4',
    '妙峰山是北京骑友的圣地，20公里连续爬坡，海拔爬升900米。\n\n## 路线详情\n起点：门头沟妙峰山牌楼\n终点：妙峰山金顶\n距离：单程20km（纯爬坡）\n平均坡度：4.5%\n\n## 分段策略\n1. 前7km（牌楼→樱桃沟）：坡度较缓，热身段\n2. 中7km（樱桃沟→涧沟村）：最陡段，部分超8%\n3. 后6km（涧沟村→金顶）：缓坡+平路段\n\n## Tips\n- 带足水，途中只有一个补水点\n- 下山注意控制速度，弯道多\n- 新手建议骑到涧沟村就折返',
    1800, 95, 42, 12, 1, '2026-06-15 08:00:00', 0, 3, 78, '40公里', '半天', 2, 2);

-- ==========================================
-- 15. strategy_image（每篇攻略 2 张图）
-- ==========================================
INSERT INTO strategy_image (id, create_time, update_time, delete_time, strategy_id, url, sort)
VALUES
(1, NOW(), NOW(), NULL, 1, 'https://picsum.photos/800/600?random=20', 0),
(2, NOW(), NOW(), NULL, 1, 'https://picsum.photos/800/600?random=21', 1),
(3, NOW(), NOW(), NULL, 2, 'https://picsum.photos/800/600?random=22', 0),
(4, NOW(), NOW(), NULL, 2, 'https://picsum.photos/800/600?random=23', 1),
(5, NOW(), NOW(), NULL, 3, 'https://picsum.photos/800/600?random=24', 0),
(6, NOW(), NOW(), NULL, 3, 'https://picsum.photos/800/600?random=25', 1);

-- ==========================================
-- 16. strategy_tag
-- ==========================================
INSERT INTO strategy_tag (id, create_time, update_time, delete_time, strategy_id, tag_id)
VALUES
(1, NOW(), NOW(), NULL, 1, 7), (2, NOW(), NOW(), NULL, 1, 5),
(3, NOW(), NOW(), NULL, 2, 3), (4, NOW(), NOW(), NULL, 2, 1),
(5, NOW(), NOW(), NULL, 3, 4), (6, NOW(), NOW(), NULL, 3, 3);

-- ==========================================
-- 17. strategy_like
-- ==========================================
INSERT INTO strategy_like (id, create_time, update_time, delete_time, strategy_id, member_id)
VALUES
(1, NOW(), NOW(), NULL, 1, 2), (2, NOW(), NOW(), NULL, 1, 3),
(3, NOW(), NOW(), NULL, 2, 2), (4, NOW(), NOW(), NULL, 2, 3),
(5, NOW(), NOW(), NULL, 3, 1);

-- ==========================================
-- 18. strategy_collect
-- ==========================================
INSERT INTO strategy_collect (id, create_time, update_time, delete_time, strategy_id, member_id)
VALUES
(1, NOW(), NOW(), NULL, 1, 3), (2, NOW(), NOW(), NULL, 1, 5),
(3, NOW(), NOW(), NULL, 2, 3), (4, NOW(), NOW(), NULL, 2, 2);

-- ==========================================
-- 19. activity（3 个活动）
-- ==========================================
INSERT INTO activity (id, create_time, update_time, delete_time, leader_member_id, title, cover_img, activity_content, address, start_time, end_time, sign_up_start, sign_up_end, max_people, current_people, price, activity_status, is_del, difficulty, activity_type, is_slide)
VALUES
(1, NOW(), NOW(), NULL, 1, '【周末】四姑娘山大峰二日登顶挑战',
    'https://picsum.photos/800/400?random=30',
    '## 活动介绍\n四姑娘山大峰海拔5025米，是入门级5000米雪山。本次活动由山鹰队长带队，适合有高海拔徒步经验的户外爱好者。\n\n## 行程安排\n**Day1** 成都→四姑娘山镇（3200m），适应海拔，装备检查\n**Day2** 凌晨3点出发→大本营（4300m）→冲顶（5025m）→下撤→返回成都\n\n## 费用包含\n- 成都往返交通\n- 一晚住宿\n- 专业领队+协作\n- 户外保险\n\n## 装备要求\n高帮登山鞋、羽绒服、冲锋衣、头灯、登山杖、雪镜、保暖手套',
    '四川省阿坝州小金县四姑娘山景区',
    '2026-08-15 07:00:00', '2026-08-16 20:00:00',
    '2026-07-25 00:00:00', '2026-08-12 23:59:59',
    20, 8, 880.00, 1, 0, 3, 1, 1),
(2, NOW(), NOW(), NULL, 4, '【国庆】梅里雪山雨崩村徒步7日',
    'https://picsum.photos/800/400?random=31',
    '## 活动介绍\n深入梅里雪山腹地，探访与世隔绝的雨崩村，感受藏区最原始的风景和人文。\n\n## 行程安排（7天）\nD1 丽江集合\nD2 丽江→飞来寺（看梅里日落）\nD3 飞来寺→西当→徒步进雨崩\nD4 冰湖徒步\nD5 神瀑徒步\nD6 雨崩→尼农峡谷→香格里拉\nD7 香格里拉→丽江解散',
    '云南省迪庆州德钦县',
    '2026-10-01 08:00:00', '2026-10-07 18:00:00',
    '2026-08-01 00:00:00', '2026-09-25 23:59:59',
    15, 12, 3280.00, 1, 0, 2, 2, 1),
(3, NOW(), NOW(), NULL, 1, '【平日】杭州西湖群山轻徒步一日',
    'https://picsum.photos/800/400?random=32',
    '## 活动介绍\n休闲级别的城市周边徒步，适合零基础新手，全程约10公里，耗时4-5小时。\n\n## 路线\n龙井村→十里琅珰→五云山→云栖竹径\n\n## 亮点\n- 茶园风光\n- 千年古树\n- 竹林步道\n- 登顶可远眺西湖全景\n\n## 费用\n包含领队费+户外保险，午餐自理（龙井村农家乐AA）',
    '浙江省杭州市西湖区龙井村',
    '2026-08-02 09:00:00', '2026-08-02 15:00:00',
    '2026-07-20 00:00:00', '2026-08-01 18:00:00',
    30, 18, 68.00, 1, 0, 1, 3, 0);

-- ==========================================
-- 20. activity_join（报名记录）
-- ==========================================
INSERT INTO activity_join (id, create_time, update_time, delete_time, activity_id, member_id, is_cancel)
VALUES
(1, NOW(), NOW(), NULL, 1, 2, 0),
(2, NOW(), NOW(), NULL, 1, 3, 0),
(3, NOW(), NOW(), NULL, 1, 5, 1),  -- 取消报名
(4, NOW(), NOW(), NULL, 2, 2, 0),
(5, NOW(), NOW(), NULL, 2, 3, 0),
(6, NOW(), NOW(), NULL, 3, 2, 0),
(7, NOW(), NOW(), NULL, 3, 3, 0),
(8, NOW(), NOW(), NULL, 3, 5, 0);

-- ==========================================
-- 21. comment（8 条评论：笔记 + 攻略）
-- ==========================================
INSERT INTO comment (id, create_time, update_time, delete_time, target_type, target_id, member_id, content, reply_to_id, like_count, status)
VALUES
-- 笔记评论
(1, NOW(), NOW(), NULL, 1, 1, 2, '太厉害了！我一直想去四姑娘山，这篇笔记给了我很大的信心', NULL, 12, 0),
(2, NOW(), NOW(), NULL, 1, 1, 3, '请问大峰对新手友好吗？需要什么装备？', NULL, 5, 0),
(3, NOW(), NOW(), NULL, 1, 1, 1, '@背包客小张 大峰是入门级雪山，有徒步经验就可以尝试，装备清单我补充在文章里了', 2, 8, 0),
-- 攻略评论
(4, NOW(), NOW(), NULL, 2, 1, 3, '川藏线自驾是我的梦想！收藏了', NULL, 15, 0),
(5, NOW(), NOW(), NULL, 2, 1, 2, '15天会不会太赶？每天开车要多久？', NULL, 3, 0),
(6, NOW(), NOW(), NULL, 2, 1, 4, '@行走的骆驼 平均每天4-6小时，不算太赶，沿途有很多观景台可以休息', 5, 6, 0),
-- 更多笔记评论
(7, NOW(), NOW(), NULL, 1, 2, 1, '牛背山的云海确实一绝，我去年也去过', NULL, 7, 0),
(8, NOW(), NOW(), NULL, 1, 5, 3, '贡嘎真的太震撼了，看了你的笔记我决定今年一定要去！', NULL, 10, 0);

-- ==========================================
-- 22. comment_like（5 条评论点赞）
-- ==========================================
INSERT INTO comment_like (id, create_time, update_time, delete_time, comment_id, member_id)
VALUES
(1, NOW(), NOW(), NULL, 1, 1), (2, NOW(), NOW(), NULL, 1, 3),
(3, NOW(), NOW(), NULL, 4, 2), (4, NOW(), NOW(), NULL, 4, 5),
(5, NOW(), NOW(), NULL, 7, 3);

-- ==========================================
-- 23. follow（关注关系）
-- ==========================================
INSERT INTO follow (id, create_time, update_time, delete_time, follower_id, followee_id)
VALUES
(1, NOW(), NOW(), NULL, 2, 1),   -- 行走的骆驼 关注 山鹰队长
(2, NOW(), NOW(), NULL, 3, 1),   -- 背包客小张 关注 山鹰队长
(3, NOW(), NOW(), NULL, 5, 1),   -- 小鱼爱旅行 关注 山鹰队长
(4, NOW(), NOW(), NULL, 2, 4),   -- 行走的骆驼 关注 极星领队
(5, NOW(), NOW(), NULL, 3, 2),   -- 背包客小张 关注 行走的骆驼
(6, NOW(), NOW(), NULL, 1, 4);   -- 山鹰队长 关注 极星领队

-- ==========================================
-- 24. msg_template（15 条消息模板，与 internal/service/web/message/notifier.go 的 templateID 对齐）
--     注意：4、7 号模板历史上未使用；新增模板时不要复用这两个 ID，避免与线上数据冲突。
-- ==========================================
INSERT INTO msg_template (id, create_time, update_time, delete_time, template_name, msg_type, content_template, fixed_tip)
VALUES
(1,  NOW(), NOW(), NULL, '点赞通知',       1, '赞了你的内容《{title}》', ''),
(2,  NOW(), NOW(), NULL, '关注通知',       1, '关注了你', ''),
(3,  NOW(), NOW(), NULL, '评论通知',       1, '评论了你的内容《{title}》', ''),
(5,  NOW(), NOW(), NULL, '系统通知',       2, '{content}', ''),
(6,  NOW(), NOW(), NULL, '报名通知',       2, '报名了你的活动《{title}》', ''),
(8,  NOW(), NOW(), NULL, '报名成功通知',   2, '你已成功报名活动《{title}》,请到活动详情界面加入群聊', ''),
(9,  NOW(), NOW(), NULL, '活动快开始提醒', 2, '距离你报名的活动《{title}》还有48小时', ''),
(10, NOW(), NOW(), NULL, '回复通知',       1, '回复了你的内容 {comment}', ''),
(11, NOW(), NOW(), NULL, '活动删除通知',   2, '删除了活动《{title}》', ''),
(12, NOW(), NOW(), NULL, '收藏通知',       1, '收藏了你的内容《{title}》', ''),
(13, NOW(), NOW(), NULL, '点赞评论通知',   1, '赞了你的评论', ''),
(14, NOW(), NOW(), NULL, '笔记删除通知',   2, '你的笔记《{title}》已被删除', ''),
(15, NOW(), NOW(), NULL, '攻略删除通知',   2, '你的攻略《{title}》已被删除', ''),
(16, NOW(), NOW(), NULL, '笔记下架通知',   2, '你的笔记《{title}》已被下架', ''),
(17, NOW(), NOW(), NULL, '攻略下架通知',   2, '你的攻略《{title}》已被下架', '');

-- ==========================================
-- 25. msg_info（4 条示例消息）
-- ==========================================
INSERT INTO msg_info (id, create_time, update_time, delete_time, msg_type, receive_obj_type, sender_id, template_id, msg_content, send_time, is_del)
VALUES
(1, NOW(), NOW(), NULL, 1, 1, 2, 1, '赞了你的内容《四姑娘山大峰登顶记》', '2026-07-12 10:30:00', 0),
(2, NOW(), NOW(), NULL, 1, 1, 3, 3, '评论了你的内容《四姑娘山大峰登顶记》', '2026-07-12 11:15:00', 0),
(3, NOW(), NOW(), NULL, 1, 1, 2, 2, '关注了你', '2026-07-14 16:40:00', 0),
(4, NOW(), NOW(), NULL, 2, 0, 0, 9, '距离你报名的活动《【周末】四姑娘山大峰二日登顶挑战》还有48小时', '2026-08-13 09:00:00', 0);

-- ==========================================
-- 26. msg_user_record（消息接收记录）
-- ==========================================
INSERT INTO msg_user_record (id, create_time, update_time, delete_time, msg_id, receive_member_id, is_read, is_user_del)
VALUES
(1, NOW(), NOW(), NULL, 1, 1, 1, 0),   -- 山鹰队长 已读
(2, NOW(), NOW(), NULL, 2, 1, 1, 0),   -- 山鹰队长 已读
(3, NOW(), NOW(), NULL, 3, 1, 0, 0),   -- 山鹰队长 未读
(4, NOW(), NOW(), NULL, 4, 2, 0, 0),   -- 行走的骆驼 未读
(5, NOW(), NOW(), NULL, 4, 3, 0, 0),   -- 背包客小张 未读
(6, NOW(), NOW(), NULL, 4, 5, 1, 0);   -- 小鱼爱旅行 已读

SET FOREIGN_KEY_CHECKS = 1;

-- ==========================================
-- 验证：各表行数
-- ==========================================
-- member: 5
-- sys_role: 3
-- sys_admin: 1
-- sys_menu: 7
-- sys_permission: 9
-- sys_role_menu: 7
-- sys_role_perm: 9
-- tag: 8
-- note: 5, note_image: 10, note_tag: 10, note_like: 8, note_collect: 5
-- strategy: 3, strategy_image: 6, strategy_tag: 6, strategy_like: 5, strategy_collect: 4
-- activity: 3, activity_join: 8
-- comment: 8, comment_like: 5
-- follow: 6
-- msg_template: 15, msg_info: 4, msg_user_record: 6
-- ==========================================
