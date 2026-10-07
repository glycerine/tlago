package tlago

import (
	"strconv"

	"github.com/glycerine/tlago/tlc"
)

// JunctionListContext tracks the source list's bullet type and alignment.
// Token columns are already supplied by the character stream.
type sanyJunctionListInfo struct {
	column int
	kind   SanyTokenKind
}

type sanyJunctionListContext struct {
	stack []sanyJunctionListInfo
}

func (c *sanyJunctionListContext) startNewJunctionList(column int, kind SanyTokenKind) {
	if !IsSanyJunctionBullet(kind) {
		panic(tlc.NewIllegalArgumentException(strconv.Itoa(int(kind))))
	}
	c.stack = append(c.stack, sanyJunctionListInfo{column, kind})
}

func (c *sanyJunctionListContext) terminateCurrentJunctionList() {
	if len(c.stack) == 0 {
		panic(tlc.NewNoSuchElementException())
	}
	c.stack = c.stack[:len(c.stack)-1]
}

func (c *sanyJunctionListContext) current() (sanyJunctionListInfo, bool) {
	if len(c.stack) == 0 {
		return sanyJunctionListInfo{}, false
	}
	return c.stack[len(c.stack)-1], true
}

func (c *sanyJunctionListContext) isNewBullet(column int, kind SanyTokenKind) bool {
	current, exists := c.current()
	return exists && IsSanyJunctionBullet(kind) && current.column == column && current.kind == kind
}

func (c *sanyJunctionListContext) isAboveCurrent(column int) bool {
	current, exists := c.current()
	return !exists || current.column < column
}
