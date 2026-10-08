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
	case 1:
		return s.jj_3_1()
	case 2:
		return s.jj_3_2()
	case 3:
		return s.jj_3_3()
	case 4:
		return s.jj_3_4()
	case 5:
		return s.jj_3_5()
	case 6:
		return s.jj_3_6()
	case 7:
		return s.jj_3_7()
	case 8:
		return s.jj_3_8()
	case 9:
		return s.jj_3_9()
	case 10:
		return s.jj_3_10()
	case 11:
		return s.jj_3_11()
	case 12:
		return s.jj_3_12()
	case 13:
		return s.jj_3_13()
	case 14:
		return s.jj_3_14()
	case 15:
		return s.jj_3_15()
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
	case 21:
		return s.jj_3_21()
	case 22:
		return s.jj_3_22()
	case 23:
		return s.jj_3_23()
	case 24:
		return s.jj_3_24()
	case 25:
		return s.jj_3_25()
	case 26:
		return s.jj_3_26()
	case 27:
		return s.jj_3_27()
	case 28:
		return s.jj_3_28()
	case 29:
		return s.jj_3_29()
	case 30:
		return s.jj_3_30()
	case 31:
		return s.jj_3_31()
	case 32:
		return s.jj_3_32()
	case 33:
		return s.jj_3_33()
	case 34:
		return s.jj_3_34()
	case 35:
		return s.jj_3_35()
	case 36:
		return s.jj_3_36()
	case 37:
		return s.jj_3_37()
	case 38:
		return s.jj_3_38()
	case 39:
		return s.jj_3_39()
	case 40:
		return s.jj_3_40()
	case 41:
		return s.jj_3_41()
	case 42:
		return s.jj_3_42()
	case 43:
		return s.jj_3_43()
	case 44:
		return s.jj_3_44()
	case 45:
		return s.jj_3_45()
	case 46:
		return s.jj_3_46()
	case 47:
		return s.jj_3_47()
	case 48:
		return s.jj_3_48()
	case 49:
		return s.jj_3_49()
	case 50:
		return s.jj_3_50()
	case 51:
		return s.jj_3_51()
	case 52:
		return s.jj_3_52()
	case 53:
		return s.jj_3_53()
	case 54:
		return s.jj_3_54()
	case 55:
		return s.jj_3_55()
	case 56:
		return s.jj_3_56()
	case 57:
		return s.jj_3_57()
	case 58:
		return s.jj_3_58()
	case 59:
		return s.jj_3_59()
	case 60:
		return s.jj_3_60()
	case 61:
		return s.jj_3_61()
	case 62:
		return s.jj_3_62()
	case 63:
		return s.jj_3_63()
	case 64:
		return s.jj_3_64()
	case 65:
		return s.jj_3_65()
	case 66:
		return s.jj_3_66()
	case 67:
		return s.jj_3_67()
	case 68:
		return s.jj_3_68()
	case 69:
		return s.jj_3_69()
	case 70:
		return s.jj_3_70()
	case 71:
		return s.jj_3_71()
	case 72:
		return s.jj_3_72()
	case 73:
		return s.jj_3_73()
	case 74:
		return s.jj_3_74()
	default:
		panic("invalid JavaCC lookahead entry point")
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

func (p *SanyParser) rescanLookaheads(expected [][]SanyTokenKind) [][]SanyTokenKind {
	rescan := &sanyLookaheadRescan{entries: expected}
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
