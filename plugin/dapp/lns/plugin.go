// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package lottery

import (
	"github.com/33cn/chain33/pluginmgr"
	"github.com/33cn/plugin/plugin/dapp/lns/commands"
	"github.com/33cn/plugin/plugin/dapp/lns/executor"
	"github.com/33cn/plugin/plugin/dapp/lns/rpc"
	"github.com/33cn/plugin/plugin/dapp/lns/types"
)

func init() {
	pluginmgr.Register(&pluginmgr.PluginBase{
		Name:     types.LnsX,
		ExecName: executor.GetName(),
		Exec:     executor.Init,
		Cmd:      commands.Cmd,
		RPC:      rpc.Init,
	})
}
