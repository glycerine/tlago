package tlago

// JavaCC scanner state is separate from the parser's consumed token position.
// Alternatives restore position but retain lastPosition and the remaining budget.
// A matching token at the budget boundary throws LookaheadSuccess immediately.
type sanyLookahead struct {
	parser       *SanyParser
	startAt      int
	position     int
	lastPosition int
	remaining    int
	rescan       *sanyLookaheadRescan
}

type sanyLookaheadSuccess struct{}

type sanyLookaheadCall struct {
	generation int
	first      *SanyToken
	budget     int
	next       *sanyLookaheadCall
}

func (s *sanyLookahead) scanToken(kind SanyTokenKind) bool {
	if s.position == s.lastPosition {
		s.remaining--
		s.position++
		s.lastPosition = s.position
	} else {
		s.position++
	}
	token := s.parser.tokenAt(s.startAt + s.position - s.parser.at)
	if s.rescan != nil {
		distance := s.startAt + s.position - s.parser.at + 1
		if distance >= 0 {
			s.rescan.addErrorToken(kind, distance)
		}
	}
	if token.Kind != kind {
		return true
	}
	if s.remaining == 0 && s.position == s.lastPosition {
		panic(sanyLookaheadSuccess{})
	}
	return false
}

func (s *sanyLookahead) run(production int) bool {
	switch production {
	case 16:
		return s.jj_3_16()
	case 17:
		return s.jj_3_17()
	case 18:
		return s.jj_3_18()
	case 19:
		return s.jj_3_19()
	case 20:
		return s.jj_3_20()
	case 22:
		return s.jj_3_22()
	default:
		panic("unported JavaCC lookahead entry point")
	}
}

func (p *SanyParser) scanLookahead(production, budget int) (matches bool) {
	scanner := &sanyLookahead{parser: p, startAt: p.at, position: -1, lastPosition: -1, remaining: budget}
	first := p.peek()
	defer func() {
		// JavaCC saves successful and failed calls, including calls ending in
		// LookaheadSuccess. Expiration depends on how far scanning reached.
		call := &p.lookaheadCalls[production-1]
		for call.generation > p.at {
			if call.next == nil {
				call.next = &sanyLookaheadCall{}
				call = call.next
				break
			}
			call = call.next
		}
		call.generation = p.at + budget - scanner.remaining
		call.first = first
		call.budget = budget
		if failure := recover(); failure != nil {
			if _, ok := failure.(sanyLookaheadSuccess); !ok {
				panic(failure)
			}
			matches = true
		}
	}()
	return !scanner.run(production)
}

// Source getToken uses jj_scanpos while a semantic predicate is being evaluated.
func (s *sanyLookahead) token(index int) *SanyToken {
	return s.parser.tokenAt(s.startAt + s.position + index - s.parser.at)
}

func (s *sanyLookahead) matchFcnConst() bool {
	return s.parser.matchFcnConstAt(s.startAt + s.position + 1 - s.parser.at)
}

func (s *sanyLookahead) preInEmptyTop() bool {
	stack := s.parser.lookaheadOperatorStack
	return stack == nil || stack.PreInEmptyTop()
}

// JavaCC accumulates error token sequences during a rescan, including scans
// that ended successfully earlier. Preserve its 100-token sequence bound.
type sanyLookaheadRescan struct {
	lastTokens  [100]SanyTokenKind
	endPosition int
	entries     [][]SanyTokenKind
}

func (r *sanyLookaheadRescan) addErrorToken(kind SanyTokenKind, position int) {
	if position >= len(r.lastTokens) {
		return
	}
	if position == r.endPosition+1 {
		r.lastTokens[r.endPosition] = kind
		r.endPosition++
	} else if r.endPosition != 0 {
		entry := append([]SanyTokenKind(nil), r.lastTokens[:r.endPosition]...)
		exists := false
		for _, old := range r.entries {
			if len(old) == len(entry) {
				exists = true
				for i := range old {
					if old[i] != entry[i] {
						exists = false
						break
					}
				}
				if exists {
					break
				}
			}
		}
		if !exists {
			r.entries = append(r.entries, entry)
		}
		if position != 0 {
			r.endPosition = position
			r.lastTokens[position-1] = kind
		}
	}
}

func (p *SanyParser) rescanLookaheads() [][]SanyTokenKind {
	rescan := &sanyLookaheadRescan{}
	// JavaCC iterates entry points in production-index order, and catches
	// LookaheadSuccess outside each entry point's entire linked call list.
	for index := range p.lookaheadCalls {
		func() {
			defer func() {
				if failure := recover(); failure != nil {
					if _, ok := failure.(sanyLookaheadSuccess); !ok {
						panic(failure)
					}
				}
			}()
			for call := &p.lookaheadCalls[index]; call != nil; call = call.next {
				if call.generation <= p.at {
					continue
				}
				start := -1
				for i, token := range p.tokens {
					if token == call.first {
						start = i
						break
					}
				}
				if start < 0 {
					panic("saved JavaCC lookahead token is no longer in the parser stream")
				}
				scanner := &sanyLookahead{parser: p, startAt: start, position: -1, lastPosition: -1, remaining: call.budget, rescan: rescan}
				scanner.run(index + 1)
			}
		}()
	}
	rescan.addErrorToken(0, 0)
	return rescan.entries
}
