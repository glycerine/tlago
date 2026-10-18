package tlc

import (
	"crypto/rand"
	"encoding/hex"
	mathrand "math/rand/v2"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	ExecutionStatisticsRandomIdentifier   = "RANDOM_IDENTIFIER"
	ExecutionStatisticsNoStatistics       = "NO_STATISTICS"
	ExecutionStatisticsIdentifierProperty = "util.ExecutionStatisticsCollector.id"
	executionStatisticsDomainProperty     = "util.ExecutionStatisticsCollector.domain"
	executionStatisticsDefaultHostname    = "tlaplus-execution-stats-collection01"
	executionStatisticsWaitProperty       = "util.ExecutionStatisticsCollector.waitForCompletion"
)

var executionStatisticsPending = struct {
	sync.Mutex
	collectors []<-chan struct{}
}{}

var executionStatisticsDefaults struct {
	sync.Once
	path     string
	hostname string
}

type ExecutionStatisticsSelection int

const (
	ExecutionStatisticsSelectionOn ExecutionStatisticsSelection = iota
	ExecutionStatisticsSelectionRandomIdentifier
	ExecutionStatisticsSelectionNoESC
)

func (s ExecutionStatisticsSelection) String() string {
	switch s {
	case ExecutionStatisticsSelectionOn:
		return executionStatisticsRandomIdentifier()
	case ExecutionStatisticsSelectionRandomIdentifier:
		return ExecutionStatisticsRandomIdentifier
	default:
		return ExecutionStatisticsNoStatistics
	}
}

type ExecutionStatisticsCollector struct {
	pathname, hostname string
	// Native representation of the source's virtual submit method.
	SubmitOverride func(string, *InsMap[string, string])
}

func NewExecutionStatisticsCollector(pathname, hostname string) *ExecutionStatisticsCollector {
	executionStatisticsInitializeDefaults()
	return &ExecutionStatisticsCollector{pathname: pathname, hostname: hostname}
}

func NewDefaultExecutionStatisticsCollector() *ExecutionStatisticsCollector {
	executionStatisticsInitializeDefaults()
	return NewExecutionStatisticsCollector(executionStatisticsDefaults.path, executionStatisticsDefaults.hostname)
}

func executionStatisticsInitializeDefaults() {
	executionStatisticsDefaults.Do(func() {
		home, ok := tlcLookupSystemProperty("user.home")
		if !ok {
			home, _ = os.UserHomeDir()
		}
		separator := string(os.PathSeparator)
		executionStatisticsDefaults.path = home + separator + ".tlaplus" + separator + "esc.txt"
		executionStatisticsDefaults.hostname = tlcGetSystemProperty(
			executionStatisticsDomainProperty, executionStatisticsDefaultHostname)
	})
}
func (c *ExecutionStatisticsCollector) require() {
	if c == nil {
		panic(NewNullPointerException())
	}
}

func (c *ExecutionStatisticsCollector) Collect(parameters *InsMap[string, string]) {
	c.require()
	wait, _ := tlcLookupSystemProperty(executionStatisticsWaitProperty)
	c.CollectAsync(parameters, javaBooleanProperty(wait))
}

func (c *ExecutionStatisticsCollector) CollectAsync(parameters *InsMap[string, string], waitForCompletion bool) {
	c.require()
	done := make(chan struct{})
	go func() {
		defer close(done)
		defer func() { _ = recover() }()
		c.Collect0(parameters)
	}()
	if waitForCompletion {
		executionStatisticsPending.Lock()
		executionStatisticsPending.collectors = append(executionStatisticsPending.collectors, done)
		executionStatisticsPending.Unlock()
	}
}

func WaitForExecutionStatisticsCollectors() {
	executionStatisticsPending.Lock()
	collectors := executionStatisticsPending.collectors
	executionStatisticsPending.collectors = nil
	executionStatisticsPending.Unlock()

	var waits sync.WaitGroup
	waits.Add(len(collectors))
	for _, collector := range collectors {
		go func(done <-chan struct{}) {
			defer waits.Done()
			select {
			case <-done:
			case <-time.After(10 * time.Second):
			}
		}(collector)
	}
	waits.Wait()
}
func executionStatisticsTrim(text string) string {
	return strings.TrimFunc(text, func(r rune) bool { return r <= 0x20 })
}
func executionStatisticsReadLine(path string) (line *string, err error) {
	file, err := openFileInputStream(path)
	if err != nil {
		return nil, err
	}
	defer func() {
		failure := recover()
		closeErr := bufferedRandomAccessFileIOError(file.Close())
		if failure != nil {
			if primary, ok := failure.(error); ok {
				javaSuppressCloseError(primary, closeErr)
			}
			panic(failure)
		}
		if err == nil {
			if closeErr != nil {
				line = nil
				err = closeErr
			}
		} else {
			javaSuppressCloseError(err, closeErr)
		}
	}()
	reader, err := newJavaReplacingLineReader(file, javaDefaultCharset())
	if err != nil {
		return nil, bufferedRandomAccessFileIOError(err)
	}
	line, err = reader.ReadLine()
	return line, bufferedRandomAccessFileIOError(err)
}
func (c *ExecutionStatisticsCollector) IdentifierFromLine(identifier *string) *string {
	c.require()
	if value, ok := tlcLookupSystemProperty(ExecutionStatisticsIdentifierProperty); ok {
		return &value
	}
	if identifier == nil || executionStatisticsTrim(*identifier) == ExecutionStatisticsNoStatistics {
		return nil
	}
	text := *identifier
	if executionStatisticsTrim(text) == ExecutionStatisticsRandomIdentifier {
		text = executionStatisticsRandomIdentifier()
	}
	units := javaStringUTF16(executionStatisticsTrim(text))
	result := javaStringFromUTF16(units[:min(len(units), 32)])
	return &result
}
func (c *ExecutionStatisticsCollector) GetIdentifier() *string {
	c.require()
	if value, ok := tlcLookupSystemProperty(ExecutionStatisticsIdentifierProperty); ok {
		return &value
	}
	if _, err := os.Stat(c.pathname); err != nil {
		return nil
	}
	line, err := executionStatisticsReadLine(c.pathname)
	if err != nil {
		return nil
	}
	return c.IdentifierFromLine(line)
}

func (c *ExecutionStatisticsCollector) escFileExists() bool {
	c.require()
	_, err := os.Stat(c.pathname)
	return err == nil
}

func (c *ExecutionStatisticsCollector) IsEnabled() bool {
	c.require()
	return c.GetIdentifier() != nil || executionStatisticsOptIn(c.hostname) != nil
}

func (c *ExecutionStatisticsCollector) Set(selection ExecutionStatisticsSelection) (err error) {
	c.require()
	executionStatisticsInitializeDefaults()
	path := executionStatisticsDefaults.path
	_ = os.MkdirAll(filepath.Dir(path), 0777)
	if err := touchFile(path); err != nil {
		return err
	}
	file, err := openFileOutputStream(path)
	if err != nil {
		return err
	}
	writer := NewJavaBufferedWriter(file)
	defer func() {
		failure := recover()
		closeErr := bufferedRandomAccessFileIOError(writer.Flush())
		fileCloseErr := bufferedRandomAccessFileIOError(file.Close())
		if closeErr == nil {
			closeErr = fileCloseErr
		} else {
			javaSuppressCloseError(closeErr, fileCloseErr)
		}
		if failure != nil {
			if primary, ok := failure.(error); ok {
				javaSuppressCloseError(primary, closeErr)
			}
			panic(failure)
		}
		if err == nil {
			err = closeErr
		} else {
			javaSuppressCloseError(err, closeErr)
		}
	}()
	_, err = writer.WriteString(selection.String() + "\n")
	return bufferedRandomAccessFileIOError(err)
}

func (c *ExecutionStatisticsCollector) Get() ExecutionStatisticsSelection {
	c.require()
	if c.IsEnabled() {
		if line, err := executionStatisticsReadLine(c.pathname); err == nil {
			if line != nil && *line == ExecutionStatisticsRandomIdentifier {
				return ExecutionStatisticsSelectionRandomIdentifier
			}
			return ExecutionStatisticsSelectionOn
		}
	}
	return ExecutionStatisticsSelectionNoESC
}

func ExecutionStatisticsPromptUser() bool {
	return !NewDefaultExecutionStatisticsCollector().escFileExists()
}

func executionStatisticsOptIn(hostname string) net.IP {
	addresses, err := net.LookupIP(hostname)
	if err != nil || len(addresses) == 0 {
		return nil
	}
	preferIPv6, _ := tlcLookupSystemProperty("java.net.preferIPv6Addresses")
	ipv4Only, _ := tlcLookupSystemProperty("java.net.preferIPv4Stack")
	for _, address := range addresses {
		ipv4 := address.To4() != nil
		if ipv4Only == "true" && !ipv4 {
			continue
		}
		if strings.EqualFold(preferIPv6, "system") || ipv4 != strings.EqualFold(preferIPv6, "true") {
			return address
		}
	}
	if ipv4Only == "true" {
		return nil
	}
	return addresses[0]
}
func executionStatisticsCanonicalHost(address net.IP) string {
	canonical := distributedNumericAddress(address)
	if names, err := net.LookupAddr(address.String()); err == nil && len(names) > 0 {
		name := strings.TrimSuffix(names[0], ".")
		if forward, err := net.LookupIP(name); err == nil {
			for _, candidate := range forward {
				if address.Equal(candidate) {
					return name
				}
			}
		}
	}
	return canonical
}
func (c *ExecutionStatisticsCollector) Collect0(parameters *InsMap[string, string]) {
	c.require()
	var line *string
	if _, err := os.Stat(c.pathname); err == nil {
		var err error
		line, err = executionStatisticsReadLine(c.pathname)
		if err == nil {
			if line != nil && executionStatisticsTrim(*line) == ExecutionStatisticsNoStatistics {
				return
			}
		} else {
			switch err.(type) {
			case *FileNotFoundException, *NoSuchFileException:
			default:
				return
			}
		}
	}
	optIn := executionStatisticsOptIn(c.hostname)
	if optIn == nil {
		if id := c.IdentifierFromLine(line); id != nil {
			if parameters == nil {
				panic(NewNullPointerException())
			}
			parameters.Set("id", *id)
			c.submit("esc01.tlapl.us", parameters)
		}
	} else {
		hostname := executionStatisticsCanonicalHost(optIn)
		if line == nil || executionStatisticsTrim(*line) == "" {
			if parameters == nil {
				panic(NewNullPointerException())
			}
			parameters.Set("id", executionStatisticsUUIDv1())
			c.submit(hostname, parameters)
			return
		}
		if id := c.IdentifierFromLine(line); id != nil {
			if parameters == nil {
				panic(NewNullPointerException())
			}
			parameters.Set("id", *id)
			c.submit(hostname, parameters)
			return
		}
	}
}
func (c *ExecutionStatisticsCollector) submit(hostname string, parameters *InsMap[string, string]) {
	c.require()
	if c.SubmitOverride != nil {
		c.SubmitOverride(hostname, parameters)
		return
	}
	parameters.Set("ts", strconv.FormatInt(time.Now().UnixMilli(), 10))
	parameters.Set("optout", "false")
	scheme := "https"
	if value, _ := tlcLookupSystemProperty("util.ExecutionStatisticsCollector.nossl"); javaBooleanProperty(value) {
		scheme = "http"
	}
	var fields []string
	for key, value := range parameters.All() {
		fields = append(fields, executionStatisticsURLEncode(key)+"="+executionStatisticsURLEncode(value))
	}
	request, err := http.NewRequest(http.MethodHead, scheme+"://"+hostname+"/?"+strings.Join(fields, ","), nil)
	if err != nil {
		return
	}
	response, err := http.DefaultClient.Do(request)
	if err == nil {
		_ = response.Body.Close()
	}
}

// java.net.URLEncoder's UTF-8 form encoding keeps '*' and escapes '~'.
func executionStatisticsURLEncode(text string) string {
	const digits = "0123456789ABCDEF"
	var encoded strings.Builder
	for _, b := range []byte(javaStringUTF8(text)) {
		switch {
		case b >= 'a' && b <= 'z', b >= 'A' && b <= 'Z', b >= '0' && b <= '9', b == '-', b == '_', b == '.', b == '*':
			encoded.WriteByte(b)
		case b == ' ':
			encoded.WriteByte('+')
		default:
			encoded.WriteByte('%')
			encoded.WriteByte(digits[b>>4])
			encoded.WriteByte(digits[b&0xf])
		}
	}
	return encoded.String()
}
func executionStatisticsRandomIdentifier() string {
	var value [16]byte
	if _, err := rand.Read(value[:]); err != nil {
		panic(err)
	}
	value[6] = (value[6] & 0x0f) | 0x40
	value[8] = (value[8] & 0x3f) | 0x80
	return hex.EncodeToString(value[:])
}
func executionStatisticsMAC() []byte {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil
	}
	for _, iface := range interfaces {
		if len(iface.HardwareAddr) > 0 {
			return iface.HardwareAddr
		}
	}
	return nil
}
func executionStatisticsUUIDv1() string {
	mac := executionStatisticsMAC()
	if mac == nil {
		return executionStatisticsRandomIdentifier()
	}
	timestamp := uint64((time.Now().UnixMilli() + 12219292800000) * 10000)
	most := ((timestamp & 0xffffffff) << 32) | (((timestamp >> 32) & 0xffff) << 16) | ((timestamp >> 48) & 0x0fff) | 0x1000
	clock := uint64(mathrand.Float64()*0x3fff) | 0x8000
	least := clock << 48
	for i := 0; i < 6; i++ {
		if i >= len(mac) {
			panic(NewArrayIndexOutOfBoundsException(i, len(mac)))
		}
		least |= uint64(mac[i]) << uint(40-8*i)
	}
	return fmtUUIDBits(most, least)
}
func fmtUUIDBits(most, least uint64) string {
	var bytes [16]byte
	for i := 0; i < 8; i++ {
		bytes[i] = byte(most >> uint(56-8*i))
		bytes[8+i] = byte(least >> uint(56-8*i))
	}
	return hex.EncodeToString(bytes[:])
}

func ExecutionStatisticsCollectorMain() {
	NewDefaultExecutionStatisticsCollector().Collect0(NewInsMap[string, string]())
}
