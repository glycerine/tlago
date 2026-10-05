// Copyright (c) 2019 Microsoft Research. All rights reserved.
package tlc

import (
	"bufio"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"io"
	mathrand "math/rand/v2"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	ExecutionStatisticsRandomIdentifier   = "RANDOM_IDENTIFIER"
	ExecutionStatisticsNoStatistics       = "NO_STATISTICS"
	ExecutionStatisticsIdentifierProperty = "util.ExecutionStatisticsCollector.id"
)

type ExecutionStatisticsCollector struct {
	pathname, hostname string
	// Native representation of the source's virtual submit method.
	SubmitOverride func(string, *InsMap[string, string])
}

func NewExecutionStatisticsCollector(pathname, hostname string) *ExecutionStatisticsCollector {
	return &ExecutionStatisticsCollector{pathname: pathname, hostname: hostname}
}
func executionStatisticsTrim(text string) string {
	return strings.TrimFunc(text, func(r rune) bool { return r <= 0x20 })
}
func executionStatisticsReadLine(path string) (*string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, distributedFileOpenException(path, err)
	}
	defer file.Close()
	if info, err := file.Stat(); err != nil {
		return nil, bufferedRandomAccessFileIOError(err)
	} else if info.IsDir() {
		return nil, NewFileNotFoundException(path + " (Is a directory)")
	}
	// Java BufferedReader recognizes LF, CR, and CRLF. All preference tokens are
	// decoded with the original platform charset, including replacement behavior.
	reader := bufio.NewReader(file)
	charset := javaDefaultCharset()
	wide := charset == "UTF-16" || charset == "UTF-16BE" || charset == "UTF-16LE"
	littleEndian := charset == "UTF-16LE"
	terminated := false
	var data []byte
	for {
		b, err := reader.ReadByte()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, bufferedRandomAccessFileIOError(err)
		}
		unit := uint16(b)
		var second byte
		if wide {
			second, err = reader.ReadByte()
			if errors.Is(err, io.EOF) {
				data = append(data, b)
				break
			}
			if err != nil {
				return nil, bufferedRandomAccessFileIOError(err)
			}
			if charset == "UTF-16" && len(data) == 0 && b == 0xff && second == 0xfe {
				littleEndian = true
			}
			unit = uint16(b)<<8 | uint16(second)
			if littleEndian {
				unit = uint16(second)<<8 | uint16(b)
			}
		}
		if unit == '\n' || unit == '\r' {
			terminated = true
			break
		}
		data = append(data, b)
		if wide {
			data = append(data, second)
			if unit >= 0xd800 && unit <= 0xdbff {
				// The UTF-16 decoder consumes the next code unit even when it is
				// malformed. A CR/LF consumed this way is a replacement character,
				// rather than a BufferedReader line terminator.
				for i := 0; i < 2; i++ {
					next, err := reader.ReadByte()
					if errors.Is(err, io.EOF) {
						break
					}
					if err != nil {
						return nil, bufferedRandomAccessFileIOError(err)
					}
					data = append(data, next)
				}
			}
		}
	}
	line, err := javaCharsetDecode(data, charset, true)
	if err != nil {
		return nil, err
	}
	if line == "" && !terminated {
		return nil, nil
	}
	return &line, nil
}
func (c *ExecutionStatisticsCollector) IdentifierFromLine(identifier *string) *string {
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
