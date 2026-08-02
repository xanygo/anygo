package xkv

// DataType 数据类型，在  Monitor.After 中会用到
type DataType string

const (
	DataTypeString DataType = "String"
	DataTypeList   DataType = "List"
	DataTypeSet    DataType = "Set"
	DataTypeHash   DataType = "Hash"
	DataTypeZSet   DataType = "ZSet"
	DataTypeKey    DataType = "Key"
)

const (
	actionGet         = "Get"
	actionSet         = "Set"
	actionSetNX       = "SetNX"
	actionGetDel      = "GetDel"
	actionGetSet      = "GetSet"
	actionIncr        = "Incr"
	actionIncrBy      = "IncrBy"
	actionIncrByFloat = "IncrByFloat"
	actionDecr        = "Decr"
)

const (
	actionLPush  = "LPush"
	actionRPush  = "RPush"
	actionLPop   = "LPop"
	actionLPopN  = "LPopN"
	actionRPop   = "RPop"
	actionRPopN  = "RPopN"
	actionLRem   = "LRem"
	actionRange  = "Range"
	actionLRange = "LRange"
	actionRRange = "RRange"
	actionLLen   = "LLen"
)

const (
	actionHSet    = "HSet"
	actionHMSet   = "HMSet"
	actionHGet    = "HGet"
	actionHMGet   = "HMGet"
	actionHDel    = "HDel"
	actionHRange  = "HRange"
	actionHGetAll = "HGetAll"
	actionHExists = "HExists"
	actionHIncrBy = "HIncrBy"
	actionHLen    = "HLen"
)

const (
	actionSAdd         = "SAdd"
	actionSRem         = "SRem"
	actionSRange       = "SRange"
	actionSMembers     = "SMembers"
	actionSCard        = "SCard"
	actionSIsMember    = "SIsMember"
	actionSMIsMember   = "SMIsMember"
	actionSPop         = "SPop"
	actionSPopN        = "SPopN"
	actionSRandMember  = "SRandMember"
	actionSRandMemberN = "SRandMemberN"
)

const (
	actionZAdd             = "ZAdd"
	actionZScore           = "ZScore"
	actionZIncrBy          = "ZIncrBy"
	actionZRange           = "ZRange"
	actionZRangeByScore    = "ZRangeByScore"
	actionZRem             = "ZRem"
	actionZRemRangeByScore = "ZRemRangeByScore"
	actionZCount           = "ZCount"
	actionZLen             = "ZLen"
	actionZRank            = "ZRank"
	actionZPopMax          = "ZPopMax"
	actionZPopMin          = "ZPopMin"
)

const (
	actionDelete = "Delete"
	actionHas    = "Has"
)

// IsReadAction 可用于 Monitor.After 回调中，判断执行动作的类型是否是只读的
func IsReadAction(dataType DataType, action string) bool {
	switch dataType {
	case DataTypeString:
		switch action {
		case actionGet:
			return true
		}
	case DataTypeList:
		switch action {
		case actionRange, actionLRange, actionRRange, actionLLen:
			return true
		}
	case DataTypeHash:
		switch action {
		case actionHGet, actionHMGet, actionHRange, actionHGetAll, actionHExists, actionHLen:
			return true
		}
	case DataTypeSet:
		switch action {
		case actionSRange, actionSMembers, actionSCard, actionSIsMember, actionSMIsMember, actionSRandMember, actionSRandMemberN:
			return true
		}
	case DataTypeZSet:
		switch action {
		case actionZScore, actionZRange, actionZRangeByScore, actionZCount, actionZLen, actionZRank:
			return true
		}
	case DataTypeKey:
		switch action {
		case actionHas:
			return true
		}
	}
	return false
}

func IsDeleteAction(dataType DataType, action string) bool {
	switch dataType {
	case DataTypeString:
		switch action {
		case actionDelete:
			return true
		}
	case DataTypeList:
		switch action {
		case actionLRem, actionLPop, actionLPopN:
			return true
		}
	case DataTypeHash:
		switch action {
		case actionHDel:
			return true
		}
	case DataTypeSet:
		switch action {
		case actionSRem, actionSPop, actionSPopN:
			return true
		}
	case DataTypeZSet:
		switch action {
		case actionZRem, actionZPopMin, actionZPopMax:
			return true
		}
	case DataTypeKey:
		switch action {
		case actionDelete:
			return true
		}
	}
	return false
}
