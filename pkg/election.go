package cheek

import (
	"fmt"
	"github.com/hashicorp/consul/api"
	election "github.com/joe-at-startupmedia/consul-leader-election"
	"time"
)

type notify struct {
	T string
}

type ElectionInterface interface {
	IsLeader() bool
	Stop()
}

func (n *notify) EventLeader(f bool) {
	if f {
		fmt.Println(n.T, "I'm the leader!")
	} else {
		fmt.Println(n.T, "I'm no longer the leader!")
	}
}

func elector(s *Schedule) *election.Election {

	conf := api.DefaultConfig()
	if len(s.ConsulAclToken) > 0 {
		conf.Token = s.ConsulAclToken
	}

	consul, _ := api.NewClient(conf)
	n := &notify{
		T: "cheek-turner",
	}

	sessionKey := "service/cheek-turner-election"

	if len(s.ConsulSessionKey) > 0 {
		sessionKey = s.ConsulSessionKey
	}

	e := election.NewElection(&election.ElectionConfig{
		CheckTimeout: 5 * time.Second,
		Client:       consul,
		Key:          sessionKey + "/leader",
		LogLevel:     election.LogDebug,
		Event:        n,
	})

	go e.Init()

	return e
}
