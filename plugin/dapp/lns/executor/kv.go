package executor

import (
	"github.com/33cn/chain33/system/dapp"
)

/*
 * 用户合约存取kv数据时，key值前缀需要满足一定规范
 * 即key = keyPrefix + userKey
 * 需要字段前缀查询时，使用’-‘作为分割符号
 */

var (
	//KeyPrefixStateDB state db key必须前缀
	KeyPrefixStateDB = "mavl-lns-"
	//KeyPrefixLocalDB local db的key必须前缀
	KeyPrefixLocalDB = "LODB-lns-"
)

func calcLnsChannelCountKey() []byte {
	return []byte(KeyPrefixStateDB + "channelIDCount")
}

func calcLnsChannelIDKey(channelID int64) []byte {

	return []byte(KeyPrefixStateDB + "channelID-" + dapp.HeightIndexStr(channelID, 0))
}

func calcLnsLocalChannelIDKey(channelID int64) []byte {

	return []byte(KeyPrefixLocalDB + "channelID-" + dapp.HeightIndexStr(channelID, 0))
}