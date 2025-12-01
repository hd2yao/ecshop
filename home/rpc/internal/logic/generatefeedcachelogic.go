package logic

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/hd2yao/ecshop/home/rpc/internal/svc"
	"github.com/hd2yao/ecshop/home/rpc/types/home"

	"github.com/zeromicro/go-zero/core/logx"
)

// ===== 缓存 Key 定义 =====

const (
	homeRecipesHashKey         = "recipes_info"
	homeFeedLatestVersionField = "latest_version"
)

type GenerateFeedCacheLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGenerateFeedCacheLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GenerateFeedCacheLogic {
	return &GenerateFeedCacheLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GenerateFeedCacheLogic) GenerateFeedCache(in *home.GenerateFeedCacheReq) (*home.GenerateFeedCacheResp, error) {
	// 1. 构造本地食谱列表（模拟推荐系统返回的食谱）
	recipes := getFeedRecipes()
	if len(recipes) == 0 {
		return &home.GenerateFeedCacheResp{
			Code:    0,
			Message: "no recipes",
		}, nil
	}

	// 2. 生成当前 feed 版本号（毫秒时间戳）
	feedVersion := strconv.FormatInt(time.Now().UnixMilli(), 10)

	ctx := l.ctx
	cache := l.svcCtx.FeedCache

	// 3. 写入首页 feed 流中存储的食谱 ID 集合（List）
	var idStrList []string
	for _, r := range recipes {
		idStrList = append(idStrList, strconv.FormatInt(r.Id, 10))
	}

	// key: home:feed:{version}
	listKey := feedVersion
	if _, err := cache.RPush(ctx, listKey, idStrList...); err != nil {
		l.Errorf("写入 feed 流食谱 ID 列表失败, version=%s, err=%v", feedVersion, err)
		return nil, err
	}

	// 4. 首页食谱信息缓存（Hash，field 为食谱 ID，value 为 JSON）
	for _, r := range recipes {
		data, err := json.Marshal(r)
		if err != nil {
			l.Errorf("序列化食谱信息失败, id=%d, err=%v", r.Id, err)
			return nil, err
		}

		if err := cache.HSet(ctx, homeRecipesHashKey, strconv.FormatInt(r.Id, 10), string(data)); err != nil {
			l.Errorf("写入食谱 Hash 缓存失败, id=%d, err=%v", r.Id, err)
			return nil, err
		}
	}

	// 5. 记录最新版本号（String: latest_version），不设置过期
	if err := cache.Set(ctx, homeFeedLatestVersionField, feedVersion, 0); err != nil {
		l.Errorf("写入最新 feed 版本号失败, version=%s, err=%v", feedVersion, err)
		return nil, err
	}

	return &home.GenerateFeedCacheResp{
		Code:        0,
		Message:     "success",
		FeedVersion: feedVersion,
		RecipesList: recipes,
	}, nil
}

// getFeedRecipes 模拟从推荐系统获取首页 feed 食谱列表
func getFeedRecipes() []*home.RecipesInfo {
	return []*home.RecipesInfo{
		{
			Id:                 1,
			RecipesName:        "新手宝，81道家常菜谱，厨艺速成1-18道。",
			RecipesUrl:         "https://sns-webpic-qc.xhscdn.com/202408300825/9657b8b3bdd1f71570c358a6945cb677/spectrum/1040g34o316r8jttags0g5pm2srcne1bpj81tprg!nd_dft_wlteh_webp_3",
			RecipesUserName:    "Yo小",
			RecipesUserAvatar:  "https://sns-avatar-qc.xhscdn.com/avatar/645b7f6e338379fac1893093.jpg?imageView2/2/w/540/format/webp|imageMogr2/strip2/blur/1x56",
			RecipesDescription: "一周的新手厨艺提升计划，专为零基础的小白量身定制，助你轻松成为厨艺高手！零基础的你，不妨下载这份菜谱，慢慢磨练技艺~同时，也别忘了分享给那些你在乎的人，让TA也能加入烹饪的行列~\n1. 辣味炒肉片（基础技能）\n2. 绿色青椒炒鸡蛋（基础技能）\n3. 乡村风味一锅香（双重美味）4. 炒五花肉配花菜（品质可靠）\n5. 蒜苔炒肉末（风味独特）\n6. 炒肉配笋干（质地上乘）\n7. 鸡蛋拌酱香（酱料配方：2勺酱油1勺香醋1勺蚝油半勺糖1勺辣椒粉1勺生粉半碗水混合均匀）\n8. 清蒸排骨带蒜香（零失误）\n9. 炒蛋配虾仁（操作简单，酱汁：1勺酱油，1勺醋，1勺蚝油，1勺糖，1勺生粉，1勺辣椒粉，半碗水，调匀备用）10. 经典川菜之辣子鸡丁\n11. 传统粤式葱油焖鸡\n12. 肉丸与白菜的炒制\n13. 香菇与滑嫩鸡肉（不辣口味）\n14. 年夜饭必备金钱蛋\n15. 辣椒与油豆腐的炒制（勾芡技巧）\n16. 土豆丝的炒制（基础技巧）\n17. 川菜中人见人爱的包菜回锅肉\n18. 入门至进阶的白菜火腿豆腐煲19.青椒搭配火腿炒制（简便）",
		},
		{
			Id:                 2,
			RecipesName:        "减脂刮油汤合集‼好喝又掉秤，越喝越瘦",
			RecipesUrl:         "https://sns-webpic-qc.xhscdn.com/202408300858/ed8ac378851570ebbd892db54aceb28d/1000g0082ean96ckgm06g4a5n7b1ua4pjfo21vt8!nd_dft_wlteh_webp_3",
			RecipesUserName:    "小毒减脂分享",
			RecipesUserAvatar:  "https://sns-avatar-qc.xhscdn.com/avatar/63e79a43687a4354f16c1233.jpg?imageView2/2/w/540/format/webp|imageMogr2/strip2/blur/1x56",
			RecipesDescription: "减脂期的姐妹们集合啦‼\n今天给大家整理了1⃣2⃣款巨好喝的减脂刮油汤，做法都非常简单，用来代替晚餐真的是绝绝子，掉秤速度杠杠的\n\n五一假期你放纵了吗，这些汤做起来，刮油神器\n\n1⃣海带豆腐汤\n2⃣冬瓜虾仁汤\n3⃣黄瓜鸡蛋汤\n4⃣番茄菌菇汤\n5⃣萝卜煎蛋汤\n6⃣丝瓜三鲜汤\n7⃣冬瓜玉米汤\n8⃣西兰花豆腐汤\n9⃣裙带菜豆芽汤\n🔟低卡酸辣汤\n1⃣1⃣冬瓜花甲汤\n1⃣2⃣巫婆瘦身汤\n详细做法参照笔记中的图片哦\n\n✂️🥣真的很简单，只要你愿意坚持下去，一步一步的完成自己的减脂计划，你一定会成功的",
		},
		{
			Id:                 3,
			RecipesName:        "家庭做饭万能公式🔥新手小白秒变大厨‼️",
			RecipesUrl:         "https://sns-webpic-qc.xhscdn.com/202408300825/9657b8b3bdd1f71570c358a6945cb677/spectrum/1040g34o316r8jttags0g5pm2srcne1bpj81tprg!nd_dft_wlteh_webp_3",
			RecipesUserName:    "木子的慢生活",
			RecipesUserAvatar:  "https://sns-avatar-qc.xhscdn.com/avatar/1040g2jo3156p4qlr0o605ph6839gudijfieog70?imageView2/2/w/540/format/webp|imageMogr2/strip2",
			RecipesDescription: "对于新手小白来说，掌握一些做饭的万能公式和调味方法，能够让烹饪变得更加简单。今天给大家整理一些实用的做饭调味方法，快来一起看看吧！\n·\n今天给大家整理了一期家庭做饭万能公式🔥厨房小知识‼️，赶紧🐎住～\n·\n🔸炒任何素菜：油+盐+蒜末\n🔸炒任何肉菜：油+蒜末+老抽+生抽+耗油+料酒+淀粉\n🔸任何凉拌菜：生抽+白糖+醋+辣椒粉+芝麻+蒜末\n🔸炖肉or红烧：盐+姜片+生抽+老抽+耗油+料酒+干辣椒+桂皮+香叶+八角\n🔸任何糖醋菜：白糖+生抽+醋+料酒+水\n🔸任何煲汤：料酒+姜片+葱段+盐+红枣+枸杞\n🔸任何辣炒类：耗油+生抽+姜片+小米辣+盐+白糖+火锅底料\n🔸麻辣拌：白糖+耗油+生抽+辣椒粉+醋+蒜末+盐+芝麻\n🔸减脂水煮菜蘸汁：蒜末+芝麻+辣椒粉+小米辣+醋+生抽+耗油+水\n🔸火锅万能底料：香油+耗油+生抽+小米辣+醋+蒜末+葱+芝麻",
		},
		{
			Id:                 4,
			RecipesName:        "我敢说，方圆十里都没有我们家的饭菜香！",
			RecipesUrl:         "https://sns-webpic-qc.xhscdn.com/202408300911/5b6fda558c066463a0867346e8e53b66/1040g2sg315m5ojf81a6g5okb66o8d93gd9ua7a8!nd_dft_wlteh_webp_3",
			RecipesUserName:    "自然知食局",
			RecipesUserAvatar:  "https://sns-avatar-qc.xhscdn.com/avatar/1040g2jo31123ck12mg6g5okb66o8d93gvbaeah0?imageView2/2/w/540/format/webp|imageMogr2/strip2",
			RecipesDescription: "1，葱姜蒜猪排\n2，蒜香鱼片\n3，沙茶咖喱酱牛腩\n4，青花椒鸡翅\n5，蒜蓉鸡翅\n#学做菜 #做菜我是认真的 #舌尖上的美食 #简单美食 #做饭 美味且创新的美食，是不是香迷糊了，欢迎关注、点赞和收藏，赶快去动手试试吧",
		},
		{
			Id:                 5,
			RecipesName:        "九款给女朋友做的美食！",
			RecipesUrl:         "https://sns-webpic-qc.xhscdn.com/202408300916/f14a335b77418c66205f47a92ab1d5ab/1000g0082q6ibftejs06g4ajjir0btj45pnas28g!nd_dft_wlteh_webp_3",
			RecipesUserName:    "小羊很爱吃",
			RecipesUserAvatar:  "https://sns-avatar-qc.xhscdn.com/avatar/6447a1e0bdd3d1091432c08d.jpg?imageView2/2/w/540/format/webp|imageMogr2/strip2/blur/1x56",
			RecipesDescription: "https://sns-webpic-qc.xhscdn.com/202408300916/f14a335b77418c66205f47a92ab1d5ab/1000g0082q6ibftejs06g4ajjir0btj45pnas28g!nd_dft_wlteh_webp_3",
		},
		{
			Id:                 10,
			RecipesName:        "新手易学，16道家常菜分享！",
			RecipesUrl:         "https://sns-webpic-qc.xhscdn.com/202408300918/5428c6ca2653dfd2e1bcceb3b5434409/spectrum/1040g0k03171jhjln0o005pm2srcne1bpq8momp0!nd_dft_wlteh_webp_3",
			RecipesUserName:    "小鹿",
			RecipesUserAvatar:  "https://sns-avatar-qc.xhscdn.com/avatar/645b7f6e338379fac1893093.jpg?imageView2/2/w/540/format/webp|imageMogr2/strip2/blur/1x56",
			RecipesDescription: "介绍十六道新手友好的家常美味。每天外出就餐选择困难？不妨尝试在家制作这十六道美味家常菜，快叫你的另一半为你烹饪吧！\n1. 酸辣爽口的白菜\n2. 香辣土豆干锅\n3. 干煸风味豆角\n4. 酸香五花肉\n5. 麻辣诱人口水鸡\n6. 香菇与鸡肉的滑嫩组合7. 脆皮豆腐配特制酱汁\n8. 花甲爆炒\n9. 鸡蛋与蟹柳的酱汁组合\n10. 手撕鸡盐焗风味\n11. 鹌鹑蛋与五花肉的炒制\n12. 鸡翅麻辣风味\n13. 干煸杏鲍菇\n14. 泡椒藕片与鸡腿的炒制\n15. 包菜配酱汁\n16. 花菜炒腊肉",
		},
		{
			Id:                 20,
			RecipesName:        "3元成本，5分钟上桌，疯抢",
			RecipesUrl:         "https://sns-webpic-qc.xhscdn.com/202408300927/4c2c7e89e2e999812244f80d7904e6ef/spectrum/1040g0k03171kp65h0q005pm2srcne1bpncfghko!nd_dft_wlteh_webp_3",
			RecipesUserName:    "豆豆子",
			RecipesUserAvatar:  "https://sns-avatar-qc.xhscdn.com/avatar/645b7f6e338379fac1893093.jpg?imageView2/2/w/540/format/webp|imageMogr2/strip2/blur/1x56",
			RecipesDescription: "3元成本，5分钟完成，立刻成为餐桌上的热门选择#分享一篇图文秘籍 极易上手的家庭常备菜肴，初学者必备的食谱来啦！5分钟内轻松完成，简单美味，一上桌就被疯抢，米饭遭殃。",
		},
		{
			Id:                 30,
			RecipesName:        "3元成本，5分钟上桌，疯抢",
			RecipesUrl:         "https://sns-webpic-qc.xhscdn.com/202408300927/4c2c7e89e2e999812244f80d7904e6ef/spectrum/1040g0k03171kp65h0q005pm2srcne1bpncfghko!nd_dft_wlteh_webp_3",
			RecipesUserName:    "华仔",
			RecipesUserAvatar:  "https://sns-avatar-qc.xhscdn.com/avatar/645b7f6e338379fac1893093.jpg?imageView2/2/w/540/format/webp|imageMogr2/strip2/blur/1x56",
			RecipesDescription: "3元成本，5分钟完成，立刻成为餐桌上的热门选择#分享一篇图文秘籍 极易上手的家庭常备菜肴，初学者必备的食谱来啦！5分钟内轻松完成，简单美味，一上桌就被疯抢，米饭遭殃。",
		},
		{
			Id:                 40,
			RecipesName:        "4元成本，3分钟上桌，疯抢",
			RecipesUrl:         "https://sns-webpic-qc.xhscdn.com/202408300927/4c2c7e89e2e999812244f80d7904e6ef/spectrum/1040g0k03171kp65h0q005pm2srcne1bpncfghko!nd_dft_wlteh_webp_3",
			RecipesUserName:    "Yo小55",
			RecipesUserAvatar:  "https://sns-avatar-qc.xhscdn.com/avatar/645b7f6e338379fac1893093.jpg?imageView2/2/w/540/format/webp|imageMogr2/strip2/blur/1x56",
			RecipesDescription: "3元成本，4分钟完成，立刻成为餐桌上的热门选择#分享一篇图文秘籍 极易上手的家庭常备菜肴，初学者必备的食谱来啦！5分钟内轻松完成，简单美味，一上桌就被疯抢，米饭遭殃。",
		},
		{
			Id:                 50,
			RecipesName:        "新手易学，16道家常菜分享！",
			RecipesUrl:         "https://sns-webpic-qc.xhscdn.com/202408300918/5428c6ca2653dfd2e1bcceb3b5434409/spectrum/1040g0k03171jhjln0o005pm2srcne1bpq8momp0!nd_dft_wlteh_webp_3",
			RecipesUserName:    "小鹿",
			RecipesUserAvatar:  "https://sns-avatar-qc.xhscdn.com/avatar/645b7f6e338379fac1893093.jpg?imageView2/2/w/540/format/webp|imageMogr2/strip2/blur/1x56",
			RecipesDescription: "介绍十六道新手友好的家常美味。每天外出就餐选择困难？不妨尝试在家制作这十六道美味家常菜，快叫你的另一半为你烹饪吧！\n1. 酸辣爽口的白菜\n2. 香辣土豆干锅\n3. 干煸风味豆角\n4. 酸香五花肉\n5. 麻辣诱人口水鸡\n6. 香菇与鸡肉的滑嫩组合7. 脆皮豆腐配特制酱汁\n8. 花甲爆炒\n9. 鸡蛋与蟹柳的酱汁组合\n10. 手撕鸡盐焗风味\n11. 鹌鹑蛋与五花肉的炒制\n12. 鸡翅麻辣风味\n13. 干煸杏鲍菇\n14. 泡椒藕片与鸡腿的炒制\n15. 包菜配酱汁\n16. 花菜炒腊肉",
		},
		{
			Id:                 60,
			RecipesName:        "新手易学，20道家常菜分享！",
			RecipesUrl:         "https://sns-webpic-qc.xhscdn.com/202408300918/5428c6ca2653dfd2e1bcceb3b5434409/spectrum/1040g0k03171jhjln0o005pm2srcne1bpq8momp0!nd_dft_wlteh_webp_3",
			RecipesUserName:    "小鹿2",
			RecipesUserAvatar:  "https://sns-avatar-qc.xhscdn.com/avatar/645b7f6e338379fac1893093.jpg?imageView2/2/w/540/format/webp|imageMogr2/strip2/blur/1x56",
			RecipesDescription: "介绍十六道新手友好的家常美味。每天外出就餐选择困难？不妨尝试在家制作这十六道美味家常菜，快叫你的另一半为你烹饪吧！\n1. 酸辣爽口的白菜\n2. 香辣土豆干锅\n3. 干煸风味豆角\n4. 酸香五花肉\n5. 麻辣诱人口水鸡\n6. 香菇与鸡肉的滑嫩组合7. 脆皮豆腐配特制酱汁\n8. 花甲爆炒\n9. 鸡蛋与蟹柳的酱汁组合\n10. 手撕鸡盐焗风味\n11. 鹌鹑蛋与五花肉的炒制\n12. 鸡翅麻辣风味\n13. 干煸杏鲍菇\n14. 泡椒藕片与鸡腿的炒制\n15. 包菜配酱汁\n16. 花菜炒腊肉",
		},
		{
			Id:                 70,
			RecipesName:        "新手易学，18道家常菜分享！",
			RecipesUrl:         "https://sns-webpic-qc.xhscdn.com/202408300918/5428c6ca2653dfd2e1bcceb3b5434409/spectrum/1040g0k03171jhjln0o005pm2srcne1bpq8momp0!nd_dft_wlteh_webp_3",
			RecipesUserName:    "小鹿3",
			RecipesUserAvatar:  "https://sns-avatar-qc.xhscdn.com/avatar/645b7f6e338379fac1893093.jpg?imageView2/2/w/540/format/webp|imageMogr2/strip2/blur/1x56",
			RecipesDescription: "介绍十六道新手友好的家常美味。每天外出就餐选择困难？不妨尝试在家制作这十六道美味家常菜，快叫你的另一半为你烹饪吧！\n1. 酸辣爽口的白菜\n2. 香辣土豆干锅\n3. 干煸风味豆角\n4. 酸香五花肉\n5. 麻辣诱人口水鸡\n6. 香菇与鸡肉的滑嫩组合7. 脆皮豆腐配特制酱汁\n8. 花甲爆炒\n9. 鸡蛋与蟹柳的酱汁组合\n10. 手撕鸡盐焗风味\n11. 鹌鹑蛋与五花肉的炒制\n12. 鸡翅麻辣风味\n13. 干煸杏鲍菇\n14. 泡椒藕片与鸡腿的炒制\n15. 包菜配酱汁\n16. 花菜炒腊肉",
		},
		{
			Id:                 80,
			RecipesName:        "新手易学，19道家常菜分享！",
			RecipesUrl:         "https://sns-webpic-qc.xhscdn.com/202408300918/5428c6ca2653dfd2e1bcceb3b5434409/spectrum/1040g0k03171jhjln0o005pm2srcne1bpq8momp0!nd_dft_wlteh_webp_3",
			RecipesUserName:    "小鹿4",
			RecipesUserAvatar:  "https://sns-avatar-qc.xhscdn.com/avatar/645b7f6e338379fac1893093.jpg?imageView2/2/w/540/format/webp|imageMogr2/strip2/blur/1x56",
			RecipesDescription: "介绍十六道新手友好的家常美味。每天外出就餐选择困难？不妨尝试在家制作这十六道美味家常菜，快叫你的另一半为你烹饪吧！\n1. 酸辣爽口的白菜\n2. 香辣土豆干锅\n3. 干煸风味豆角\n4. 酸香五花肉\n5. 麻辣诱人口水鸡\n6. 香菇与鸡肉的滑嫩组合7. 脆皮豆腐配特制酱汁\n8. 花甲爆炒\n9. 鸡蛋与蟹柳的酱汁组合\n10. 手撕鸡盐焗风味\n11. 鹌鹑蛋与五花肉的炒制\n12. 鸡翅麻辣风味\n13. 干煸杏鲍菇\n14. 泡椒藕片与鸡腿的炒制\n15. 包菜配酱汁\n16. 花菜炒腊肉",
		},
		{
			Id:                 90,
			RecipesName:        "新手易学，21道家常菜分享！",
			RecipesUrl:         "https://sns-webpic-qc.xhscdn.com/202408300918/5428c6ca2653dfd2e1bcceb3b5434409/spectrum/1040g0k03171jhjln0o005pm2srcne1bpq8momp0!nd_dft_wlteh_webp_3",
			RecipesUserName:    "小鹿5",
			RecipesUserAvatar:  "https://sns-avatar-qc.xhscdn.com/avatar/645b7f6e338379fac1893093.jpg?imageView2/2/w/540/format/webp|imageMogr2/strip2/blur/1x56",
			RecipesDescription: "介绍十六道新手友好的家常美味。每天外出就餐选择困难？不妨尝试在家制作这十六道美味家常菜，快叫你的另一半为你烹饪吧！\n1. 酸辣爽口的白菜\n2. 香辣土豆干锅\n3. 干煸风味豆角\n4. 酸香五花肉\n5. 麻辣诱人口水鸡\n6. 香菇与鸡肉的滑嫩组合7. 脆皮豆腐配特制酱汁\n8. 花甲爆炒\n9. 鸡蛋与蟹柳的酱汁组合\n10. 手撕鸡盐焗风味\n11. 鹌鹑蛋与五花肉的炒制\n12. 鸡翅麻辣风味\n13. 干煸杏鲍菇\n14. 泡椒藕片与鸡腿的炒制\n15. 包菜配酱汁\n16. 花菜炒腊肉",
		},
		{
			Id:                 100,
			RecipesName:        "新手宝，81道家常菜谱，厨艺速成1-18道。",
			RecipesUrl:         "https://sns-webpic-qc.xhscdn.com/202408300825/9657b8b3bdd1f71570c358a6945cb677/spectrum/1040g34o316r8jttags0g5pm2srcne1bpj81tprg!nd_dft_wlteh_webp_3",
			RecipesUserName:    "Yo小",
			RecipesUserAvatar:  "https://sns-avatar-qc.xhscdn.com/avatar/645b7f6e338379fac1893093.jpg?imageView2/2/w/540/format/webp|imageMogr2/strip2/blur/1x56",
			RecipesDescription: "一周的新手厨艺提升计划，专为零基础的小白量身定制，助你轻松成为厨艺高手！零基础的你，不妨下载这份菜谱，慢慢磨练技艺~同时，也别忘了分享给那些你在乎的人，让TA也能加入烹饪的行列~\n1. 辣味炒肉片（基础技能）\n2. 绿色青椒炒鸡蛋（基础技能）\n3. 乡村风味一锅香（双重美味）4. 炒五花肉配花菜（品质可靠）\n5. 蒜苔炒肉末（风味独特）\n6. 炒肉配笋干（质地上乘）\n7. 鸡蛋拌酱香（酱料配方：2勺酱油1勺香醋1勺蚝油半勺糖1勺辣椒粉1勺生粉半碗水混合均匀）\n8. 清蒸排骨带蒜香（零失误）\n9. 炒蛋配虾仁（操作简单，酱汁：1勺酱油，1勺醋，1勺蚝油，1勺糖，1勺生粉，1勺辣椒粉，半碗水，调匀备用）10. 经典川菜之辣子鸡丁\n11. 传统粤式葱油焖鸡\n12. 肉丸与白菜的炒制\n13. 香菇与滑嫩鸡肉（不辣口味）\n14. 年夜饭必备金钱蛋\n15. 辣椒与油豆腐的炒制（勾芡技巧）\n16. 土豆丝的炒制（基础技巧）\n17. 川菜中人见人爱的包菜回锅肉\n18. 入门至进阶的白菜火腿豆腐煲19.青椒搭配火腿炒制（简便）",
		},
		{
			Id:                 110,
			RecipesName:        "新手宝，81道家常菜谱，厨艺速成2-36道。",
			RecipesUrl:         "https://sns-webpic-qc.xhscdn.com/202408300825/9657b8b3bdd1f71570c358a6945cb677/spectrum/1040g34o316r8jttags0g5pm2srcne1bpj81tprg!nd_dft_wlteh_webp_3",
			RecipesUserName:    "Yo小",
			RecipesUserAvatar:  "https://sns-avatar-qc.xhscdn.com/avatar/645b7f6e338379fac1893093.jpg?imageView2/2/w/540/format/webp|imageMogr2/strip2/blur/1x56",
			RecipesDescription: "一周的新手厨艺提升计划，专为零基础的小白量身定制，助你轻松成为厨艺高手！零基础的你，不妨下载这份菜谱，慢慢磨练技艺~同时，也别忘了分享给那些你在乎的人，让TA也能加入烹饪的行列~\n1. 辣味炒肉片（基础技能）\n2. 绿色青椒炒鸡蛋（基础技能）\n3. 乡村风味一锅香（双重美味）4. 炒五花肉配花菜（品质可靠）\n5. 蒜苔炒肉末（风味独特）\n6. 炒肉配笋干（质地上乘）\n7. 鸡蛋拌酱香（酱料配方：2勺酱油1勺香醋1勺蚝油半勺糖1勺辣椒粉1勺生粉半碗水混合均匀）\n8. 清蒸排骨带蒜香（零失误）\n9. 炒蛋配虾仁（操作简单，酱汁：1勺酱油，1勺醋，1勺蚝油，1勺糖，1勺生粉，1勺辣椒粉，半碗水，调匀备用）10. 经典川菜之辣子鸡丁\n11. 传统粤式葱油焖鸡\n12. 肉丸与白菜的炒制\n13. 香菇与滑嫩鸡肉（不辣口味）\n14. 年夜饭必备金钱蛋\n15. 辣椒与油豆腐的炒制（勾芡技巧）\n16. 土豆丝的炒制（基础技巧）\n17. 川菜中人见人爱的包菜回锅肉\n18. 入门至进阶的白菜火腿豆腐煲19.青椒搭配火腿炒制（简便）",
		},
		{
			Id:                 120,
			RecipesName:        "新手宝，81道家常菜谱，厨艺速成37-54道。",
			RecipesUrl:         "https://sns-webpic-qc.xhscdn.com/202408300825/9657b8b3bdd1f71570c358a6945cb677/spectrum/1040g34o316r8jttags0g5pm2srcne1bpj81tprg!nd_dft_wlteh_webp_3",
			RecipesUserName:    "Yo小",
			RecipesUserAvatar:  "https://sns-avatar-qc.xhscdn.com/avatar/645b7f6e338379fac1893093.jpg?imageView2/2/w/540/format/webp|imageMogr2/strip2/blur/1x56",
			RecipesDescription: "一周的新手厨艺提升计划，专为零基础的小白量身定制，助你轻松成为厨艺高手！零基础的你，不妨下载这份菜谱，慢慢磨练技艺~同时，也别忘了分享给那些你在乎的人，让TA也能加入烹饪的行列~\n1. 辣味炒肉片（基础技能）\n2. 绿色青椒炒鸡蛋（基础技能）\n3. 乡村风味一锅香（双重美味）4. 炒五花肉配花菜（品质可靠）\n5. 蒜苔炒肉末（风味独特）\n6. 炒肉配笋干（质地上乘）\n7. 鸡蛋拌酱香（酱料配方：2勺酱油1勺香醋1勺蚝油半勺糖1勺辣椒粉1勺生粉半碗水混合均匀）\n8. 清蒸排骨带蒜香（零失误）\n9. 炒蛋配虾仁（操作简单，酱汁：1勺酱油，1勺醋，1勺蚝油，1勺糖，1勺生粉，1勺辣椒粉，半碗水，调匀备用）10. 经典川菜之辣子鸡丁\n11. 传统粤式葱油焖鸡\n12. 肉丸与白菜的炒制\n13. 香菇与滑嫩鸡肉（不辣口味）\n14. 年夜饭必备金钱蛋\n15. 辣椒与油豆腐的炒制（勾芡技巧）\n16. 土豆丝的炒制（基础技巧）\n17. 川菜中人见人爱的包菜回锅肉\n18. 入门至进阶的白菜火腿豆腐煲19.青椒搭配火腿炒制（简便）",
		},
		{
			Id:                 130,
			RecipesName:        "新手宝，81道家常菜谱，厨艺速成55-68道。",
			RecipesUrl:         "https://sns-webpic-qc.xhscdn.com/202408300825/9657b8b3bdd1f71570c358a6945cb677/spectrum/1040g34o316r8jttags0g5pm2srcne1bpj81tprg!nd_dft_wlteh_webp_3",
			RecipesUserName:    "Yo小",
			RecipesUserAvatar:  "https://sns-avatar-qc.xhscdn.com/avatar/645b7f6e338379fac1893093.jpg?imageView2/2/w/540/format/webp|imageMogr2/strip2/blur/1x56",
			RecipesDescription: "一周的新手厨艺提升计划，专为零基础的小白量身定制，助你轻松成为厨艺高手！零基础的你，不妨下载这份菜谱，慢慢磨练技艺~同时，也别忘了分享给那些你在乎的人，让TA也能加入烹饪的行列~\n1. 辣味炒肉片（基础技能）\n2. 绿色青椒炒鸡蛋（基础技能）\n3. 乡村风味一锅香（双重美味）4. 炒五花肉配花菜（品质可靠）\n5. 蒜苔炒肉末（风味独特）\n6. 炒肉配笋干（质地上乘）\n7. 鸡蛋拌酱香（酱料配方：2勺酱油1勺香醋1勺蚝油半勺糖1勺辣椒粉1勺生粉半碗水混合均匀）\n8. 清蒸排骨带蒜香（零失误）\n9. 炒蛋配虾仁（操作简单，酱汁：1勺酱油，1勺醋，1勺蚝油，1勺糖，1勺生粉，1勺辣椒粉，半碗水，调匀备用）10. 经典川菜之辣子鸡丁\n11. 传统粤式葱油焖鸡\n12. 肉丸与白菜的炒制\n13. 香菇与滑嫩鸡肉（不辣口味）\n14. 年夜饭必备金钱蛋\n15. 辣椒与油豆腐的炒制（勾芡技巧）\n16. 土豆丝的炒制（基础技巧）\n17. 川菜中人见人爱的包菜回锅肉\n18. 入门至进阶的白菜火腿豆腐煲19.青椒搭配火腿炒制（简便）",
		},
		{
			Id:                 140,
			RecipesName:        "新手宝，81道家常菜谱，厨艺速成68-81道。",
			RecipesUrl:         "https://sns-webpic-qc.xhscdn.com/202408300825/9657b8b3bdd1f71570c358a6945cb677/spectrum/1040g34o316r8jttags0g5pm2srcne1bpj81tprg!nd_dft_wlteh_webp_3",
			RecipesUserName:    "Yo小",
			RecipesUserAvatar:  "https://sns-avatar-qc.xhscdn.com/avatar/645b7f6e338379fac1893093.jpg?imageView2/2/w/540/format/webp|imageMogr2/strip2/blur/1x56",
			RecipesDescription: "一周的新手厨艺提升计划，专为零基础的小白量身定制，助你轻松成为厨艺高手！零基础的你，不妨下载这份菜谱，慢慢磨练技艺~同时，也别忘了分享给那些你在乎的人，让TA也能加入烹饪的行列~\n1. 辣味炒肉片（基础技能）\n2. 绿色青椒炒鸡蛋（基础技能）\n3. 乡村风味一锅香（双重美味）4. 炒五花肉配花菜（品质可靠）\n5. 蒜苔炒肉末（风味独特）\n6. 炒肉配笋干（质地上乘）\n7. 鸡蛋拌酱香（酱料配方：2勺酱油1勺香醋1勺蚝油半勺糖1勺辣椒粉1勺生粉半碗水混合均匀）\n8. 清蒸排骨带蒜香（零失误）\n9. 炒蛋配虾仁（操作简单，酱汁：1勺酱油，1勺醋，1勺蚝油，1勺糖，1勺生粉，1勺辣椒粉，半碗水，调匀备用）10. 经典川菜之辣子鸡丁\n11. 传统粤式葱油焖鸡\n12. 肉丸与白菜的炒制\n13. 香菇与滑嫩鸡肉（不辣口味）\n14. 年夜饭必备金钱蛋\n15. 辣椒与油豆腐的炒制（勾芡技巧）\n16. 土豆丝的炒制（基础技巧）\n17. 川菜中人见人爱的包菜回锅肉\n18. 入门至进阶的白菜火腿豆腐煲19.青椒搭配火腿炒制（简便）",
		},
	}
}
