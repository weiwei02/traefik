package compass

import (
	"context"
	pb "github.com/containous/traefik/v2/api/compass/virtualHost"
	"github.com/containous/traefik/v2/pkg/config/dynamic"
	"github.com/containous/traefik/v2/pkg/log"
	"github.com/containous/traefik/v2/pkg/tls"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"testing"
	"time"
)

//@author Wang Weiwei
//@since 2020/6/12

var config = &dynamic.Configuration{
	HTTP: &dynamic.HTTPConfiguration{
		Routers:     make(map[string]*dynamic.Router),
		Middlewares: make(map[string]*dynamic.Middleware),
		Services:    make(map[string]*dynamic.Service),
	},
	TCP: &dynamic.TCPConfiguration{
		Routers:  make(map[string]*dynamic.TCPRouter),
		Services: make(map[string]*dynamic.TCPService),
	},
	TLS: &dynamic.TLSConfiguration{
		Stores:  make(map[string]tls.Store),
		Options: make(map[string]tls.Options),
	},
	UDP: &dynamic.UDPConfiguration{
		Routers:  make(map[string]*dynamic.UDPRouter),
		Services: make(map[string]*dynamic.UDPService),
	},
}

// 测试从compass获取数据
func TestGetDataFromCompass(t *testing.T) {
	conn, err := grpc.Dial("10.160.137.160:9091", grpc.WithInsecure())
	require.NoError(t, err)
	time.Sleep(10 * time.Second)
	conn.ResetConnectBackoff()
	client := pb.NewNeoDiscoveryServiceClient(conn)
	stream, err := client.DiscoveryData(context.Background(), &pb.QueryConfig{})
	require.NoError(t, err)
	//for {
	res, err := stream.Recv()
	require.NoError(t, err)
	require.NotNilf(t, res, "compass data can not null")
	t.Log(res.String())
	for _, vhost := range res.Vhosts {
		log.WithoutContext().WithField(log.ProviderName, providerName).Infof("receive compass virtual update data %s", vhost.String())
		transVhostToHTTPConfig(config.HTTP, vhost)
		require.Truef(t, len(config.HTTP.Services) > 0, "compass 动态配置服务为空")
		require.Truef(t, len(config.HTTP.Routers) > 0, "compass 动态配置路由为空")
		t.Log(config)
	}
	//}
}
