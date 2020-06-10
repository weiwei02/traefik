package compass

import (
	"context"
	"fmt"
	"github.com/containous/traefik/v2/pkg/config/dynamic"
	"github.com/containous/traefik/v2/pkg/log"
	"github.com/containous/traefik/v2/pkg/safe"
	"github.com/containous/traefik/v2/pkg/tls"
	"google.golang.org/grpc"
)

//@author Wang Weiwei
//@since 2020/6/10
var providerName = "compass"

type Provider struct {
	GRPCAddress               string `description:"compass grpc server address" json:"grpcAddress,omitempty" toml:"grpcAddress,omitempty" yaml:"grpcAddress,omitempty" export:"true"`
	RestAddress               string `description:"compass rest server address" json:"restAddress,omitempty" toml:"restAddress,omitempty" yaml:"restAddress,omitempty" export:"true"`
	DebugLogGeneratedTemplate bool   `description:"Enable debug logging of generated configuration template." json:"debugLogGeneratedTemplate,omitempty" toml:"debugLogGeneratedTemplate,omitempty" yaml:"debugLogGeneratedTemplate,omitempty" export:"true"`
	conn                      *grpc.ClientConn
}

// Init the provider.
func (p *Provider) Init() error {
	if p.GRPCAddress == "" {

	}
	conn, err := grpc.Dial(p.GRPCAddress, grpc.WithInsecure())
	if err != nil {
		return err
	}
	p.conn = conn
	return nil
}

func (p *Provider) Provide(configurationChan chan<- dynamic.Message, pool *safe.Pool) error {
	pool.GoCtx(func(ctx context.Context) {
		// 失败后尝试重新连接
	LIS:
		p.conn.ResetConnectBackoff()
		client := NewNeoDiscoveryServiceClient(p.conn)
		stream, err := client.DiscoveryData(context.Background(), &QueryConfig{})
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
					configurationChan <-  dynamic.Message{
						ProviderName: providerName,
						Configuration: &dynamic.Configuration{
							HTTP: transVhostToHTTPConfig(vhost),
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
						},
					}
				}
			}


		}
	})
	return nil
}


// 创建http虚拟主机配置
// compass返回的数据：
// 每个虚拟主机下至少有一个路由，默认是 default ，其它路由代表是有特定的前缀
func transVhostToHTTPConfig(vhost *VirtualHost) *dynamic.HTTPConfiguration {
	routePrefix := vhost.VirtualHost + "_"
	conf := &dynamic.HTTPConfiguration{
		Routers:     make(map[string]*dynamic.Router),
		Middlewares: make(map[string]*dynamic.Middleware),
		Services:    make(map[string]*dynamic.Service),
	}
	for prefix, ser := range vhost.Routes {


		// add routes
		conf.Routers[routePrefix + prefix] = &dynamic.Router{
			EntryPoints: []string{"web"},
			Middlewares: nil,
			Service:     ser,
			Rule: func() string {
				if prefix == "default" {
					return fmt.Sprintf("Host(`%s`)", vhost.VirtualHost)
				}
				return fmt.Sprintf("Host(`%s`) && PathPrefix(`%s`)", vhost.VirtualHost, prefix)
			}(),
		}

		// target service
		for _, service := range vhost.Services {
			conf.Services[routePrefix + service.Name] = &dynamic.Service{
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
	return conf
}
