package compass

import (
	"github.com/containous/traefik/v2/pkg/config/dynamic"
	"testing"
)

//@author Wang Weiwei
//@since 2020/6/18

// 模拟构建一个依赖compass的动态配置
func NewCompassDynamicConfig() *dynamic.Configuration {
	conf := &dynamic.Configuration{
		HTTP: &dynamic.HTTPConfiguration{
			Routers: func() map[string]*dynamic.Router {
				router := make(map[string]*dynamic.Router)
				router["dsfv2.t.17usoft.com"] = &dynamic.Router{
					EntryPoints: []string{"web"},
					Middlewares: make([]string, 0),
					Service:     "neo-compass-qa.elong.com",
					Rule:        "Host(`dsfv2.t.17usoft.com`)",
					Priority:    0,
					TLS:         nil,
				}
				return router
			}(),
			Services: func() map[string]*dynamic.Service {
				ss := make(map[string]*dynamic.Service)
				ss["neo-compass-qa.elong.com"] = &dynamic.Service{
					LoadBalancer: &dynamic.ServersLoadBalancer{
						Servers: []dynamic.Server{dynamic.Server{
							URL: "127.0.0.1:8080",
						}},
					},
					Weighted:  nil,
					Mirroring: nil,
				}
				return ss
			}(),
			Middlewares: make(map[string]*dynamic.Middleware),
		},
	}
	return conf
}

func TestProvider_updateRule(t *testing.T) {
	type fields struct {
		GRPCAddress               string
		RestAddress               string
		DebugLogGeneratedTemplate bool
		Configuration             *dynamic.Configuration
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{name: "测试规则更新", fields: fields{
			RestAddress:   "http://10.160.137.160:8083",
			Configuration: NewCompassDynamicConfig(),
		},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Provider{
				GRPCAddress:               tt.fields.GRPCAddress,
				RestAddress:               tt.fields.RestAddress,
				DebugLogGeneratedTemplate: tt.fields.DebugLogGeneratedTemplate,
				Configuration:             tt.fields.Configuration,
			}
			p.updateRule()
		})
	}
}
