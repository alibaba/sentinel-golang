package grpc

import (
	"context"
	"errors"
	"fmt"
	"github.com/alibaba/sentinel-golang/core/route"
	"github.com/alibaba/sentinel-golang/core/route/base"
	"google.golang.org/grpc"
	"net"
	"strings"
)

var (
	connToBaggage map[string]map[string]string = make(map[string]map[string]string)
	cm            *route.ClusterManager        = nil
)

type Baggage map[string]string

func NewDialer(b Baggage) func(context.Context, string) (net.Conn, error) {
	return func(ctx context.Context, addr string) (net.Conn, error) {
		parts := strings.Split(addr, "/")
		if len(parts) != 2 {
			return nil, errors.New("invalid address format")
		}
		tc := &base.TrafficContext{
			ServiceName: parts[0],
			MethodName:  parts[1],
			Headers:     make(map[string]string),
		}

		instance, err := cm.GetOne(tc)

		if err != nil {
			return nil, err
		}
		if instance == nil {
			return nil, errors.New("no matched provider")
		}
		conn, err := net.Dial("tcp", fmt.Sprintf("%s:%v", instance.Host, instance.Port))
		if err != nil {
			return nil, err
		}
		b = tc.Baggage

		return conn, nil
	}
}

func NewTrafficUnaryIntercepter(baggage Baggage) grpc.DialOption {
	return grpc.WithUnaryInterceptor(
		func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
			newCtx := ctx
			// TODO modify the request with baggage
			return invoker(newCtx, method, req, reply, cc, opts...)
		})
}

func NewTrafficStreamIntercepter(baggage Baggage) grpc.DialOption {
	return grpc.WithStreamInterceptor(
		func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
			newCtx := ctx
			// TODO modify the request with baggage
			return streamer(newCtx, desc, cc, method, opts...)
		})
}

func Dial(addr string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	var b Baggage = make(map[string]string)
	opts = append(opts, grpc.WithContextDialer(NewDialer(b)))
	opts = append(opts, NewTrafficUnaryIntercepter(b))
	opts = append(opts, NewTrafficStreamIntercepter(b))
	return grpc.Dial(addr, opts...)
}
