package logic

// 缓存 Key 定义
const (
	homeRecipesHashKey         = "recipes_info"
	homeFeedLatestVersionField = "latest_version"
)

// 分页 & 下拉刷新
const (
	defaultPageStart     = 0
	defaultPageSize      = 20
	pullRefreshPageStart = 10
	pullRefreshPageEnd   = 50
)
