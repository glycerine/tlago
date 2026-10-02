package tlc

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf16"
)

type UniqueString struct {
	s   string
	tok int
	loc int
}

var internTable = NewInternTable(1024)

func UniqueStringInitialize() {
	UniqueStringInitializeWithSource(nil)
}

func UniqueStringOf(s string) *UniqueString {
	return internTable.Put(s)
}

func UniqueStringByToken(tok int) *UniqueString {
	return internTable.Get(tok)
}

func SetUniqueStringVariableCount(n int) {
	internTable.mu.Lock()
	defer internTable.mu.Unlock()
	internTable.varCount = n
}

func ResetUniqueStringLocations() {
	internTable.mu.Lock()
	defer internTable.mu.Unlock()
	for _, us := range internTable.table {
		if us != nil {
			us.loc = -1
		}
	}
	internTable.varCount = 0
}

func UniqueStringVariableCount() int {
	internTable.mu.Lock()
	defer internTable.mu.Unlock()
	return internTable.varCount
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
	return int(int32(u.tok) - int32(other.tok))
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

func (t *InternTable) BeginChkpt(metadir string) error {
	return t.beginChkptWithVarCount(metadir, UniqueStringVariableCount())
}

func (t *InternTable) beginChkptWithVarCount(metadir string, varCount int) error {
	if t == nil {
		panic(NewNullPointerException())
	}
	path := uniqueStringChkptName(metadir, "tmp")
	file, err := os.Create(path)
	if err != nil {
		return distributedFileOpenException(path, err)
	}
	out := NewValueOutputStream(file)
	t.dataMu.RLock()
	tokenCnt := t.tokenCnt
	t.dataMu.RUnlock()
	if err := out.WriteInt(tokenCnt); err != nil {
		_ = out.Close()
		return err
	}
	for i := 0; ; i++ {
		us, present := t.slotAt(i)
		if !present {
			break
		}
		if us == nil {
			continue
		}
		if err := writeJavaUniqueStringWithVarCount(out, us, varCount); err != nil {
			_ = out.Close()
			return err
		}
	}
	return out.Close()
}

func (t *InternTable) CommitChkpt(metadir string) error {
	if t == nil {
		panic(NewNullPointerException())
	}
	oldChkpt := uniqueStringChkptName(metadir, "chkpt")
	newChkpt := uniqueStringChkptName(metadir, "tmp")
	if err := os.Remove(oldChkpt); err != nil && !errors.Is(err, os.ErrNotExist) {
		return NewIOException(fmt.Sprintf("InternTable.commitChkpt: cannot delete %s", oldChkpt))
	}
	if err := os.Rename(newChkpt, oldChkpt); err != nil {
		return NewIOException(fmt.Sprintf("InternTable.commitChkpt: cannot delete %s", oldChkpt))
	}
	return nil
}

func (t *InternTable) Recover(metadir string) error {
	if t == nil {
		panic(NewNullPointerException())
	}
	file, err := os.Open(uniqueStringChkptName(metadir, "chkpt"))
	if err != nil {
		return distributedFileOpenException(uniqueStringChkptName(metadir, "chkpt"), err)
	}
	reader := bufio.NewReader(file)
	in := NewValueInputStream(reader)
	defer file.Close()
	tokenCnt, err := in.ReadInt()
	if err != nil {
		_ = in.Close()
		return err
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	t.dataMu.Lock()
	t.tokenCnt = tokenCnt
	t.dataMu.Unlock()
	for {
		if _, err := reader.Peek(1); err == io.EOF {
			break
		} else if err != nil {
			return err
		}
		us, err := readJavaUniqueString(in)
		if errors.Is(err, io.EOF) {
			failure := newTLCErrorCodeNullable(ECSystemCheckpointRecoveryCorrupt, javaThrowableDetailMessage(err))
			failure.Runtime = true
			panic(failure)
		}
		if err != nil {
			_ = in.Close()
			return err
		}
		func() {
			t.dataMu.Lock()
			defer t.dataMu.Unlock()
			t.putValue(us)
		}()
	}
	return file.Close()
}

func writeJavaUniqueString(out *ValueOutputStream, us *UniqueString) error {
	varCount := UniqueStringVariableCount()
	return writeJavaUniqueStringWithVarCount(out, us, varCount)
}

func writeJavaUniqueStringWithVarCount(out *ValueOutputStream, us *UniqueString, varCount int) error {
	if us == nil {
		panic(NewNullPointerException())
	}
	if err := out.WriteInt(int32(us.tok)); err != nil {
		return err
	}
	loc := -1
	if us.loc < varCount {
		loc = us.loc
	}
	if err := out.WriteInt(int32(loc)); err != nil {
		return err
	}
	bytes := javaLegacyStringBytes(us.s)
	if err := out.WriteInt(int32(len(bytes))); err != nil {
		return err
	}
	_, err := out.WriteRaw(bytes)
	return err
}

func readJavaUniqueString(in *ValueInputStream) (*UniqueString, error) {
	tok, err := in.ReadInt()
	if err != nil {
		return nil, err
	}
	loc, err := in.ReadInt()
	if err != nil {
		return nil, err
	}
	length, err := in.ReadInt()
	if err != nil {
		return nil, err
	}
	bytes := make([]byte, valueStreamArrayLength(length))
	if err := in.ReadFully(bytes); err != nil {
		return nil, err
	}
	return &UniqueString{s: javaLegacyStringFromBytes(bytes), tok: int(tok), loc: int(loc)}, nil
}

func readExternalJavaUniqueString(in *ValueInputStream) (*UniqueString, error) {
	if _, err := in.ReadInt(); err != nil {
		return nil, err
	}
	if _, err := in.ReadInt(); err != nil {
		return nil, err
	}
	length, err := in.ReadInt()
	if err != nil {
		return nil, err
	}
	bytes := make([]byte, valueStreamArrayLength(length))
	if err := in.ReadFully(bytes); err != nil {
		return nil, err
	}
	return UniqueStringOf(javaLegacyStringFromBytes(bytes)), nil
}

func javaLegacyStringBytes(s string) []byte {
	units := utf16.Encode([]rune(s))
	bytes := make([]byte, len(units))
	for i, unit := range units {
		bytes[i] = byte(unit)
	}
	return bytes
}

func javaLegacyStringFromBytes(bytes []byte) string {
	runes := make([]rune, len(bytes))
	for i, b := range bytes {
		runes[i] = rune(uint16(int16(int8(b))))
	}
	return string(runes)
}

func uniqueStringChkptName(metadir string, ext string) string {
	// Java concatenates the separator even for an empty directory, and File
	// collapses repeated separators without resolving . or .. components.
	sep := string(filepath.Separator)
	path := metadir + sep + "vars." + ext
	prefix := ""
	if filepath.Separator == '\\' {
		path = strings.ReplaceAll(path, "/", sep)
		if strings.HasPrefix(path, sep+sep) {
			prefix, path = sep, path[1:]
		}
	}
	for strings.Contains(path, sep+sep) {
		path = strings.ReplaceAll(path, sep+sep, sep)
	}
	return prefix + path
}
