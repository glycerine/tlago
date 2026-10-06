package tlago

// belchDEF ports TLAplusParser.belchDEF. Java temporarily reverses token links
// while looking ahead; indices into the same lazy token stream give us the
// reverse traversal without disturbing the forward links. The stop conditions
// and operator token ranges are those of the source, including its distinction
// between ASSUME and ASSUMPTION.
func (p *SanyParser) belchDEF() {
	floor := p.at - 1
	at := func(index int) *SanyToken {
		if index < 0 {
			return &SanyToken{Kind: SanyTokenEOF}
		}
		return p.tokenAt(index - p.at)
	}
	isOp := func(t *SanyToken) bool { return t.Kind >= SanyTokenOp57 && t.Kind <= SanyTokenOp119 }
	isInfix := func(t *SanyToken) bool { return t.Kind >= SanyTokenOp1 && t.Kind <= SanyTokenOp119 }
	isPostfix := func(t *SanyToken) bool { return t.Kind >= SanyTokenOp57 && t.Kind <= SanyTokenOp70 }
	for current := p.at; ; current++ {
		currentT, nextT := at(current), at(current+1)
		if nextT.Kind == SanyTokenEOF || nextT.Kind == SanyTokenTheorem || nextT.Kind == SanyTokenProposition ||
			nextT.Kind == SanyTokenAssumption || nextT.Kind == SanyTokenEndModule ||
			currentT.Kind == SanyTokenTheorem || currentT.Kind == SanyTokenProposition {
			return
		}
		if currentT.Kind != SanyTokenDef {
			continue
		}
		t := current - 1
		switch at(t).Kind {
		case SanyTokenRbr, SanyTokenRsb:
			depth := 1
			for t > floor {
				t--
				kind := at(t).Kind
				if (kind == SanyTokenLbr || kind == SanyTokenLsb) && depth == 1 {
					break
				}
				if kind == SanyTokenLbr || kind == SanyTokenLsb {
					depth--
				}
				if kind == SanyTokenRbr || kind == SanyTokenRsb {
					depth++
				}
			}
			if t == floor {
				return
			}
			if at(t-1).Kind == SanyTokenIdentifier {
				t--
			}
		case SanyTokenIdentifier:
			identifier := t
			if t > floor && isOp(at(t-1)) && !isPostfix(at(t-1)) &&
				(!isInfix(at(t-1)) || (t-1 > floor && at(t-2).Kind == SanyTokenIdentifier)) {
				t--
				if at(t-1).Kind == SanyTokenSubstitute {
					t = identifier
				} else if isInfix(at(t)) {
					t--
				}
			}
		default:
			if isOp(at(t)) {
				t--
			}
		}
		if t > floor && at(t-1).Kind == SanyTokenDefbreak {
			return
		}
		token := at(t)
		marker := &SanyToken{Kind: SanyTokenDefbreak, Image: "Beginning of definition", Begin: token.Begin, End: token.End, Next: token}
		p.tokens = append(p.tokens, nil)
		copy(p.tokens[t+1:], p.tokens[t:])
		p.tokens[t] = marker
		if t > 0 {
			p.tokens[t-1].Next = marker
		}
		if t < p.at {
			p.at++
		}
		return
	}
}
