package model

import _ "embed"

//go:embed redis_scripts/Attentions.lua
var luaAttentions string

//go:embed redis_scripts/AttentionsZset.lua
var luaAttentionsZset string

//go:embed redis_scripts/Followers.lua
var luaFollowers string

//go:embed redis_scripts/FollowersZset.lua
var luaFollowersZset string


