package kratos

import (
	"context"
	"strings"

	"github.com/go-kratos/kratos/v2/transport"
)

func ServiceNameExtract(ctx context.Context) string {
	if v, ok := transport.FromClientContext(ctx); ok {
		res := v.Endpoint()
		if strings.HasPrefix(res, "discovery:///") {
			return strings.TrimPrefix(res, "discovery:///")
		}
	}
	return ""
}
