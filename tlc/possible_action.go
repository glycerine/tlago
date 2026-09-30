package tlc

type PossibleAction struct {
	*Action
	UserPredicate *OpDefNode
}

func NewPossibleAction(pred SemanticNode, con *Context, userPredicate *OpDefNode) *PossibleAction {
	name := ""
	if userPredicate != nil && userPredicate.Name != nil {
		name = userPredicate.Name.String()
	}
	return &PossibleAction{
		Action:        NewAction(pred, con, name),
		UserPredicate: userPredicate,
	}
}

func (a *PossibleAction) GetPred() SemanticNode {
	if a != nil && a.UserPredicate != nil {
		return a.UserPredicate.Body
	}
	if a == nil || a.Action == nil {
		return nil
	}
	return a.Action.GetPred()
}
