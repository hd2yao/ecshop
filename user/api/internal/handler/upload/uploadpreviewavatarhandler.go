package upload

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"

	"github.com/hd2yao/ecshop/user/api/internal/logic/upload"
	"github.com/hd2yao/ecshop/user/api/internal/svc"
)

// UploadPreviewAvatarHandler 头像上传预览（注册和更新通用）
func UploadPreviewAvatarHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := upload.NewUploadPreviewAvatarLogic(r.Context(), svcCtx)
		resp, err := l.UploadPreviewAvatar(r)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
