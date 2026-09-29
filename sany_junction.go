package tlago

import "fmt"

type SanyJunctionType int

const (
	SanyJunctionConjunction SanyJunctionType = iota
	SanyJunctionDisjunction
)

type SanyJunctionListInfo struct {
	Type   SanyJunctionType
	Column int
}

type SanyJunctionListContext struct {
	stack []SanyJunctionListInfo
}

func IsSanyJunctionBullet(kind SanyTokenKind) bool {
	return kind == SanyTokenAND || kind == SanyTokenOR
}

func (c *SanyJunctionListContext) StartNewJunctionList(column int, kind SanyTokenKind) error {
	typ, err := sanyJunctionType(kind)
	if err != nil {
		return err
	}
	c.stack = append(c.stack, SanyJunctionListInfo{Type: typ, Column: column})
	return nil
}

func (c *SanyJunctionListContext) TerminateCurrentJunctionList() error {
	if len(c.stack) == 0 {
		return fmt.Errorf("not inside a junction list")
	}
	c.stack = c.stack[:len(c.stack)-1]
	return nil
}

func (c *SanyJunctionListContext) IsNewBullet(column int, kind SanyTokenKind) bool {
	if len(c.stack) == 0 || !IsSanyJunctionBullet(kind) {
		return false
	}
	head := c.stack[len(c.stack)-1]
	typ, err := sanyJunctionType(kind)
	return err == nil && head.Column == column && head.Type == typ
}

func (c *SanyJunctionListContext) IsAboveCurrent(column int) bool {
	if len(c.stack) == 0 {
		return true
	}
	return c.stack[len(c.stack)-1].Column < column
}

func sanyJunctionType(kind SanyTokenKind) (SanyJunctionType, error) {
	switch kind {
	case SanyTokenAND:
		return SanyJunctionConjunction, nil
	case SanyTokenOR:
		return SanyJunctionDisjunction, nil
	default:
		return 0, fmt.Errorf("%s is not a junction-list bullet", kind.JavaName())
	}
}
