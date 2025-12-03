DROP TABLE IF EXISTS `food`;
CREATE TABLE `food`  (
     `id` bigint(0) NOT NULL AUTO_INCREMENT,
     `user_id` bigint(0) NOT NULL COMMENT '用户id',
     `food_name` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '美食名称',
     `food_des` text CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '美食描述',
     `food_cate_tag` tinyint(1) NOT NULL DEFAULT 1 COMMENT '美食分类 1、私房菜  2、家常菜  3、硬菜 4、小吃 5、开胃菜 6、甜点 7、减脂餐 8、素菜 9、养生菜',
     `food_url` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '美食主图链接',
     `food_video_url` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '美食视频链接',
     `food_time` tinyint(1) NOT NULL DEFAULT 0 COMMENT '美食制作时长  1、10分钟以内  2、10-30分钟  3、30-60分钟  4、1小时以上',
     `food_difficulty` tinyint(1) NOT NULL DEFAULT 0 COMMENT '美食制作难度 1、简单  2、一般  3、较难  4、极难',
     `food_detail` json NOT NULL COMMENT '美食制作详情 [{\"step\":\"步骤\", \"content\":\"步骤内容\", \"img\":\"步骤图片\"}]',
     `food_list` json NOT NULL COMMENT '美食清单',
     `food_status` tinyint(1) NOT NULL DEFAULT 0 COMMENT '美食状态 0、有效 1、删除',
     `food_skuIds` json NULL COMMENT '美食关联的商品ids',
     `food_createtime` datetime(0) NULL DEFAULT CURRENT_TIMESTAMP(0) COMMENT '创建时间',
     `food_updatetime` datetime(0) NULL DEFAULT CURRENT_TIMESTAMP(0) ON UPDATE CURRENT_TIMESTAMP(0) COMMENT '更新时间',
     PRIMARY KEY (`id`) USING BTREE,
     INDEX `uid`(`user_id`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 64 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT = '美食基础信息表' ROW_FORMAT = Dynamic;