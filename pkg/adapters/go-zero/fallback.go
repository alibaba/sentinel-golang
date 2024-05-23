package go_zero

import (
	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/alibaba/sentinel-golang/core/fallback"
	"github.com/zeromicro/go-zero/rest/httpx"

	//"github.com/zeromicro/go-zero/rest/httpx"
	"net/http"
)

func sentinelFallback(w *http.ResponseWriter, req *http.Request, resource string, blockType base.BlockType) {
	if w == nil || req == nil {
		return
	}

	behavior, exist := fallback.GetWebFallbackBehavior(resource, blockType)
	if !exist || behavior == nil {
		http.Error(*w, "Blocked by Sentinel", http.StatusInternalServerError)
		return
	}

	switch behavior.WebFallbackMode {
	case 0: // return
		if behavior.WebRespContentType != 0 && behavior.WebRespContentType != 1 { // text
			http.Error(*w, "Blocked by Sentinel", http.StatusInternalServerError)
			return
		}

		if behavior.WebRespContentType == 0 { // text
			http.Error(*w, behavior.WebRespMessage, int(behavior.WebRespStatusCode))
			return
		}
		if behavior.WebRespContentType == 1 { // json
			httpx.WriteJson(*w, int(behavior.WebRespStatusCode), behavior.WebRespMessage)
			return
		}
	case 1: // redirect
		http.Redirect(*w, req, behavior.WebRedirectUrl, int(behavior.WebRespStatusCode))
		return
	}
}
