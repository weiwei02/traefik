package compass

import (
	"context"
	"fmt"
	pb "github.com/containous/traefik/v2/api/compass/virtualHost"
	"github.com/containous/traefik/v2/pkg/config/dynamic"
	"github.com/containous/traefik/v2/pkg/log"
	"github.com/containous/traefik/v2/pkg/safe"
)

//@author Wang Weiwei
//@since 2020/6/11
const defaultPath = "default"

func (p *Provider) VHostDiscovery(configurationChan chan<- dynamic.Message, pool *safe.Pool) {
	pool.GoCtx(func(ctx context.Context) {
		// 失败后尝试重新连接
	LIS:
		p.conn.ResetConnectBackoff()
		client := pb.NewNeoDiscoveryServiceClient(p.conn)
		stream, err := client.DiscoveryData(context.Background(), &pb.QueryConfig{})
		if err != nil {
			log.WithoutContext().WithField(log.ProviderName, providerName).Errorf("create stream error: %s", err)
			goto LIS
		}
		for {
			res, err := stream.Recv()
			if err != nil {
				log.WithoutContext().WithField(log.ProviderName, providerName).Errorf("receive stream error: %s", err)
				goto LIS
			}
			if len(res.Vhosts) > 0 {
				for _, vhost := range res.Vhosts {
					log.WithoutContext().WithField(log.ProviderName, providerName).Info("receive compass virtual update data %s", vhost.String())
					transVhostToHTTPConfig(p.Configuration.HTTP, vhost)
					configurationChan <- dynamic.Message{
						ProviderName:  providerName,
						Configuration: p.Configuration.DeepCopy(),
					}
				}
			}
		}
	})
}

// 创建http虚拟主机配置
// compass返回的数据：
// 每个虚拟主机下至少有一个路由，默认是 default ，其它路由代表是有特定的前缀
// 如果虚拟主机已有相关路由，则只更新路由下的服务。如果虚拟主机下没有相关路由，则需同时更新服务与中间件
func transVhostToHTTPConfig(conf *dynamic.HTTPConfiguration, vhost *pb.VirtualHost)  {
	routePrefix := vhost.VirtualHost + "@"
	for prefix, ser := range vhost.Routes {
		// add routes
		if route, ok := conf.Routers[routePrefix + prefix]; ok {
			route.Service = ser
		}else {
			conf.Routers[routePrefix + prefix] = &dynamic.Router{
				EntryPoints: []string{"web"},
				Middlewares: make([]string,0),
				Service:     ser,
				Rule: func() string {
					if prefix == defaultPath {
						return fmt.Sprintf("Host(`%s`)", vhost.VirtualHost)
					}
					return fmt.Sprintf("Host(`%s`) && PathPrefix(`%s`)", vhost.VirtualHost, prefix)
				}(),
			}
		}

		// target service
		for _, service := range vhost.Services {
			conf.Services[service.Name] = &dynamic.Service{
				LoadBalancer: &dynamic.ServersLoadBalancer{
					Servers: func() []dynamic.Server{
						ss := make([]dynamic.Server, 0)
						for _, endpoint := range service.Endpoints {
							ss = append(ss,
								dynamic.Server{
									URL: fmt.Sprintf("http://%s%s", endpoint, service.BasePath),
									// todo healthcheck
								})
						}
						return ss
					}(),
				},
			}
		}
	}
}