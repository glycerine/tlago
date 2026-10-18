package tlago

import (
	"io"
	"runtime"

	"github.com/glycerine/tlago/tlc"
)

// SanyLogLevel preserves tla2sany.output.LogLevel's ordinal routing.
type SanyLogLevel int

const (
	SanyLogTrace SanyLogLevel = iota
	SanyLogDebug
	SanyLogInfo
	SanyLogWarning
	SanyLogError
)

// SanyOutput controls parser output. Parser format arguments are strings;
// the driver's additional numeric/Object formatting remains a separate port.
type SanyOutput interface {
	Log(SanyLogLevel, string, ...string)
	GetStream(SanyLogLevel) io.Writer
}

type SimpleSanyOutput struct {
	out      io.Writer
	logLevel SanyLogLevel
}

func NewSimpleSanyOutput(out io.Writer, level SanyLogLevel) *SimpleSanyOutput {
	if out == nil {
		panic(tlc.NewIllegalArgumentException("out stream cannot be null"))
	}
	return &SimpleSanyOutput{out: out, logLevel: level}
}

func (out *SimpleSanyOutput) GetStream(level SanyLogLevel) io.Writer {
	if level >= out.logLevel {
		return out.out
	}
	return io.Discard
}

func (out *SimpleSanyOutput) Log(level SanyLogLevel, format string, args ...string) {
	sanyOutputPrintln(out.GetStream(level), format, args)
}

type SilentSanyOutput struct{}

func (*SilentSanyOutput) Log(SanyLogLevel, string, ...string) {}
func (*SilentSanyOutput) GetStream(SanyLogLevel) io.Writer    { return io.Discard }

type OutErrSanyOutput struct{ outStreams [5]io.Writer }

func NewOutErrSanyOutput(out, err io.Writer, outputLevel, errorLevel SanyLogLevel) *OutErrSanyOutput {
	if out == nil {
		panic(tlc.NewIllegalArgumentException("out stream cannot be null"))
	}
	if err == nil {
		panic(tlc.NewIllegalArgumentException("err stream cannot be null"))
	}
	result := &OutErrSanyOutput{}
	for level := SanyLogTrace; level <= SanyLogError; level++ {
		stream := io.Writer(io.Discard)
		if level >= outputLevel {
			stream = out
			if level >= errorLevel {
				stream = err
			}
		}
		result.outStreams[level] = stream
	}
	return result
}

func (out *OutErrSanyOutput) GetStream(level SanyLogLevel) io.Writer { return out.outStreams[level] }
func (out *OutErrSanyOutput) Log(level SanyLogLevel, format string, args ...string) {
	sanyOutputPrintln(out.GetStream(level), format, args)
}

func sanyOutputPrintln(out io.Writer, format string, args []string) {
	message := format
	if len(args) != 0 {
		var err error
		message, err = tlc.JavaFormatStrings(format, args...)
		if err != nil {
			panic(err)
		}
	}
	separator := "\n"
	if runtime.GOOS == "windows" {
		separator = "\r\n"
	}
	// Java PrintStream records I/O failure instead of throwing it to the parser.
	_, _ = io.WriteString(out, message+separator)
}

// ParseSanySyntaxWithOutput follows TLAplusParser.parse's reported-error loop.
// It parses bytes supplied by the caller without module loading or semantics.
func ParseSanySyntaxWithOutput(file, source string, out SanyOutput) (*SanySyntaxNode, Diagnostics) {
	node, _, diags := parseSanySyntaxUsingOutput(file, source, out)
	return node, diags
}
