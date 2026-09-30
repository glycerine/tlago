package tlc

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"unicode/utf16"
)

type UniqueString struct {
	s   string
	tok int
	loc int
}

type uniqueStringTable struct {
	mu       sync.Mutex
	byString map[string]*UniqueString
	byToken  map[int]*UniqueString
	tokenCnt int
	varCount int
}

var internTable = &uniqueStringTable{
	byString: make(map[string]*UniqueString),
	byToken:  make(map[int]*UniqueString),
}

func UniqueStringInitialize() {
	internTable.mu.Lock()
	internTable.byString = make(map[string]*UniqueString)
	internTable.byToken = make(map[int]*UniqueString)
	internTable.tokenCnt = 0
	internTable.varCount = 0
	internTable.mu.Unlock()
	initBuiltInOPs()
}

func UniqueStringOf(s string) *UniqueString {
	internTable.mu.Lock()
	defer internTable.mu.Unlock()
	if us := internTable.byString[s]; us != nil {
		return us
	}
	internTable.tokenCnt++
	us := &UniqueString{s: s, tok: internTable.tokenCnt, loc: -1}
	internTable.byString[s] = us
	internTable.byToken[us.tok] = us
	return us
}

func UniqueStringByToken(tok int) *UniqueString {
	internTable.mu.Lock()
	defer internTable.mu.Unlock()
	return internTable.byToken[tok]
}

func SetUniqueStringVariableCount(n int) {
	internTable.mu.Lock()
	defer internTable.mu.Unlock()
	internTable.varCount = n
}

func UniqueStringJoin(delim string, values ...*UniqueString) *UniqueString {
	parts := make([]string, 0, len(values))
	for _, value := range values {
		if value != nil {
			parts = append(parts, value.String())
		}
	}
	return UniqueStringOf(strings.Join(parts, "!"))
}

func (u *UniqueString) String() string {
	if u == nil {
		return ""
	}
	return u.s
}

func (u *UniqueString) Token() int {
	if u == nil {
		return 0
	}
	return u.tok
}

func (u *UniqueString) SetLoc(loc int) {
	if u != nil {
		u.loc = loc
	}
}

func (u *UniqueString) VarLoc() int {
	if u == nil {
		return -1
	}
	internTable.mu.Lock()
	defer internTable.mu.Unlock()
	if u.loc < internTable.varCount {
		return u.loc
	}
	return -1
}

func (u *UniqueString) DefnLoc() int {
	if u == nil {
		return -1
	}
	internTable.mu.Lock()
	defer internTable.mu.Unlock()
	if u.loc < internTable.varCount {
		return -1
	}
	return u.loc
}

func (u *UniqueString) Length() int {
	if u == nil {
		return 0
	}
	return len(utf16.Encode([]rune(u.s)))
}

func (u *UniqueString) Compare(other *UniqueString) int {
	if u == nil {
		if other == nil {
			return 0
		}
		return -other.tok
	}
	if other == nil {
		return u.tok
	}
	return u.tok - other.tok
}

func (u *UniqueString) Equal(other *UniqueString) bool {
	if u == nil || other == nil {
		return u == other
	}
	return u.tok == other.tok
}

func (u *UniqueString) FingerPrint(fp uint64) uint64 {
	if u == nil {
		return fp
	}
	return FP64ExtendInt(fp, int32(u.tok))
}

func (u *UniqueString) StartsWith(prefix string) bool {
	return u != nil && strings.HasPrefix(u.s, prefix)
}

func (u *UniqueString) Substring(begin int) string {
	if u == nil {
		return ""
	}
	return u.s[begin:]
}

func BeginChkptUniqueStrings(metadir string) error {
	return internTable.BeginChkpt(metadir)
}

func CommitChkptUniqueStrings(metadir string) error {
	return internTable.CommitChkpt(metadir)
}

func RecoverUniqueStrings(metadir string) error {
	return internTable.Recover(metadir)
}

func (t *uniqueStringTable) BeginChkpt(metadir string) error {
	if t == nil || metadir == "" {
		return nil
	}
	path := uniqueStringChkptName(metadir, "tmp")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	out := NewValueOutputStream(file)
	t.mu.Lock()
	defer t.mu.Unlock()
	if err := out.WriteInt(int32(t.tokenCnt)); err != nil {
		_ = out.Close()
		return err
	}
	if err := out.WriteInt(int32(t.varCount)); err != nil {
		_ = out.Close()
		return err
	}
	count := int32(len(t.byToken))
	if err := out.WriteInt(count); err != nil {
		_ = out.Close()
		return err
	}
	for tok := 1; tok <= t.tokenCnt; tok++ {
		us := t.byToken[tok]
		if us == nil {
			continue
		}
		if err := out.WriteInt(int32(us.tok)); err != nil {
			_ = out.Close()
			return err
		}
		if err := out.WriteInt(int32(us.loc)); err != nil {
			_ = out.Close()
			return err
		}
		if err := out.WriteUniqueString(us); err != nil {
			_ = out.Close()
			return err
		}
	}
	return out.Close()
}

func (t *uniqueStringTable) CommitChkpt(metadir string) error {
	if t == nil || metadir == "" {
		return nil
	}
	oldChkpt := uniqueStringChkptName(metadir, "chkpt")
	newChkpt := uniqueStringChkptName(metadir, "tmp")
	if err := os.Remove(oldChkpt); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return os.Rename(newChkpt, oldChkpt)
}

func (t *uniqueStringTable) Recover(metadir string) error {
	if t == nil || metadir == "" {
		return nil
	}
	file, err := os.Open(uniqueStringChkptName(metadir, "chkpt"))
	if err != nil {
		return err
	}
	in := NewValueInputStream(file)
	tokenCnt, err := in.ReadInt()
	if err != nil {
		_ = in.Close()
		return err
	}
	varCount, err := in.ReadInt()
	if err != nil {
		_ = in.Close()
		return err
	}
	count, err := in.ReadInt()
	if err != nil {
		_ = in.Close()
		return err
	}
	byString := make(map[string]*UniqueString, int(count))
	byToken := make(map[int]*UniqueString, int(count))
	for i := int32(0); i < count; i++ {
		tok, err := in.ReadInt()
		if err != nil {
			_ = in.Close()
			return err
		}
		loc, err := in.ReadInt()
		if err != nil {
			_ = in.Close()
			return err
		}
		str, err := in.readExternalUniqueString()
		if err != nil {
			_ = in.Close()
			return err
		}
		us := &UniqueString{s: str.String(), tok: int(tok), loc: int(loc)}
		byString[us.s] = us
		byToken[us.tok] = us
	}
	if err := in.Close(); err != nil {
		return err
	}
	t.mu.Lock()
	t.byString = byString
	t.byToken = byToken
	t.tokenCnt = int(tokenCnt)
	t.varCount = int(varCount)
	t.mu.Unlock()
	return nil
}

func uniqueStringChkptName(metadir string, ext string) string {
	return filepath.Join(metadir, "vars."+ext)
}
