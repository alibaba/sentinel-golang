package http

import (
	"errors"
	"fmt"
	sentinelroute "github.com/alibaba/sentinel-golang/core/route"
	"net/http"
)

func ClientRoute(req *http.Request) (*http.Request, error) {
	if req == nil {
		return req, errors.New("req is nil")
	}

	trafficContext := &sentinelroute.TrafficContext{
		Ctx:      req.Context(),
		HostName: req.URL.Hostname(),
		Port:     req.URL.Port(),
		Method:   req.Method,
		Path:     req.URL.Path,
		Scheme:   req.URL.Scheme,
	}
	header := make(map[string]string)
	for k, v := range req.Header {
		if len(v) != 0 {
			header[k] = v[0]
		}
	}
	trafficContext.Header = header

	routeDestination, err := trafficContext.Route()
	if err != nil {
		return nil, err
	}

	if routeDestination != nil {
		if routeDestination.HostUpdated {
			req.URL.Host = fmt.Sprintf("%s:%s", routeDestination.HostName, routeDestination.Port)
		}
		req.Header.Set(sentinelroute.TrafficTagHeader, routeDestination.TrafficTag)
		if routeDestination.TagUpdated && routeDestination.Ctx != nil {
			req = req.WithContext(routeDestination.Ctx)
		}
	}

	return req, nil
}
