package tlc

import (
	"sync"
	"time"
)

const DefaultCheckpointDurationMillis = (30 * 60 * 1000) + 42

var Globals = struct {
	sync.Mutex
	EnumBound                  int
	SetBound                   int
	NumWorkers                 int
	LivenessThreshold          float64
	LivenessGraphSizeThreshold float64
	LivenessRatio              float64
	LNCheck                    string
	CoverageInterval           int
	DFIDMax                    int
	Continuation               bool
	Expand                     bool
	PrintDiffsOnly             bool
	Warn                       bool
	CheckpointDurationMillis   int64
	ForceCheckpoint            bool
	LastCheckpoint             time.Time
	MainChecker                *ModelChecker
	Simulator                  *Simulator
	StartTime                  time.Time
}{
	EnumBound:                  2000,
	SetBound:                   1000000,
	NumWorkers:                 1,
	LivenessThreshold:          0.1,
	LivenessGraphSizeThreshold: 0.1,
	LivenessRatio:              0.2,
	LNCheck:                    "default",
	CoverageInterval:           -1,
	DFIDMax:                    -1,
	Expand:                     true,
	Warn:                       true,
	CheckpointDurationMillis:   DefaultCheckpointDurationMillis,
	LastCheckpoint:             time.Now(),
	StartTime:                  time.Now(),
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

func ForceCheckpoint() {
	Globals.Lock()
	Globals.ForceCheckpoint = true
	Globals.Unlock()
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
