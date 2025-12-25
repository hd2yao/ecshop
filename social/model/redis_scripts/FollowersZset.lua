-- 在脚本开头调用 redis.replicate_commands() 切换为单个命令的复制模式
redis.replicate_commands()
-- 从 Redis 中获取粉丝用户 ID
local followerId = ARGV[1]
-- 获取当前时间戳作为关注/取关事件触发的时间
local timestamp = redis.call('TIME')[1]
-- 粉丝用户 ID 存在，将其添加到 sorted set 集合中
redis.call('ZADD', KEYS[1], timestamp, followerId)

-- 判断集合长度是否大于等于 10000
local zsetsize = redis.call('ZCARD', KEYS[1])
if zsetsize >= 10000 then
    -- 集合长度超过 10000，从右侧弹出一个元素
    redis.call('ZPOPMIN', KEYS[1])
end

-- 返回列表当前长度
return zsetsize


