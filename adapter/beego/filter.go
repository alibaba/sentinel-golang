package beego

import (
	"net/http"

	sentinel "github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/alibaba/sentinel-golang/logging"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context"
)

var logger = logging.GetDefaultLogger()

// SentinelFilter return new FilterFunc
func SentinelFilters(opts ...Option) (beforeExec beego.FilterFunc, finishRouter beego.FilterFunc) {
	options := evaluateOptions(opts)
	return func(ctx *context.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Memory leaks may occur if BeforeExec exec fails
				// because of FinishRouter won't be executed and let entry escape.
				// The most likely place to panic is your custom fallback.
				// So keep your custom fallback safe.
				// Here is the last guarantee.
				ctx.ResponseWriter.WriteHeader(http.StatusInternalServerError)
			}
		}()
		var resourceName = ctx.Input.Method() + ":" + ctx.Input.URL()
		// todo(gorexlv):
		// here will not get RouterPattern because this filter
		// exec at BeforeExec hook pointer, which before setting RouterPattern
		// ref: https://github.com/astaxie/beego/issues/3949
		if routerPattern, ok := ctx.Input.GetData("RouterPattern").(string); ok {
			resourceName = ctx.Input.Method() + ":" + routerPattern
		}

		if options.resourceExtract != nil {
			resourceName = options.resourceExtract(ctx)
		}

		entry, err := sentinel.Entry(
			resourceName,
			sentinel.WithResourceType(base.ResTypeWeb),
			sentinel.WithTrafficType(base.Inbound),
		)

		if err != nil {
			if options.blockFallback != nil {
				options.blockFallback(ctx)
			} else {
				ctx.ResponseWriter.WriteHeader(http.StatusTooManyRequests)
			}
			return
		}
		ctx.Input.SetData("SentinelEntry", entry)
		// todo(gorexlv) how to check finishRouter filter
	}, func(ctx *context.Context) {
		if entryData := ctx.Input.GetData("SentinelEntry"); entryData != nil {
			if entry, ok := entryData.(*base.SentinelEntry); ok {
				entry.Exit()
				return
			}
		}

		if beego.BConfig.RunMode == beego.DEV {
			// check BeforeExec filter
			logger.Panic("no beforeExec filter found.")
		}
	}
}

