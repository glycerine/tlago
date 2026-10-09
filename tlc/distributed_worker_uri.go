package tlc

import (
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf16"

	"golang.org/x/text/unicode/norm"
)

// DistributedWorkerAddress supplies the hostname and exported endpoint port
// used by Java's TLCWorker constructor. Local workers use port zero until a
// transport supplies an endpoint; this is also Java getPort's fallback value.
type DistributedWorkerAddress struct {
	Hostname string
	Port     int
	// Native networking supplies its validated TCP address before construction.
	// Keep endpoint URLs within Go's address/URL grammar, including zone names.
	nativeAddress string
}

type distributedWorkerURIValue struct {
	raw  string
	host *string
}

func (u *distributedWorkerURIValue) asciiString() string {
	for _, c := range u.raw {
		if c >= 128 {
			var result strings.Builder
			const hex = "0123456789ABCDEF"
			for _, b := range []byte(javaURINFC(u.raw)) {
				if b < 128 {
					result.WriteByte(b)
				} else {
					result.WriteByte('%')
					result.WriteByte(hex[b>>4])
					result.WriteByte(hex[b&15])
				}
			}
			return result.String()
		}
	}
	return u.raw
}

// x/text implements stream-safe normalization, which may insert CGJ after
// thirty non-starters. Java Normalizer does not insert it. Keep the ordinary
// fast path, then decompose/order/compose without segment-size limits if CGJ
// insertion occurred. Original CGJs remain ordinary CCC-zero input characters.
func javaURINFC(input string) string {
	normalized := norm.NFC.String(input)
	if strings.Count(normalized, "\u034f") == strings.Count(input, "\u034f") {
		return normalized
	}
	type character struct {
		value rune
		ccc   uint8
	}
	var decomposed []character
	for _, c := range input {
		// One canonical rune decomposition is always shorter than the
		// stream-safe limit; it cannot introduce an artificial CGJ.
		for _, d := range norm.NFD.String(string(c)) {
			decomposed = append(decomposed, character{d, norm.NFD.PropertiesString(string(d)).CCC()})
		}
	}
	for index := 0; index < len(decomposed); {
		if decomposed[index].ccc == 0 {
			index++
			continue
		}
		start := index
		for index < len(decomposed) && decomposed[index].ccc != 0 {
			index++
		}
		segment := decomposed[start:index]
		sort.SliceStable(segment, func(i, j int) bool { return segment[i].ccc < segment[j].ccc })
	}
	var result []rune
	starter := -1
	var lastCCC uint8
	for _, c := range decomposed {
		if starter >= 0 && (lastCCC == 0 || lastCCC < c.ccc) {
			pair := []rune(norm.NFC.String(string([]rune{result[starter], c.value})))
			if len(pair) == 1 {
				result[starter] = pair[0]
				continue
			}
		}
		if c.ccc == 0 {
			starter = len(result)
		}
		result = append(result, c.value)
		lastCCC = c.ccc
	}
	return string(result)
}

func newDistributedWorkerURI(address DistributedWorkerAddress, threadID int) *distributedWorkerURIValue {
	return newDistributedWorkerEndpointURI(address, fmtInt(threadID))
}

func newDistributedWorkerEndpointURI(address DistributedWorkerAddress, object string) *distributedWorkerURIValue {
	if address.nativeAddress != "" {
		host := address.Hostname
		raw := (&url.URL{Scheme: "tcp", Host: address.nativeAddress, Path: "/" + object}).String()
		return &distributedWorkerURIValue{raw: raw, host: &host}
	}
	raw := fmt.Sprintf("tcp://%s:%d/%s", address.Hostname, address.Port, object)
	parser := workerURIParser{raw: raw, chars: utf16.Encode([]rune(raw))}
	host, err := parser.parse()
	if err != nil {
		panic(NewIllegalArgumentExceptionWithCause(err.GetMessage(), err))
	}
	return &distributedWorkerURIValue{raw: raw, host: host}
}

// These masks and the hierarchical/server parsing below mirror java.net.URI.
// Only the tcp:// URI constructed by the worker is needed; scheme selection,
// relative resolution, normalization and URI comparison are not exposed here.
type workerURIMask struct{ low, high uint64 }

var (
	workerURIPath          = workerURIMask{3458764316252045313, 5188146764422578175}
	workerURIURIC          = workerURIMask{12682136353106821121, 5188146765093666815}
	workerURIRegName       = workerURIMask{3458623578763689985, 5188146764422578175}
	workerURIServer        = workerURIMask{3458623578763689985, 5188146765093666815}
	workerURIServerPercent = workerURIMask{3458623716202643457, 5188146765093666815}
	workerURIUserInfo      = workerURIMask{3458623578763689985, 5188146764422578174}
	workerURIScope         = workerURIMask{288019269919178752, 576460745995190270}
	workerURIHex           = workerURIMask{287948901175001088, 541165879422}
	workerURIAlphanum      = workerURIMask{287948901175001088, 576460743847706622}
	workerURIHostname      = workerURIMask{287984085547089920, 576460743847706622}
)

func (m workerURIMask) matches(c uint16) bool {
	if c > 0 && c < 64 {
		return m.low&(uint64(1)<<c) != 0
	}
	if c < 128 && c >= 64 {
		return m.high&(uint64(1)<<(c-64)) != 0
	}
	return false
}

type workerURIParser struct {
	raw       string
	chars     []uint16
	ipv6bytes int
}

func (p *workerURIParser) failure(reason string, index int) *URISyntaxException {
	return NewURISyntaxException(p.raw, reason, index)
}
func (p *workerURIParser) text(start, end int) string {
	return string(utf16.Decode(p.chars[start:end]))
}
func (p *workerURIParser) at(pos, end int, c uint16) bool { return pos < end && p.chars[pos] == c }
func (p *workerURIParser) until(pos, end int, stop string) int {
	for pos < end && !strings.ContainsRune(stop, rune(p.chars[pos])) {
		pos++
	}
	return pos
}

func (p *workerURIParser) scan(pos, end int, mask workerURIMask) (int, *URISyntaxException) {
	for pos < end {
		c := p.chars[pos]
		if mask.matches(c) {
			pos++
			continue
		}
		if mask.low&1 == 0 {
			break
		}
		if c == '%' {
			if pos+3 > end || !workerURIHex.matches(p.chars[pos+1]) || !workerURIHex.matches(p.chars[pos+2]) {
				return pos, p.failure("Malformed escape pair", pos)
			}
			pos += 3
		} else if c > 128 && !(c >= 128 && c <= 159) && !unicode.Is(unicode.Zs, rune(c)) && !unicode.Is(unicode.Zl, rune(c)) && !unicode.Is(unicode.Zp, rune(c)) {
			pos++
		} else {
			break
		}
	}
	return pos, nil
}

func (p *workerURIParser) check(start, end int, mask workerURIMask, component string) *URISyntaxException {
	pos, err := p.scan(start, end, mask)
	if err != nil {
		return err
	}
	if pos < end {
		return p.failure("Illegal character in "+component, pos)
	}
	return nil
}

func (p *workerURIParser) parse() (*string, *URISyntaxException) {
	end := len(p.chars)
	pos := 6 // tcp://, already supplied by the constructor
	authorityEnd := p.until(pos, end, "/?#")
	var host *string
	if authorityEnd > pos {
		var err *URISyntaxException
		host, err = p.authority(pos, authorityEnd)
		if err != nil {
			return nil, err
		}
		pos = authorityEnd
	} else if authorityEnd == end {
		return nil, p.failure("Expected authority", pos)
	}
	pathEnd := p.until(pos, end, "?#")
	if err := p.check(pos, pathEnd, workerURIPath, "path"); err != nil {
		return nil, err
	}
	pos = pathEnd
	if p.at(pos, end, '?') {
		queryEnd := p.until(pos+1, end, "#")
		if err := p.check(pos+1, queryEnd, workerURIURIC, "query"); err != nil {
			return nil, err
		}
		pos = queryEnd
	}
	if p.at(pos, end, '#') {
		if err := p.check(pos+1, end, workerURIURIC, "fragment"); err != nil {
			return nil, err
		}
	}
	return host, nil
}

func (p *workerURIParser) authority(start, end int) (*string, *URISyntaxException) {
	mask := workerURIServer
	if p.until(start, end, "]") > start {
		mask = workerURIServerPercent
	}
	serverEnd, err := p.scan(start, end, mask)
	if err != nil {
		return nil, err
	}
	regEnd, err := p.scan(start, end, workerURIRegName)
	if err != nil {
		return nil, err
	}
	serverValid, regValid := serverEnd == end, regEnd == end
	if regValid && !serverValid {
		return nil, nil
	}
	pos := start
	var failure *URISyntaxException
	if serverValid {
		var host *string
		host, pos, failure = p.server(start, end, regValid)
		if failure == nil && pos == end {
			return host, nil
		}
	}
	if regValid {
		return nil, nil
	}
	if failure != nil {
		return nil, failure
	}
	if !serverValid {
		pos = regEnd
	}
	return nil, p.failure("Illegal character in authority", pos)
}

func (p *workerURIParser) server(start, end int, skip bool) (*string, int, *URISyntaxException) {
	pos := start
	if userEnd := p.until(pos, end, "@"); userEnd < end {
		if err := p.check(pos, userEnd, workerURIUserInfo, "user info"); err != nil {
			return nil, pos, err
		}
		pos = userEnd + 1
	}
	hostStart := pos
	var host string
	if p.at(pos, end, '[') {
		pos++
		close := p.until(pos, end, "]")
		if close <= pos || close == end {
			return nil, pos, p.failure("Expected closing bracket for IPv6 address", close)
		}
		addressEnd := p.until(pos, close, "%")
		if err := p.ipv6(pos, addressEnd); err != nil {
			return nil, pos, err
		}
		if addressEnd < close {
			if addressEnd+1 == close {
				return nil, pos, p.failure("scope id expected", -1)
			}
			if err := p.check(addressEnd+1, close, workerURIScope, "scope id"); err != nil {
				return nil, pos, err
			}
		}
		host = p.text(hostStart, close+1)
		pos = close + 1
	} else {
		v4, _ := p.ipv4(pos, end, false)
		if v4 > pos && (v4 == end || p.at(v4, end, ':')) {
			pos = v4
		} else {
			var err *URISyntaxException
			pos, err = p.hostname(pos, end, skip)
			if err != nil {
				return nil, pos, err
			}
			if pos < end && !p.at(pos, end, ':') && skip {
				return nil, pos, nil
			}
		}
		host = p.text(hostStart, pos)
	}
	if p.at(pos, end, ':') {
		pos++
		if pos < end {
			if err := p.check(pos, end, workerURIMask{287948901175001088, 0}, "port number"); err != nil {
				return nil, pos, err
			}
			if _, err := strconv.ParseInt(p.text(pos, end), 10, 32); err != nil {
				return nil, pos, p.failure("Malformed port number", pos)
			}
			pos = end
		}
	} else if pos < end && skip {
		return nil, pos, nil
	}
	if pos < end {
		return nil, pos, p.failure("Expected port number", pos)
	}
	return &host, pos, nil
}

func (p *workerURIParser) hostname(start, end int, skip bool) (int, *URISyntaxException) {
	pos, last := start, -1
	for pos < end {
		next, _ := p.scan(pos, end, workerURIAlphanum)
		if next <= pos {
			break
		}
		last, pos = pos, next
		next, _ = p.scan(pos, end, workerURIHostname)
		if next > pos {
			if p.chars[next-1] == '-' {
				return next, p.failure("Illegal character in hostname", next-1)
			}
			pos = next
		}
		if !p.at(pos, end, '.') {
			break
		}
		pos++
	}
	if pos < end && !p.at(pos, end, ':') {
		if skip {
			return pos, nil
		}
		return pos, p.failure("Illegal character in hostname", pos)
	}
	if last < 0 {
		return pos, p.failure("Expected hostname", start)
	}
	if last > start && !(workerURIMask{0, 576460743847706622}).matches(p.chars[last]) {
		return pos, p.failure("Illegal character in hostname", last)
	}
	return pos, nil
}

func (p *workerURIParser) ipv4(start, end int, strict bool) (int, *URISyntaxException) {
	limit, _ := p.scan(start, end, workerURIMask{288019269919178752, 0})
	if limit <= start || strict && limit != end {
		return -1, nil
	}
	pos, next := start, start
	for part := 0; part < 4; part++ {
		next, _ = p.scan(pos, limit, workerURIMask{287948901175001088, 0})
		if next <= pos {
			break
		}
		significant := pos
		for significant < next && p.chars[significant] == '0' {
			significant++
		}
		if next-significant > 3 {
			next = pos
			break
		}
		if next-significant == 3 {
			value, _ := strconv.ParseInt(p.text(pos, next), 10, 32)
			if value > 255 {
				next = pos
				break
			}
		}
		pos = next
		if part == 3 {
			if pos == limit {
				return pos, nil
			}
			break
		}
		if !p.at(pos, limit, '.') {
			break
		}
		pos++
	}
	if strict {
		return -1, p.failure("Malformed IPv4 address", next)
	}
	return -1, nil
}

func (p *workerURIParser) takeIPv4(start, end int, expected string) (int, *URISyntaxException) {
	pos, err := p.ipv4(start, end, true)
	if err != nil {
		return pos, err
	}
	if pos <= start {
		return pos, p.failure("Expected "+expected, start)
	}
	return pos, nil
}

func (p *workerURIParser) hexSeq(start, end int) (int, *URISyntaxException) {
	pos, _ := p.scan(start, end, workerURIHex)
	if pos <= start || p.at(pos, end, '.') {
		return -1, nil
	}
	if pos > start+4 {
		return pos, p.failure("IPv6 hexadecimal digit sequence too long", start)
	}
	p.ipv6bytes += 2
	for p.at(pos, end, ':') && !p.at(pos+1, end, ':') {
		pos++
		next, _ := p.scan(pos, end, workerURIHex)
		if next <= pos {
			return next, p.failure("Expected digits for an IPv6 address", pos)
		}
		if p.at(next, end, '.') {
			pos--
			break
		}
		if next > pos+4 {
			return next, p.failure("IPv6 hexadecimal digit sequence too long", pos)
		}
		p.ipv6bytes += 2
		pos = next
	}
	return pos, nil
}

func (p *workerURIParser) hexPost(start, end int) (int, *URISyntaxException) {
	if start == end {
		return start, nil
	}
	pos, err := p.hexSeq(start, end)
	if err != nil {
		return pos, err
	}
	if pos > start {
		if !p.at(pos, end, ':') {
			return pos, nil
		}
		pos++
	} else {
		pos = start
	}
	pos, err = p.takeIPv4(pos, end, "hex digits or IPv4 address")
	if err == nil {
		p.ipv6bytes += 4
	}
	return pos, err
}

func (p *workerURIParser) ipv6(start, end int) *URISyntaxException {
	pos, err := p.hexSeq(start, end)
	if err != nil {
		return err
	}
	if pos <= start {
		pos = start
	}
	compressed := p.at(pos, end, ':') && p.at(pos+1, end, ':')
	if compressed {
		pos, err = p.hexPost(pos+2, end)
	} else if pos > start && p.at(pos, end, ':') {
		pos, err = p.takeIPv4(pos+1, end, "IPv4 address")
		if err == nil {
			p.ipv6bytes += 4
		}
	}
	if err != nil {
		return err
	}
	if pos < end {
		return p.failure("Malformed IPv6 address", start)
	}
	if p.ipv6bytes > 16 {
		return p.failure("IPv6 address too long", start)
	}
	if !compressed && p.ipv6bytes < 16 {
		return p.failure("IPv6 address too short", start)
	}
	if compressed && p.ipv6bytes == 16 {
		return p.failure("Malformed IPv6 address", start)
	}
	return nil
}
