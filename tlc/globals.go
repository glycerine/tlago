package tlc

import (
	"os"
	"runtime/debug"
	"strconv"
	"sync"
	"time"
)

const DefaultCheckpointDurationMillis = (30 * 60 * 1000) + 42
const DefaultProgressIntervalMillis = 60 * 1000
const CoverageIndent = '|'
const MetaRoot = "states"

var Globals = struct {
	sync.Mutex
	EnumBound                  int
	SetBound                   int
	NumWorkers                 int
	LivenessThreshold          float64
	LivenessGraphSizeThreshold float64
	LivenessRatio              float64
	LNCheck                    string
	ProgressIntervalMillis     int64
	CoverageInterval           int
	CoverageFlags              int
	DFIDMax                    int
	Continuation               bool
	Expand                     bool
	PrintDiffsOnly             bool
	Warn                       bool
	Cdot                       bool
	Probabilistic              bool
	CheckpointDurationMillis   int64
	ForceCheckpoint            bool
	LastCheckpoint             time.Time
	MainChecker                *ModelChecker
	Simulator                  *Simulator
	MetaDir                    string
	UseView                    bool
	UseGZIP                    bool
	Debug                      bool
	Tool                       bool
	SuppressedMessages         *InsMap[int, bool]
	MessagesAsErrors           *InsMap[int, bool]
	StartTime                  time.Time
}{
	EnumBound:                  2000,
	SetBound:                   1000000,
	NumWorkers:                 1,
	LivenessThreshold:          0.1,
	LivenessGraphSizeThreshold: 0.1,
	LivenessRatio:              0.2,
	LNCheck:                    "default",
	ProgressIntervalMillis:     DefaultProgressIntervalMillis,
	CoverageInterval:           -1,
	CoverageFlags:              initialCoverageFlags(),
	DFIDMax:                    -1,
	Expand:                     true,
	Warn:                       true,
	CheckpointDurationMillis:   DefaultCheckpointDurationMillis,
	MetaDir:                    "",
	LastCheckpoint:             time.Now(),
	SuppressedMessages:         NewInsMap[int, bool](),
	MessagesAsErrors:           NewInsMap[int, bool](),
	StartTime:                  time.Now(),
}

var tlcVersionNumber struct {
	sync.Once
	value string
}

func SetMainChecker(checker *ModelChecker) {
	Globals.Lock()
	defer Globals.Unlock()
	Globals.MainChecker = checker
}

func MainChecker() *ModelChecker {
	Globals.Lock()
	defer Globals.Unlock()
	return Globals.MainChecker
}

func SetSimulator(simulator *Simulator) {
	Globals.Lock()
	defer Globals.Unlock()
	Globals.Simulator = simulator
}

func CurrentSimulator() *Simulator {
	Globals.Lock()
	defer Globals.Unlock()
	return Globals.Simulator
}

func TLCStartTime() time.Time {
	Globals.Lock()
	defer Globals.Unlock()
	return Globals.StartTime
}

func TLCVersionNumber() string {
	tlcVersionNumber.Do(func() {
		tlcVersionNumber.value = TLCBuildDate().UTC().Format("2006.01.02.150405")
	})
	return tlcVersionNumber.value
}

func TLCVersion() string {
	return "Version " + TLCVersionNumber()
}

func TLCRevision() string {
	if rev := os.Getenv("TLAGO_REVISION"); rev != "" {
		return rev
	}
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			if setting.Key == "vcs.revision" && setting.Value != "" {
				if len(setting.Value) > 7 {
					return setting.Value[:7]
				}
				return setting.Value
			}
		}
	}
	return ""
}

func TLCRevisionOrDev() string {
	if rev := TLCRevision(); rev != "" {
		return rev
	}
	return "development"
}

func TLCBuildDate() time.Time {
	for _, key := range []string{"TLAGO_BUILD_TIMESTAMP", "SOURCE_DATE_EPOCH"} {
		value := os.Getenv(key)
		if value == "" {
			continue
		}
		if key == "SOURCE_DATE_EPOCH" {
			if seconds, err := strconv.ParseInt(value, 10, 64); err == nil {
				return time.Unix(seconds, 0).UTC()
			}
			continue
		}
		for _, layout := range []string{time.RFC3339Nano, time.RFC3339, "2006-01-02T15:04:05.0Z"} {
			if parsed, err := time.Parse(layout, value); err == nil {
				return parsed.UTC()
			}
		}
	}
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			if setting.Key == "vcs.time" && setting.Value != "" {
				if parsed, err := time.Parse(time.RFC3339, setting.Value); err == nil {
					return parsed.UTC()
				}
			}
		}
	}
	return time.Now().UTC()
}

func TLCScmCommits() int {
	for _, key := range []string{"TLAGO_COMMITS", "TLAGO_SCM_COMMITS"} {
		if value := os.Getenv(key); value != "" {
			if commits, err := strconv.Atoi(value); err == nil {
				return commits
			}
		}
	}
	return 0
}

func TLCInstallLocation() string {
	if exe, err := os.Executable(); err == nil && exe != "" {
		return exe
	}
	return "unknown"
}

func IsValidSetSize(bound int) bool {
	return bound >= 1
}

func SetNumWorkers(n int) {
	Globals.Lock()
	defer Globals.Unlock()
	Globals.NumWorkers = n
}

func NumWorkers() int {
	Globals.Lock()
	defer Globals.Unlock()
	return Globals.NumWorkers
}

func IncNumWorkers(n int) {
	Globals.Lock()
	defer Globals.Unlock()
	Globals.NumWorkers += n
}

func DecNumWorkers() {
	IncNumWorkers(-1)
}

func DoLiveness() bool {
	Globals.Lock()
	defer Globals.Unlock()
	return Globals.LNCheck != "final" && Globals.LNCheck != "seqfinal" && Globals.LNCheck != "off"
}

func DoSequentialLiveness() bool {
	Globals.Lock()
	defer Globals.Unlock()
	return len(Globals.LNCheck) >= 3 && Globals.LNCheck[:3] == "seq"
}

func CoverageEnabled() bool {
	Globals.Lock()
	defer Globals.Unlock()
	return Globals.CoverageInterval >= 0
}

func CoverageVariableEnabled() bool {
	Globals.Lock()
	defer Globals.Unlock()
	return Globals.CoverageInterval >= 0 || (Globals.CoverageFlags&2) > 0
}

func CoverageActionEnabled() bool {
	Globals.Lock()
	defer Globals.Unlock()
	return Globals.CoverageInterval >= 0 || (Globals.CoverageFlags&1) > 0
}

func CoverageAnyEnabled() bool {
	Globals.Lock()
	defer Globals.Unlock()
	return Globals.CoverageInterval >= 0 || Globals.CoverageFlags > 0
}

func initialCoverageFlags() int {
	for _, key := range []string{"tlc2.TLCGlobals.coverage", "TLAGO_COVERAGE"} {
		value := os.Getenv(key)
		if value == "" {
			continue
		}
		flags, err := strconv.Atoi(value)
		if err == nil {
			return flags
		}
	}
	return 0
}

func ProgressInterval() time.Duration {
	Globals.Lock()
	defer Globals.Unlock()
	if Globals.ProgressIntervalMillis <= 0 {
		return time.Duration(DefaultProgressIntervalMillis) * time.Millisecond
	}
	return time.Duration(Globals.ProgressIntervalMillis) * time.Millisecond
}

func LivenessRatio() float64 {
	Globals.Lock()
	defer Globals.Unlock()
	return Globals.LivenessRatio
}

func SetMetaDir(dir string) {
	Globals.Lock()
	Globals.MetaDir = dir
	Globals.Unlock()
}

func MetaDir() string {
	Globals.Lock()
	defer Globals.Unlock()
	return Globals.MetaDir
}

func SetUseView(enabled bool) {
	Globals.Lock()
	Globals.UseView = enabled
	Globals.Unlock()
}

func UseView() bool {
	Globals.Lock()
	defer Globals.Unlock()
	return Globals.UseView
}

func SetUseGZIP(enabled bool) {
	Globals.Lock()
	Globals.UseGZIP = enabled
	Globals.Unlock()
}

func UseGZIP() bool {
	Globals.Lock()
	defer Globals.Unlock()
	return Globals.UseGZIP
}

func ForceCheckpoint() {
	Globals.Lock()
	Globals.ForceCheckpoint = true
	Globals.Unlock()
}

func SuppressTLCMessage(code int) {
	Globals.Lock()
	if Globals.SuppressedMessages == nil {
		Globals.SuppressedMessages = NewInsMap[int, bool]()
	}
	Globals.SuppressedMessages.Set(code, true)
	Globals.Unlock()
}

func TreatTLCMessageAsError(code int) {
	Globals.Lock()
	if Globals.MessagesAsErrors == nil {
		Globals.MessagesAsErrors = NewInsMap[int, bool]()
	}
	Globals.MessagesAsErrors.Set(code, true)
	Globals.Unlock()
}

func messageControlFor(code int) (suppressed bool, asError bool, warn bool) {
	Globals.Lock()
	defer Globals.Unlock()
	warn = Globals.Warn
	if Globals.SuppressedMessages != nil {
		_, suppressed = Globals.SuppressedMessages.Get2(code)
	}
	if Globals.MessagesAsErrors != nil {
		_, asError = Globals.MessagesAsErrors.Get2(code)
	}
	return suppressed, asError, warn
}

func CheckpointExplicitlyEnabled() bool {
	Globals.Lock()
	defer Globals.Unlock()
	return Globals.CheckpointDurationMillis > 0 && Globals.CheckpointDurationMillis != DefaultCheckpointDurationMillis
}

func DoCheckPoint() bool {
	Globals.Lock()
	defer Globals.Unlock()
	if Globals.ForceCheckpoint {
		Globals.ForceCheckpoint = false
		return true
	}
	if Globals.CheckpointDurationMillis == 0 {
		return false
	}
	now := time.Now()
	if now.Sub(Globals.LastCheckpoint) >= time.Duration(Globals.CheckpointDurationMillis)*time.Millisecond {
		Globals.LastCheckpoint = now
		return true
	}
	return false
}
