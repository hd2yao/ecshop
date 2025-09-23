package upload

import (
	"net/http"

	"github.com/hd2yao/ecshop/common/app"
	"github.com/hd2yao/ecshop/common/errcode"
	"github.com/hd2yao/ecshop/user/api/internal/logic/upload"
	"github.com/hd2yao/ecshop/user/api/internal/svc"
)

// UploadPreviewAvatarHandler 头像上传预览（注册和更新通用）
func UploadPreviewAvatarHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := upload.NewUploadPreviewAvatarLogic(r.Context(), svcCtx)
		resp, err := l.UploadPreviewAvatar(r)

		response := app.NewResponse(r.Context(), w)
		if err != nil {
			response.Error(errcode.CommonServerError.WithCause(err))
		} else {
			response.Success(resp)
		}
	}
}
