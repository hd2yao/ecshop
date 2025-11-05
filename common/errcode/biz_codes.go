package errcode

import "net/http"

var codes = map[int]struct{}{}

/**
 * 整个状态码总共 7 位，前三位表示「业务微服务」状态码，后4位表示「服务内部接口」状态码，后续需要再进行拆分
 * 公共操作：110
 * 用户服务：210，其中验证码：2101 开头，用户账号：2102 开头
 * 购物车服务：310
 * 商品服务：410
 * 优惠券服务：510
 * 订单服务：610
 * 支付服务：710
 */

/**
 * 通用操作码
 */
var (
	Success                   = newError(110000, "success")
	CommonOpRepeat            = newError(110001, "请不要重复操作")
	CommonParamError          = newError(110002, "参数错误")
	CommonServerError         = newError(110003, "服务异常")
	CommonTooManyTry          = newError(110004, "当前访问人数过多，请稍候再试...")
	CommonNetworkAddressError = newError(110005, "网络地址错误")
	CommonServerBusyError     = newError(110006, "系统繁忙，请稍后重试")
	CommonPanicErr            = newError(110007, "(*^__^*)系统开小差了,请稍后重试")
)

/**
 * 用户微服务验证码相关  2101 开头
 */
var (
	UserPhoneError       = newError(2101001, "手机号不合法")
	UserCodeFastLimited  = newError(2101002, "验证码发送太快了")
	UserCodeError        = newError(2101003, "验证码错误")
	UserCodeCaptchaError = newError(2101004, "图形验证码错误")
	UserCodeEmailError   = newError(2101005, "邮箱验证码错误")
)

/**
 * 用户微服务账号相关 2102 开头
 */
var (
	UserAccountExist      = newError(2102001, "用户账号已存在")
	UserAccountUnregister = newError(2102002, "用户账号未注册")
	UserAccountPwdError   = newError(2102003, "用户账号或密码错误")
)

/**
 * 用户微服务 Token 相关 2103 开头
 */
var (
	UserTokenInvalid        = newError(2103001, "Token 无效")
	UserTokenExpired        = newError(2103002, "Token 已过期")
	UserTokenMalformed      = newError(2103003, "Token 格式错误")
	UserRefreshTokenInvalid = newError(2103004, "Refresh Token 无效或已失效")
)

// HttpStatusCode 返回 HTTP 状态码
func (e *AppError) HttpStatusCode() int {
	switch e.Code() {
	case Success.Code():
		return http.StatusOK
	case CommonServerError.Code(), CommonServerBusyError.Code(), CommonPanicErr.Code():
		return http.StatusInternalServerError
	case CommonOpRepeat.Code(), CommonParamError.Code(), UserCodeFastLimited.Code(), UserCodeError.Code(), UserCodeCaptchaError.Code(), UserCodeEmailError.Code(),
		UserAccountExist.Code(), UserAccountUnregister.Code(), UserAccountPwdError.Code():
		return http.StatusBadRequest
	case CommonNetworkAddressError.Code():
		return http.StatusNotFound
	case CommonTooManyTry.Code():
		return http.StatusTooManyRequests
	case UserPhoneError.Code():
		return http.StatusForbidden
	case UserTokenInvalid.Code(), UserTokenExpired.Code(), UserTokenMalformed.Code(), UserRefreshTokenInvalid.Code():
		return http.StatusUnauthorized
	default:
		return http.StatusInternalServerError
	}
}
