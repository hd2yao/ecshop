# 用户关系表，用来维护用户的关注数、粉丝数，后续可能会迁移到计数服务中
CREATE TABLE `user_relation` (
    `id` int NOT NULL AUTO_INCREMENT COMMENT '主键id',
    `user_id` int NOT NULL DEFAULT '0' COMMENT '用户id',
    `attention_count` int NOT NULL DEFAULT '0' COMMENT '关注数',
    `follower_count` int NOT NULL DEFAULT '0' COMMENT '粉丝数',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uid` (`user_id`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='用户关系表';

#用户关注表 用来实现关注列表
CREATE TABLE `user_attention` (
    `id` int NOT NULL AUTO_INCREMENT COMMENT '主键id',
    `user_id` int NOT NULL COMMENT '当前博主用户id',
    `attention_id` int NOT NULL COMMENT '关注博主用户id',
    `create_time` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '创建设计，会按这个字段排序',
    `is_del` tinyint(1) NOT NULL DEFAULT '0' COMMENT '是否删除，0 正常 1 删除',
    PRIMARY KEY (`id`),
    KEY `uid` (`user_id`,`create_time`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='用户关注表';

# 用户粉丝表，用来实现粉丝列表
CREATE TABLE `user_follower` (
    `id` int NOT NULL AUTO_INCREMENT COMMENT '主键id',
    `user_id` int NOT NULL COMMENT '博主id',
    `follower_id` int NOT NULL COMMENT '粉丝id',
    `create_time` datetime DEFAULT NULL COMMENT '创建时间，按时间排序',
    `is_del` tinyint(1) NOT NULL DEFAULT '0' COMMENT '是否删除 0 正常 1删除',
    PRIMARY KEY (`id`),
    KEY `uid` (`user_id`,`create_time`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='粉丝关系表';