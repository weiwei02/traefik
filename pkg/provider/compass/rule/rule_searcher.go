package rule

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"time"
)

//@author Wang Weiwei
//@since 2019/11/18

// 搜索服务配置参数
type RuleConfigSearcher struct {
	ClientName string
	ClientIns  string
	ServerName string
	ServerIns  string
	// address of compass
	baseUri string
}

func NewRuleConfigSearcher(serverName, baseUri string) *RuleConfigSearcher {
	return &RuleConfigSearcher{
		ClientName: "neo.service.oceanxinjian1.neo.sidecar.k8s",
		ClientIns:  "[\"all\"]",
		ServerName: serverName,
		ServerIns:  "[\"all\"]",
		baseUri:    baseUri,
	}
}

// 搜索参数转化成远程请求所需参数
// 返回参数是compass配置查询接口所需的参数
func (searcher *RuleConfigSearcher) toRemoteParam() []*RemoteRuleQueryParam {
	var params = make([]*RemoteRuleQueryParam, 0)
	clientParam := RemoteRuleQueryParam{
		ServiceName: searcher.ClientName,
		Instance:    []string{searcher.ClientIns},
		Client:      []string{"*"},
		Server:      &RemoteRuleQueryParamServer{ApiConfig: []string{"*"}},
	}
	serverParam := RemoteRuleQueryParam{
		ServiceName: searcher.ServerName,
		Instance:    []string{searcher.ServerIns},
		Client:      []string{"*"},
		Server:      &RemoteRuleQueryParamServer{ApiConfig: []string{"*"}},
	}
	if searcher.ClientName != "" {
		params = append(params, &clientParam)
	}
	if searcher.ServerName != "" {
		params = append(params, &serverParam)
	}
	return params
}

func (searcher *RuleConfigSearcher) FindRule() (map[string]*RuleInvokerType, error) {
	invokerMap := make(map[string]*RuleInvokerType)
	result, err := searcher.findRuleConfig()
	if err != nil {
		return nil, err
	}
	for _, server := range result.Server {
		if server.Instance == "all" && server.Server != nil {
			for api, _ := range server.Server.ApiConfig {
				invokerMap[api] = convertConfigToRuleType(server.Server, api)
			}
			invokerMap["all"] = convertConfigToRuleType(server.Server, "")
		}
	}
	return invokerMap, err
}

// 同时查询客户端和服务端配置
func (searcher *RuleConfigSearcher) findRuleConfig() (*RuleQueryResult, error) {
	res, err := searcher.findFromRemote(searcher.toRemoteParam())
	if err != nil {
		res = DefaultRuleConfig(searcher.ServerName)
	}
	return &RuleQueryResult{
		Server:   res,
		Searcher: searcher,
	}, nil
}

// 远程到compass查询服务配置所需参数
type RemoteRuleQueryParam struct {
	ServiceName string `json:"serviceName"`
	//实例名,此配置作用于哪些实例,如果是all代表作用于所有实例
	Instance []string `json:"instance"`
	// 要获取调用策略配置的目标服务和接口
	Client []string `json:"client"`
	// 需要获取的作为服务端时的配置，该参数可以为空
	Server *RemoteRuleQueryParamServer `json:"server"`
}

type RemoteRuleQueryParamServer struct {
	ApiConfig []string `json:"apiConfig"`
}

// 从远程compass获取服务配置，并将服务配置缓存
func (searcher *RuleConfigSearcher) findFromRemote(param []*RemoteRuleQueryParam) ([]*ServiceInstanceConfig, error) {
	if len(param) == 0 {
		return nil, errors.New("规则查询参数不足")
	}
	url := searcher.baseUri + "/ruleConfig/query"
	body, err := json.Marshal(param)
	if err != nil {
		return nil, err
	} else {
		head := make(map[string]string)
		head["Last-Modified"] = ""

		request, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
		if err != nil {
			return nil, err
		}
		request.Header.Add("Content-Type", "application/json")
		request.Header.Add("Last-Modified", "")
		ctx, _ := context.WithTimeout(context.Background(), time.Second*5)
		request.WithContext(ctx)
		resp, err := http.DefaultClient.Do(request)
		if err != nil {
			return nil, err
		} else if resp.StatusCode != 200 {
			return nil, errors.New("http status is " + string(resp.StatusCode))
		}
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		configList := make([]*ServiceInstanceConfig, 0)
		err = json.Unmarshal(body, &configList)
		if err != nil {
			return nil, err
		}
		return configList, nil
	}
}
