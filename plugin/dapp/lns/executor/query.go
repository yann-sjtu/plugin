// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"github.com/33cn/chain33/types"
	lnstypes "github.com/33cn/plugin/plugin/dapp/lns/types"
)

//
func (l *lns) Query_GetChannel(in *lnstypes.ReqGetChannel) (types.Message, error) {

	channel := &lnstypes.Channel{}
	chanKey := calcLnsChannelIDKey(in.ChannelID)

	err := getDBAndDecode(l.GetStateDB(), chanKey, channel)
	if err != nil {
		elog.Error("Query_GetChannelChannel", "GetChannelErr", err)
		return nil, err
	}
	return channel, nil
}

func (l *lns) Query_GetChannelCount(reqNil *types.ReqNil) (types.Message, error) {

	chanCount := &lnstypes.ChannelCount{}
	countKey := calcLnsChannelCountKey()

	err := getDBAndDecode(l.GetStateDB(), countKey, chanCount)

	if err != nil {
		elog.Error("Query_GetChannelCountChannel", "GetChannelCountErr", err)
		return nil, err
	}
	return chanCount, nil
}
