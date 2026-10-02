package tlc

import (
	"fmt"
	"io"
	"os"
	"os/user"
	"regexp"
	"sort"
	"strings"
	"time"
)

const ResultMailAddressProperty = "result.mail.address"

// MailInternetAddress retains InternetAddress's parsed mailbox and separate
// personal/encoded-personal fields. Parsing stores EncodedPersonal; lazy RFC2047
// personal-name decoding and MIME formatting belong to the MIME provider port.
type MailInternetAddress struct {
	Address         string
	Personal        *string
	EncodedPersonal *string
}

type MailFileList struct{ Files []*TLAFile }

type MailRequest struct {
	From, To      *MailInternetAddress
	Subject, Body string
	Files         []*TLAFile
}

type MailMXRecord struct {
	Weight   int32
	Hostname string
}

// MailSenderEnvironment supplies native JavaMail/JNDI/process boundaries.
// ParseAddresses and NewAddress default to the JavaMail 1.6.8 grammar port.
// Transmit owns Session/MIME/attachment/SMTP construction, using
// the same live system properties exposed here. LookupMX supplies the raw MX
// attribute: nil means no attribute, whereas a nonnil empty attribute has no
// hosts. InstallOut/Err assign ToolIO streams in source construction order.
type MailSenderEnvironment struct {
	LoadProperties         func()
	GetProperty            func(string) (string, bool)
	ContainsProperty       func(string) bool
	SetProperty            func(string, string)
	ParseAddresses         func(string) ([]*MailInternetAddress, error)
	NewAddress             func(string) (*MailInternetAddress, error)
	LocalHostName          func() (string, error)
	OpenLog                func(string) (io.WriteCloser, error)
	InstallOut, InstallErr func(*MailLogPrintStream)
	SystemOut, SystemErr   io.Writer
	OpenInput              func(string) (io.ReadCloser, error)
	LookupMX               func(string) (*[]any, error)
	Transmit               func(MailRequest) error
	Sleep                  func(time.Duration) error
	Date                   func() string
	PrintStackTrace        func(error)
	LowerCase              func(string) string
}

type MailSender struct {
	modelName, specName  *string
	toAddresses          []*MailInternetAddress
	from, fromAlt        *MailInternetAddress
	out, err             *TLAFile
	OutStream, ErrStream *MailLogPrintStream
	env                  MailSenderEnvironment
}

func NewMailSender(env MailSenderEnvironment, mainFile ...*string) (sender *MailSender, err error) {
	defer func() {
		if failure := recover(); failure != nil {
			sender = nil
			err = panicValueAsError(failure)
		}
	}()
	env = mailSenderEnvironment(env)
	m := &MailSender{modelName: javaString("unknown model"), specName: javaString("unknown spec"), env: env}
	env.LoadProperties()
	if mailto, ok := env.GetProperty(ResultMailAddressProperty); ok {
		m.toAddresses, err = env.ParseAddresses(mailto)
		if err != nil {
			return nil, err
		}
		if m.toAddresses == nil {
			return nil, NewNullPointerException()
		}
		if len(m.toAddresses) == 0 {
			return nil, NewArrayIndexOutOfBoundsException(0, 0)
		}
		if m.toAddresses[0] == nil {
			return nil, NewNullPointerException()
		}
		m.from, err = env.NewAddress("TLC - The friendly model checker <" + m.toAddresses[0].Address + ">")
		if err != nil {
			return nil, err
		}
		username := mailNullableProperty(env, "user.name")
		hostname, e := env.LocalHostName()
		if e != nil {
			return nil, e
		}
		m.fromAlt, err = env.NewAddress("TLC - The friendly model checker <" + username + "@" + hostname + ">")
		if err != nil {
			return nil, err
		}
		tmpdir := mailNullableProperty(env, "java.io.tmpdir")
		m.out = NewTLAFile(filenameNormalizeFile(tmpdir+string(os.PathSeparator)+"MC.out"), false, nil)
		file, e := env.OpenLog(m.out.GetPath())
		if e != nil {
			return nil, e
		}
		m.OutStream = NewMailLogPrintStream(file, env.SystemOut)
		if env.InstallOut != nil {
			env.InstallOut(m.OutStream)
		}
		m.err = NewTLAFile(filenameNormalizeFile(tmpdir+string(os.PathSeparator)+"MC.err"), false, nil)
		file, e = env.OpenLog(m.err.GetPath())
		if e != nil {
			return nil, e
		}
		m.ErrStream = NewMailLogPrintStream(file, env.SystemErr)
		if env.InstallErr != nil {
			env.InstallErr(m.ErrStream)
		}
	}
	if len(mainFile) > 0 {
		m.SetModelName(mainFile[0])
	}
	return m, nil
}

func mailSenderEnvironment(env MailSenderEnvironment) MailSenderEnvironment {
	if env.ParseAddresses == nil {
		env.ParseAddresses = func(value string) ([]*MailInternetAddress, error) { return ParseMailInternetAddresses(value) }
	}
	if env.NewAddress == nil {
		env.NewAddress = func(value string) (*MailInternetAddress, error) { return NewMailInternetAddress(value) }
	}
	if env.LoadProperties == nil {
		env.LoadProperties = func() { NewModelInJar().LoadProperties() }
	}
	if env.GetProperty == nil {
		env.GetProperty = func(key string) (string, bool) {
			if value, ok := tlcLookupSystemProperty(key); ok {
				return value, true
			}
			switch key {
			case "java.io.tmpdir":
				return os.TempDir(), true
			case "user.name":
				if current, err := user.Current(); err == nil {
					return current.Username, true
				}
			}
			return "", false
		}
	}
	if env.SetProperty == nil {
		env.SetProperty = tlcSetSystemProperty
	}
	if env.ContainsProperty == nil {
		env.ContainsProperty = func(key string) bool { _, present := env.GetProperty(key); return present }
	}
	if env.LocalHostName == nil {
		env.LocalHostName = distributedLocalHostName
	}
	if env.SystemOut == nil {
		env.SystemOut = os.Stdout
	}
	if env.SystemErr == nil {
		env.SystemErr = os.Stderr
	}
	if env.OpenLog == nil {
		env.OpenLog = func(path string) (io.WriteCloser, error) {
			file, err := os.Create(path)
			if err != nil {
				return nil, distributedFileOpenException(path, err)
			}
			return file, nil
		}
	}
	if env.OpenInput == nil {
		env.OpenInput = func(path string) (io.ReadCloser, error) {
			file, err := os.Open(path)
			if err != nil {
				return nil, distributedFileOpenException(path, err)
			}
			return file, nil
		}
	}
	if env.Sleep == nil {
		env.Sleep = func(duration time.Duration) error { time.Sleep(duration); return nil }
	}
	if env.Date == nil {
		env.Date = func() string { return time.Now().Format("Mon Jan 02 15:04:05 MST 2006") }
	}
	if env.PrintStackTrace == nil {
		env.PrintStackTrace = func(err error) { fmt.Fprint(env.SystemErr, javaThrowableStackTrace(err)) }
	}
	if env.LowerCase == nil {
		env.LowerCase = strings.ToLower
	}
	return env
}
func mailNullableProperty(env MailSenderEnvironment, key string) string {
	if value, ok := env.GetProperty(key); ok {
		return value
	}
	return "null"
}
func (m *MailSender) SetModelName(name *string) {
	if m == nil {
		panic(NewNullPointerException())
	}
	m.modelName = copyJavaMessage(name)
}
func (m *MailSender) SetSpecName(name *string) {
	if m == nil {
		panic(NewNullPointerException())
	}
	m.specName = copyJavaMessage(name)
}
func (m *MailSender) GetModelName() *string      { return copyJavaMessage(m.modelName) }
func (m *MailSender) GetSpecName() *string       { return copyJavaMessage(m.specName) }
func (m *MailSender) SendDefault() (bool, error) { return m.Send(&MailFileList{}) }

// Send mutates its caller's list by prepending output, then nonempty error.
// It attempts every recipient even after one succeeds, using the alternate
// sender only on failure and rereading the body before each attempt.
func (m *MailSender) Send(files *MailFileList) (success bool, err error) {
	defer func() {
		if failure := recover(); failure != nil {
			success = false
			err = panicValueAsError(failure)
		}
	}()
	if m == nil {
		return false, NewNullPointerException()
	}
	if m.toAddresses == nil {
		return true, nil
	}
	if files == nil {
		return false, NewNullPointerException()
	}
	files.Files = append([]*TLAFile{m.out}, files.Files...)
	if info, e := os.Stat(m.err.GetPath()); e == nil && info.Size() != 0 {
		files.Files = append([]*TLAFile{m.err}, files.Files...)
	}
	for _, to := range m.toAddresses {
		subject := "Model Checking result for " + javaNullableString(m.modelName) + " with spec " + javaNullableString(m.specName)
		body, e := m.ExtractBody(m.out)
		if e != nil {
			return false, e
		}
		ok, e := SendMailRequest(MailRequest{m.from, to, subject, body, append([]*TLAFile(nil), files.Files...)}, m.env)
		if e != nil {
			return false, e
		}
		if ok {
			success = true
			continue
		}
		subject = "Model Checking result for " + javaNullableString(m.modelName) + " with spec " + javaNullableString(m.specName)
		body, e = m.ExtractBody(m.out)
		if e != nil {
			return false, e
		}
		ok, e = SendMailRequest(MailRequest{m.fromAlt, to, subject, body, append([]*TLAFile(nil), files.Files...)}, m.env)
		if e != nil {
			return false, e
		}
		if ok {
			success = true
		}
	}
	return success, nil
}

func SendMailRequest(request MailRequest, env MailSenderEnvironment) (success bool, err error) {
	defer func() {
		if failure := recover(); failure != nil {
			success = false
			err = panicValueAsError(failure)
		}
	}()
	env = mailSenderEnvironment(env)
	if request.To == nil {
		return false, NewNullPointerException()
	}
	address := request.To.Address
	if !strings.Contains(address, "@") {
		return false, nil
	}
	if !strings.HasSuffix(address, "localhost") {
		env.SetProperty("mail.smtp.starttls.enable", "true")
	}
	parts := strings.Split(address, "@")
	for len(parts) > 0 && parts[len(parts)-1] == "" {
		parts = parts[:len(parts)-1]
	}
	if len(parts) < 2 {
		return false, NewArrayIndexOutOfBoundsException(1, len(parts))
	}
	if env.LookupMX == nil {
		return false, NewUnsupportedOperationException("JNDI MX provider is not configured")
	}
	attribute, e := env.LookupMX(parts[1])
	if e != nil {
		if _, ok := e.(*NamingException); ok {
			env.PrintStackTrace(e)
			return false, nil
		}
		return false, e
	}
	hosts, e := MailMXRecords(parts[1], attribute)
	if e != nil {
		return false, e
	}
	for i := 0; i < len(hosts); i++ {
		host := hosts[i].Hostname
		env.SetProperty("mail.smtp.host", host)
		if env.Transmit == nil {
			return false, NewUnsupportedOperationException("JavaMail transport provider is not configured")
		}
		failure := invokeDistributedServerOperation(func() error { return env.Transmit(request) })
		if failure == nil {
			return true, nil
		}
		if failed, ok := failure.(*SendFailedException); ok {
			next := failed.GetNextException()
			greylisted := false
			if next != nil && javaThrowableDetailMessage(next) != nil && strings.Contains(env.LowerCase(*javaThrowableDetailMessage(next)), "greylist") {
				greylisted = env.ContainsProperty(mailNullableProperty(env, "mail.smtp.host") + ".greylisted")
				if !greylisted {
					env.SetProperty(mailNullableProperty(env, "mail.smtp.host")+".greylisted", "true")
					if e = mailThrottleRetry(fmt.Sprintf("%s EMail Report: Detected greylisting when sending to %s at %s, will retry in %s minutes...", env.Date(), address, host, "10"), 10, env); e != nil {
						return false, e
					}
					i--
					continue
				}
			}
			if e = mailThrottleRetry(fmt.Sprintf("%s EMail Report: Slowing down due to errors when sending to %s at %s, will continue in %d minute...", env.Date(), address, host, 1), 1, env); e != nil {
				return false, e
			}
		} else if javaMessagingException(failure) != nil || isJavaIOException(failure) {
			env.PrintStackTrace(failure)
		} else {
			return false, failure
		}
	}
	return false, nil
}

var mailMXSplit = regexp.MustCompile(`[\t\n\x0B\f\r ]+`)

func MailMXRecords(domain string, attribute *[]any) (records []MailMXRecord, err error) {
	if attribute == nil {
		return []MailMXRecord{{0, domain}}, nil
	}
	records = []MailMXRecord{}
	for _, entry := range *attribute {
		value, ok := entry.(string)
		if !ok {
			continue
		}
		parts := mailMXSplit.Split(value, -1)
		for len(parts) > 0 && parts[len(parts)-1] == "" {
			parts = parts[:len(parts)-1]
		}
		if len(parts) != 2 {
			continue
		}
		weight, ok := javaParseDecimalInt(parts[0])
		if !ok {
			return nil, NewNumberFormatException(parts[0])
		}
		records = append(records, MailMXRecord{weight, parts[1]})
	}
	sort.SliceStable(records, func(i, j int) bool { return records[i].Weight < records[j].Weight })
	return records, nil
}
func mailThrottleRetry(message string, minutes int64, env MailSenderEnvironment) error {
	_, _ = fmt.Fprintln(env.SystemErr, message)
	_, _ = fmt.Fprintln(env.SystemOut, message)
	if err := env.Sleep(time.Duration(minutes*60*1000) * time.Millisecond); err != nil {
		if _, ok := err.(*InterruptedException); ok {
			env.PrintStackTrace(err)
			return nil
		}
		return err
	}
	return nil
}

// ExtractBody follows Scanner.hasNext followed by nextLine: interior blank
// lines survive, but a suffix containing only Java whitespace is not consumed.
func (m *MailSender) ExtractBody(out *TLAFile) (body string, err error) {
	defer func() {
		if failure := recover(); failure != nil {
			body = ""
			err = panicValueAsError(failure)
		}
	}()
	if m == nil || out == nil {
		return "", NewNullPointerException()
	}
	stream, e := m.env.OpenInput(out.GetPath())
	if e != nil {
		if _, ok := e.(*FileNotFoundException); !ok {
			return "", e
		}
		m.env.PrintStackTrace(e)
		return "Failed to find file " + out.GetAbsolutePath(), nil
	}
	body = mailExtractScannedBody(stream)
	if closeErr := stream.Close(); closeErr != nil {
		if _, typed := closeErr.(interface{ GetMessage() *string }); typed && !isJavaIOException(closeErr) {
			return "", closeErr
		}
	}
	return body, nil
}

func mailJavaWhitespace(r rune) bool {
	return r >= '\t' && r <= '\r' || r >= 0x1c && r <= 0x20 || r == 0x1680 || r >= 0x2000 && r <= 0x2006 || r >= 0x2008 && r <= 0x200a || r == 0x2028 || r == 0x2029 || r == 0x205f || r == 0x3000
}

// DistributedServerMail connects this sender's source report flow to the server
// command's existing boundary. The provider must also install its ToolIO
// streams; complete MP/SANY console-to-PrintStream routing remains separate.
func (m *MailSender) DistributedServerMail() *DistributedServerMail {
	if m == nil {
		panic(NewNullPointerException())
	}
	mail := &DistributedServerMail{ModelName: javaNullableString(m.modelName), SpecName: javaNullableString(m.specName)}
	mail.Deliver = func(files []*TLAFile) (bool, error) {
		m.SetModelName(&mail.ModelName)
		m.SetSpecName(&mail.SpecName)
		return m.Send(&MailFileList{Files: files})
	}
	return mail
}
