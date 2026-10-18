package tlago

import (
	"strconv"

	"github.com/glycerine/tlago/tlc"
)

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

func (c *SanyJunctionListContext) require() {
	if c == nil {
		panic(tlc.NewNullPointerException())
	}
}

func IsSanyJunctionBullet(kind SanyTokenKind) bool {
	return kind == SanyTokenAND || kind == SanyTokenOR
}

func (c *SanyJunctionListContext) StartNewJunctionList(column int, kind SanyTokenKind) {
	c.require()
	typ := sanyJunctionType(kind)
	c.stack = append(c.stack, SanyJunctionListInfo{Type: typ, Column: column})
}

func (c *SanyJunctionListContext) TerminateCurrentJunctionList() {
	c.require()
	if len(c.stack) == 0 {
		panic(tlc.NewNoSuchElementException())
	}
	c.stack = c.stack[:len(c.stack)-1]
}

func (c *SanyJunctionListContext) IsNewBullet(column int, kind SanyTokenKind) bool {
	c.require()
	if len(c.stack) == 0 || !IsSanyJunctionBullet(kind) {
		return false
	}
	head := c.stack[len(c.stack)-1]
	return head.Column == column && head.Type == sanyJunctionType(kind)
}

func (c *SanyJunctionListContext) IsAboveCurrent(column int) bool {
	c.require()
	if len(c.stack) == 0 {
		return true
	}
	return c.stack[len(c.stack)-1].Column < column
}

func (c *SanyJunctionListContext) current() (SanyJunctionListInfo, bool) {
	c.require()
	if len(c.stack) == 0 {
		return SanyJunctionListInfo{}, false
	}
	return c.stack[len(c.stack)-1], true
}

func sanyJunctionType(kind SanyTokenKind) SanyJunctionType {
	switch kind {
	case SanyTokenAND:
		return SanyJunctionConjunction
	case SanyTokenOR:
		return SanyJunctionDisjunction
	default:
		panic(tlc.NewIllegalArgumentException(strconv.Itoa(int(kind))))
	}
}
