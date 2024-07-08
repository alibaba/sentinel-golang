package gin

import (
	"encoding/json"
	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/alibaba/sentinel-golang/core/fallback"
	"github.com/alibaba/sentinel-golang/logging"
	"github.com/gin-gonic/gin"
	"net/http"
)

func sentinelFallback(ctx *gin.Context, resource string, blockType base.BlockType) bool {
	behavior, exist := fallback.GetWebFallbackBehavior(resource, blockType)
	if !exist || behavior == nil {
		ctx.AbortWithStatus(http.StatusTooManyRequests)
		return false
	}

	switch behavior.WebFallbackMode {
	case 0: // return
		if behavior.WebRespContentType != 0 && behavior.WebRespContentType != 1 { // text
			return false
		}

		if behavior.WebRespContentType == 0 { // text
			ctx.String(int(behavior.WebRespStatusCode), behavior.WebRespMessage)
			ctx.Abort()
			return true
		}
		if behavior.WebRespContentType == 1 { // json
			var message interface{}
			err := json.Unmarshal([]byte(behavior.WebRespMessage), &message)
			if err != nil {
				logging.Error(err, "[sentinelFallback] unmarshal web fall back behavior failed")
				return false
			}
			ctx.AbortWithStatusJSON(int(behavior.WebRespStatusCode), message)
			return true
		}
	case 1: // redirect
		ctx.Redirect(int(behavior.WebRespStatusCode), behavior.WebRedirectUrl)
		return true
	}
	return false
}
