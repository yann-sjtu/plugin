package executor

import (
	"testing"

	"github.com/33cn/chain33/types"
	lnstypes "github.com/33cn/plugin/plugin/dapp/lns/types"
)

func Test_test(t *testing.T) {

	lnsAction1 := &lnstypes.LnsAction{}
	var open *lnstypes.LnsAction_Open
	lnsAction1.Value = open

	msg := types.Encode(lnsAction1)

	lnsAction := &lnstypes.LnsAction{}
	_ = types.Decode(msg, lnsAction)

	if _, ok := lnsAction.GetValue().(*lnstypes.LnsAction_Open); ok {
		println("ok")
	}
	if lnsAction.GetOpen() == nil {
		println("nil")
	}

}
