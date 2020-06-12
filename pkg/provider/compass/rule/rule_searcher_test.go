package rule

import (
	"github.com/stretchr/testify/require"
	"testing"
)

//@author Wang Weiwei
//@since 2020/6/12

func TestFindRuleConfig(t *testing.T) {
	searcher := NewRuleConfigSearcher("neo.service.oceanxinjian1.neo.mix.server", "http://")
	config, err := searcher.findRuleConfig()
	require.NoError(t, err)
	t.Log(config)
}

func TestFindRule(t *testing.T) {
	searcher := NewRuleConfigSearcher("neo.service.oceanxinjian1.neo.mix.server", "http://")
	config, err := searcher.FindRule()
	require.NoError(t, err)
	require.Truef(t, len(config) > 0, "查询规则数据为空")
	t.Log(config)
}
