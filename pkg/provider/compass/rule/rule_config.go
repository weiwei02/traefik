package rule

// 规则配置对象
//@author Wang Weiwei
//@since 2019/11/18

// 服务实例规则配置
// 每一个实例都应该最少匹配一个规则
type ServiceInstanceConfig struct {
	// 服务名
	ServiceName string `json:"serviceName"`
	//实例名,此配置作用于哪些实例,如果是all代表作用于所有实例
	Instance string `json:"instance"`
	// 作为调用端的配置
	Client *RulePointConfig `json:"client"`
	//作为服务端的配置
	Server *RulePointConfig `json:"server"`
}


// 补齐配置对象
func (config *ServiceInstanceConfig) fillConfig() {
	if config.Client == nil {
		config.Client = defaultRulePointConfig()
	}
	if config.Client.GlobalConfig == nil {
		config.Client.GlobalConfig = defaultRulePointGlobalConfig()
	}
	if config.Client.ApiConfig == nil {
		config.Client.ApiConfig = make(map[string]*RuleApiConfig)
	}
	if config.Server == nil {
		config.Server = defaultRulePointConfig()
	}
	if config.Server.GlobalConfig == nil {
		config.Server.GlobalConfig = defaultRulePointGlobalConfig()
	}
	if config.Server.ApiConfig == nil {
		config.Server.ApiConfig = make(map[string]*RuleApiConfig)
	}

}

// 端点规则配置
// 端点规则配置包括客户端规则配置或服务端规则配置
type RulePointConfig struct {
	//整体配置
	GlobalConfig *RulePointGlobalConfig `json:"globalConfig"`
	// api 配置
	ApiConfig map[string]*RuleApiConfig `json:"apiConfig"`
}

func defaultRulePointConfig() *RulePointConfig {
	return &RulePointConfig{
		GlobalConfig: defaultRulePointGlobalConfig(),
		ApiConfig:    make(map[string]*RuleApiConfig),
	}
}

// 端点整体配置，如果没有api级配置，应使用这个全局配置
type RulePointGlobalConfig struct {
	// 超时时间， 默认为5000ms
	Timeout string `json:"timeout"`
	// 重试次数 默认为0
	Retry string `json:"retry"`
	// 最大请求数量(qps)，-1 代表不限流
	MaxRequestCount string `json:"maxRequestCount"`
	// naming 环境配置
	Naming *RuleNamingConfig `json:"naming"`
}

// naming 中的服务配置
type RuleNamingConfig struct {
	// 机器环境信息
	Env []string `json:"env"`
	// 节点是否可用 0 ：否；1：是
	IsActive string `json:"isActive"`
	// 是否在负载中 0 ：否；1：是
	IsLoad string `json:"isLoad"`
}

// api级别配置
type RuleApiConfig struct {
	RulePointGlobalConfig
	// 请求的最小版本(客户端配置,请求服务的最小版本(naming))
	Version string `json:"version"`
	// 熔断配置
	CircuitBreak *RuleCircuitBreakConfig `json:"circuitBreak"`
}

// 熔断配置
type RuleCircuitBreakConfig struct {
	// 开启熔断  0：关闭；1：开启
	SwitchFlag string `json:"switchFlag"`
	// 触发熔断最小请求次数,小于该值不会触发熔断，默认为10
	InitCount string `json:"initCount"`
	// 熔断窗口，默认为60秒
	ComputeWindow string `json:"computeWindow"`
	// 熔断错误阈值,默认错误率达到100%时才熔断
	ErrorPercent string `json:"errorPercent"`
	// 熔断后返回内容
	DefaultValue string `json:"defaultValue"`
}

// 默认熔断配置
func DefaultRuleCircuitBreakConfig() RuleCircuitBreakConfig {
	return RuleCircuitBreakConfig{
		SwitchFlag:    "1",
		InitCount:     "10",
		ErrorPercent:  "1",
		ComputeWindow: "60",
		DefaultValue:  ""}
}

// 判断熔断限流的配置是否默认配置
func (config *RuleCircuitBreakConfig) IsDefault() bool {
	return config.SwitchFlag == "0" && config.ComputeWindow == "60" &&
		config.ErrorPercent == "1" && config.DefaultValue == ""
}


func defaultRulePointGlobalConfig() *RulePointGlobalConfig {
	return &RulePointGlobalConfig{
		Timeout:         "15000",
		Retry:           "1",
		MaxRequestCount: "-1",
	}
}

// 默认的服务配置
// 当服务没有配置时，应该使用这个默认配置
func DefaultRuleConfig(serviceName string) []*ServiceInstanceConfig {
	rule := ServiceInstanceConfig{
		ServiceName: serviceName,
		Instance:    "all",
		Client:      defaultRulePointConfig(),
		Server:      defaultRulePointConfig(),
	}
	return []*ServiceInstanceConfig{&rule}
}


// 规则结果处理器
// 规则查询结果
type RuleQueryResult struct {
	Client   []*ServiceInstanceConfig
	Server   []*ServiceInstanceConfig
	Searcher *RuleConfigSearcher
}

type RuleInvokerType struct {
	Timeout  string
	Retry    string
	Version  string
	Env      []string
	Instance string
	// api 映射
	RouterMap    string
	Map          string
	Protocol     string
	Platform     string
	CircuitBreak RuleCircuitBreakConfig
	LimitRate	string
}

// 将端点配置转化成规则类型
func convertConfigToRuleType(pointConfig *RulePointConfig, api string) *RuleInvokerType {
	ruleType := &RuleInvokerType{
		Timeout:      pointConfig.GlobalConfig.Timeout,
		Retry:        pointConfig.GlobalConfig.Retry,
		LimitRate:		pointConfig.GlobalConfig.MaxRequestCount,
		CircuitBreak: DefaultRuleCircuitBreakConfig(),
	}
	if pointConfig.GlobalConfig.Naming != nil {
		ruleType.Env = pointConfig.GlobalConfig.Naming.Env
	}
	if pointConfig.ApiConfig == nil {
		return ruleType
	}
	apiConfig, ok := pointConfig.ApiConfig[api]
	if ok {
		if apiConfig.CircuitBreak != nil {
			ruleType.CircuitBreak = *apiConfig.CircuitBreak
		}
		ruleType.Version = apiConfig.Version
		if apiConfig.Timeout != "" {
			ruleType.Timeout = apiConfig.Timeout
		}
		if apiConfig.Retry != "" {
			ruleType.Retry = apiConfig.Retry
		}
		if apiConfig.MaxRequestCount != "" {
			ruleType.LimitRate = apiConfig.MaxRequestCount
		}
	}
	return ruleType
}
