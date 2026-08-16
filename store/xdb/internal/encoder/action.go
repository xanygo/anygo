package encoder

type Action uint8

const (
	ActionUnset Action = iota
	ActionInsert
	ActionUpdate
	ActionUpsert
	ActionDelete
	ActionSelect
	ActionOther
)

func (a Action) IsInsert() bool {
	switch a {
	case ActionInsert, ActionUpsert:
		return true
	default:
		return false
	}
}

func (a Action) IsUpdate() bool {
	switch a {
	case ActionUpdate, ActionUpsert:
		return true
	default:
		return false
	}
}

func (a Action) IsInsertOrUpdate() bool {
	switch a {
	case ActionInsert, ActionUpsert, ActionUpdate:
		return true
	default:
		return false
	}
}
