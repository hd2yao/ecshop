package upload

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"

	"github.com/hd2yao/ecshop/user/api/internal/logic/upload"
	"github.com/hd2yao/ecshop/user/api/internal/svc"
	"github.com/hd2yao/ecshop/user/api/internal/types"
)

// UploadAvatarHandler 上传用户头像
func UploadAvatarHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 解析multipart表单
		err := r.ParseMultipartForm(10 << 20) // 10MB
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		// 获取表单参数
		var req types.UploadAvatarRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		// 获取上传的文件
		file, fileHeader, err := r.FormFile("file")
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		defer file.Close()

		// 将文件句柄添加到请求中，传递给logic层
		l := upload.NewUploadAvatarLogic(r.Context(), svcCtx)
		resp, err := l.UploadAvatar(&req, file, fileHeader)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
