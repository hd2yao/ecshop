-- 从 Redis 中获取粉丝用户 ID
local followerId = ARGV[1]

-- 判断用户 ID 是否存在 KEYS[1] = user_follower:uid
-- local exists = redis.call('EXISTS', KEYS[1])

-- if exists == 0 then
    -- 粉丝用户 ID 不存在，返回0
    -- redis.call('LPUSH', KEYS[1], followerId)
-- else
    -- 粉丝用户 ID 存在，将其添加到列表最左侧
    redis.call('LPUSH', KEYS[1], followerId)

    -- 判断列表长度是否大于等于 1000
    local listsize = redis.call('LLEN', KEYS[1])
    if listsize >= 2000 then
        -- 列表长度超过 2000，从右侧弹出一个元素
        redis.call('RPOP', KEYS[1])
    end

    -- 返回列表当前长度
    return listsize
-- end


