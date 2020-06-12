package compass

import (
	"github.com/containous/traefik/v2/pkg/config/dynamic"
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
	Configuration			*dynamic.Configuration
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
	p.Configuration = &dynamic.Configuration{
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
	return nil
}


// 提供虚拟主机发现检查和规则发现检查功能
func (p *Provider) Provide(configurationChan chan<- dynamic.Message, pool *safe.Pool) error {
	p.VHostDiscovery(configurationChan, pool)
	p.CompassRuleWatcher(configurationChan, pool)
	return nil
}



