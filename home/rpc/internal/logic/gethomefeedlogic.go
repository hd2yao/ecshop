package logic

import (
	"context"
	"encoding/json"
	"math/rand"
	"time"

	"github.com/zeromicro/go-zero/core/logx"

	redisPool "github.com/hd2yao/ecshop/common/redis"
	"github.com/hd2yao/ecshop/home/rpc/internal/svc"
	"github.com/hd2yao/ecshop/home/rpc/types/home"
)

var seededRand = rand.New(rand.NewSource(time.Now().UnixNano()))

type GetHomeFeedLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

type recipesListResult struct {
	feedVersion string
	currPage    int
	hasNext     bool
	recipes     []*home.RecipesInfo
}

func NewGetHomeFeedLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetHomeFeedLogic {
	return &GetHomeFeedLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GetHomeFeed 获取首页 feed 食谱列表
func (l *GetHomeFeedLogic) GetHomeFeed(in *home.HomeFeedReq) (*home.HomeFeedResp, error) {
	var (
		result *recipesListResult
		err    error
	)

	if isMatchRecommend(in.UserId) {
		result, err = l.getRecommendRecipesList(in)
	} else {
		result, err = l.getCacheRecipesList(in)
	}

	if err != nil {
		return nil, err
	}

	if result == nil {
		result = &recipesListResult{}
	}

	return &home.HomeFeedResp{
		Code:        0,
		Message:     "success",
		FeedVersion: result.feedVersion,
		CurrPage:    int32(result.currPage),
		HasNextPage: result.hasNext,
		RecipesList: result.recipes,
	}, nil
}

// getRecommendRecipesList 从推荐系统进行获取
func (l *GetHomeFeedLogic) getRecommendRecipesList(in *home.HomeFeedReq) (*recipesListResult, error) {
	ids, hasMore := getRecommendRecipesIds(in.UserId)
	recipes, err := getRecipesFromCache(l.ctx, l.svcCtx.FeedCache, ids)
	if err != nil {
		return nil, err
	}

	return &recipesListResult{
		feedVersion: in.FeedVersion,
		currPage:    int(in.Page),
		hasNext:     hasMore,
		recipes:     recipes,
	}, nil
}

// getCacheRecipesList 从本地缓存进行获取
func (l *GetHomeFeedLogic) getCacheRecipesList(in *home.HomeFeedReq) (*recipesListResult, error) {
	cache := l.svcCtx.FeedCache
	ctx := l.ctx

	feedVersion := in.FeedVersion
	if feedVersion == "" || !isHomeFeedCached(ctx, cache, feedVersion) {
		latest, err := getHomeFeedLatestVersion(ctx, cache)
		if err != nil {
			return nil, err
		}
		feedVersion = latest
	}

	if feedVersion == "" {
		return &recipesListResult{}, nil
	}

	page := normalizePage(in.Page, in.PullRefresh)
	pageSize := normalizePageSize(in.PageSize)

	ids, err := getHomeFeedFromCacheByVersionAndPage(ctx, cache, feedVersion, page, pageSize)
	if err != nil {
		return nil, err
	}

	recipes, err := getRecipesFromCache(ctx, cache, ids)
	if err != nil {
		return nil, err
	}

	total, err := getHomeFeedSizeFromCacheByVersion(ctx, cache, feedVersion)
	if err != nil {
		return nil, err
	}

	hasNext := false
	if total > 0 {
		hasNext = int64((page+1)*pageSize) < total
	}

	return &recipesListResult{
		feedVersion: feedVersion,
		currPage:    page,
		hasNext:     hasNext,
		recipes:     recipes,
	}, nil
}

func normalizePage(page int32, pullRefresh bool) int {
	current := int(page)
	if current < defaultPageStart {
		current = defaultPageStart
	}

	if pullRefresh {
		return pullRefreshPageStart + seededRand.Intn(pullRefreshPageEnd-pullRefreshPageStart+1)
	}

	return current
}

func normalizePageSize(pageSize int32) int {
	if pageSize <= 0 {
		return defaultPageSize
	}
	return int(pageSize)
}

// ===== 推荐相关（简单模拟）=====

// isMatchRecommend 模拟用户是否命中推荐（仅 userId==2 命中）
func isMatchRecommend(userId int64) bool {
	return userId == 2
}

// getRecommendRecipesIds 模拟推荐返回的食谱 ID 列表和是否有下一页
func getRecommendRecipesIds(userId int64) ([]string, bool) {
	if userId == 2 {
		return []string{"1", "2", "3", "4", "5", "10", "20"}, false
	}
	return nil, false
}

// getRecipesFromCache 根据食谱 ID 列表从 Hash 中批量获取食谱详情
func getRecipesFromCache(ctx context.Context, cache *redisPool.RedisCache, recipesIdList []string) ([]*home.RecipesInfo, error) {
	result := make([]*home.RecipesInfo, 0, len(recipesIdList))

	if len(recipesIdList) == 0 {
		return result, nil
	}

	// 批量获取 Hash 字段
	values, err := cache.HMGet(ctx, homeRecipesHashKey, recipesIdList...)
	if err != nil {
		return nil, err
	}

	for _, jsonStr := range values {
		if jsonStr == "" {
			continue
		}
		var info home.RecipesInfo
		if err := json.Unmarshal([]byte(jsonStr), &info); err != nil {
			continue
		}
		result = append(result, &info)
	}

	return result, nil
}

// isHomeFeedCached 检测是否已经缓存过首页 feed 流
func isHomeFeedCached(ctx context.Context, cache *redisPool.RedisCache, feedVersion string) bool {
	if feedVersion == "" {
		return false
	}

	// 判断列表是否存在
	exists, err := cache.Exists(ctx, feedVersion)
	if err != nil {
		return false
	}
	return exists
}

// getHomeFeedLatestVersion 获取首页 feed 流最新版本号
func getHomeFeedLatestVersion(ctx context.Context, cache *redisPool.RedisCache) (string, error) {
	var version string
	if err := cache.Get(ctx, homeFeedLatestVersionField, &version); err != nil {
		return "", err
	}
	return version, nil
}

// getHomeFeedFromCacheByVersionAndPage 从 Redis 分页获取 feed 流缓存
func getHomeFeedFromCacheByVersionAndPage(
	ctx context.Context,
	cache *redisPool.RedisCache,
	feedVersion string,
	page, pageSize int,
) ([]string, error) {
	start := int64(page * pageSize)
	end := int64(start + int64(pageSize) - 1)

	return cache.LRange(ctx, feedVersion, start, end)
}

// getHomeFeedSizeFromCacheByVersion 从 Redis 获取 feed 流缓存大小
func getHomeFeedSizeFromCacheByVersion(
	ctx context.Context,
	cache *redisPool.RedisCache,
	feedVersion string,
) (int64, error) {
	return cache.LLen(ctx, feedVersion)
}
