package compass

import (
	"fmt"
	"github.com/containous/traefik/v2/pkg/config/dynamic"
	"github.com/containous/traefik/v2/pkg/log"
	"github.com/containous/traefik/v2/pkg/provider/compass/rule"
	"github.com/containous/traefik/v2/pkg/types"
	"strconv"
	"strings"
	"time"
)

//@author Wang Weiwei
//@since 2020/6/12

func (p *Provider) CompassRuleWatcher(configurationChan chan<- dynamic.Message) {
	ticker := time.NewTicker(time.Second * 30)
	for {
		select {
		case _ = <-ticker.C:
			{
				p.updateRule()
				configurationChan <- dynamic.Message{
					ProviderName:  providerName,
					Configuration: p.Configuration.DeepCopy(),
				}

			}

		}
	}
}

func (p *Provider) updateRule() {
	for key, _ := range p.Configuration.HTTP.Services {
		ruleMap, err := rule.NewRuleConfigSearcher(key, p.RestAddress).FindRule()
		if err != nil {
			log.WithoutContext().WithField(log.ProviderName, providerName).Errorf("query rule %s  error: %s", key, err)
			continue
		}
		for path, ruleType := range ruleMap {
			// rate limit middleware
			if t, err := strconv.Atoi(ruleType.LimitRate); err == nil && t > 0 {
				t := int64(t)
				middlewareName := "limitRate-" + ruleType.LimitRate
				if _, ok := p.Configuration.HTTP.Middlewares[middlewareName]; !ok {
					p.Configuration.HTTP.Middlewares[middlewareName] =
						&dynamic.Middleware{RateLimit: &dynamic.RateLimit{
							Average:         t,
							Period:          types.Duration(time.Second),
							Burst:           t,
							SourceCriterion: &dynamic.SourceCriterion{RequestHost: true},
						}}
				}
				p.addMiddlewareToRouter(key, path, middlewareName)
			}

			// retry middleware
			if t, err := strconv.Atoi(ruleType.Retry); err == nil && t > 0 {
				middlewareName := "retry-" + ruleType.Retry
				if _, ok := p.Configuration.HTTP.Middlewares[middlewareName]; !ok {
					p.Configuration.HTTP.Middlewares[middlewareName] =
						&dynamic.Middleware{Retry: &dynamic.Retry{Attempts: t}}
				}
				p.addMiddlewareToRouter(key, path, middlewareName)
			}

			// circuit breaker middleware
			if ruleType.CircuitBreak.SwitchFlag == "1" {
				if !strings.Contains(ruleType.CircuitBreak.ErrorPercent, ".") {
					ruleType.CircuitBreak.ErrorPercent += ".0"
				}
				middlewareName := "CircuitBreak-" + ruleType.CircuitBreak.ErrorPercent
				if _, ok := p.Configuration.HTTP.Middlewares[middlewareName]; !ok {
					p.Configuration.HTTP.Middlewares[middlewareName] =
						&dynamic.Middleware{CircuitBreaker: &dynamic.CircuitBreaker{
							Expression: fmt.Sprintf("NetworkErrorRatio() >= %s || ResponseCodeRatio(500, 600, 0, 600) >= %s", ruleType.CircuitBreak.ErrorPercent, ruleType.CircuitBreak.ErrorPercent),
						}}
				}
				p.addMiddlewareToRouter(key, path, middlewareName)
			}
		}
	}
}

// add middleware to router
func (p *Provider) addMiddlewareToRouter(serviceName string, path string, middlewareName string) {

	for routerKey, router := range p.Configuration.HTTP.Routers {
		if router.Service == serviceName {
			if path != defaultPath {
				if strings.HasSuffix(routerKey, path) {
					router.Middlewares = append(router.Middlewares, middlewareName)
				}
			} else {
				router.Middlewares = append(router.Middlewares, middlewareName)
			}
		}
	}
}
