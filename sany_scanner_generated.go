// Copyright (c) 2003 Compaq Corporation. All rights reserved.
// Portions Copyright (c) 2003 Microsoft Corporation. All rights reserved.
// Mechanical port of the generated TLAplusParserTokenManager DFA/NFA methods.
package tlago

func sanyScannerPostIncrement(v *int32) int32            { old := *v; *v++; return old }
func sanyScannerPreIncrement(v *int32) int32             { *v++; return *v }
func sanyScannerPreDecrement(v *int32) int32             { *v--; return *v }
func sanyScannerSetInt32(v *int32, next int32) int32     { *v = next; return next }
func sanyScannerAndUint64(v *uint64, mask uint64) uint64 { *v &= mask; return *v }

func (tm *SanyTokenManager) jjStopStringLiteralDfa_0(pos int32, active0 uint64) int32 {
	switch pos {
	case 0:
		if (active0 & 0x4) != 0 {
			return 10
		}
		return -1
	case 1:
		if (active0 & 0x4) != 0 {
			return 9
		}
		return -1
	case 2:
		if (active0 & 0x4) != 0 {
			return 0
		}
		return -1
	default:
		return -1
	}
}

func (tm *SanyTokenManager) jjStartNfa_0(pos int32, active0 uint64) int32 {
	return tm.jjMoveNfa_0(tm.jjStopStringLiteralDfa_0(pos, active0), pos+1)
}

func (tm *SanyTokenManager) jjStopAtPos(pos int32, kind int32) int32 {
	tm.jjmatchedKind = kind
	tm.jjmatchedPos = pos
	return pos + 1
}

func (tm *SanyTokenManager) jjStartNfaWithStates_0(pos int32, kind int32, state int32) int32 {
	tm.jjmatchedKind = kind
	tm.jjmatchedPos = pos
	if !tm.readScanChar() {
		return pos + 1
	}
	return tm.jjMoveNfa_0(state, pos+1)
}

func (tm *SanyTokenManager) jjMoveStringLiteralDfa0_0() int32 {
	switch tm.curChar {
	case 45:
		return tm.jjMoveStringLiteralDfa1_0(0x4)
	default:
		return tm.jjMoveNfa_0(11, 0)
	}
}

func (tm *SanyTokenManager) jjMoveStringLiteralDfa1_0(active0 uint64) int32 {
	if !tm.readScanChar() {
		tm.jjStopStringLiteralDfa_0(0, active0)
		return 1
	}
	switch tm.curChar {
	case 45:
		return tm.jjMoveStringLiteralDfa2_0(active0, 0x4)
	default:
		break
	}
	return tm.jjStartNfa_0(0, active0)
}

func (tm *SanyTokenManager) jjMoveStringLiteralDfa2_0(old0 uint64, active0 uint64) int32 {
	if (sanyScannerAndUint64(&active0, old0)) == 0 {
		return tm.jjStartNfa_0(0, old0)
	}
	if !tm.readScanChar() {
		tm.jjStopStringLiteralDfa_0(1, active0)
		return 2
	}
	switch tm.curChar {
	case 45:
		return tm.jjMoveStringLiteralDfa3_0(active0, 0x4)
	default:
		break
	}
	return tm.jjStartNfa_0(1, active0)
}

func (tm *SanyTokenManager) jjMoveStringLiteralDfa3_0(old0 uint64, active0 uint64) int32 {
	if (sanyScannerAndUint64(&active0, old0)) == 0 {
		return tm.jjStartNfa_0(1, old0)
	}
	if !tm.readScanChar() {
		tm.jjStopStringLiteralDfa_0(2, active0)
		return 3
	}
	switch tm.curChar {
	case 62:
		if (active0 & 0x4) != 0 {
			return tm.jjStopAtPos(3, 2)
		}
		break
	default:
		break
	}
	return tm.jjStartNfa_0(2, active0)
}

func (tm *SanyTokenManager) jjMoveNfa_0(startState int32, curPos int32) int32 {

	var startsAt int32 = int32(0)
	_ = startsAt
	tm.jjnewStateCnt = 12
	var i int32 = int32(1)
	_ = i
	tm.jjstateSet[0] = startState
	var j int32
	_ = j
	var kind int32 = int32(0x7fffffff)
	_ = kind
	for {
		{
			if sanyScannerPreIncrement(&tm.jjround) == 0x7fffffff {
				tm.reInitRounds()
			}
			if tm.curChar < 64 {
				{
					var l uint64 = uint64(uint64(1) << tm.curChar)
					_ = l
					for {
						{
							switch tm.jjstateSet[sanyScannerPreDecrement(&i)] {
							case 0, 1:
								if tm.curChar == 45 {
									tm.jjCheckNAddStates(0, 2)
								}
								break
							case 2:
								if tm.curChar == 32 {
									tm.jjCheckNAddTwoStates(2, 8)
								}
								break
							case 9:
								if tm.curChar == 45 {
									tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 0
								}
								break
							case 10:
								if tm.curChar == 45 {
									tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 9
								}
								break
							case 11:
								if tm.curChar == 45 {
									tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 10
								}
								break
							default:
								break
							}
						}
						if !(i != startsAt) {
							break
						}
					}
				}
			} else {
				if tm.curChar < 128 {
					{
						var l uint64 = uint64(uint64(1) << (tm.curChar & 077))
						_ = l
						for {
							{
								switch tm.jjstateSet[sanyScannerPreDecrement(&i)] {
								case 3:
									if tm.curChar == 69 {
										kind = 3
									}
									break
								case 4:
									if tm.curChar == 76 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 3
									}
									break
								case 5:
									if tm.curChar == 85 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 4
									}
									break
								case 6:
									if tm.curChar == 68 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 5
									}
									break
								case 7:
									if tm.curChar == 79 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 6
									}
									break
								case 8:
									if tm.curChar == 77 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 7
									}
									break
								default:
									break
								}
							}
							if !(i != startsAt) {
								break
							}
						}
					}
				} else {
					{
						var hiByte int32 = int32(int32(tm.curChar >> 8))
						_ = hiByte
						var i1 int32 = int32(hiByte >> 6)
						_ = i1
						var l1 uint64 = uint64(uint64(1) << (hiByte & 077))
						_ = l1
						var i2 int32 = int32((tm.curChar & 0xff) >> 6)
						_ = i2
						var l2 uint64 = uint64(uint64(1) << (tm.curChar & 077))
						_ = l2
						for {
							{
								switch tm.jjstateSet[sanyScannerPreDecrement(&i)] {
								default:
									break
								}
							}
							if !(i != startsAt) {
								break
							}
						}
					}
				}
			}
			if kind != 0x7fffffff {
				{
					tm.jjmatchedKind = kind
					tm.jjmatchedPos = curPos
					kind = 0x7fffffff
				}
			}
			sanyScannerPreIncrement(&curPos)
			if (sanyScannerSetInt32(&i, tm.jjnewStateCnt)) == (sanyScannerSetInt32(&startsAt, 12-sanyScannerSetInt32(&tm.jjnewStateCnt, startsAt))) {
				return curPos
			}
			if !tm.readScanChar() {
				return curPos
			}
		}
	}
}

func (tm *SanyTokenManager) jjStopStringLiteralDfa_4(pos int32, active0 uint64) int32 {
	switch pos {
	default:
		return -1
	}
}

func (tm *SanyTokenManager) jjStartNfa_4(pos int32, active0 uint64) int32 {
	return tm.jjMoveNfa_4(tm.jjStopStringLiteralDfa_4(pos, active0), pos+1)
}

func (tm *SanyTokenManager) jjStartNfaWithStates_4(pos int32, kind int32, state int32) int32 {
	tm.jjmatchedKind = kind
	tm.jjmatchedPos = pos
	if !tm.readScanChar() {
		return pos + 1
	}
	return tm.jjMoveNfa_4(state, pos+1)
}

func (tm *SanyTokenManager) jjMoveStringLiteralDfa0_4() int32 {
	switch tm.curChar {
	case 42:
		return tm.jjMoveStringLiteralDfa1_4(0x80000000)
	default:
		return tm.jjMoveNfa_4(0, 0)
	}
}

func (tm *SanyTokenManager) jjMoveStringLiteralDfa1_4(active0 uint64) int32 {
	if !tm.readScanChar() {
		tm.jjStopStringLiteralDfa_4(0, active0)
		return 1
	}
	switch tm.curChar {
	case 41:
		if (active0 & 0x80000000) != 0 {
			return tm.jjStopAtPos(1, 31)
		}
		break
	default:
		break
	}
	return tm.jjStartNfa_4(0, active0)
}

func (tm *SanyTokenManager) jjMoveNfa_4(startState int32, curPos int32) int32 {

	var startsAt int32 = int32(0)
	_ = startsAt
	tm.jjnewStateCnt = 4
	var i int32 = int32(1)
	_ = i
	tm.jjstateSet[0] = startState
	var j int32
	_ = j
	var kind int32 = int32(0x7fffffff)
	_ = kind
	for {
		{
			if sanyScannerPreIncrement(&tm.jjround) == 0x7fffffff {
				tm.reInitRounds()
			}
			if tm.curChar < 64 {
				{
					var l uint64 = uint64(uint64(1) << tm.curChar)
					_ = l
					for {
						{
							switch tm.jjstateSet[sanyScannerPreDecrement(&i)] {
							case 0:
								if tm.curChar == 40 {
									tm.jjAddStates(3, 4)
								}
								break
							case 1:
								if tm.curChar == 46 {
									kind = 32
								}
								break
							case 2:
								if tm.curChar == 42 {
									tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 1
								}
								break
							case 3:
								if tm.curChar == 42 && kind > 32 {
									kind = 32
								}
								break
							default:
								break
							}
						}
						if !(i != startsAt) {
							break
						}
					}
				}
			} else {
				if tm.curChar < 128 {
					{
						var l uint64 = uint64(uint64(1) << (tm.curChar & 077))
						_ = l
						for {
							{
								switch tm.jjstateSet[sanyScannerPreDecrement(&i)] {
								default:
									break
								}
							}
							if !(i != startsAt) {
								break
							}
						}
					}
				} else {
					{
						var hiByte int32 = int32(int32(tm.curChar >> 8))
						_ = hiByte
						var i1 int32 = int32(hiByte >> 6)
						_ = i1
						var l1 uint64 = uint64(uint64(1) << (hiByte & 077))
						_ = l1
						var i2 int32 = int32((tm.curChar & 0xff) >> 6)
						_ = i2
						var l2 uint64 = uint64(uint64(1) << (tm.curChar & 077))
						_ = l2
						for {
							{
								switch tm.jjstateSet[sanyScannerPreDecrement(&i)] {
								default:
									break
								}
							}
							if !(i != startsAt) {
								break
							}
						}
					}
				}
			}
			if kind != 0x7fffffff {
				{
					tm.jjmatchedKind = kind
					tm.jjmatchedPos = curPos
					kind = 0x7fffffff
				}
			}
			sanyScannerPreIncrement(&curPos)
			if (sanyScannerSetInt32(&i, tm.jjnewStateCnt)) == (sanyScannerSetInt32(&startsAt, 4-sanyScannerSetInt32(&tm.jjnewStateCnt, startsAt))) {
				return curPos
			}
			if !tm.readScanChar() {
				return curPos
			}
		}
	}
}

func (tm *SanyTokenManager) jjMoveStringLiteralDfa0_1() int32 {
	return tm.jjMoveNfa_1(0, 0)
}

func (tm *SanyTokenManager) jjMoveNfa_1(startState int32, curPos int32) int32 {

	var startsAt int32 = int32(0)
	_ = startsAt
	tm.jjnewStateCnt = 58
	var i int32 = int32(1)
	_ = i
	tm.jjstateSet[0] = startState
	var j int32
	_ = j
	var kind int32 = int32(0x7fffffff)
	_ = kind
	for {
		{
			if sanyScannerPreIncrement(&tm.jjround) == 0x7fffffff {
				tm.reInitRounds()
			}
			if tm.curChar < 64 {
				{
					var l uint64 = uint64(uint64(1) << tm.curChar)
					_ = l
					for {
						{
							switch tm.jjstateSet[sanyScannerPreDecrement(&i)] {
							case 0:
								if (0x3ff000000000000 & l) != 0 {
									{
										if kind > 20 {
											kind = 20
										}
										tm.jjCheckNAddStates(5, 13)
									}
								} else {
									if tm.curChar == 45 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 11
									}
								}
								if tm.curChar == 48 {
									{
										if kind > 20 {
											kind = 20
										}
									}
								}
								break
							case 1, 2:
								if tm.curChar == 45 {
									tm.jjCheckNAddStates(14, 16)
								}
								break
							case 3:
								if tm.curChar == 32 {
									tm.jjCheckNAddTwoStates(3, 9)
								}
								break
							case 10:
								if tm.curChar == 45 {
									tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 1
								}
								break
							case 11:
								if tm.curChar == 45 {
									tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 10
								}
								break
							case 12:
								if tm.curChar == 45 {
									tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 11
								}
								break
							case 14:
								if (0x3ff000000000000 & l) != 0 {
									tm.jjAddStates(17, 18)
								}
								break
							case 16:
								if (0x3ff000000000000 & l) == 0 {
									break
								}
								if kind > 289 {
									kind = 289
								}
								tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 16
								break
							case 18, 19:
								if (0x3ff000000000000 & l) == 0 {
									break
								}
								if kind > 289 {
									kind = 289
								}
								tm.jjCheckNAdd(19)
								break
							case 23:
								if (0x3ff000000000000 & l) == 0 {
									break
								}
								if kind > 20 {
									kind = 20
								}
								tm.jjCheckNAddStates(5, 13)
								break
							case 24:
								if (0x3ff000000000000 & l) == 0 {
									break
								}
								if kind > 20 {
									kind = 20
								}
								tm.jjCheckNAdd(24)
								break
							case 25:
								if (0x3ff000000000000 & l) != 0 {
									tm.jjCheckNAddTwoStates(25, 28)
								}
								break
							case 27, 42, 52:
								if tm.curChar == 47 {
									tm.jjCheckNAdd(26)
								}
								break
							case 28:
								if tm.curChar == 46 {
									tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 27
								}
								break
							case 29:
								if (0x3ff000000000000 & l) != 0 {
									tm.jjCheckNAddTwoStates(29, 32)
								}
								break
							case 30:
								if tm.curChar == 47 && kind > 112 {
									kind = 112
								}
								break
							case 32:
								if tm.curChar == 46 {
									tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 31
								}
								break
							case 33:
								if (0x3ff000000000000 & l) != 0 {
									tm.jjCheckNAddTwoStates(33, 34)
								}
								break
							case 35:
								if (0x3ff000000000000 & l) == 0 {
									break
								}
								if kind > 289 {
									kind = 289
								}
								tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 35
								break
							case 36:
								if (0x3ff000000000000 & l) != 0 {
									tm.jjCheckNAddTwoStates(36, 37)
								}
								break
							case 38:
								if (0x3ff000000000000 & l) == 0 {
									break
								}
								if kind > 289 {
									kind = 289
								}
								tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 38
								break
							case 40, 41:
								if (0x3ff000000000000 & l) != 0 {
									tm.jjCheckNAddTwoStates(41, 43)
								}
								break
							case 43:
								if tm.curChar == 46 {
									tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 42
								}
								break
							case 44, 45:
								if (0x3ff000000000000 & l) != 0 {
									tm.jjCheckNAddTwoStates(45, 47)
								}
								break
							case 47:
								if tm.curChar == 46 {
									tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 46
								}
								break
							case 48, 49:
								if (0x3ff000000000000 & l) == 0 {
									break
								}
								if kind > 289 {
									kind = 289
								}
								tm.jjCheckNAdd(49)
								break
							case 51:
								if (0x3ff000000000000 & l) != 0 {
									tm.jjAddStates(19, 20)
								}
								break
							case 53:
								if tm.curChar == 46 {
									tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 52
								}
								break
							case 54:
								if (0x3ff000000000000 & l) != 0 {
									tm.jjAddStates(21, 22)
								}
								break
							case 56:
								if tm.curChar == 46 {
									tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 55
								}
								break
							case 57:
								if (0x3ff000000000000 & l) == 0 {
									break
								}
								if kind > 289 {
									kind = 289
								}
								tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 57
								break
							default:
								break
							}
						}
						if !(i != startsAt) {
							break
						}
					}
				}
			} else {
				if tm.curChar < 128 {
					{
						var l uint64 = uint64(uint64(1) << (tm.curChar & 077))
						_ = l
						for {
							{
								switch tm.jjstateSet[sanyScannerPreDecrement(&i)] {
								case 0:
									if (0x7fffffe0777fffe & l) != 0 {
										{
											if kind > 289 {
												kind = 289
											}
											tm.jjCheckNAddStates(23, 27)
										}
									} else {
										if (0x880000 & l) != 0 {
											{
												if kind > 289 {
													kind = 289
												}
												tm.jjAddStates(28, 30)
											}
										} else {
											if tm.curChar == 64 {
												{
													if kind > 289 {
														kind = 289
													}
												}
											} else {
												if tm.curChar == 95 {
													tm.jjCheckNAddTwoStates(14, 15)
												}
											}
										}
									}
									if tm.curChar == 83 {
										tm.jjCheckNAdd(17)
									} else {
										if tm.curChar == 87 {
											tm.jjCheckNAdd(17)
										}
									}
									break
								case 4:
									if tm.curChar == 69 && kind > 21 {
										kind = 21
									}
									break
								case 5:
									if tm.curChar == 76 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 4
									}
									break
								case 6:
									if tm.curChar == 85 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 5
									}
									break
								case 7:
									if tm.curChar == 68 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 6
									}
									break
								case 8:
									if tm.curChar == 79 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 7
									}
									break
								case 9:
									if tm.curChar == 77 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 8
									}
									break
								case 13:
									if tm.curChar == 95 {
										tm.jjCheckNAddTwoStates(14, 15)
									}
									break
								case 14:
									if (0x7fffffe87fffffe & l) != 0 {
										tm.jjCheckNAddTwoStates(14, 15)
									}
									break
								case 15:
									if (0x7fffffe07fffffe & l) == 0 {
										break
									}
									if kind > 289 {
										kind = 289
									}
									tm.jjCheckNAdd(16)
									break
								case 16:
									if (0x7fffffe87fffffe & l) == 0 {
										break
									}
									if kind > 289 {
										kind = 289
									}
									tm.jjCheckNAdd(16)
									break
								case 17:
									if tm.curChar != 70 {
										break
									}
									if kind > 289 {
										kind = 289
									}
									tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 18
									break
								case 18:
									if (0x7fffffe07fffffe & l) == 0 {
										break
									}
									if kind > 289 {
										kind = 289
									}
									tm.jjCheckNAdd(19)
									break
								case 19:
									if (0x7fffffe87fffffe & l) == 0 {
										break
									}
									if kind > 289 {
										kind = 289
									}
									tm.jjCheckNAdd(19)
									break
								case 20:
									if tm.curChar == 87 {
										tm.jjCheckNAdd(17)
									}
									break
								case 21:
									if tm.curChar == 83 {
										tm.jjCheckNAdd(17)
									}
									break
								case 22:
									if tm.curChar == 64 {
										kind = 289
									}
									break
								case 25:
									if (0x7fffffe07fffffe & l) != 0 {
										tm.jjAddStates(31, 32)
									}
									break
								case 26:
									if tm.curChar == 92 && kind > 111 {
										kind = 111
									}
									break
								case 29:
									if (0x7fffffe07fffffe & l) != 0 {
										tm.jjAddStates(33, 34)
									}
									break
								case 31, 46, 55:
									if tm.curChar == 92 {
										tm.jjCheckNAdd(30)
									}
									break
								case 33:
									if (0x7fffffe87fffffe & l) != 0 {
										tm.jjAddStates(35, 36)
									}
									break
								case 34:
									if (0x7fffffe07fffffe & l) == 0 {
										break
									}
									if kind > 289 {
										kind = 289
									}
									tm.jjCheckNAdd(35)
									break
								case 35:
									if (0x7fffffe87fffffe & l) == 0 {
										break
									}
									if kind > 289 {
										kind = 289
									}
									tm.jjCheckNAdd(35)
									break
								case 37:
									if (0x7fffffe07fffffe & l) == 0 {
										break
									}
									if kind > 289 {
										kind = 289
									}
									tm.jjCheckNAdd(38)
									break
								case 38:
									if (0x7fffffe87fffffe & l) == 0 {
										break
									}
									if kind > 289 {
										kind = 289
									}
									tm.jjCheckNAdd(38)
									break
								case 39:
									if (0x880000 & l) == 0 {
										break
									}
									if kind > 289 {
										kind = 289
									}
									tm.jjAddStates(28, 30)
									break
								case 40:
									if (0x7fffffe07ffffbe & l) != 0 {
										tm.jjCheckNAddTwoStates(41, 43)
									}
									break
								case 41:
									if (0x7fffffe07fffffe & l) != 0 {
										tm.jjCheckNAddTwoStates(41, 43)
									}
									break
								case 44:
									if (0x7fffffe07ffffbe & l) != 0 {
										tm.jjCheckNAddTwoStates(45, 47)
									}
									break
								case 45:
									if (0x7fffffe07fffffe & l) != 0 {
										tm.jjCheckNAddTwoStates(45, 47)
									}
									break
								case 48:
									if (0x7fffffe87ffffbe & l) == 0 {
										break
									}
									if kind > 289 {
										kind = 289
									}
									tm.jjCheckNAdd(49)
									break
								case 49:
									if (0x7fffffe87fffffe & l) == 0 {
										break
									}
									if kind > 289 {
										kind = 289
									}
									tm.jjCheckNAdd(49)
									break
								case 50:
									if (0x7fffffe0777fffe & l) == 0 {
										break
									}
									if kind > 289 {
										kind = 289
									}
									tm.jjCheckNAddStates(23, 27)
									break
								case 51:
									if (0x7fffffe07fffffe & l) != 0 {
										tm.jjCheckNAddTwoStates(51, 53)
									}
									break
								case 54:
									if (0x7fffffe07fffffe & l) != 0 {
										tm.jjCheckNAddTwoStates(54, 56)
									}
									break
								case 57:
									if (0x7fffffe87fffffe & l) == 0 {
										break
									}
									if kind > 289 {
										kind = 289
									}
									tm.jjCheckNAdd(57)
									break
								default:
									break
								}
							}
							if !(i != startsAt) {
								break
							}
						}
					}
				} else {
					{
						var hiByte int32 = int32(int32(tm.curChar >> 8))
						_ = hiByte
						var i1 int32 = int32(hiByte >> 6)
						_ = i1
						var l1 uint64 = uint64(uint64(1) << (hiByte & 077))
						_ = l1
						var i2 int32 = int32((tm.curChar & 0xff) >> 6)
						_ = i2
						var l2 uint64 = uint64(uint64(1) << (tm.curChar & 077))
						_ = l2
						for {
							{
								switch tm.jjstateSet[sanyScannerPreDecrement(&i)] {
								case 0:
									if tm.jjCanMove_0(hiByte, i1, i2, l1, l2) && kind > 289 {
										kind = 289
									}
									break
								default:
									break
								}
							}
							if !(i != startsAt) {
								break
							}
						}
					}
				}
			}
			if kind != 0x7fffffff {
				{
					tm.jjmatchedKind = kind
					tm.jjmatchedPos = curPos
					kind = 0x7fffffff
				}
			}
			sanyScannerPreIncrement(&curPos)
			if (sanyScannerSetInt32(&i, tm.jjnewStateCnt)) == (sanyScannerSetInt32(&startsAt, 58-sanyScannerSetInt32(&tm.jjnewStateCnt, startsAt))) {
				return curPos
			}
			if !tm.readScanChar() {
				return curPos
			}
		}
	}
}

func (tm *SanyTokenManager) jjMoveStringLiteralDfa0_5() int32 {
	return tm.jjMoveNfa_5(0, 0)
}

func (tm *SanyTokenManager) jjMoveNfa_5(startState int32, curPos int32) int32 {

	var startsAt int32 = int32(0)
	_ = startsAt
	tm.jjnewStateCnt = 3
	var i int32 = int32(1)
	_ = i
	tm.jjstateSet[0] = startState
	var j int32
	_ = j
	var kind int32 = int32(0x7fffffff)
	_ = kind
	for {
		{
			if sanyScannerPreIncrement(&tm.jjround) == 0x7fffffff {
				tm.reInitRounds()
			}
			if tm.curChar < 64 {
				{
					var l uint64 = uint64(uint64(1) << tm.curChar)
					_ = l
					for {
						{
							switch tm.jjstateSet[sanyScannerPreDecrement(&i)] {
							case 0:
								if (0x2400 & l) != 0 {
									{
										if kind > 33 {
											kind = 33
										}
									}
								}
								if tm.curChar == 13 {
									tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 1
								}
								break
							case 1:
								if tm.curChar == 10 && kind > 33 {
									kind = 33
								}
								break
							case 2:
								if tm.curChar == 13 {
									tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 1
								}
								break
							default:
								break
							}
						}
						if !(i != startsAt) {
							break
						}
					}
				}
			} else {
				if tm.curChar < 128 {
					{
						var l uint64 = uint64(uint64(1) << (tm.curChar & 077))
						_ = l
						for {
							{
								switch tm.jjstateSet[sanyScannerPreDecrement(&i)] {
								default:
									break
								}
							}
							if !(i != startsAt) {
								break
							}
						}
					}
				} else {
					{
						var hiByte int32 = int32(int32(tm.curChar >> 8))
						_ = hiByte
						var i1 int32 = int32(hiByte >> 6)
						_ = i1
						var l1 uint64 = uint64(uint64(1) << (hiByte & 077))
						_ = l1
						var i2 int32 = int32((tm.curChar & 0xff) >> 6)
						_ = i2
						var l2 uint64 = uint64(uint64(1) << (tm.curChar & 077))
						_ = l2
						for {
							{
								switch tm.jjstateSet[sanyScannerPreDecrement(&i)] {
								default:
									break
								}
							}
							if !(i != startsAt) {
								break
							}
						}
					}
				}
			}
			if kind != 0x7fffffff {
				{
					tm.jjmatchedKind = kind
					tm.jjmatchedPos = curPos
					kind = 0x7fffffff
				}
			}
			sanyScannerPreIncrement(&curPos)
			if (sanyScannerSetInt32(&i, tm.jjnewStateCnt)) == (sanyScannerSetInt32(&startsAt, 3-sanyScannerSetInt32(&tm.jjnewStateCnt, startsAt))) {
				return curPos
			}
			if !tm.readScanChar() {
				return curPos
			}
		}
	}
}

func (tm *SanyTokenManager) jjStopStringLiteralDfa_2(pos int32, active0 uint64, active1 uint64, active2 uint64, active3 uint64, active4 uint64) int32 {
	switch pos {
	case 0:
		if (active1&0x40000000) != 0 || (active3&0x7c0000000) != 0 {
			return 108
		}
		if (active0 & 0x8000000000) != 0 {
			{
				tm.jjmatchedKind = 289
				return 249
			}
		}
		if (active1&0x2000000) != 0 || (active3&0x38000000) != 0 {
			return 18
		}
		if (active1&0x20000100000000) != 0 || (active2&0xc000000000000000) != 0 || (active3&0x3) != 0 {
			return 121
		}
		if (active0&0x83ad600000000000) != 0 || (active1&0x580000000000c1f1) != 0 {
			{
				tm.jjmatchedKind = 289
				return 269
			}
		}
		if (active2 & 0xd) != 0 {
			return 46
		}
		if (active1 & 0x8000000000000002) != 0 {
			{
				tm.jjmatchedKind = 289
				return 102
			}
		}
		if (active1 & 0x41800) != 0 {
			{
				tm.jjmatchedKind = 289
				return 11
			}
		}
		if (active2 & 0x3c0000000000000) != 0 {
			return 131
		}
		if (active1&0x18000000000000) != 0 || (active2&0x3000000000000000) != 0 {
			return 38
		}
		if (active1 & 0x10000000) != 0 {
			return 270
		}
		if (active0&0x800000000000000) != 0 || (active1&0x2000000000d90000) != 0 {
			{
				tm.jjmatchedKind = 289
				return 181
			}
		}
		if (active0 & 0xc0000000000) != 0 {
			{
				tm.jjmatchedKind = 289
				return 140
			}
		}
		if (active0&0x3000000010000000) != 0 || (active1&0x80000000000000) != 0 || (active2&0x1ffffffff7ffe0) != 0 || (active3&0x800000000) != 0 || (active4&0x180000000) != 0 {
			return 200
		}
		if (active3 & 0xc00) != 0 {
			return 162
		}
		if (active0&0x10000000000) != 0 || (active1&0x200002000) != 0 {
			return 42
		}
		if (active3 & 0x2000000) != 0 {
			{
				tm.jjmatchedKind = 289
				return -1
			}
		}
		if (active3 & 0x3c000) != 0 {
			return 29
		}
		if (active0&0x50000000000000) != 0 || (active1&0x200) != 0 {
			{
				tm.jjmatchedKind = 289
				return 16
			}
		}
		if (active3 & 0x3c0) != 0 {
			return 158
		}
		if (active0&0x4000000000000000) != 0 || (active1&0x408) != 0 {
			{
				tm.jjmatchedKind = 289
				return 91
			}
		}
		return -1
	case 1:
		if (active2 & 0x100000) != 0 {
			return 219
		}
		if (active3 & 0x10000) != 0 {
			return 28
		}
		if (active2 & 0x3c00000) != 0 {
			return 216
		}
		if (active1 & 0x1800) != 0 {
			{
				if tm.jjmatchedPos != 1 {
					{
						tm.jjmatchedKind = 289
						tm.jjmatchedPos = 1
					}
				}
				return 10
			}
		}
		if (active1 & 0x2000000000990000) != 0 {
			{
				if tm.jjmatchedPos != 1 {
					{
						tm.jjmatchedKind = 289
						tm.jjmatchedPos = 1
					}
				}
				return 271
			}
		}
		if (active0 & 0x1000000000000000) != 0 {
			{
				if tm.jjmatchedPos != 1 {
					{
						tm.jjmatchedKind = 47
						tm.jjmatchedPos = 1
					}
				}
				return -1
			}
		}
		if (active0&0x800000000000000) != 0 || (active1&0x400000) != 0 {
			{
				if tm.jjmatchedPos != 1 {
					{
						tm.jjmatchedKind = 289
						tm.jjmatchedPos = 1
					}
				}
				return 56
			}
		}
		if (active3 & 0x20000000) != 0 {
			{
				if tm.jjmatchedPos != 1 {
					{
						tm.jjmatchedKind = 90
						tm.jjmatchedPos = 1
					}
				}
				return -1
			}
		}
		if (active2 & 0x10000) != 0 {
			return 205
		}
		if (active0&0x43c16c0000000000) != 0 || (active1&0xd80000000004c7f9) != 0 {
			{
				if tm.jjmatchedPos != 1 {
					{
						tm.jjmatchedKind = 289
						tm.jjmatchedPos = 1
					}
				}
				return 269
			}
		}
		if (active0 & 0x802c000000000000) != 0 {
			return 269
		}
		if (active0 & 0x8000000000) != 0 {
			{
				if tm.jjmatchedPos != 1 {
					{
						tm.jjmatchedKind = 289
						tm.jjmatchedPos = 1
					}
				}
				return 263
			}
		}
		if (active2 & 0xfc000000) != 0 {
			return 63
		}
		if (active1 & 0x2) != 0 {
			{
				if tm.jjmatchedPos != 1 {
					{
						tm.jjmatchedKind = 289
						tm.jjmatchedPos = 1
					}
				}
				return 101
			}
		}
		if (active0&0x10000000000) != 0 || (active1&0x2000) != 0 {
			{
				if tm.jjmatchedPos != 1 {
					{
						tm.jjmatchedKind = 121
						tm.jjmatchedPos = 1
					}
				}
				return -1
			}
		}
		if (active0 & 0x2000000000000000) != 0 {
			{
				if tm.jjmatchedPos != 1 {
					{
						tm.jjmatchedKind = 49
						tm.jjmatchedPos = 1
					}
				}
				return -1
			}
		}
		if (active2 & 0x4000000000000000) != 0 {
			return 120
		}
		if (active3 & 0x2000000) != 0 {
			{
				if tm.jjmatchedPos == 0 {
					{
						tm.jjmatchedKind = 289
						tm.jjmatchedPos = 0
					}
				}
				return -1
			}
		}
		if (active2 & 0x180) != 0 {
			return 65
		}
		if (active0 & 0x10000000000000) != 0 {
			{
				if tm.jjmatchedPos != 1 {
					{
						tm.jjmatchedKind = 289
						tm.jjmatchedPos = 1
					}
				}
				return 15
			}
		}
		return -1
	case 2:
		if (active0&0x10000000000) != 0 || (active1&0x2000) != 0 {
			{
				if tm.jjmatchedPos < 1 {
					{
						tm.jjmatchedKind = 121
						tm.jjmatchedPos = 1
					}
				}
				return -1
			}
		}
		if (active2 & 0x100000) != 0 {
			{
				tm.jjmatchedKind = 147
				tm.jjmatchedPos = 2
				return -1
			}
		}
		if (active0&0x110000000000000) != 0 || (active1&0x4010) != 0 {
			return 269
		}
		if (active0&0x42c96c0000000000) != 0 || (active1&0xd8000000000487e9) != 0 {
			{
				tm.jjmatchedKind = 289
				tm.jjmatchedPos = 2
				return 269
			}
		}
		if (active0 & 0x1000000000000000) != 0 {
			{
				if tm.jjmatchedPos < 1 {
					{
						tm.jjmatchedKind = 47
						tm.jjmatchedPos = 1
					}
				}
				return -1
			}
		}
		if (active1 & 0x2000000000990000) != 0 {
			{
				tm.jjmatchedKind = 289
				tm.jjmatchedPos = 2
				return 271
			}
		}
		if (active0 & 0x8000000000) != 0 {
			{
				tm.jjmatchedKind = 289
				tm.jjmatchedPos = 2
				return 262
			}
		}
		if (active1 & 0x1800) != 0 {
			{
				tm.jjmatchedKind = 289
				tm.jjmatchedPos = 2
				return 9
			}
		}
		if (active3 & 0x20000000) != 0 {
			{
				if tm.jjmatchedPos < 1 {
					{
						tm.jjmatchedKind = 90
						tm.jjmatchedPos = 1
					}
				}
				return -1
			}
		}
		if (active1 & 0x2) != 0 {
			{
				tm.jjmatchedKind = 66
				tm.jjmatchedPos = 2
				return 103
			}
		}
		if (active0 & 0x2000000000000000) != 0 {
			{
				if tm.jjmatchedPos < 1 {
					{
						tm.jjmatchedKind = 49
						tm.jjmatchedPos = 1
					}
				}
				return -1
			}
		}
		return -1
	case 3:
		if (active2 & 0x100000) != 0 {
			{
				if tm.jjmatchedPos < 2 {
					{
						tm.jjmatchedKind = 147
						tm.jjmatchedPos = 2
					}
				}
				return -1
			}
		}
		if (active0&0x10000000000) != 0 || (active1&0x2000) != 0 {
			{
				if tm.jjmatchedPos < 1 {
					{
						tm.jjmatchedKind = 121
						tm.jjmatchedPos = 1
					}
				}
				return -1
			}
		}
		if (active0&0x2c9480000000000) != 0 || (active1&0xd800000000009b8a) != 0 {
			{
				if tm.jjmatchedPos != 3 {
					{
						tm.jjmatchedKind = 289
						tm.jjmatchedPos = 3
					}
				}
				return 269
			}
		}
		if (active1 & 0x2000000000190000) != 0 {
			{
				if tm.jjmatchedPos != 3 {
					{
						tm.jjmatchedKind = 289
						tm.jjmatchedPos = 3
					}
				}
				return 271
			}
		}
		if (active0&0x4000240000000000) != 0 || (active1&0x40461) != 0 {
			return 269
		}
		if (active1 & 0x800000) != 0 {
			return 271
		}
		if (active0 & 0x8000000000) != 0 {
			{
				if tm.jjmatchedPos != 3 {
					{
						tm.jjmatchedKind = 289
						tm.jjmatchedPos = 3
					}
				}
				return 261
			}
		}
		return -1
	case 4:
		if (active2 & 0x100000) != 0 {
			{
				if tm.jjmatchedPos < 2 {
					{
						tm.jjmatchedKind = 147
						tm.jjmatchedPos = 2
					}
				}
				return -1
			}
		}
		if (active0 & 0x8000000000) != 0 {
			{
				if tm.jjmatchedPos != 4 {
					{
						tm.jjmatchedKind = 289
						tm.jjmatchedPos = 4
					}
				}
				return 260
			}
		}
		if (active1 & 0x2000000000180000) != 0 {
			{
				if tm.jjmatchedPos != 4 {
					{
						tm.jjmatchedKind = 289
						tm.jjmatchedPos = 4
					}
				}
				return 271
			}
		}
		if (active0&0x89480000000000) != 0 || (active1&0x980000000000838a) != 0 {
			{
				if tm.jjmatchedPos != 4 {
					{
						tm.jjmatchedKind = 289
						tm.jjmatchedPos = 4
					}
				}
				return 269
			}
		}
		if (active0&0x10000000000) != 0 || (active1&0x2000) != 0 {
			{
				if tm.jjmatchedPos < 1 {
					{
						tm.jjmatchedKind = 121
						tm.jjmatchedPos = 1
					}
				}
				return -1
			}
		}
		if (active0&0x240000000000000) != 0 || (active1&0x4000000000001800) != 0 {
			return 269
		}
		if (active1 & 0x10000) != 0 {
			return 271
		}
		return -1
	case 5:
		if (active2 & 0x100000) != 0 {
			{
				if tm.jjmatchedPos < 2 {
					{
						tm.jjmatchedKind = 147
						tm.jjmatchedPos = 2
					}
				}
				return -1
			}
		}
		if (active0&0x9000000000000) != 0 || (active1&0x1800000000008188) != 0 {
			{
				tm.jjmatchedKind = 289
				tm.jjmatchedPos = 5
				return 269
			}
		}
		if (active0&0x10000000000) != 0 || (active1&0x2000) != 0 {
			{
				if tm.jjmatchedPos < 1 {
					{
						tm.jjmatchedKind = 121
						tm.jjmatchedPos = 1
					}
				}
				return -1
			}
		}
		if (active0&0x80488000000000) != 0 || (active1&0x8000000000000202) != 0 {
			return 269
		}
		if (active1 & 0x2000000000000000) != 0 {
			return 271
		}
		if (active1 & 0x180000) != 0 {
			{
				tm.jjmatchedKind = 289
				tm.jjmatchedPos = 5
				return 271
			}
		}
		return -1
	case 6:
		if (active2 & 0x100000) != 0 {
			{
				if tm.jjmatchedPos < 2 {
					{
						tm.jjmatchedKind = 147
						tm.jjmatchedPos = 2
					}
				}
				return -1
			}
		}
		if (active0&0x10000000000) != 0 || (active1&0x2000) != 0 {
			{
				if tm.jjmatchedPos < 1 {
					{
						tm.jjmatchedKind = 121
						tm.jjmatchedPos = 1
					}
				}
				return -1
			}
		}
		if (active1 & 0x100000) != 0 {
			{
				if tm.jjmatchedPos != 6 {
					{
						tm.jjmatchedKind = 289
						tm.jjmatchedPos = 6
					}
				}
				return 271
			}
		}
		if (active0&0x8000000000000) != 0 || (active1&0x1000000000008000) != 0 {
			{
				if tm.jjmatchedPos != 6 {
					{
						tm.jjmatchedKind = 289
						tm.jjmatchedPos = 6
					}
				}
				return 269
			}
		}
		if (active0&0x1000000000000) != 0 || (active1&0x800000000000188) != 0 {
			return 269
		}
		if (active1 & 0x80000) != 0 {
			return 271
		}
		return -1
	case 7:
		if (active2 & 0x100000) != 0 {
			{
				if tm.jjmatchedPos < 2 {
					{
						tm.jjmatchedKind = 147
						tm.jjmatchedPos = 2
					}
				}
				return -1
			}
		}
		if (active1 & 0x1000000000008000) != 0 {
			{
				tm.jjmatchedKind = 289
				tm.jjmatchedPos = 7
				return 269
			}
		}
		if (active0 & 0x10000000000) != 0 {
			{
				if tm.jjmatchedPos < 1 {
					{
						tm.jjmatchedKind = 121
						tm.jjmatchedPos = 1
					}
				}
				return -1
			}
		}
		if (active0 & 0x8000000000000) != 0 {
			return 269
		}
		if (active1 & 0x100000) != 0 {
			return 271
		}
		return -1
	case 8:
		if (active2 & 0x100000) != 0 {
			{
				if tm.jjmatchedPos < 2 {
					{
						tm.jjmatchedKind = 147
						tm.jjmatchedPos = 2
					}
				}
				return -1
			}
		}
		if (active1 & 0x1000000000008000) != 0 {
			return 269
		}
		return -1
	case 9:
		if (active2 & 0x100000) != 0 {
			{
				if tm.jjmatchedPos < 2 {
					{
						tm.jjmatchedKind = 147
						tm.jjmatchedPos = 2
					}
				}
				return -1
			}
		}
		return -1
	default:
		return -1
	}
}

func (tm *SanyTokenManager) jjStartNfa_2(pos int32, active0 uint64, active1 uint64, active2 uint64, active3 uint64, active4 uint64) int32 {
	return tm.jjMoveNfa_2(tm.jjStopStringLiteralDfa_2(pos, active0, active1, active2, active3, active4), pos+1)
}

func (tm *SanyTokenManager) jjStartNfaWithStates_2(pos int32, kind int32, state int32) int32 {
	tm.jjmatchedKind = kind
	tm.jjmatchedPos = pos
	if !tm.readScanChar() {
		return pos + 1
	}
	return tm.jjMoveNfa_2(state, pos+1)
}

func (tm *SanyTokenManager) jjMoveStringLiteralDfa0_2() int32 {
	switch tm.curChar {
	case 33:
		tm.jjmatchedKind = 105
		return tm.jjMoveStringLiteralDfa1_2(0x0, 0x0, 0x0, 0x4000000, 0x0)
	case 35:
		tm.jjmatchedKind = 187
		return tm.jjMoveStringLiteralDfa1_2(0x0, 0x0, 0x400000000000000, 0x0, 0x0)
	case 36:
		tm.jjmatchedKind = 213
		return tm.jjMoveStringLiteralDfa1_2(0x0, 0x0, 0x0, 0x100000, 0x0)
	case 37:
		tm.jjmatchedKind = 216
		return tm.jjMoveStringLiteralDfa1_2(0x0, 0x0, 0x0, 0x800000, 0x0)
	case 38:
		tm.jjmatchedKind = 211
		return tm.jjMoveStringLiteralDfa1_2(0x0, 0x0, 0x0, 0x40000, 0x0)
	case 39:
		return tm.jjStopAtPos(0, 113)
	case 40:
		tm.jjmatchedKind = 94
		return tm.jjMoveStringLiteralDfa1_2(0x0, 0x0, 0x0, 0x7c0000000, 0x0)
	case 41:
		return tm.jjStopAtPos(0, 95)
	case 42:
		tm.jjmatchedKind = 195
		return tm.jjMoveStringLiteralDfa1_2(0x0, 0x0, 0x0, 0x4, 0x0)
	case 43:
		tm.jjmatchedKind = 197
		return tm.jjMoveStringLiteralDfa1_2(0x0, 0x0, 0x0, 0x10, 0x0)
	case 44:
		return tm.jjStopAtPos(0, 88)
	case 45:
		tm.jjmatchedKind = 193
		return tm.jjMoveStringLiteralDfa1_2(0x0, 0x20000100000000, 0xc000000000000000, 0x1, 0x0)
	case 46:
		tm.jjmatchedKind = 91
		return tm.jjMoveStringLiteralDfa1_2(0x0, 0x0, 0x0, 0x3000, 0x0)
	case 47:
		tm.jjmatchedKind = 131
		return tm.jjMoveStringLiteralDfa1_2(0x0, 0x0, 0x5, 0x0, 0x0)
	case 58:
		tm.jjmatchedKind = 89
		return tm.jjMoveStringLiteralDfa1_2(0x0, 0x0, 0x0, 0x38000000, 0x0)
	case 60:
		tm.jjmatchedKind = 201
		return tm.jjMoveStringLiteralDfa1_2(0x0, 0x0, 0x0, 0x1c0, 0x0)
	case 61:
		tm.jjmatchedKind = 185
		return tm.jjMoveStringLiteralDfa1_2(0x0, 0x0, 0x1c0000000000000, 0x0, 0x0)
	case 62:
		tm.jjmatchedKind = 203
		return tm.jjMoveStringLiteralDfa1_2(0x0, 0x0, 0x0, 0x400, 0x0)
	case 63:
		return tm.jjMoveStringLiteralDfa1_2(0x0, 0x0, 0x0, 0x400000, 0x0)
	case 64:
		return tm.jjMoveStringLiteralDfa1_2(0x0, 0x0, 0x0, 0x2000000, 0x0)
	case 65:
		return tm.jjMoveStringLiteralDfa1_2(0x8000000000, 0x0, 0x0, 0x0, 0x0)
	case 66:
		return tm.jjMoveStringLiteralDfa1_2(0x8000000000000000, 0x0, 0x0, 0x0, 0x0)
	case 67:
		return tm.jjMoveStringLiteralDfa1_2(0xc0000000000, 0x0, 0x0, 0x0, 0x0)
	case 68:
		return tm.jjMoveStringLiteralDfa1_2(0x0, 0x8000000000000002, 0x0, 0x0, 0x0)
	case 69:
		return tm.jjMoveStringLiteralDfa1_2(0x1600000000000, 0x800000000000000, 0x0, 0x0, 0x0)
	case 72:
		return tm.jjMoveStringLiteralDfa1_2(0x0, 0x60, 0x0, 0x0, 0x0)
	case 73:
		return tm.jjMoveStringLiteralDfa1_2(0x2c000000000000, 0x0, 0x0, 0x0, 0x0)
	case 76:
		return tm.jjMoveStringLiteralDfa1_2(0x50000000000000, 0x200, 0x0, 0x0, 0x0)
	case 77:
		return tm.jjMoveStringLiteralDfa1_2(0x80000000000000, 0x0, 0x0, 0x0, 0x0)
	case 78:
		return tm.jjMoveStringLiteralDfa1_2(0x100000000000000, 0x0, 0x0, 0x0, 0x0)
	case 79:
		return tm.jjMoveStringLiteralDfa1_2(0x200000000000000, 0x181, 0x0, 0x0, 0x0)
	case 80:
		return tm.jjMoveStringLiteralDfa1_2(0x0, 0x41800, 0x0, 0x0, 0x0)
	case 81:
		return tm.jjMoveStringLiteralDfa1_2(0x0, 0x4000, 0x0, 0x0, 0x0)
	case 82:
		return tm.jjMoveStringLiteralDfa1_2(0x0, 0x8000, 0x0, 0x0, 0x0)
	case 83:
		return tm.jjMoveStringLiteralDfa1_2(0x800000000000000, 0x2000000000110000, 0x0, 0x0, 0x0)
	case 84:
		return tm.jjMoveStringLiteralDfa1_2(0x4000000000000000, 0x408, 0x0, 0x0, 0x0)
	case 85:
		return tm.jjMoveStringLiteralDfa1_2(0x0, 0x5000000000000010, 0x0, 0x0, 0x0)
	case 87:
		return tm.jjMoveStringLiteralDfa1_2(0x0, 0xc80000, 0x0, 0x0, 0x0)
	case 91:
		tm.jjmatchedKind = 97
		return tm.jjMoveStringLiteralDfa1_2(0x10000000000, 0x2000, 0x0, 0x0, 0x0)
	case 92:
		tm.jjmatchedKind = 180
		return tm.jjMoveStringLiteralDfa1_2(0x3000000010000000, 0x80000000000000, 0xffffffff7ffe0, 0x800000000, 0x180000000)
	case 93:
		tm.jjmatchedKind = 99
		return tm.jjMoveStringLiteralDfa1_2(0x0, 0x400000000, 0x0, 0x0, 0x0)
	case 94:
		tm.jjmatchedKind = 189
		return tm.jjMoveStringLiteralDfa1_2(0x0, 0x18000000000000, 0x1000000000000000, 0x0, 0x0)
	case 95:
		return tm.jjStartNfaWithStates_2(0, 92, 270)
	case 123:
		return tm.jjStopAtPos(0, 100)
	case 124:
		tm.jjmatchedKind = 207
		return tm.jjMoveStringLiteralDfa1_2(0x0, 0x0, 0x0, 0x34000, 0x0)
	case 125:
		return tm.jjStopAtPos(0, 101)
	case 126:
		tm.jjmatchedKind = 120
		return tm.jjMoveStringLiteralDfa1_2(0x0, 0x0, 0x20000000000000, 0x0, 0x0)
	case 215:
		return tm.jjStopAtPos(0, 283)
	case 247:
		return tm.jjStopAtPos(0, 239)
	case 8214:
		return tm.jjStopAtPos(0, 285)
	case 8229:
		return tm.jjStopAtPos(0, 241)
	case 8230:
		return tm.jjStopAtPos(0, 242)
	case 8252:
		return tm.jjStopAtPos(0, 244)
	case 8263:
		return tm.jjStopAtPos(0, 265)
	case 8605:
		return tm.jjStopAtPos(0, 250)
	case 8658:
		return tm.jjStopAtPos(0, 248)
	case 8660:
		return tm.jjStopAtPos(0, 247)
	case 8696:
		return tm.jjStopAtPos(0, 261)
	case 8713:
		return tm.jjStopAtPos(0, 255)
	case 8728:
		return tm.jjStopAtPos(0, 236)
	case 8733:
		return tm.jjStopAtPos(0, 264)
	case 8745:
		return tm.jjStopAtPos(0, 234)
	case 8746:
		return tm.jjStopAtPos(0, 238)
	case 8764:
		return tm.jjStopAtPos(0, 268)
	case 8768:
		return tm.jjStopAtPos(0, 286)
	case 8771:
		return tm.jjStopAtPos(0, 269)
	case 8773:
		return tm.jjStopAtPos(0, 237)
	case 8776:
		return tm.jjStopAtPos(0, 228)
	case 8781:
		return tm.jjStopAtPos(0, 230)
	case 8784:
		return tm.jjStopAtPos(0, 240)
	case 8788:
		return tm.jjStopAtPos(0, 229)
	case 8800:
		return tm.jjStopAtPos(0, 254)
	case 8801:
		return tm.jjStopAtPos(0, 243)
	case 8804:
		return tm.jjStopAtPos(0, 251)
	case 8805:
		return tm.jjStopAtPos(0, 245)
	case 8810:
		return tm.jjStopAtPos(0, 252)
	case 8811:
		return tm.jjStopAtPos(0, 246)
	case 8826:
		return tm.jjStopAtPos(0, 262)
	case 8827:
		return tm.jjStopAtPos(0, 279)
	case 8834:
		return tm.jjStopAtPos(0, 277)
	case 8835:
		return tm.jjStopAtPos(0, 281)
	case 8838:
		return tm.jjStopAtPos(0, 278)
	case 8839:
		return tm.jjStopAtPos(0, 282)
	case 8846:
		return tm.jjStopAtPos(0, 284)
	case 8847:
		return tm.jjStopAtPos(0, 272)
	case 8848:
		return tm.jjStopAtPos(0, 274)
	case 8849:
		return tm.jjStopAtPos(0, 273)
	case 8850:
		return tm.jjStopAtPos(0, 275)
	case 8851:
		return tm.jjStopAtPos(0, 270)
	case 8852:
		return tm.jjStopAtPos(0, 271)
	case 8853:
		return tm.jjStopAtPos(0, 258)
	case 8854:
		return tm.jjStopAtPos(0, 257)
	case 8855:
		return tm.jjStopAtPos(0, 260)
	case 8856:
		return tm.jjStopAtPos(0, 259)
	case 8857:
		return tm.jjStopAtPos(0, 256)
	case 8866:
		return tm.jjStopAtPos(0, 267)
	case 8867:
		return tm.jjStopAtPos(0, 253)
	case 8872:
		return tm.jjStopAtPos(0, 266)
	case 8901:
		return tm.jjStopAtPos(0, 235)
	case 8902:
		return tm.jjStopAtPos(0, 276)
	case 9679:
		return tm.jjStopAtPos(0, 233)
	case 9711:
		return tm.jjStopAtPos(0, 231)
	case 10868:
		return tm.jjStopAtPos(0, 232)
	case 10927:
		return tm.jjStopAtPos(0, 263)
	case 10928:
		return tm.jjStopAtPos(0, 280)
	case 10980:
		return tm.jjStopAtPos(0, 249)
	default:
		return tm.jjMoveNfa_2(0, 0)
	}
}

func (tm *SanyTokenManager) jjMoveStringLiteralDfa1_2(active0 uint64, active1 uint64, active2 uint64, active3 uint64, active4 uint64) int32 {
	if !tm.readScanChar() {
		tm.jjStopStringLiteralDfa_2(0, active0, active1, active2, active3, active4)
		return 1
	}
	switch tm.curChar {
	case 33:
		if (active3 & 0x4000000) != 0 {
			return tm.jjStopAtPos(1, 218)
		}
		break
	case 35:
		if (active1 & 0x10000000000000) != 0 {
			return tm.jjStopAtPos(1, 116)
		} else {
			if (active2 & 0x400000000000000) != 0 {
				return tm.jjStopAtPos(1, 186)
			}
		}
		break
	case 36:
		if (active3 & 0x100000) != 0 {
			return tm.jjStopAtPos(1, 212)
		}
		break
	case 37:
		if (active3 & 0x800000) != 0 {
			return tm.jjStopAtPos(1, 215)
		}
		break
	case 38:
		if (active3 & 0x40000) != 0 {
			return tm.jjStopAtPos(1, 210)
		}
		break
	case 42:
		if (active0 & 0x10000000) != 0 {
			return tm.jjStopAtPos(1, 28)
		} else {
			if (active1 & 0x8000000000000) != 0 {
				return tm.jjStopAtPos(1, 115)
			} else {
				if (active3 & 0x4) != 0 {
					return tm.jjStopAtPos(1, 194)
				}
			}
		}
		break
	case 43:
		if (active3 & 0x10) != 0 {
			return tm.jjStopAtPos(1, 196)
		}
		return tm.jjMoveStringLiteralDfa2_2(active0, 0, active1, 0, active2, 0, active3, 0x40000001, active4, 0)
	case 45:
		if (active2 & 0x4000000000000000) != 0 {
			return tm.jjStartNfaWithStates_2(1, 190, 120)
		} else {
			if (active3 & 0x10000) != 0 {
				return tm.jjStartNfaWithStates_2(1, 208, 28)
			}
		}
		return tm.jjMoveStringLiteralDfa2_2(active0, 0, active1, 0, active2, 0, active3, 0x80000000, active4, 0)
	case 46:
		if (active1 & 0x20000000000000) != 0 {
			return tm.jjStopAtPos(1, 117)
		} else {
			if (active3 & 0x2000) != 0 {
				{
					tm.jjmatchedKind = 205
					tm.jjmatchedPos = 1
				}
			}
		}
		return tm.jjMoveStringLiteralDfa2_2(active0, 0, active1, 0, active2, 0, active3, 0x100001000, active4, 0)
	case 47:
		if (active2 & 0x1) != 0 {
			return tm.jjStopAtPos(1, 128)
		}
		return tm.jjMoveStringLiteralDfa2_2(active0, 0, active1, 0, active2, 0, active3, 0x200000000, active4, 0)
	case 58:
		if (active3 & 0x80) != 0 {
			return tm.jjStopAtPos(1, 199)
		}
		return tm.jjMoveStringLiteralDfa2_2(active0, 0, active1, 0, active2, 0, active3, 0x20000000, active4, 0)
	case 60:
		if (active2 & 0x80000000000000) != 0 {
			return tm.jjStopAtPos(1, 183)
		}
		break
	case 61:
		if (active2 & 0x4) != 0 {
			return tm.jjStopAtPos(1, 130)
		} else {
			if (active3 & 0x100) != 0 {
				{
					tm.jjmatchedKind = 200
					tm.jjmatchedPos = 1
				}
			} else {
				if (active3 & 0x400) != 0 {
					return tm.jjStopAtPos(1, 202)
				} else {
					if (active3 & 0x20000) != 0 {
						return tm.jjStopAtPos(1, 209)
					} else {
						if (active3 & 0x10000000) != 0 {
							return tm.jjStopAtPos(1, 220)
						}
					}
				}
			}
		}
		return tm.jjMoveStringLiteralDfa2_2(active0, 0, active1, 0, active2, 0, active3, 0x40, active4, 0)
	case 62:
		if (active2 & 0x20000000000000) != 0 {
			return tm.jjStopAtPos(1, 181)
		} else {
			if (active2 & 0x40000000000000) != 0 {
				return tm.jjStopAtPos(1, 182)
			} else {
				if (active3 & 0x8000000) != 0 {
					return tm.jjStopAtPos(1, 219)
				}
			}
		}
		break
	case 63:
		if (active3 & 0x400000) != 0 {
			return tm.jjStopAtPos(1, 214)
		}
		break
	case 64:
		if (active3 & 0x2000000) != 0 {
			return tm.jjStopAtPos(1, 217)
		}
		break
	case 65:
		return tm.jjMoveStringLiteralDfa2_2(active0, 0x2000040000000000, active1, 0x640, active2, 0, active3, 0, active4, 0)
	case 66:
		return tm.jjMoveStringLiteralDfa2_2(active0, 0, active1, 0x80, active2, 0, active3, 0, active4, 0)
	case 69:
		return tm.jjMoveStringLiteralDfa2_2(active0, 0x1110000000000000, active1, 0xc002, active2, 0, active3, 0, active4, 0)
	case 70:
		if (active0 & 0x4000000000000) != 0 {
			return tm.jjStartNfaWithStates_2(1, 50, 269)
		}
		return tm.jjMoveStringLiteralDfa2_2(active0, 0x800000000000000, active1, 0x400000, active2, 0, active3, 0, active4, 0)
	case 72:
		return tm.jjMoveStringLiteralDfa2_2(active0, 0x4000080000000000, active1, 0x8, active2, 0, active3, 0, active4, 0)
	case 73:
		return tm.jjMoveStringLiteralDfa2_2(active0, 0, active1, 0x8c0020, active2, 0, active3, 0, active4, 0)
	case 76:
		return tm.jjMoveStringLiteralDfa2_2(active0, 0x200000000000, active1, 0, active2, 0, active3, 0, active4, 0)
	case 77:
		return tm.jjMoveStringLiteralDfa2_2(active0, 0, active1, 0x100, active2, 0, active3, 0, active4, 0)
	case 78:
		if (active0 & 0x20000000000000) != 0 {
			{
				tm.jjmatchedKind = 53
				tm.jjmatchedPos = 1
			}
		}
		return tm.jjMoveStringLiteralDfa2_2(active0, 0x8000000000000, active1, 0x5800000000000001, active2, 0, active3, 0, active4, 0)
	case 79:
		return tm.jjMoveStringLiteralDfa2_2(active0, 0xc0000000000000, active1, 0x8000000000000000, active2, 0, active3, 0, active4, 0)
	case 82:
		return tm.jjMoveStringLiteralDfa2_2(active0, 0, active1, 0x1800, active2, 0, active3, 0, active4, 0)
	case 83:
		return tm.jjMoveStringLiteralDfa2_2(active0, 0x8000000000, active1, 0x10, active2, 0, active3, 0, active4, 0)
	case 84:
		return tm.jjMoveStringLiteralDfa2_2(active0, 0x200000000000000, active1, 0x10000, active2, 0, active3, 0, active4, 0)
	case 85:
		return tm.jjMoveStringLiteralDfa2_2(active0, 0, active1, 0x2000000000100000, active2, 0, active3, 0, active4, 0)
	case 88:
		if (active4 & 0x100000000) != 0 {
			return tm.jjStopAtPos(1, 288)
		}
		return tm.jjMoveStringLiteralDfa2_2(active0, 0x1400000000000, active1, 0, active2, 0, active3, 0, active4, 0)
	case 89:
		if (active0 & 0x8000000000000000) != 0 {
			return tm.jjStartNfaWithStates_2(1, 63, 269)
		}
		break
	case 92:
		return tm.jjMoveStringLiteralDfa2_2(active0, 0, active1, 0, active2, 0, active3, 0x400000000, active4, 0)
	case 93:
		return tm.jjMoveStringLiteralDfa2_2(active0, 0x10000000000, active1, 0x2000, active2, 0, active3, 0, active4, 0)
	case 94:
		if (active2 & 0x1000000000000000) != 0 {
			return tm.jjStopAtPos(1, 188)
		}
		break
	case 95:
		if (active1 & 0x400000000) != 0 {
			return tm.jjStopAtPos(1, 98)
		}
		break
	case 97:
		return tm.jjMoveStringLiteralDfa2_2(active0, 0, active1, 0, active2, 0x60, active3, 0, active4, 0)
	case 98:
		return tm.jjMoveStringLiteralDfa2_2(active0, 0, active1, 0, active2, 0x180, active3, 0, active4, 0)
	case 99:
		return tm.jjMoveStringLiteralDfa2_2(active0, 0, active1, 0, active2, 0x3e00, active3, 0, active4, 0)
	case 100:
		return tm.jjMoveStringLiteralDfa2_2(active0, 0, active1, 0, active2, 0xc000, active3, 0, active4, 0)
	case 101:
		return tm.jjMoveStringLiteralDfa2_2(active0, 0, active1, 0, active2, 0x10000, active3, 0, active4, 0)
	case 103:
		return tm.jjMoveStringLiteralDfa2_2(active0, 0, active1, 0, active2, 0x60000, active3, 0, active4, 0)
	case 105:
		return tm.jjMoveStringLiteralDfa2_2(active0, 0, active1, 0, active2, 0x100000, active3, 0, active4, 0)
	case 108:
		return tm.jjMoveStringLiteralDfa2_2(active0, 0, active1, 0, active2, 0x3c00000, active3, 0, active4, 0)
	case 110:
		return tm.jjMoveStringLiteralDfa2_2(active0, 0, active1, 0x80000000000000, active2, 0, active3, 0x800000000, active4, 0)
	case 111:
		if (active2 & 0x4000000) != 0 {
			{
				tm.jjmatchedKind = 154
				tm.jjmatchedPos = 1
			}
		}
		return tm.jjMoveStringLiteralDfa2_2(active0, 0, active1, 0, active2, 0xf8000000, active3, 0, active4, 0)
	case 112:
		return tm.jjMoveStringLiteralDfa2_2(active0, 0, active1, 0, active2, 0x700000000, active3, 0, active4, 0)
	case 115:
		return tm.jjMoveStringLiteralDfa2_2(active0, 0, active1, 0, active2, 0x3fff800000000, active3, 0, active4, 0)
	case 116:
		return tm.jjMoveStringLiteralDfa2_2(active0, 0, active1, 0, active2, 0, active3, 0, active4, 0x80000000)
	case 117:
		return tm.jjMoveStringLiteralDfa2_2(active0, 0, active1, 0, active2, 0x4000000200000, active3, 0, active4, 0)
	case 119:
		return tm.jjMoveStringLiteralDfa2_2(active0, 0, active1, 0, active2, 0x8000000000000, active3, 0, active4, 0)
	case 124:
		if (active2 & 0x100000000000000) != 0 {
			return tm.jjStopAtPos(1, 184)
		} else {
			if (active2 & 0x8000000000000000) != 0 {
				{
					tm.jjmatchedKind = 191
					tm.jjmatchedPos = 1
				}
			} else {
				if (active3 & 0x4000) != 0 {
					return tm.jjStopAtPos(1, 206)
				}
			}
		}
		return tm.jjMoveStringLiteralDfa2_2(active0, 0, active1, 0x100000000, active2, 0, active3, 0, active4, 0)
	default:
		break
	}
	return tm.jjStartNfa_2(0, active0, active1, active2, active3, active4)
}

func (tm *SanyTokenManager) jjMoveStringLiteralDfa2_2(old0 uint64, active0 uint64, old1 uint64, active1 uint64, old2 uint64, active2 uint64, old3 uint64, active3 uint64, old4 uint64, active4 uint64) int32 {
	if ((sanyScannerAndUint64(&active0, old0)) | (sanyScannerAndUint64(&active1, old1)) | (sanyScannerAndUint64(&active2, old2)) | (sanyScannerAndUint64(&active3, old3)) | (sanyScannerAndUint64(&active4, old4))) == 0 {
		return tm.jjStartNfa_2(0, old0, old1, old2, old3, old4)
	}
	if !tm.readScanChar() {
		tm.jjStopStringLiteralDfa_2(1, active0, active1, active2, active3, active4)
		return 2
	}
	switch tm.curChar {
	case 41:
		if (active3 & 0x40000000) != 0 {
			return tm.jjStopAtPos(2, 222)
		} else {
			if (active3 & 0x80000000) != 0 {
				return tm.jjStopAtPos(2, 223)
			} else {
				if (active3 & 0x100000000) != 0 {
					return tm.jjStopAtPos(2, 224)
				} else {
					if (active3 & 0x200000000) != 0 {
						return tm.jjStopAtPos(2, 225)
					}
				}
			}
		}
		break
	case 45:
		if (active1 & 0x100000000) != 0 {
			return tm.jjStopAtPos(2, 96)
		}
		return tm.jjMoveStringLiteralDfa3_2(active0, 0, active1, 0, active2, 0, active3, 0x1, active4, 0)
	case 46:
		if (active3 & 0x1000) != 0 {
			return tm.jjStopAtPos(2, 204)
		}
		break
	case 61:
		if (active3 & 0x20000000) != 0 {
			return tm.jjStopAtPos(2, 221)
		}
		break
	case 62:
		if (active3 & 0x40) != 0 {
			return tm.jjStopAtPos(2, 198)
		}
		break
	case 65:
		if (active0 & 0x2000000000000000) != 0 {
			return tm.jjStopAtPos(2, 61)
		}
		return tm.jjMoveStringLiteralDfa3_2(active0, 0x10000000000, active1, 0x800000000010000, active2, 0, active3, 0, active4, 0)
	case 66:
		return tm.jjMoveStringLiteralDfa3_2(active0, 0, active1, 0x2000000000000000, active2, 0, active3, 0, active4, 0)
	case 67:
		return tm.jjMoveStringLiteralDfa3_2(active0, 0x40400000000000, active1, 0x1000000000048000, active2, 0, active3, 0, active4, 0)
	case 68:
		if (active1 & 0x4000) != 0 {
			return tm.jjStartNfaWithStates_2(2, 78, 269)
		}
		return tm.jjMoveStringLiteralDfa3_2(active0, 0x80000000000000, active1, 0x20, active2, 0, active3, 0, active4, 0)
	case 69:
		if (active0 & 0x1000000000000000) != 0 {
			return tm.jjStopAtPos(2, 60)
		} else {
			if (active1 & 0x10) != 0 {
				return tm.jjStartNfaWithStates_2(2, 68, 269)
			}
		}
		return tm.jjMoveStringLiteralDfa3_2(active0, 0x4000000000000000, active1, 0x8, active2, 0, active3, 0, active4, 0)
	case 70:
		return tm.jjMoveStringLiteralDfa3_2(active0, 0, active1, 0x100002, active2, 0, active3, 0, active4, 0)
	case 72:
		return tm.jjMoveStringLiteralDfa3_2(active0, 0x200000000000000, active1, 0, active2, 0, active3, 0, active4, 0)
	case 73:
		return tm.jjMoveStringLiteralDfa3_2(active0, 0, active1, 0x4000000000000100, active2, 0, active3, 0, active4, 0)
	case 75:
		return tm.jjMoveStringLiteralDfa3_2(active0, 0, active1, 0x400, active2, 0, active3, 0, active4, 0)
	case 76:
		return tm.jjMoveStringLiteralDfa3_2(active0, 0, active1, 0x1, active2, 0, active3, 0, active4, 0)
	case 77:
		return tm.jjMoveStringLiteralDfa3_2(active0, 0, active1, 0x8000000000000200, active2, 0, active3, 0, active4, 0)
	case 79:
		return tm.jjMoveStringLiteralDfa3_2(active0, 0x80000000000, active1, 0x1800, active2, 0, active3, 0, active4, 0)
	case 80:
		return tm.jjMoveStringLiteralDfa3_2(active0, 0, active1, 0x2000, active2, 0, active3, 0, active4, 0)
	case 83:
		return tm.jjMoveStringLiteralDfa3_2(active0, 0x8248000000000, active1, 0, active2, 0, active3, 0, active4, 0)
	case 84:
		if (active0 & 0x10000000000000) != 0 {
			return tm.jjStartNfaWithStates_2(2, 52, 269)
		}
		return tm.jjMoveStringLiteralDfa3_2(active0, 0x1000000000000, active1, 0x880000, active2, 0, active3, 0, active4, 0)
	case 86:
		return tm.jjMoveStringLiteralDfa3_2(active0, 0, active1, 0xc0, active2, 0, active3, 0, active4, 0)
	case 87:
		if (active0 & 0x100000000000000) != 0 {
			return tm.jjStartNfaWithStates_2(2, 56, 269)
		}
		break
	case 88:
		return tm.jjMoveStringLiteralDfa3_2(active0, 0, active1, 0, active2, 0, active3, 0x400000000, active4, 0)
	case 95:
		if (active0 & 0x800000000000000) != 0 {
			return tm.jjStopAtPos(2, 59)
		} else {
			if (active1 & 0x400000) != 0 {
				return tm.jjStopAtPos(2, 86)
			}
		}
		break
	case 97:
		return tm.jjMoveStringLiteralDfa3_2(active0, 0, active1, 0, active2, 0x400200, active3, 0, active4, 0)
	case 100:
		return tm.jjMoveStringLiteralDfa3_2(active0, 0, active1, 0, active2, 0x8000400, active3, 0, active4, 0)
	case 101:
		return tm.jjMoveStringLiteralDfa3_2(active0, 0, active1, 0x80000000000000, active2, 0x820000, active3, 0, active4, 0)
	case 103:
		if (active2 & 0x40000) != 0 {
			return tm.jjStopAtPos(2, 146)
		}
		break
	case 105:
		return tm.jjMoveStringLiteralDfa3_2(active0, 0, active1, 0, active2, 0x1800004880, active3, 0, active4, 0x80000000)
	case 108:
		if (active2 & 0x1000000) != 0 {
			return tm.jjStopAtPos(2, 152)
		}
		break
	case 109:
		return tm.jjMoveStringLiteralDfa3_2(active0, 0, active1, 0, active2, 0x10000000, active3, 0, active4, 0)
	case 110:
		return tm.jjMoveStringLiteralDfa3_2(active0, 0, active1, 0, active2, 0x300000, active3, 0, active4, 0)
	case 111:
		return tm.jjMoveStringLiteralDfa3_2(active0, 0, active1, 0, active2, 0x2009000, active3, 0x800000000, active4, 0)
	case 112:
		return tm.jjMoveStringLiteralDfa3_2(active0, 0, active1, 0, active2, 0x4000020000020, active3, 0, active4, 0)
	case 113:
		return tm.jjMoveStringLiteralDfa3_2(active0, 0, active1, 0, active2, 0x7e000010000, active3, 0, active4, 0)
	case 114:
		if (active2 & 0x8000000000000) != 0 {
			return tm.jjStopAtPos(2, 179)
		}
		return tm.jjMoveStringLiteralDfa3_2(active0, 0, active1, 0, active2, 0x700000000, active3, 0, active4, 0)
	case 115:
		return tm.jjMoveStringLiteralDfa3_2(active0, 0, active1, 0, active2, 0x40000040, active3, 0, active4, 0)
	case 116:
		return tm.jjMoveStringLiteralDfa3_2(active0, 0, active1, 0, active2, 0x80080000000, active3, 0, active4, 0)
	case 117:
		return tm.jjMoveStringLiteralDfa3_2(active0, 0, active1, 0, active2, 0x3f00000002100, active3, 0, active4, 0)
	default:
		break
	}
	return tm.jjStartNfa_2(1, active0, active1, active2, active3, active4)
}

func (tm *SanyTokenManager) jjMoveStringLiteralDfa3_2(old0 uint64, active0 uint64, old1 uint64, active1 uint64, old2 uint64, active2 uint64, old3 uint64, active3 uint64, old4 uint64, active4 uint64) int32 {
	if ((sanyScannerAndUint64(&active0, old0)) | (sanyScannerAndUint64(&active1, old1)) | (sanyScannerAndUint64(&active2, old2)) | (sanyScannerAndUint64(&active3, old3)) | (sanyScannerAndUint64(&active4, old4))) == 0 {
		return tm.jjStartNfa_2(1, old0, old1, old2, old3, old4)
	}
	if !tm.readScanChar() {
		tm.jjStopStringLiteralDfa_2(2, active0, active1, active2, active3, active4)
		return 3
	}
	switch tm.curChar {
	case 41:
		if (active3 & 0x400000000) != 0 {
			return tm.jjStopAtPos(3, 226)
		}
		break
	case 62:
		if (active3 & 0x1) != 0 {
			return tm.jjStopAtPos(3, 192)
		}
		break
	case 65:
		return tm.jjMoveStringLiteralDfa4_2(active0, 0x40000000000000, active1, 0x8000000000000000, active2, 0, active3, 0, active4, 0)
	case 66:
		return tm.jjMoveStringLiteralDfa4_2(active0, 0, active1, 0x800000000000200, active2, 0, active3, 0, active4, 0)
	case 69:
		if (active0 & 0x40000000000) != 0 {
			return tm.jjStartNfaWithStates_2(3, 42, 269)
		} else {
			if (active0 & 0x200000000000) != 0 {
				return tm.jjStartNfaWithStates_2(3, 45, 269)
			} else {
				if (active1 & 0x20) != 0 {
					return tm.jjStartNfaWithStates_2(3, 69, 269)
				} else {
					if (active1 & 0x40) != 0 {
						return tm.jjStartNfaWithStates_2(3, 70, 269)
					} else {
						if (active1 & 0x400) != 0 {
							return tm.jjStartNfaWithStates_2(3, 74, 269)
						}
					}
				}
			}
		}
		return tm.jjMoveStringLiteralDfa4_2(active0, 0x201400000000000, active1, 0, active2, 0, active3, 0, active4, 0)
	case 70:
		return tm.jjMoveStringLiteralDfa4_2(active0, 0, active1, 0x100000, active2, 0, active3, 0, active4, 0)
	case 72:
		if (active1 & 0x800000) != 0 {
			return tm.jjStartNfaWithStates_2(3, 87, 271)
		}
		return tm.jjMoveStringLiteralDfa4_2(active0, 0, active1, 0x1000000000000000, active2, 0, active3, 0, active4, 0)
	case 73:
		return tm.jjMoveStringLiteralDfa4_2(active0, 0, active1, 0x82, active2, 0, active3, 0, active4, 0)
	case 75:
		if (active1 & 0x40000) != 0 {
			return tm.jjStartNfaWithStates_2(3, 82, 269)
		}
		break
	case 78:
		if (active0 & 0x4000000000000000) != 0 {
			return tm.jjStartNfaWithStates_2(3, 62, 269)
		}
		return tm.jjMoveStringLiteralDfa4_2(active0, 0, active1, 0x80000, active2, 0, active3, 0, active4, 0)
	case 79:
		return tm.jjMoveStringLiteralDfa4_2(active0, 0x80000000000, active1, 0x4000000000000808, active2, 0, active3, 0, active4, 0)
	case 82:
		return tm.jjMoveStringLiteralDfa4_2(active0, 0, active1, 0x2000, active2, 0, active3, 0, active4, 0)
	case 83:
		return tm.jjMoveStringLiteralDfa4_2(active0, 0x10000000000, active1, 0x2000000000000000, active2, 0, active3, 0, active4, 0)
	case 84:
		return tm.jjMoveStringLiteralDfa4_2(active0, 0x8000000000000, active1, 0x10100, active2, 0, active3, 0, active4, 0)
	case 85:
		return tm.jjMoveStringLiteralDfa4_2(active0, 0x80008000000000, active1, 0x8000, active2, 0, active3, 0, active4, 0)
	case 86:
		return tm.jjMoveStringLiteralDfa4_2(active0, 0, active1, 0x1000, active2, 0, active3, 0, active4, 0)
	case 89:
		if (active1 & 0x1) != 0 {
			return tm.jjStartNfaWithStates_2(3, 64, 269)
		}
		break
	case 97:
		return tm.jjMoveStringLiteralDfa4_2(active0, 0, active1, 0, active2, 0x80000000000, active3, 0, active4, 0)
	case 98:
		return tm.jjMoveStringLiteralDfa4_2(active0, 0, active1, 0, active2, 0x300000000000, active3, 0, active4, 0)
	case 99:
		return tm.jjMoveStringLiteralDfa4_2(active0, 0, active1, 0, active2, 0xc06000000000, active3, 0, active4, 0)
	case 101:
		return tm.jjMoveStringLiteralDfa4_2(active0, 0, active1, 0, active2, 0x300000000, active3, 0, active4, 0)
	case 103:
		if (active1 & 0x80000000000000) != 0 {
			return tm.jjStopAtPos(3, 119)
		}
		return tm.jjMoveStringLiteralDfa4_2(active0, 0, active1, 0, active2, 0x80, active3, 0, active4, 0)
	case 105:
		return tm.jjMoveStringLiteralDfa4_2(active0, 0, active1, 0, active2, 0x90200000, active3, 0, active4, 0)
	case 108:
		return tm.jjMoveStringLiteralDfa4_2(active0, 0, active1, 0, active2, 0x4000060000100, active3, 0, active4, 0)
	case 109:
		if (active2 & 0x800000000) != 0 {
			{
				tm.jjmatchedKind = 163
				tm.jjmatchedPos = 3
			}
		}
		return tm.jjMoveStringLiteralDfa4_2(active0, 0, active1, 0, active2, 0x1000000000, active3, 0, active4, 0x80000000)
	case 110:
		return tm.jjMoveStringLiteralDfa4_2(active0, 0, active1, 0, active2, 0x401000, active3, 0, active4, 0)
	case 111:
		return tm.jjMoveStringLiteralDfa4_2(active0, 0, active1, 0, active2, 0x408000400, active3, 0, active4, 0)
	case 112:
		if (active2 & 0x200) != 0 {
			return tm.jjStopAtPos(3, 137)
		} else {
			if (active2 & 0x2000) != 0 {
				return tm.jjStopAtPos(3, 141)
			}
		}
		return tm.jjMoveStringLiteralDfa4_2(active0, 0, active1, 0, active2, 0x3000000000020, active3, 0, active4, 0)
	case 113:
		if (active2 & 0x20000) != 0 {
			return tm.jjStopAtPos(3, 145)
		} else {
			if (active2 & 0x800000) != 0 {
				return tm.jjStopAtPos(3, 151)
			}
		}
		break
	case 114:
		if (active2 & 0x2000000) != 0 {
			return tm.jjStopAtPos(3, 153)
		}
		return tm.jjMoveStringLiteralDfa4_2(active0, 0, active1, 0, active2, 0x800, active3, 0, active4, 0)
	case 115:
		return tm.jjMoveStringLiteralDfa4_2(active0, 0, active1, 0, active2, 0x78000000000, active3, 0, active4, 0)
	case 116:
		return tm.jjMoveStringLiteralDfa4_2(active0, 0, active1, 0, active2, 0x108000, active3, 0x800000000, active4, 0)
	case 117:
		return tm.jjMoveStringLiteralDfa4_2(active0, 0, active1, 0, active2, 0x10000, active3, 0, active4, 0)
	case 118:
		if (active2 & 0x4000) != 0 {
			return tm.jjStopAtPos(3, 142)
		}
		break
	case 121:
		return tm.jjMoveStringLiteralDfa4_2(active0, 0, active1, 0, active2, 0x40, active3, 0, active4, 0)
	default:
		break
	}
	return tm.jjStartNfa_2(2, active0, active1, active2, active3, active4)
}

func (tm *SanyTokenManager) jjMoveStringLiteralDfa4_2(old0 uint64, active0 uint64, old1 uint64, active1 uint64, old2 uint64, active2 uint64, old3 uint64, active3 uint64, old4 uint64, active4 uint64) int32 {
	if ((sanyScannerAndUint64(&active0, old0)) | (sanyScannerAndUint64(&active1, old1)) | (sanyScannerAndUint64(&active2, old2)) | (sanyScannerAndUint64(&active3, old3)) | (sanyScannerAndUint64(&active4, old4))) == 0 {
		return tm.jjStartNfa_2(2, old0, old1, old2, old3, old4)
	}
	if !tm.readScanChar() {
		tm.jjStopStringLiteralDfa_2(3, active0, active1, active2, active3, active4)
		return 4
	}
	switch tm.curChar {
	case 65:
		return tm.jjMoveStringLiteralDfa5_2(active0, 0x8000000000000, active1, 0x1000000000000000, active2, 0, active3, 0, active4, 0)
	case 68:
		return tm.jjMoveStringLiteralDfa5_2(active0, 0, active1, 0x200, active2, 0, active3, 0, active4, 0)
	case 69:
		if (active1 & 0x1000) != 0 {
			return tm.jjStartNfaWithStates_2(4, 76, 269)
		} else {
			if (active1 & 0x10000) != 0 {
				return tm.jjStartNfaWithStates_2(4, 80, 271)
			}
		}
		return tm.jjMoveStringLiteralDfa5_2(active0, 0, active1, 0x2000000000080000, active2, 0, active3, 0, active4, 0)
	case 70:
		if (active1 & 0x800) != 0 {
			return tm.jjStartNfaWithStates_2(4, 75, 269)
		}
		break
	case 73:
		return tm.jjMoveStringLiteralDfa5_2(active0, 0, active1, 0x8000000000100000, active2, 0, active3, 0, active4, 0)
	case 76:
		if (active0 & 0x40000000000000) != 0 {
			return tm.jjStartNfaWithStates_2(4, 54, 269)
		}
		return tm.jjMoveStringLiteralDfa5_2(active0, 0x80000000000000, active1, 0x800000000000000, active2, 0, active3, 0, active4, 0)
	case 77:
		return tm.jjMoveStringLiteralDfa5_2(active0, 0x8000000000, active1, 0, active2, 0, active3, 0, active4, 0)
	case 78:
		if (active1 & 0x4000000000000000) != 0 {
			return tm.jjStartNfaWithStates_2(4, 126, 269)
		}
		return tm.jjMoveStringLiteralDfa5_2(active0, 0x1000000000000, active1, 0x2, active2, 0, active3, 0, active4, 0)
	case 79:
		return tm.jjMoveStringLiteralDfa5_2(active0, 0, active1, 0x2080, active2, 0, active3, 0, active4, 0)
	case 80:
		return tm.jjMoveStringLiteralDfa5_2(active0, 0x400000000000, active1, 0, active2, 0, active3, 0, active4, 0)
	case 82:
		if (active0 & 0x200000000000000) != 0 {
			return tm.jjStartNfaWithStates_2(4, 57, 269)
		}
		return tm.jjMoveStringLiteralDfa5_2(active0, 0, active1, 0x8008, active2, 0, active3, 0, active4, 0)
	case 83:
		return tm.jjMoveStringLiteralDfa5_2(active0, 0x90000000000, active1, 0, active2, 0, active3, 0, active4, 0)
	case 84:
		return tm.jjMoveStringLiteralDfa5_2(active0, 0, active1, 0x100, active2, 0, active3, 0, active4, 0)
	case 97:
		return tm.jjMoveStringLiteralDfa5_2(active0, 0, active1, 0, active2, 0x2040000000, active3, 0, active4, 0)
	case 99:
		if (active2 & 0x800) != 0 {
			return tm.jjStopAtPos(4, 139)
		} else {
			if (active2 & 0x100000000) != 0 {
				{
					tm.jjmatchedKind = 160
					tm.jjmatchedPos = 4
				}
			} else {
				if (active2 & 0x400000000000) != 0 {
					{
						tm.jjmatchedKind = 174
						tm.jjmatchedPos = 4
					}
				}
			}
		}
		return tm.jjMoveStringLiteralDfa5_2(active0, 0, active1, 0, active2, 0x800200000080, active3, 0, active4, 0)
	case 100:
		if (active2 & 0x400000) != 0 {
			return tm.jjStopAtPos(4, 150)
		}
		break
	case 101:
		return tm.jjMoveStringLiteralDfa5_2(active0, 0, active1, 0, active2, 0x1000108000, active3, 0, active4, 0x80000000)
	case 103:
		if (active2 & 0x1000) != 0 {
			return tm.jjStopAtPos(4, 140)
		}
		break
	case 105:
		return tm.jjMoveStringLiteralDfa5_2(active0, 0, active1, 0, active2, 0x10000, active3, 0x800000000, active4, 0)
	case 108:
		return tm.jjMoveStringLiteralDfa5_2(active0, 0, active1, 0, active2, 0x100, active3, 0, active4, 0)
	case 109:
		return tm.jjMoveStringLiteralDfa5_2(active0, 0, active1, 0, active2, 0x80000040, active3, 0, active4, 0)
	case 110:
		return tm.jjMoveStringLiteralDfa5_2(active0, 0, active1, 0, active2, 0x10000000, active3, 0, active4, 0)
	case 111:
		return tm.jjMoveStringLiteralDfa5_2(active0, 0, active1, 0, active2, 0x200000, active3, 0, active4, 0)
	case 112:
		return tm.jjMoveStringLiteralDfa5_2(active0, 0, active1, 0, active2, 0x400000000, active3, 0, active4, 0)
	case 114:
		if (active2 & 0x80000000000) != 0 {
			return tm.jjStopAtPos(4, 171)
		}
		return tm.jjMoveStringLiteralDfa5_2(active0, 0, active1, 0, active2, 0x20, active3, 0, active4, 0)
	case 115:
		return tm.jjMoveStringLiteralDfa5_2(active0, 0, active1, 0, active2, 0x3300000000000, active3, 0, active4, 0)
	case 116:
		if (active2 & 0x400) != 0 {
			return tm.jjStopAtPos(4, 138)
		} else {
			if (active2 & 0x8000000) != 0 {
				return tm.jjStopAtPos(4, 155)
			}
		}
		break
	case 117:
		return tm.jjMoveStringLiteralDfa5_2(active0, 0, active1, 0, active2, 0x407c020000000, active3, 0, active4, 0)
	default:
		break
	}
	return tm.jjStartNfa_2(3, active0, active1, active2, active3, active4)
}

func (tm *SanyTokenManager) jjMoveStringLiteralDfa5_2(old0 uint64, active0 uint64, old1 uint64, active1 uint64, old2 uint64, active2 uint64, old3 uint64, active3 uint64, old4 uint64, active4 uint64) int32 {
	if ((sanyScannerAndUint64(&active0, old0)) | (sanyScannerAndUint64(&active1, old1)) | (sanyScannerAndUint64(&active2, old2)) | (sanyScannerAndUint64(&active3, old3)) | (sanyScannerAndUint64(&active4, old4))) == 0 {
		return tm.jjStartNfa_2(3, old0, old1, old2, old3, old4)
	}
	if !tm.readScanChar() {
		tm.jjStopStringLiteralDfa_2(4, active0, active1, active2, active3, active4)
		return 5
	}
	switch tm.curChar {
	case 65:
		if (active1 & 0x200) != 0 {
			return tm.jjStartNfaWithStates_2(5, 73, 269)
		}
		break
	case 67:
		return tm.jjMoveStringLiteralDfa6_2(active0, 0, active1, 0x100000, active2, 0, active3, 0, active4, 0)
	case 68:
		return tm.jjMoveStringLiteralDfa6_2(active0, 0x1000000000000, active1, 0, active2, 0, active3, 0, active4, 0)
	case 69:
		if (active0 & 0x8000000000) != 0 {
			return tm.jjStartNfaWithStates_2(5, 39, 269)
		} else {
			if (active0 & 0x80000000000) != 0 {
				return tm.jjStartNfaWithStates_2(5, 43, 269)
			} else {
				if (active0 & 0x80000000000000) != 0 {
					return tm.jjStartNfaWithStates_2(5, 55, 269)
				} else {
					if (active1 & 0x2) != 0 {
						return tm.jjStartNfaWithStates_2(5, 65, 269)
					}
				}
			}
		}
		return tm.jjMoveStringLiteralDfa6_2(active0, 0, active1, 0x800000000000108, active2, 0, active3, 0, active4, 0)
	case 78:
		if (active1 & 0x8000000000000000) != 0 {
			return tm.jjStartNfaWithStates_2(5, 127, 269)
		}
		return tm.jjMoveStringLiteralDfa6_2(active0, 0x8000000000000, active1, 0x1000000000000000, active2, 0, active3, 0, active4, 0)
	case 83:
		return tm.jjMoveStringLiteralDfa6_2(active0, 0, active1, 0x88000, active2, 0, active3, 0, active4, 0)
	case 84:
		if (active0 & 0x400000000000) != 0 {
			return tm.jjStartNfaWithStates_2(5, 46, 269)
		} else {
			if (active1 & 0x2000000000000000) != 0 {
				return tm.jjStartNfaWithStates_2(5, 125, 271)
			}
		}
		break
	case 85:
		return tm.jjMoveStringLiteralDfa6_2(active0, 0x10000000000, active1, 0x80, active2, 0, active3, 0, active4, 0)
	case 86:
		return tm.jjMoveStringLiteralDfa6_2(active0, 0, active1, 0x2000, active2, 0, active3, 0, active4, 0)
	case 98:
		return tm.jjMoveStringLiteralDfa6_2(active0, 0, active1, 0, active2, 0x28000000000, active3, 0, active4, 0)
	case 101:
		return tm.jjMoveStringLiteralDfa6_2(active0, 0, active1, 0, active2, 0x3b00280000100, active3, 0, active4, 0)
	case 105:
		return tm.jjMoveStringLiteralDfa6_2(active0, 0, active1, 0, active2, 0x80, active3, 0, active4, 0)
	case 110:
		if (active2 & 0x200000) != 0 {
			return tm.jjStopAtPos(5, 149)
		} else {
			if (active3 & 0x800000000) != 0 {
				return tm.jjStopAtPos(5, 227)
			}
		}
		break
	case 111:
		return tm.jjMoveStringLiteralDfa6_2(active0, 0, active1, 0, active2, 0x20, active3, 0, active4, 0)
	case 112:
		if (active2 & 0x40) != 0 {
			return tm.jjStopAtPos(5, 134)
		} else {
			if (active2 & 0x2000000000) != 0 {
				return tm.jjStopAtPos(5, 165)
			} else {
				if (active2 & 0x4000000000) != 0 {
					return tm.jjStopAtPos(5, 166)
				}
			}
		}
		return tm.jjMoveStringLiteralDfa6_2(active0, 0, active1, 0, active2, 0x50000000000, active3, 0, active4, 0)
	case 113:
		if (active2 & 0x8000) != 0 {
			return tm.jjStopAtPos(5, 143)
		} else {
			if (active2 & 0x1000000000) != 0 {
				return tm.jjStopAtPos(5, 164)
			}
		}
		break
	case 114:
		return tm.jjMoveStringLiteralDfa6_2(active0, 0, active1, 0, active2, 0x100000, active3, 0, active4, 0)
	case 115:
		if (active2 & 0x20000000) != 0 {
			return tm.jjStopAtPos(5, 157)
		} else {
			if (active2 & 0x4000000000000) != 0 {
				return tm.jjStopAtPos(5, 178)
			} else {
				if (active4 & 0x80000000) != 0 {
					return tm.jjStopAtPos(5, 287)
				}
			}
		}
		return tm.jjMoveStringLiteralDfa6_2(active0, 0, active1, 0, active2, 0x40000000, active3, 0, active4, 0)
	case 116:
		return tm.jjMoveStringLiteralDfa6_2(active0, 0, active1, 0, active2, 0x400000000, active3, 0, active4, 0)
	case 117:
		return tm.jjMoveStringLiteralDfa6_2(active0, 0, active1, 0, active2, 0x10000000, active3, 0, active4, 0)
	case 118:
		if (active2 & 0x10000) != 0 {
			return tm.jjStopAtPos(5, 144)
		}
		break
	default:
		break
	}
	return tm.jjStartNfa_2(4, active0, active1, active2, active3, active4)
}

func (tm *SanyTokenManager) jjMoveStringLiteralDfa6_2(old0 uint64, active0 uint64, old1 uint64, active1 uint64, old2 uint64, active2 uint64, old3 uint64, active3 uint64, old4 uint64, active4 uint64) int32 {
	if ((sanyScannerAndUint64(&active0, old0)) | (sanyScannerAndUint64(&active1, old1)) | (sanyScannerAndUint64(&active2, old2)) | (sanyScannerAndUint64(&active3, old3)) | (sanyScannerAndUint64(&active4, old4))) == 0 {
		return tm.jjStartNfa_2(4, old0, old1, old2, old3, old4)
	}
	if !tm.readScanChar() {
		tm.jjStopStringLiteralDfa_2(5, active0, active1, active2, 0, 0)
		return 6
	}
	switch tm.curChar {
	case 67:
		return tm.jjMoveStringLiteralDfa7_2(active0, 0x8000000000000, active1, 0, active2, 0)
	case 68:
		if (active1 & 0x100) != 0 {
			return tm.jjStartNfaWithStates_2(6, 72, 269)
		} else {
			if (active1 & 0x800000000000000) != 0 {
				return tm.jjStartNfaWithStates_2(6, 123, 269)
			}
		}
		break
	case 69:
		if (active1 & 0x2000) != 0 {
			return tm.jjStopAtPos(6, 77)
		}
		return tm.jjMoveStringLiteralDfa7_2(active0, 0, active1, 0x100000, active2, 0)
	case 71:
		return tm.jjMoveStringLiteralDfa7_2(active0, 0, active1, 0x1000000000000000, active2, 0)
	case 73:
		return tm.jjMoveStringLiteralDfa7_2(active0, 0, active1, 0x8000, active2, 0)
	case 77:
		if (active1 & 0x8) != 0 {
			return tm.jjStartNfaWithStates_2(6, 67, 269)
		}
		return tm.jjMoveStringLiteralDfa7_2(active0, 0x10000000000, active1, 0, active2, 0)
	case 83:
		if (active0 & 0x1000000000000) != 0 {
			return tm.jjStartNfaWithStates_2(6, 48, 269)
		} else {
			if (active1 & 0x80) != 0 {
				return tm.jjStartNfaWithStates_2(6, 71, 269)
			} else {
				if (active1 & 0x80000) != 0 {
					return tm.jjStartNfaWithStates_2(6, 83, 271)
				}
			}
		}
		break
	case 104:
		if (active2 & 0x40000000) != 0 {
			return tm.jjStopAtPos(6, 158)
		}
		break
	case 111:
		if (active2 & 0x400000000) != 0 {
			return tm.jjStopAtPos(6, 162)
		}
		break
	case 113:
		if (active2 & 0x200000000) != 0 {
			return tm.jjStopAtPos(6, 161)
		} else {
			if (active2 & 0x800000000000) != 0 {
				return tm.jjStopAtPos(6, 175)
			}
		}
		break
	case 114:
		return tm.jjMoveStringLiteralDfa7_2(active0, 0, active1, 0, active2, 0x80)
	case 115:
		if (active2 & 0x10000000) != 0 {
			return tm.jjStopAtPos(6, 156)
		} else {
			if (active2 & 0x80000000) != 0 {
				return tm.jjStopAtPos(6, 159)
			}
		}
		return tm.jjMoveStringLiteralDfa7_2(active0, 0, active1, 0, active2, 0x78000100000)
	case 116:
		if (active2 & 0x100) != 0 {
			return tm.jjStopAtPos(6, 136)
		} else {
			if (active2 & 0x100000000000) != 0 {
				{
					tm.jjmatchedKind = 172
					tm.jjmatchedPos = 6
				}
			} else {
				if (active2 & 0x1000000000000) != 0 {
					{
						tm.jjmatchedKind = 176
						tm.jjmatchedPos = 6
					}
				}
			}
		}
		return tm.jjMoveStringLiteralDfa7_2(active0, 0, active1, 0, active2, 0x2200000000000)
	case 120:
		if (active2 & 0x20) != 0 {
			return tm.jjStopAtPos(6, 133)
		}
		break
	default:
		break
	}
	return tm.jjStartNfa_2(5, active0, active1, active2, 0, 0)
}

func (tm *SanyTokenManager) jjMoveStringLiteralDfa7_2(old0 uint64, active0 uint64, old1 uint64, active1 uint64, old2 uint64, active2 uint64) int32 {
	if ((sanyScannerAndUint64(&active0, old0)) | (sanyScannerAndUint64(&active1, old1)) | (sanyScannerAndUint64(&active2, old2))) == 0 {
		return tm.jjStartNfa_2(5, old0, old1, old2, 0, 0)
	}
	if !tm.readScanChar() {
		tm.jjStopStringLiteralDfa_2(6, active0, active1, active2, 0, 0)
		return 7
	}
	switch tm.curChar {
	case 69:
		if (active0 & 0x10000000000) != 0 {
			return tm.jjStopAtPos(7, 40)
		} else {
			if (active0 & 0x8000000000000) != 0 {
				return tm.jjStartNfaWithStates_2(7, 51, 269)
			}
		}
		return tm.jjMoveStringLiteralDfa8_2(active0, 0, active1, 0x1000000000000000, active2, 0)
	case 83:
		if (active1 & 0x100000) != 0 {
			return tm.jjStartNfaWithStates_2(7, 84, 271)
		}
		break
	case 86:
		return tm.jjMoveStringLiteralDfa8_2(active0, 0, active1, 0x8000, active2, 0)
	case 99:
		if (active2 & 0x80) != 0 {
			return tm.jjStopAtPos(7, 135)
		}
		break
	case 101:
		return tm.jjMoveStringLiteralDfa8_2(active0, 0, active1, 0, active2, 0x2278000100000)
	default:
		break
	}
	return tm.jjStartNfa_2(6, active0, active1, active2, 0, 0)
}

func (tm *SanyTokenManager) jjMoveStringLiteralDfa8_2(old0 uint64, active0 uint64, old1 uint64, active1 uint64, old2 uint64, active2 uint64) int32 {
	if ((sanyScannerAndUint64(&active0, old0)) | (sanyScannerAndUint64(&active1, old1)) | (sanyScannerAndUint64(&active2, old2))) == 0 {
		return tm.jjStartNfa_2(6, old0, old1, old2, 0, 0)
	}
	if !tm.readScanChar() {
		tm.jjStopStringLiteralDfa_2(7, 0, active1, active2, 0, 0)
		return 8
	}
	switch tm.curChar {
	case 68:
		if (active1 & 0x1000000000000000) != 0 {
			return tm.jjStartNfaWithStates_2(8, 124, 269)
		}
		break
	case 69:
		if (active1 & 0x8000) != 0 {
			return tm.jjStartNfaWithStates_2(8, 79, 269)
		}
		break
	case 99:
		return tm.jjMoveStringLiteralDfa9_2(active1, 0, active2, 0x100000)
	case 113:
		if (active2 & 0x200000000000) != 0 {
			return tm.jjStopAtPos(8, 173)
		} else {
			if (active2 & 0x2000000000000) != 0 {
				return tm.jjStopAtPos(8, 177)
			}
		}
		break
	case 116:
		if (active2 & 0x8000000000) != 0 {
			{
				tm.jjmatchedKind = 167
				tm.jjmatchedPos = 8
			}
		} else {
			if (active2 & 0x10000000000) != 0 {
				{
					tm.jjmatchedKind = 168
					tm.jjmatchedPos = 8
				}
			}
		}
		return tm.jjMoveStringLiteralDfa9_2(active1, 0, active2, 0x60000000000)
	default:
		break
	}
	return tm.jjStartNfa_2(7, 0, active1, active2, 0, 0)
}

func (tm *SanyTokenManager) jjMoveStringLiteralDfa9_2(old1 uint64, active1 uint64, old2 uint64, active2 uint64) int32 {
	if ((sanyScannerAndUint64(&active1, old1)) | (sanyScannerAndUint64(&active2, old2))) == 0 {
		return tm.jjStartNfa_2(7, 0, old1, old2, 0, 0)
	}
	if !tm.readScanChar() {
		tm.jjStopStringLiteralDfa_2(8, 0, 0, active2, 0, 0)
		return 9
	}
	switch tm.curChar {
	case 101:
		return tm.jjMoveStringLiteralDfa10_2(active2, 0x60000000000)
	case 116:
		if (active2 & 0x100000) != 0 {
			return tm.jjStopAtPos(9, 148)
		}
		break
	default:
		break
	}
	return tm.jjStartNfa_2(8, 0, 0, active2, 0, 0)
}

func (tm *SanyTokenManager) jjMoveStringLiteralDfa10_2(old2 uint64, active2 uint64) int32 {
	if (sanyScannerAndUint64(&active2, old2)) == 0 {
		return tm.jjStartNfa_2(8, 0, 0, old2, 0, 0)
	}
	if !tm.readScanChar() {
		tm.jjStopStringLiteralDfa_2(9, 0, 0, active2, 0, 0)
		return 10
	}
	switch tm.curChar {
	case 113:
		if (active2 & 0x20000000000) != 0 {
			return tm.jjStopAtPos(10, 169)
		} else {
			if (active2 & 0x40000000000) != 0 {
				return tm.jjStopAtPos(10, 170)
			}
		}
		break
	default:
		break
	}
	return tm.jjStartNfa_2(9, 0, 0, active2, 0, 0)
}

func (tm *SanyTokenManager) jjMoveNfa_2(startState int32, curPos int32) int32 {

	var startsAt int32 = int32(0)
	_ = startsAt
	tm.jjnewStateCnt = 269
	var i int32 = int32(1)
	_ = i
	tm.jjstateSet[0] = startState
	var j int32
	_ = j
	var kind int32 = int32(0x7fffffff)
	_ = kind
	for {
		{
			if sanyScannerPreIncrement(&tm.jjround) == 0x7fffffff {
				tm.reInitRounds()
			}
			if tm.curChar < 64 {
				{
					var l uint64 = uint64(uint64(1) << tm.curChar)
					_ = l
					for {
						{
							switch tm.jjstateSet[sanyScannerPreDecrement(&i)] {
							case 262:
								if (0x3ff000000000000 & l) != 0 {
									{
										if kind > 289 {
											kind = 289
										}
										tm.jjCheckNAdd(198)
									}
								} else {
									if tm.curChar == 46 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 196
									}
								}
								if (0x3ff000000000000 & l) != 0 {
									tm.jjCheckNAddTwoStates(195, 197)
								} else {
									if tm.curChar == 46 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 193
									}
								}
								if (0x3ff000000000000 & l) != 0 {
									tm.jjCheckNAddTwoStates(192, 194)
								}
								break
							case 10:
								if (0x3ff000000000000 & l) != 0 {
									{
										if kind > 289 {
											kind = 289
										}
										tm.jjCheckNAdd(198)
									}
								} else {
									if tm.curChar == 46 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 196
									}
								}
								if (0x3ff000000000000 & l) != 0 {
									tm.jjCheckNAddTwoStates(195, 197)
								} else {
									if tm.curChar == 46 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 193
									}
								}
								if (0x3ff000000000000 & l) != 0 {
									tm.jjCheckNAddTwoStates(192, 194)
								}
								break
							case 16:
								if (0x3ff000000000000 & l) != 0 {
									{
										if kind > 289 {
											kind = 289
										}
										tm.jjCheckNAdd(198)
									}
								} else {
									if tm.curChar == 46 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 196
									}
								}
								if (0x3ff000000000000 & l) != 0 {
									tm.jjCheckNAddTwoStates(195, 197)
								} else {
									if tm.curChar == 46 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 193
									}
								}
								if (0x3ff000000000000 & l) != 0 {
									tm.jjCheckNAddTwoStates(192, 194)
								}
								break
							case 108:
								if tm.curChar == 42 {
									{
										if kind > 27 {
											kind = 27
										}
									}
								}
								if tm.curChar == 42 {
									tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 107
								}
								break
							case 158:
								if (0x3ff000000000000 & l) != 0 {
									tm.jjCheckNAddTwoStates(236, 237)
								} else {
									if (0xc0000000000 & l) != 0 {
										tm.jjCheckNAdd(237)
									} else {
										if tm.curChar == 62 {
											{
												if kind > 122 {
													kind = 122
												}
											}
										} else {
											if tm.curChar == 45 {
												{
													if kind > 107 {
														kind = 107
													}
												}
											} else {
												if tm.curChar == 60 {
													{
														if kind > 102 {
															kind = 102
														}
													}
												}
											}
										}
									}
								}
								if (0x3ff000000000000 & l) != 0 {
									tm.jjCheckNAddTwoStates(233, 234)
								} else {
									if (0xc0000000000 & l) != 0 {
										tm.jjCheckNAdd(234)
									}
								}
								if (0x3ff000000000000 & l) != 0 {
									tm.jjCheckNAddTwoStates(228, 229)
								} else {
									if (0xc0000000000 & l) != 0 {
										tm.jjCheckNAdd(229)
									}
								}
								if (0x3ff000000000000 & l) != 0 {
									tm.jjCheckNAddTwoStates(222, 223)
								} else {
									if (0xc0000000000 & l) != 0 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 226
									}
								}
								break
							case 140:
								if (0x3ff000000000000 & l) != 0 {
									{
										if kind > 289 {
											kind = 289
										}
										tm.jjCheckNAdd(198)
									}
								} else {
									if tm.curChar == 46 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 196
									}
								}
								if (0x3ff000000000000 & l) != 0 {
									tm.jjCheckNAddTwoStates(195, 197)
								} else {
									if tm.curChar == 46 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 193
									}
								}
								if (0x3ff000000000000 & l) != 0 {
									tm.jjCheckNAddTwoStates(192, 194)
								}
								break
							case 181:
								if (0x3ff000000000000 & l) != 0 {
									{
										if kind > 289 {
											kind = 289
										}
										tm.jjCheckNAdd(190)
									}
								}
								if (0x3ff000000000000 & l) != 0 {
									tm.jjCheckNAddTwoStates(186, 188)
								}
								if (0x3ff000000000000 & l) != 0 {
									tm.jjCheckNAddTwoStates(182, 184)
								}
								break
							case 0:
								if (0x3ff000000000000 & l) != 0 {
									{
										if kind > 109 {
											kind = 109
										}
										tm.jjCheckNAddStates(37, 45)
									}
								} else {
									if tm.curChar == 60 {
										tm.jjCheckNAddStates(46, 53)
									} else {
										if tm.curChar == 62 {
											tm.jjAddStates(54, 55)
										} else {
											if tm.curChar == 61 {
												tm.jjAddStates(56, 57)
											} else {
												if tm.curChar == 45 {
													tm.jjAddStates(58, 60)
												} else {
													if tm.curChar == 40 {
														tm.jjAddStates(61, 62)
													} else {
														if tm.curChar == 47 {
															tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 46
														} else {
															if tm.curChar == 34 {
																tm.jjCheckNAddStates(63, 65)
															} else {
																if tm.curChar == 58 {
																	tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 18
																}
															}
														}
													}
												}
											}
										}
									}
								}
								if tm.curChar == 60 {
									tm.jjAddStates(66, 68)
								} else {
									if tm.curChar == 48 {
										{
											if kind > 109 {
												kind = 109
											}
										}
									}
								}
								break
							case 103:
								if (0x3ff000000000000 & l) != 0 {
									{
										if kind > 289 {
											kind = 289
										}
										tm.jjCheckNAdd(198)
									}
								} else {
									if tm.curChar == 46 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 196
									}
								}
								if (0x3ff000000000000 & l) != 0 {
									tm.jjCheckNAddTwoStates(195, 197)
								} else {
									if tm.curChar == 46 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 193
									}
								}
								if (0x3ff000000000000 & l) != 0 {
									tm.jjCheckNAddTwoStates(192, 194)
								}
								break
							case 270, 52:
								if (0x3ff000000000000 & l) != 0 {
									tm.jjCheckNAddTwoStates(52, 53)
								}
								break
							case 162:
								if tm.curChar == 62 {
									{
										if kind > 104 {
											kind = 104
										}
									}
								}
								if tm.curChar == 62 {
									tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 23
								}
								break
							case 260:
								if (0x3ff000000000000 & l) != 0 {
									{
										if kind > 289 {
											kind = 289
										}
										tm.jjCheckNAdd(198)
									}
								} else {
									if tm.curChar == 46 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 196
									}
								}
								if (0x3ff000000000000 & l) != 0 {
									tm.jjCheckNAddTwoStates(195, 197)
								} else {
									if tm.curChar == 46 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 193
									}
								}
								if (0x3ff000000000000 & l) != 0 {
									tm.jjCheckNAddTwoStates(192, 194)
								}
								break
							case 131:
								if tm.curChar == 61 {
									{
										if kind > 93 {
											kind = 93
										}
									}
								}
								if tm.curChar == 61 {
									tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 130
								}
								break
							case 263:
								if (0x3ff000000000000 & l) != 0 {
									{
										if kind > 289 {
											kind = 289
										}
										tm.jjCheckNAdd(198)
									}
								} else {
									if tm.curChar == 46 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 196
									}
								}
								if (0x3ff000000000000 & l) != 0 {
									tm.jjCheckNAddTwoStates(195, 197)
								} else {
									if tm.curChar == 46 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 193
									}
								}
								if (0x3ff000000000000 & l) != 0 {
									tm.jjCheckNAddTwoStates(192, 194)
								}
								break
							case 11:
								if (0x3ff000000000000 & l) != 0 {
									{
										if kind > 289 {
											kind = 289
										}
										tm.jjCheckNAdd(198)
									}
								} else {
									if tm.curChar == 46 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 196
									}
								}
								if (0x3ff000000000000 & l) != 0 {
									tm.jjCheckNAddTwoStates(195, 197)
								} else {
									if tm.curChar == 46 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 193
									}
								}
								if (0x3ff000000000000 & l) != 0 {
									tm.jjCheckNAddTwoStates(192, 194)
								}
								break
							case 249:
								if (0x3ff000000000000 & l) != 0 {
									{
										if kind > 289 {
											kind = 289
										}
										tm.jjCheckNAdd(198)
									}
								} else {
									if tm.curChar == 46 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 196
									}
								}
								if (0x3ff000000000000 & l) != 0 {
									tm.jjCheckNAddTwoStates(195, 197)
								} else {
									if tm.curChar == 46 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 193
									}
								}
								if (0x3ff000000000000 & l) != 0 {
									tm.jjCheckNAddTwoStates(192, 194)
								}
								break
							case 9:
								if (0x3ff000000000000 & l) != 0 {
									{
										if kind > 289 {
											kind = 289
										}
										tm.jjCheckNAdd(198)
									}
								} else {
									if tm.curChar == 46 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 196
									}
								}
								if (0x3ff000000000000 & l) != 0 {
									tm.jjCheckNAddTwoStates(195, 197)
								} else {
									if tm.curChar == 46 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 193
									}
								}
								if (0x3ff000000000000 & l) != 0 {
									tm.jjCheckNAddTwoStates(192, 194)
								}
								break
							case 101:
								if (0x3ff000000000000 & l) != 0 {
									{
										if kind > 289 {
											kind = 289
										}
										tm.jjCheckNAdd(198)
									}
								} else {
									if tm.curChar == 46 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 196
									}
								}
								if (0x3ff000000000000 & l) != 0 {
									tm.jjCheckNAddTwoStates(195, 197)
								} else {
									if tm.curChar == 46 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 193
									}
								}
								if (0x3ff000000000000 & l) != 0 {
									tm.jjCheckNAddTwoStates(192, 194)
								}
								break
							case 121:
								if tm.curChar == 62 {
									{
										if kind > 106 {
											kind = 106
										}
									}
								} else {
									if tm.curChar == 45 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 124
									}
								}
								if tm.curChar == 45 {
									tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 120
								}
								break
							case 261:
								if (0x3ff000000000000 & l) != 0 {
									{
										if kind > 289 {
											kind = 289
										}
										tm.jjCheckNAdd(198)
									}
								} else {
									if tm.curChar == 46 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 196
									}
								}
								if (0x3ff000000000000 & l) != 0 {
									tm.jjCheckNAddTwoStates(195, 197)
								} else {
									if tm.curChar == 46 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 193
									}
								}
								if (0x3ff000000000000 & l) != 0 {
									tm.jjCheckNAddTwoStates(192, 194)
								}
								break
							case 15:
								if (0x3ff000000000000 & l) != 0 {
									{
										if kind > 289 {
											kind = 289
										}
										tm.jjCheckNAdd(198)
									}
								} else {
									if tm.curChar == 46 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 196
									}
								}
								if (0x3ff000000000000 & l) != 0 {
									tm.jjCheckNAddTwoStates(195, 197)
								} else {
									if tm.curChar == 46 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 193
									}
								}
								if (0x3ff000000000000 & l) != 0 {
									tm.jjCheckNAddTwoStates(192, 194)
								}
								break
							case 200:
								if tm.curChar == 47 && kind > 132 {
									kind = 132
								}
								break
							case 271:
								if (0x3ff000000000000 & l) != 0 {
									{
										if kind > 289 {
											kind = 289
										}
										tm.jjCheckNAdd(190)
									}
								} else {
									if tm.curChar == 46 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 187
									}
								}
								if (0x3ff000000000000 & l) != 0 {
									tm.jjCheckNAddTwoStates(186, 188)
								} else {
									if tm.curChar == 46 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 183
									}
								}
								if (0x3ff000000000000 & l) != 0 {
									tm.jjCheckNAddTwoStates(182, 184)
								}
								break
							case 120:
								if tm.curChar == 45 {
									tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 122
								}
								if tm.curChar == 45 {
									tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 111
								}
								break
							case 102:
								if (0x3ff000000000000 & l) != 0 {
									{
										if kind > 289 {
											kind = 289
										}
										tm.jjCheckNAdd(198)
									}
								} else {
									if tm.curChar == 46 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 196
									}
								}
								if (0x3ff000000000000 & l) != 0 {
									tm.jjCheckNAddTwoStates(195, 197)
								} else {
									if tm.curChar == 46 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 193
									}
								}
								if (0x3ff000000000000 & l) != 0 {
									tm.jjCheckNAddTwoStates(192, 194)
								}
								break
							case 91:
								if (0x3ff000000000000 & l) != 0 {
									{
										if kind > 289 {
											kind = 289
										}
										tm.jjCheckNAdd(198)
									}
								} else {
									if tm.curChar == 46 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 196
									}
								}
								if (0x3ff000000000000 & l) != 0 {
									tm.jjCheckNAddTwoStates(195, 197)
								} else {
									if tm.curChar == 46 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 193
									}
								}
								if (0x3ff000000000000 & l) != 0 {
									tm.jjCheckNAddTwoStates(192, 194)
								}
								break
							case 269:
								if (0x3ff000000000000 & l) != 0 {
									{
										if kind > 289 {
											kind = 289
										}
										tm.jjCheckNAdd(198)
									}
								} else {
									if tm.curChar == 46 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 196
									}
								}
								if (0x3ff000000000000 & l) != 0 {
									tm.jjCheckNAddTwoStates(195, 197)
								} else {
									if tm.curChar == 46 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 193
									}
								}
								if (0x3ff000000000000 & l) != 0 {
									tm.jjCheckNAddTwoStates(192, 194)
								}
								break
							case 18:
								if tm.curChar == 58 && kind > 90 {
									kind = 90
								}
								break
							case 19:
								if tm.curChar == 58 {
									tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 18
								}
								break
							case 28:
								if tm.curChar == 62 && kind > 108 {
									kind = 108
								}
								break
							case 29:
								if tm.curChar == 45 {
									tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 28
								}
								break
							case 32:
								if tm.curChar == 48 && kind > 109 {
									kind = 109
								}
								break
							case 33, 36:
								if tm.curChar == 34 {
									tm.jjCheckNAddStates(63, 65)
								}
								break
							case 34:
								if (0xfffffffbffffdbff & l) != 0 {
									tm.jjCheckNAddStates(63, 65)
								}
								break
							case 37:
								if tm.curChar == 34 && kind > 110 {
									kind = 110
								}
								break
							case 38:
								if tm.curChar == 43 && kind > 114 {
									kind = 114
								}
								break
							case 47:
								if tm.curChar == 47 {
									tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 46
								}
								break
							case 54:
								if (0x3ff000000000000 & l) == 0 {
									break
								}
								if kind > 289 {
									kind = 289
								}
								tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 54
								break
							case 56, 57:
								if (0x3ff000000000000 & l) == 0 {
									break
								}
								if kind > 289 {
									kind = 289
								}
								tm.jjCheckNAdd(57)
								break
							case 63:
								if (0xff000000000000 & l) == 0 {
									break
								}
								if kind > 109 {
									kind = 109
								}
								tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 63
								break
							case 65:
								if (0x3000000000000 & l) == 0 {
									break
								}
								if kind > 109 {
									kind = 109
								}
								tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 65
								break
							case 67:
								if (0x3ff000000000000 & l) == 0 {
									break
								}
								if kind > 109 {
									kind = 109
								}
								tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 67
								break
							case 106:
								if tm.curChar == 40 {
									tm.jjAddStates(61, 62)
								}
								break
							case 107:
								if tm.curChar == 46 && kind > 27 {
									kind = 27
								}
								break
							case 109:
								if tm.curChar == 42 && kind > 27 {
									kind = 27
								}
								break
							case 110:
								if tm.curChar == 45 {
									tm.jjAddStates(58, 60)
								}
								break
							case 111, 112:
								if tm.curChar == 45 {
									tm.jjCheckNAddStates(69, 71)
								}
								break
							case 113:
								if tm.curChar == 32 {
									tm.jjCheckNAddTwoStates(113, 119)
								}
								break
							case 122, 123:
								if tm.curChar != 45 {
									break
								}
								if kind > 36 {
									kind = 36
								}
								tm.jjCheckNAdd(123)
								break
							case 124:
								if tm.curChar == 45 {
									tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 122
								}
								break
							case 125:
								if tm.curChar == 45 {
									tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 124
								}
								break
							case 126:
								if tm.curChar == 62 && kind > 106 {
									kind = 106
								}
								break
							case 127:
								if tm.curChar == 61 {
									tm.jjAddStates(56, 57)
								}
								break
							case 128, 129:
								if tm.curChar != 61 {
									break
								}
								if kind > 37 {
									kind = 37
								}
								tm.jjCheckNAdd(129)
								break
							case 130:
								if tm.curChar == 61 {
									tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 128
								}
								break
							case 132:
								if tm.curChar == 61 && kind > 93 {
									kind = 93
								}
								break
							case 157:
								if tm.curChar == 60 {
									tm.jjAddStates(66, 68)
								}
								break
							case 159:
								if tm.curChar == 45 && kind > 107 {
									kind = 107
								}
								break
							case 160:
								if tm.curChar == 62 && kind > 122 {
									kind = 122
								}
								break
							case 161:
								if tm.curChar == 62 {
									tm.jjAddStates(54, 55)
								}
								break
							case 163:
								if tm.curChar == 62 && kind > 104 {
									kind = 104
								}
								break
							case 164:
								if (0x3ff000000000000 & l) == 0 {
									break
								}
								if kind > 109 {
									kind = 109
								}
								tm.jjCheckNAddStates(37, 45)
								break
							case 165:
								if (0x3ff000000000000 & l) == 0 {
									break
								}
								if kind > 109 {
									kind = 109
								}
								tm.jjCheckNAdd(165)
								break
							case 166:
								if (0x3ff000000000000 & l) != 0 {
									tm.jjCheckNAddTwoStates(166, 169)
								}
								break
							case 168, 183, 193:
								if tm.curChar == 47 {
									tm.jjCheckNAdd(167)
								}
								break
							case 169:
								if tm.curChar == 46 {
									tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 168
								}
								break
							case 170:
								if (0x3ff000000000000 & l) != 0 {
									tm.jjCheckNAddTwoStates(170, 173)
								}
								break
							case 171:
								if tm.curChar == 47 && kind > 112 {
									kind = 112
								}
								break
							case 173:
								if tm.curChar == 46 {
									tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 172
								}
								break
							case 174:
								if (0x3ff000000000000 & l) != 0 {
									tm.jjCheckNAddTwoStates(174, 175)
								}
								break
							case 176:
								if (0x3ff000000000000 & l) == 0 {
									break
								}
								if kind > 289 {
									kind = 289
								}
								tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 176
								break
							case 177:
								if (0x3ff000000000000 & l) != 0 {
									tm.jjCheckNAddTwoStates(177, 178)
								}
								break
							case 179:
								if (0x3ff000000000000 & l) == 0 {
									break
								}
								if kind > 289 {
									kind = 289
								}
								tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 179
								break
							case 182:
								if (0x3ff000000000000 & l) != 0 {
									tm.jjCheckNAddTwoStates(182, 184)
								}
								break
							case 184:
								if tm.curChar == 46 {
									tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 183
								}
								break
							case 185:
								if (0x3ff000000000000 & l) != 0 {
									tm.jjCheckNAddTwoStates(186, 188)
								}
								break
							case 186:
								if (0x3ff000000000000 & l) != 0 {
									tm.jjCheckNAddTwoStates(186, 188)
								}
								break
							case 188:
								if tm.curChar == 46 {
									tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 187
								}
								break
							case 189:
								if (0x3ff000000000000 & l) == 0 {
									break
								}
								if kind > 289 {
									kind = 289
								}
								tm.jjCheckNAdd(190)
								break
							case 190:
								if (0x3ff000000000000 & l) == 0 {
									break
								}
								if kind > 289 {
									kind = 289
								}
								tm.jjCheckNAdd(190)
								break
							case 192:
								if (0x3ff000000000000 & l) != 0 {
									tm.jjCheckNAddTwoStates(192, 194)
								}
								break
							case 194:
								if tm.curChar == 46 {
									tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 193
								}
								break
							case 195:
								if (0x3ff000000000000 & l) != 0 {
									tm.jjCheckNAddTwoStates(195, 197)
								}
								break
							case 197:
								if tm.curChar == 46 {
									tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 196
								}
								break
							case 198:
								if (0x3ff000000000000 & l) == 0 {
									break
								}
								if kind > 289 {
									kind = 289
								}
								tm.jjCheckNAdd(198)
								break
							case 221:
								if tm.curChar == 60 {
									tm.jjCheckNAddStates(46, 53)
								}
								break
							case 222:
								if (0x3ff000000000000 & l) != 0 {
									tm.jjCheckNAddTwoStates(222, 223)
								}
								break
							case 223:
								if tm.curChar == 62 {
									tm.jjCheckNAdd(224)
								}
								break
							case 224:
								if (0x3ff000000000000 & l) == 0 {
									break
								}
								if kind > 290 {
									kind = 290
								}
								tm.jjCheckNAdd(224)
								break
							case 225:
								if (0xc0000000000 & l) != 0 {
									tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 226
								}
								break
							case 226:
								if tm.curChar == 62 {
									tm.jjCheckNAdd(227)
								}
								break
							case 227:
								if (0x3ff000000000000 & l) == 0 {
									break
								}
								if kind > 291 {
									kind = 291
								}
								tm.jjCheckNAdd(227)
								break
							case 228:
								if (0x3ff000000000000 & l) != 0 {
									tm.jjCheckNAddTwoStates(228, 229)
								}
								break
							case 229:
								if tm.curChar == 62 {
									tm.jjCheckNAdd(230)
								}
								break
							case 230:
								if (0x3ff000000000000 & l) != 0 {
									tm.jjCheckNAddTwoStates(230, 231)
								}
								break
							case 231:
								if tm.curChar != 46 {
									break
								}
								if kind > 292 {
									kind = 292
								}
								tm.jjCheckNAdd(231)
								break
							case 232:
								if (0xc0000000000 & l) != 0 {
									tm.jjCheckNAdd(229)
								}
								break
							case 233:
								if (0x3ff000000000000 & l) != 0 {
									tm.jjCheckNAddTwoStates(233, 234)
								}
								break
							case 234:
								if tm.curChar == 62 && kind > 293 {
									kind = 293
								}
								break
							case 235:
								if (0xc0000000000 & l) != 0 {
									tm.jjCheckNAdd(234)
								}
								break
							case 236:
								if (0x3ff000000000000 & l) != 0 {
									tm.jjCheckNAddTwoStates(236, 237)
								}
								break
							case 237:
								if tm.curChar == 62 {
									tm.jjCheckNAddStates(72, 74)
								}
								break
							case 238:
								if tm.curChar != 42 {
									break
								}
								if kind > 294 {
									kind = 294
								}
								tm.jjCheckNAdd(239)
								break
							case 239:
								if tm.curChar != 46 {
									break
								}
								if kind > 294 {
									kind = 294
								}
								tm.jjCheckNAdd(239)
								break
							case 240:
								if tm.curChar != 45 {
									break
								}
								if kind > 294 {
									kind = 294
								}
								tm.jjCheckNAdd(241)
								break
							case 241:
								if tm.curChar != 46 {
									break
								}
								if kind > 294 {
									kind = 294
								}
								tm.jjCheckNAdd(241)
								break
							case 242:
								if tm.curChar != 46 {
									break
								}
								if kind > 294 {
									kind = 294
								}
								tm.jjCheckNAdd(242)
								break
							case 243:
								if (0xc0000000000 & l) != 0 {
									tm.jjCheckNAdd(237)
								}
								break
							default:
								break
							}
						}
						if !(i != startsAt) {
							break
						}
					}
				}
			} else {
				if tm.curChar < 128 {
					{
						var l uint64 = uint64(uint64(1) << (tm.curChar & 077))
						_ = l
						for {
							{
								switch tm.jjstateSet[sanyScannerPreDecrement(&i)] {
								case 262:
									if (0x7fffffe87fffffe & l) != 0 {
										{
											if kind > 289 {
												kind = 289
											}
											tm.jjCheckNAdd(198)
										}
									}
									if (0x7fffffe07fffffe & l) != 0 {
										tm.jjCheckNAddTwoStates(195, 197)
									}
									if (0x7fffffe07fffffe & l) != 0 {
										tm.jjCheckNAddTwoStates(192, 194)
									}
									if tm.curChar == 85 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 261
									}
									break
								case 10:
									if (0x7fffffe87fffffe & l) != 0 {
										{
											if kind > 289 {
												kind = 289
											}
											tm.jjCheckNAdd(198)
										}
									}
									if (0x7fffffe07fffffe & l) != 0 {
										tm.jjCheckNAddTwoStates(195, 197)
									}
									if (0x7fffffe07fffffe & l) != 0 {
										tm.jjCheckNAddTwoStates(192, 194)
									}
									if tm.curChar == 79 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 9
									}
									break
								case 16:
									if (0x7fffffe87fffffe & l) != 0 {
										{
											if kind > 289 {
												kind = 289
											}
											tm.jjCheckNAdd(198)
										}
									}
									if (0x7fffffe07fffffe & l) != 0 {
										tm.jjCheckNAddTwoStates(195, 197)
									}
									if (0x7fffffe07fffffe & l) != 0 {
										tm.jjCheckNAddTwoStates(192, 194)
									}
									if tm.curChar == 69 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 15
									}
									break
								case 140:
									if (0x7fffffe87fffffe & l) != 0 {
										{
											if kind > 289 {
												kind = 289
											}
											tm.jjCheckNAdd(198)
										}
									}
									if (0x7fffffe07fffffe & l) != 0 {
										tm.jjCheckNAddTwoStates(195, 197)
									}
									if (0x7fffffe07fffffe & l) != 0 {
										tm.jjCheckNAddTwoStates(192, 194)
									}
									if tm.curChar == 79 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 155
									}
									if tm.curChar == 79 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 147
									}
									if tm.curChar == 79 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 139
									}
									break
								case 181:
									if (0x7fffffe87ffffbe & l) != 0 {
										{
											if kind > 289 {
												kind = 289
											}
											tm.jjCheckNAdd(190)
										}
									} else {
										if tm.curChar == 70 {
											{
												if kind > 289 {
													kind = 289
												}
												tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 56
											}
										}
									}
									if (0x7fffffe07ffffbe & l) != 0 {
										tm.jjCheckNAddTwoStates(186, 188)
									}
									if (0x7fffffe07ffffbe & l) != 0 {
										tm.jjCheckNAddTwoStates(182, 184)
									}
									break
								case 0:
									if (0x7fffffe0777fffe & l) != 0 {
										{
											if kind > 289 {
												kind = 289
											}
											tm.jjCheckNAddStates(75, 79)
										}
									} else {
										if (0x880000 & l) != 0 {
											{
												if kind > 289 {
													kind = 289
												}
												tm.jjAddStates(80, 82)
											}
										} else {
											if tm.curChar == 92 {
												tm.jjAddStates(83, 89)
											} else {
												if tm.curChar == 64 {
													{
														if kind > 289 {
															kind = 289
														}
													}
												} else {
													if tm.curChar == 95 {
														tm.jjCheckNAddTwoStates(52, 53)
													} else {
														if tm.curChar == 91 {
															tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 42
														} else {
															if tm.curChar == 94 {
																tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 38
															} else {
																if tm.curChar == 124 {
																	tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 29
																}
															}
														}
													}
												}
											}
										}
									}
									if tm.curChar == 65 {
										tm.jjAddStates(90, 93)
									} else {
										if tm.curChar == 67 {
											tm.jjAddStates(94, 96)
										} else {
											if tm.curChar == 68 {
												tm.jjAddStates(97, 98)
											} else {
												if tm.curChar == 84 {
													tm.jjAddStates(99, 100)
												} else {
													if tm.curChar == 86 {
														tm.jjAddStates(101, 102)
													} else {
														if tm.curChar == 92 {
															tm.jjAddStates(103, 105)
														} else {
															if tm.curChar == 83 {
																tm.jjCheckNAdd(55)
															} else {
																if tm.curChar == 87 {
																	tm.jjCheckNAdd(55)
																} else {
																	if tm.curChar == 76 {
																		tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 16
																	} else {
																		if tm.curChar == 80 {
																			tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 11
																		}
																	}
																}
															}
														}
													}
												}
											}
										}
									}
									break
								case 103:
									if (0x7fffffe87fffffe & l) != 0 {
										{
											if kind > 289 {
												kind = 289
											}
											tm.jjCheckNAdd(198)
										}
									}
									if (0x7fffffe07fffffe & l) != 0 {
										tm.jjCheckNAddTwoStates(195, 197)
									}
									if (0x7fffffe07fffffe & l) != 0 {
										tm.jjCheckNAddTwoStates(192, 194)
									}
									if tm.curChar == 83 {
										{
											if kind > 66 {
												kind = 66
											}
										}
									}
									break
								case 270:
									if (0x7fffffe87fffffe & l) != 0 {
										tm.jjCheckNAddTwoStates(52, 53)
									}
									if (0x7fffffe07fffffe & l) != 0 {
										{
											if kind > 289 {
												kind = 289
											}
											tm.jjCheckNAdd(54)
										}
									}
									break
								case 260:
									if (0x7fffffe87fffffe & l) != 0 {
										{
											if kind > 289 {
												kind = 289
											}
											tm.jjCheckNAdd(198)
										}
									}
									if (0x7fffffe07fffffe & l) != 0 {
										tm.jjCheckNAddTwoStates(195, 197)
									}
									if (0x7fffffe07fffffe & l) != 0 {
										tm.jjCheckNAddTwoStates(192, 194)
									}
									if tm.curChar == 80 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 259
									}
									break
								case 263:
									if (0x7fffffe87fffffe & l) != 0 {
										{
											if kind > 289 {
												kind = 289
											}
											tm.jjCheckNAdd(198)
										}
									}
									if (0x7fffffe07fffffe & l) != 0 {
										tm.jjCheckNAddTwoStates(195, 197)
									}
									if (0x7fffffe07fffffe & l) != 0 {
										tm.jjCheckNAddTwoStates(192, 194)
									}
									if tm.curChar == 83 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 262
									}
									break
								case 11:
									if (0x7fffffe87fffffe & l) != 0 {
										{
											if kind > 289 {
												kind = 289
											}
											tm.jjCheckNAdd(198)
										}
									}
									if (0x7fffffe07fffffe & l) != 0 {
										tm.jjCheckNAddTwoStates(195, 197)
									}
									if (0x7fffffe07fffffe & l) != 0 {
										tm.jjCheckNAddTwoStates(192, 194)
									}
									if tm.curChar == 82 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 10
									}
									break
								case 249:
									if (0x7fffffe87fffffe & l) != 0 {
										{
											if kind > 289 {
												kind = 289
											}
											tm.jjCheckNAdd(198)
										}
									}
									if (0x7fffffe07fffffe & l) != 0 {
										tm.jjCheckNAddTwoStates(195, 197)
									}
									if (0x7fffffe07fffffe & l) != 0 {
										tm.jjCheckNAddTwoStates(192, 194)
									}
									if tm.curChar == 88 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 267
									} else {
										if tm.curChar == 83 {
											tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 263
										} else {
											if tm.curChar == 67 {
												tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 254
											}
										}
									}
									if tm.curChar == 67 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 248
									}
									break
								case 9:
									if (0x7fffffe87fffffe & l) != 0 {
										{
											if kind > 289 {
												kind = 289
											}
											tm.jjCheckNAdd(198)
										}
									}
									if (0x7fffffe07fffffe & l) != 0 {
										tm.jjCheckNAddTwoStates(195, 197)
									}
									if (0x7fffffe07fffffe & l) != 0 {
										tm.jjCheckNAddTwoStates(192, 194)
									}
									if tm.curChar == 80 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 8
									}
									break
								case 101:
									if (0x7fffffe87fffffe & l) != 0 {
										{
											if kind > 289 {
												kind = 289
											}
											tm.jjCheckNAdd(198)
										}
									}
									if (0x7fffffe07fffffe & l) != 0 {
										tm.jjCheckNAddTwoStates(195, 197)
									}
									if (0x7fffffe07fffffe & l) != 0 {
										tm.jjCheckNAddTwoStates(192, 194)
									}
									if tm.curChar == 70 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 103
									}
									if tm.curChar == 70 {
										{
											if kind > 66 {
												kind = 66
											}
										}
									}
									break
								case 261:
									if (0x7fffffe87fffffe & l) != 0 {
										{
											if kind > 289 {
												kind = 289
											}
											tm.jjCheckNAdd(198)
										}
									}
									if (0x7fffffe07fffffe & l) != 0 {
										tm.jjCheckNAddTwoStates(195, 197)
									}
									if (0x7fffffe07fffffe & l) != 0 {
										tm.jjCheckNAddTwoStates(192, 194)
									}
									if tm.curChar == 77 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 260
									}
									break
								case 15:
									if (0x7fffffe87fffffe & l) != 0 {
										{
											if kind > 289 {
												kind = 289
											}
											tm.jjCheckNAdd(198)
										}
									}
									if (0x7fffffe07fffffe & l) != 0 {
										tm.jjCheckNAddTwoStates(195, 197)
									}
									if (0x7fffffe07fffffe & l) != 0 {
										tm.jjCheckNAddTwoStates(192, 194)
									}
									if tm.curChar == 77 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 14
									}
									break
								case 200:
									if (0x10000000100 & l) != 0 {
										tm.jjCheckNAdd(67)
									} else {
										if (0x400000004 & l) != 0 {
											tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 65
										} else {
											if (0x800000008000 & l) != 0 {
												tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 63
											} else {
												if tm.curChar == 105 {
													tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 219
												} else {
													if tm.curChar == 108 {
														tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 216
													} else {
														if tm.curChar == 102 {
															tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 212
														} else {
															if tm.curChar == 65 {
																{
																	if kind > 49 {
																		kind = 49
																	}
																}
															} else {
																if tm.curChar == 101 {
																	tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 205
																} else {
																	if tm.curChar == 69 {
																		{
																			if kind > 47 {
																				kind = 47
																			}
																		}
																	}
																}
															}
														}
													}
												}
											}
										}
									}
									break
								case 271:
									if (0x7fffffe87fffffe & l) != 0 {
										{
											if kind > 289 {
												kind = 289
											}
											tm.jjCheckNAdd(190)
										}
									}
									if (0x7fffffe07fffffe & l) != 0 {
										tm.jjCheckNAddTwoStates(186, 188)
									}
									if (0x7fffffe07fffffe & l) != 0 {
										tm.jjCheckNAddTwoStates(182, 184)
									}
									break
								case 102:
									if (0x7fffffe87fffffe & l) != 0 {
										{
											if kind > 289 {
												kind = 289
											}
											tm.jjCheckNAdd(198)
										}
									}
									if (0x7fffffe07fffffe & l) != 0 {
										tm.jjCheckNAddTwoStates(195, 197)
									}
									if (0x7fffffe07fffffe & l) != 0 {
										tm.jjCheckNAddTwoStates(192, 194)
									}
									if tm.curChar == 69 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 104
									}
									if tm.curChar == 69 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 101
									}
									break
								case 91:
									if (0x7fffffe87fffffe & l) != 0 {
										{
											if kind > 289 {
												kind = 289
											}
											tm.jjCheckNAdd(198)
										}
									}
									if (0x7fffffe07fffffe & l) != 0 {
										tm.jjCheckNAddTwoStates(195, 197)
									}
									if (0x7fffffe07fffffe & l) != 0 {
										tm.jjCheckNAddTwoStates(192, 194)
									}
									if tm.curChar == 69 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 98
									}
									if tm.curChar == 69 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 90
									}
									break
								case 269:
									if (0x7fffffe87fffffe & l) != 0 {
										{
											if kind > 289 {
												kind = 289
											}
											tm.jjCheckNAdd(198)
										}
									}
									if (0x7fffffe07fffffe & l) != 0 {
										tm.jjCheckNAddTwoStates(195, 197)
									}
									if (0x7fffffe07fffffe & l) != 0 {
										tm.jjCheckNAddTwoStates(192, 194)
									}
									break
								case 2:
									if tm.curChar == 78 && kind > 58 {
										kind = 58
									}
									break
								case 3:
									if tm.curChar == 79 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 2
									}
									break
								case 4:
									if tm.curChar == 73 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 3
									}
									break
								case 5:
									if tm.curChar == 84 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 4
									}
									break
								case 6:
									if tm.curChar == 73 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 5
									}
									break
								case 7:
									if tm.curChar == 83 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 6
									}
									break
								case 8:
									if tm.curChar == 79 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 7
									}
									break
								case 12:
									if tm.curChar == 80 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 11
									}
									break
								case 13:
									if tm.curChar == 65 && kind > 58 {
										kind = 58
									}
									break
								case 14:
									if tm.curChar == 77 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 13
									}
									break
								case 17:
									if tm.curChar == 76 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 16
									}
									break
								case 23:
									if tm.curChar == 95 && kind > 103 {
										kind = 103
									}
									break
								case 30:
									if tm.curChar == 124 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 29
									}
									break
								case 34:
									if (0xffffffffefffffff & l) != 0 {
										tm.jjCheckNAddStates(63, 65)
									}
									break
								case 35:
									if tm.curChar == 92 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 36
									}
									break
								case 36:
									if (0x14404010000000 & l) != 0 {
										tm.jjCheckNAddStates(63, 65)
									}
									break
								case 39:
									if tm.curChar == 94 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 38
									}
									break
								case 42:
									if tm.curChar == 93 && kind > 121 {
										kind = 121
									}
									break
								case 43:
									if tm.curChar == 91 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 42
									}
									break
								case 46:
									if tm.curChar == 92 && kind > 129 {
										kind = 129
									}
									break
								case 51:
									if tm.curChar == 95 {
										tm.jjCheckNAddTwoStates(52, 53)
									}
									break
								case 52:
									if (0x7fffffe87fffffe & l) != 0 {
										tm.jjCheckNAddTwoStates(52, 53)
									}
									break
								case 53:
									if (0x7fffffe07fffffe & l) == 0 {
										break
									}
									if kind > 289 {
										kind = 289
									}
									tm.jjCheckNAdd(54)
									break
								case 54:
									if (0x7fffffe87fffffe & l) == 0 {
										break
									}
									if kind > 289 {
										kind = 289
									}
									tm.jjCheckNAdd(54)
									break
								case 55:
									if tm.curChar != 70 {
										break
									}
									if kind > 289 {
										kind = 289
									}
									tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 56
									break
								case 56:
									if (0x7fffffe07fffffe & l) == 0 {
										break
									}
									if kind > 289 {
										kind = 289
									}
									tm.jjCheckNAdd(57)
									break
								case 57:
									if (0x7fffffe87fffffe & l) == 0 {
										break
									}
									if kind > 289 {
										kind = 289
									}
									tm.jjCheckNAdd(57)
									break
								case 58:
									if tm.curChar == 87 {
										tm.jjCheckNAdd(55)
									}
									break
								case 59:
									if tm.curChar == 83 {
										tm.jjCheckNAdd(55)
									}
									break
								case 60:
									if tm.curChar == 64 && kind > 289 {
										kind = 289
									}
									break
								case 61:
									if tm.curChar == 92 {
										tm.jjAddStates(103, 105)
									}
									break
								case 62:
									if (0x800000008000 & l) != 0 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 63
									}
									break
								case 64:
									if (0x400000004 & l) != 0 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 65
									}
									break
								case 66:
									if (0x10000000100 & l) != 0 {
										tm.jjCheckNAdd(67)
									}
									break
								case 67:
									if (0x7e0000007e & l) == 0 {
										break
									}
									if kind > 109 {
										kind = 109
									}
									tm.jjCheckNAdd(67)
									break
								case 68:
									if tm.curChar == 86 {
										tm.jjAddStates(101, 102)
									}
									break
								case 69:
									if tm.curChar == 69 && kind > 85 {
										kind = 85
									}
									break
								case 70:
									if tm.curChar == 76 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 69
									}
									break
								case 71:
									if tm.curChar == 66 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 70
									}
									break
								case 72:
									if tm.curChar == 65 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 71
									}
									break
								case 73:
									if tm.curChar == 73 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 72
									}
									break
								case 74:
									if tm.curChar == 82 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 73
									}
									break
								case 75:
									if tm.curChar == 65 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 74
									}
									break
								case 76:
									if tm.curChar == 83 && kind > 85 {
										kind = 85
									}
									break
								case 77:
									if tm.curChar == 69 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 76
									}
									break
								case 78:
									if tm.curChar == 76 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 77
									}
									break
								case 79:
									if tm.curChar == 66 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 78
									}
									break
								case 80:
									if tm.curChar == 65 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 79
									}
									break
								case 81:
									if tm.curChar == 73 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 80
									}
									break
								case 82:
									if tm.curChar == 82 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 81
									}
									break
								case 83:
									if tm.curChar == 65 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 82
									}
									break
								case 84:
									if tm.curChar == 84 {
										tm.jjAddStates(99, 100)
									}
									break
								case 85:
									if tm.curChar == 76 && kind > 81 {
										kind = 81
									}
									break
								case 86:
									if tm.curChar == 65 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 85
									}
									break
								case 87:
									if tm.curChar == 82 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 86
									}
									break
								case 88:
									if tm.curChar == 79 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 87
									}
									break
								case 89:
									if tm.curChar == 80 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 88
									}
									break
								case 90:
									if tm.curChar == 77 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 89
									}
									break
								case 92:
									if tm.curChar == 83 && kind > 81 {
										kind = 81
									}
									break
								case 93:
									if tm.curChar == 76 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 92
									}
									break
								case 94:
									if tm.curChar == 65 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 93
									}
									break
								case 95:
									if tm.curChar == 82 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 94
									}
									break
								case 96:
									if tm.curChar == 79 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 95
									}
									break
								case 97:
									if tm.curChar == 80 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 96
									}
									break
								case 98:
									if tm.curChar == 77 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 97
									}
									break
								case 99:
									if tm.curChar == 69 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 98
									}
									break
								case 100:
									if tm.curChar == 68 {
										tm.jjAddStates(97, 98)
									}
									break
								case 104:
									if tm.curChar == 70 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 103
									}
									break
								case 105:
									if tm.curChar == 69 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 104
									}
									break
								case 114:
									if tm.curChar == 69 && kind > 35 {
										kind = 35
									}
									break
								case 115:
									if tm.curChar == 76 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 114
									}
									break
								case 116:
									if tm.curChar == 85 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 115
									}
									break
								case 117:
									if tm.curChar == 68 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 116
									}
									break
								case 118:
									if tm.curChar == 79 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 117
									}
									break
								case 119:
									if tm.curChar == 77 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 118
									}
									break
								case 133:
									if tm.curChar == 67 {
										tm.jjAddStates(94, 96)
									}
									break
								case 134:
									if tm.curChar == 84 && kind > 44 {
										kind = 44
									}
									break
								case 135:
									if tm.curChar == 78 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 134
									}
									break
								case 136:
									if tm.curChar == 65 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 135
									}
									break
								case 137:
									if tm.curChar == 84 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 136
									}
									break
								case 138:
									if tm.curChar == 83 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 137
									}
									break
								case 139:
									if tm.curChar == 78 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 138
									}
									break
								case 141:
									if tm.curChar == 83 && kind > 44 {
										kind = 44
									}
									break
								case 142:
									if tm.curChar == 84 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 141
									}
									break
								case 143:
									if tm.curChar == 78 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 142
									}
									break
								case 144:
									if tm.curChar == 65 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 143
									}
									break
								case 145:
									if tm.curChar == 84 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 144
									}
									break
								case 146:
									if tm.curChar == 83 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 145
									}
									break
								case 147:
									if tm.curChar == 78 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 146
									}
									break
								case 148:
									if tm.curChar == 79 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 147
									}
									break
								case 149:
									if tm.curChar == 89 && kind > 58 {
										kind = 58
									}
									break
								case 150:
									if tm.curChar == 82 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 149
									}
									break
								case 151:
									if tm.curChar == 65 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 150
									}
									break
								case 152:
									if tm.curChar == 76 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 151
									}
									break
								case 153:
									if tm.curChar == 76 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 152
									}
									break
								case 154:
									if tm.curChar == 79 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 153
									}
									break
								case 155:
									if tm.curChar == 82 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 154
									}
									break
								case 156:
									if tm.curChar == 79 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 155
									}
									break
								case 166:
									if (0x7fffffe07fffffe & l) != 0 {
										tm.jjAddStates(106, 107)
									}
									break
								case 167:
									if tm.curChar == 92 && kind > 111 {
										kind = 111
									}
									break
								case 170:
									if (0x7fffffe07fffffe & l) != 0 {
										tm.jjAddStates(108, 109)
									}
									break
								case 172, 187, 196:
									if tm.curChar == 92 {
										tm.jjCheckNAdd(171)
									}
									break
								case 174:
									if (0x7fffffe87fffffe & l) != 0 {
										tm.jjAddStates(110, 111)
									}
									break
								case 175:
									if (0x7fffffe07fffffe & l) == 0 {
										break
									}
									if kind > 289 {
										kind = 289
									}
									tm.jjCheckNAdd(176)
									break
								case 176:
									if (0x7fffffe87fffffe & l) == 0 {
										break
									}
									if kind > 289 {
										kind = 289
									}
									tm.jjCheckNAdd(176)
									break
								case 178:
									if (0x7fffffe07fffffe & l) == 0 {
										break
									}
									if kind > 289 {
										kind = 289
									}
									tm.jjCheckNAdd(179)
									break
								case 179:
									if (0x7fffffe87fffffe & l) == 0 {
										break
									}
									if kind > 289 {
										kind = 289
									}
									tm.jjCheckNAdd(179)
									break
								case 180:
									if (0x880000 & l) == 0 {
										break
									}
									if kind > 289 {
										kind = 289
									}
									tm.jjAddStates(80, 82)
									break
								case 182:
									if (0x7fffffe07fffffe & l) != 0 {
										tm.jjCheckNAddTwoStates(182, 184)
									}
									break
								case 185:
									if (0x7fffffe07ffffbe & l) != 0 {
										tm.jjCheckNAddTwoStates(186, 188)
									}
									break
								case 186:
									if (0x7fffffe07fffffe & l) != 0 {
										tm.jjCheckNAddTwoStates(186, 188)
									}
									break
								case 189:
									if (0x7fffffe87ffffbe & l) == 0 {
										break
									}
									if kind > 289 {
										kind = 289
									}
									tm.jjCheckNAdd(190)
									break
								case 190:
									if (0x7fffffe87fffffe & l) == 0 {
										break
									}
									if kind > 289 {
										kind = 289
									}
									tm.jjCheckNAdd(190)
									break
								case 191:
									if (0x7fffffe0777fffe & l) == 0 {
										break
									}
									if kind > 289 {
										kind = 289
									}
									tm.jjCheckNAddStates(75, 79)
									break
								case 192:
									if (0x7fffffe07fffffe & l) != 0 {
										tm.jjCheckNAddTwoStates(192, 194)
									}
									break
								case 195:
									if (0x7fffffe07fffffe & l) != 0 {
										tm.jjCheckNAddTwoStates(195, 197)
									}
									break
								case 198:
									if (0x7fffffe87fffffe & l) == 0 {
										break
									}
									if kind > 289 {
										kind = 289
									}
									tm.jjCheckNAdd(198)
									break
								case 199:
									if tm.curChar == 92 {
										tm.jjAddStates(83, 89)
									}
									break
								case 201:
									if tm.curChar == 115 && kind > 47 {
										kind = 47
									}
									break
								case 202:
									if tm.curChar == 116 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 201
									}
									break
								case 203:
									if tm.curChar == 115 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 202
									}
									break
								case 204:
									if tm.curChar == 105 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 203
									}
									break
								case 205:
									if tm.curChar == 120 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 204
									}
									break
								case 206:
									if tm.curChar == 101 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 205
									}
									break
								case 207:
									if tm.curChar == 65 && kind > 49 {
										kind = 49
									}
									break
								case 208:
									if tm.curChar == 108 && kind > 49 {
										kind = 49
									}
									break
								case 209:
									if tm.curChar == 108 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 208
									}
									break
								case 210:
									if tm.curChar == 97 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 209
									}
									break
								case 211:
									if tm.curChar == 114 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 210
									}
									break
								case 212:
									if tm.curChar == 111 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 211
									}
									break
								case 213:
									if tm.curChar == 102 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 212
									}
									break
								case 214:
									if tm.curChar == 116 && kind > 118 {
										kind = 118
									}
									break
								case 215:
									if tm.curChar == 111 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 214
									}
									break
								case 216:
									if tm.curChar == 110 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 215
									}
									break
								case 217:
									if tm.curChar == 108 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 216
									}
									break
								case 219:
									if tm.curChar == 110 && kind > 147 {
										kind = 147
									}
									break
								case 220:
									if tm.curChar == 105 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 219
									}
									break
								case 224:
									if (0x7fffffe87fffffe & l) == 0 {
										break
									}
									if kind > 290 {
										kind = 290
									}
									tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 224
									break
								case 227:
									if (0x7fffffe07fffffe & l) == 0 {
										break
									}
									if kind > 291 {
										kind = 291
									}
									tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 227
									break
								case 230:
									if (0x7fffffe87fffffe & l) != 0 {
										tm.jjAddStates(112, 113)
									}
									break
								case 244:
									if tm.curChar == 65 {
										tm.jjAddStates(90, 93)
									}
									break
								case 245:
									if tm.curChar == 78 && kind > 38 {
										kind = 38
									}
									break
								case 246:
									if tm.curChar == 79 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 245
									}
									break
								case 247:
									if tm.curChar == 73 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 246
									}
									break
								case 248:
									if tm.curChar == 84 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 247
									}
									break
								case 250:
									if tm.curChar == 83 && kind > 38 {
										kind = 38
									}
									break
								case 251:
									if tm.curChar == 78 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 250
									}
									break
								case 252:
									if tm.curChar == 79 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 251
									}
									break
								case 253:
									if tm.curChar == 73 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 252
									}
									break
								case 254:
									if tm.curChar == 84 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 253
									}
									break
								case 255:
									if tm.curChar == 67 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 254
									}
									break
								case 256:
									if tm.curChar == 78 && kind > 41 {
										kind = 41
									}
									break
								case 257:
									if tm.curChar == 79 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 256
									}
									break
								case 258:
									if tm.curChar == 73 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 257
									}
									break
								case 259:
									if tm.curChar == 84 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 258
									}
									break
								case 264:
									if tm.curChar == 83 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 263
									}
									break
								case 265:
									if tm.curChar == 77 && kind > 41 {
										kind = 41
									}
									break
								case 266:
									if tm.curChar == 79 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 265
									}
									break
								case 267:
									if tm.curChar == 73 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 266
									}
									break
								case 268:
									if tm.curChar == 88 {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 267
									}
									break
								default:
									break
								}
							}
							if !(i != startsAt) {
								break
							}
						}
					}
				} else {
					{
						var hiByte int32 = int32(int32(tm.curChar >> 8))
						_ = hiByte
						var i1 int32 = int32(hiByte >> 6)
						_ = i1
						var l1 uint64 = uint64(uint64(1) << (hiByte & 077))
						_ = l1
						var i2 int32 = int32((tm.curChar & 0xff) >> 6)
						_ = i2
						var l2 uint64 = uint64(uint64(1) << (tm.curChar & 077))
						_ = l2
						for {
							{
								switch tm.jjstateSet[sanyScannerPreDecrement(&i)] {
								case 0:
									if tm.jjCanMove_1(hiByte, i1, i2, l1, l2) {
										{
											if kind > 47 {
												kind = 47
											}
										}
									}
									if tm.jjCanMove_2(hiByte, i1, i2, l1, l2) {
										{
											if kind > 49 {
												kind = 49
											}
										}
									}
									if tm.jjCanMove_3(hiByte, i1, i2, l1, l2) {
										{
											if kind > 90 {
												kind = 90
											}
										}
									}
									if tm.jjCanMove_4(hiByte, i1, i2, l1, l2) {
										{
											if kind > 93 {
												kind = 93
											}
										}
									}
									if tm.jjCanMove_5(hiByte, i1, i2, l1, l2) {
										{
											if kind > 102 {
												kind = 102
											}
										}
									}
									if tm.jjCanMove_6(hiByte, i1, i2, l1, l2) {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 23
									}
									if tm.jjCanMove_6(hiByte, i1, i2, l1, l2) {
										{
											if kind > 104 {
												kind = 104
											}
										}
									}
									if tm.jjCanMove_7(hiByte, i1, i2, l1, l2) {
										{
											if kind > 106 {
												kind = 106
											}
										}
									}
									if tm.jjCanMove_8(hiByte, i1, i2, l1, l2) {
										{
											if kind > 107 {
												kind = 107
											}
										}
									}
									if tm.jjCanMove_9(hiByte, i1, i2, l1, l2) {
										{
											if kind > 108 {
												kind = 108
											}
										}
									}
									if tm.jjCanMove_11(hiByte, i1, i2, l1, l2) {
										{
											if kind > 114 {
												kind = 114
											}
										}
									}
									if tm.jjCanMove_12(hiByte, i1, i2, l1, l2) {
										{
											if kind > 118 {
												kind = 118
											}
										}
									}
									if tm.jjCanMove_13(hiByte, i1, i2, l1, l2) {
										{
											if kind > 121 {
												kind = 121
											}
										}
									}
									if tm.jjCanMove_14(hiByte, i1, i2, l1, l2) {
										{
											if kind > 122 {
												kind = 122
											}
										}
									}
									if tm.jjCanMove_15(hiByte, i1, i2, l1, l2) {
										{
											if kind > 129 {
												kind = 129
											}
										}
									}
									if tm.jjCanMove_16(hiByte, i1, i2, l1, l2) {
										{
											if kind > 132 {
												kind = 132
											}
										}
									}
									if tm.jjCanMove_17(hiByte, i1, i2, l1, l2) {
										{
											if kind > 147 {
												kind = 147
											}
										}
									}
									if tm.jjCanMove_0(hiByte, i1, i2, l1, l2) {
										{
											if kind > 289 {
												kind = 289
											}
										}
									}
									break
								case 1:
									if tm.jjCanMove_2(hiByte, i1, i2, l1, l2) && kind > 49 {
										kind = 49
									}
									break
								case 20:
									if tm.jjCanMove_3(hiByte, i1, i2, l1, l2) && kind > 90 {
										kind = 90
									}
									break
								case 21:
									if tm.jjCanMove_4(hiByte, i1, i2, l1, l2) && kind > 93 {
										kind = 93
									}
									break
								case 22:
									if tm.jjCanMove_5(hiByte, i1, i2, l1, l2) && kind > 102 {
										kind = 102
									}
									break
								case 24:
									if tm.jjCanMove_6(hiByte, i1, i2, l1, l2) {
										tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 23
									}
									break
								case 25:
									if tm.jjCanMove_6(hiByte, i1, i2, l1, l2) && kind > 104 {
										kind = 104
									}
									break
								case 26:
									if tm.jjCanMove_7(hiByte, i1, i2, l1, l2) && kind > 106 {
										kind = 106
									}
									break
								case 27:
									if tm.jjCanMove_8(hiByte, i1, i2, l1, l2) && kind > 107 {
										kind = 107
									}
									break
								case 31:
									if tm.jjCanMove_9(hiByte, i1, i2, l1, l2) && kind > 108 {
										kind = 108
									}
									break
								case 34:
									if tm.jjCanMove_10(hiByte, i1, i2, l1, l2) {
										tm.jjAddStates(63, 65)
									}
									break
								case 40:
									if tm.jjCanMove_11(hiByte, i1, i2, l1, l2) && kind > 114 {
										kind = 114
									}
									break
								case 41:
									if tm.jjCanMove_12(hiByte, i1, i2, l1, l2) && kind > 118 {
										kind = 118
									}
									break
								case 44:
									if tm.jjCanMove_13(hiByte, i1, i2, l1, l2) && kind > 121 {
										kind = 121
									}
									break
								case 45:
									if tm.jjCanMove_14(hiByte, i1, i2, l1, l2) && kind > 122 {
										kind = 122
									}
									break
								case 48:
									if tm.jjCanMove_15(hiByte, i1, i2, l1, l2) && kind > 129 {
										kind = 129
									}
									break
								case 49:
									if tm.jjCanMove_16(hiByte, i1, i2, l1, l2) && kind > 132 {
										kind = 132
									}
									break
								case 50:
									if tm.jjCanMove_17(hiByte, i1, i2, l1, l2) && kind > 147 {
										kind = 147
									}
									break
								case 60:
									if tm.jjCanMove_0(hiByte, i1, i2, l1, l2) && kind > 289 {
										kind = 289
									}
									break
								default:
									break
								}
							}
							if !(i != startsAt) {
								break
							}
						}
					}
				}
			}
			if kind != 0x7fffffff {
				{
					tm.jjmatchedKind = kind
					tm.jjmatchedPos = curPos
					kind = 0x7fffffff
				}
			}
			sanyScannerPreIncrement(&curPos)
			if (sanyScannerSetInt32(&i, tm.jjnewStateCnt)) == (sanyScannerSetInt32(&startsAt, 269-sanyScannerSetInt32(&tm.jjnewStateCnt, startsAt))) {
				return curPos
			}
			if !tm.readScanChar() {
				return curPos
			}
		}
	}
}

func (tm *SanyTokenManager) jjStopStringLiteralDfa_3(pos int32, active0 uint64) int32 {
	switch pos {
	default:
		return -1
	}
}

func (tm *SanyTokenManager) jjStartNfa_3(pos int32, active0 uint64) int32 {
	return tm.jjMoveNfa_3(tm.jjStopStringLiteralDfa_3(pos, active0), pos+1)
}

func (tm *SanyTokenManager) jjStartNfaWithStates_3(pos int32, kind int32, state int32) int32 {
	tm.jjmatchedKind = kind
	tm.jjmatchedPos = pos
	if !tm.readScanChar() {
		return pos + 1
	}
	return tm.jjMoveNfa_3(state, pos+1)
}

func (tm *SanyTokenManager) jjMoveStringLiteralDfa0_3() int32 {
	switch tm.curChar {
	case 42:
		return tm.jjMoveStringLiteralDfa1_3(0x40000000)
	default:
		return tm.jjMoveNfa_3(0, 0)
	}
}

func (tm *SanyTokenManager) jjMoveStringLiteralDfa1_3(active0 uint64) int32 {
	if !tm.readScanChar() {
		tm.jjStopStringLiteralDfa_3(0, active0)
		return 1
	}
	switch tm.curChar {
	case 41:
		if (active0 & 0x40000000) != 0 {
			return tm.jjStopAtPos(1, 30)
		}
		break
	default:
		break
	}
	return tm.jjStartNfa_3(0, active0)
}

func (tm *SanyTokenManager) jjMoveNfa_3(startState int32, curPos int32) int32 {

	var startsAt int32 = int32(0)
	_ = startsAt
	tm.jjnewStateCnt = 4
	var i int32 = int32(1)
	_ = i
	tm.jjstateSet[0] = startState
	var j int32
	_ = j
	var kind int32 = int32(0x7fffffff)
	_ = kind
	for {
		{
			if sanyScannerPreIncrement(&tm.jjround) == 0x7fffffff {
				tm.reInitRounds()
			}
			if tm.curChar < 64 {
				{
					var l uint64 = uint64(uint64(1) << tm.curChar)
					_ = l
					for {
						{
							switch tm.jjstateSet[sanyScannerPreDecrement(&i)] {
							case 0:
								if tm.curChar == 40 {
									tm.jjAddStates(3, 4)
								}
								break
							case 1:
								if tm.curChar == 46 {
									kind = 29
								}
								break
							case 2:
								if tm.curChar == 42 {
									tm.jjstateSet[sanyScannerPostIncrement(&tm.jjnewStateCnt)] = 1
								}
								break
							case 3:
								if tm.curChar == 42 && kind > 29 {
									kind = 29
								}
								break
							default:
								break
							}
						}
						if !(i != startsAt) {
							break
						}
					}
				}
			} else {
				if tm.curChar < 128 {
					{
						var l uint64 = uint64(uint64(1) << (tm.curChar & 077))
						_ = l
						for {
							{
								switch tm.jjstateSet[sanyScannerPreDecrement(&i)] {
								default:
									break
								}
							}
							if !(i != startsAt) {
								break
							}
						}
					}
				} else {
					{
						var hiByte int32 = int32(int32(tm.curChar >> 8))
						_ = hiByte
						var i1 int32 = int32(hiByte >> 6)
						_ = i1
						var l1 uint64 = uint64(uint64(1) << (hiByte & 077))
						_ = l1
						var i2 int32 = int32((tm.curChar & 0xff) >> 6)
						_ = i2
						var l2 uint64 = uint64(uint64(1) << (tm.curChar & 077))
						_ = l2
						for {
							{
								switch tm.jjstateSet[sanyScannerPreDecrement(&i)] {
								default:
									break
								}
							}
							if !(i != startsAt) {
								break
							}
						}
					}
				}
			}
			if kind != 0x7fffffff {
				{
					tm.jjmatchedKind = kind
					tm.jjmatchedPos = curPos
					kind = 0x7fffffff
				}
			}
			sanyScannerPreIncrement(&curPos)
			if (sanyScannerSetInt32(&i, tm.jjnewStateCnt)) == (sanyScannerSetInt32(&startsAt, 4-sanyScannerSetInt32(&tm.jjnewStateCnt, startsAt))) {
				return curPos
			}
			if !tm.readScanChar() {
				return curPos
			}
		}
	}
}

func (tm *SanyTokenManager) jjCanMove_0(hiByte int32, i1 int32, i2 int32, l1 uint64, l2 uint64) bool {
	switch hiByte {
	case 33:
		return ((sanyScannerBitVec0[i2] & l2) != 0)
	default:
		return false
	}
}

func (tm *SanyTokenManager) jjCanMove_1(hiByte int32, i1 int32, i2 int32, l1 uint64, l2 uint64) bool {
	switch hiByte {
	case 34:
		return ((sanyScannerBitVec1[i2] & l2) != 0)
	default:
		return false
	}
}

func (tm *SanyTokenManager) jjCanMove_2(hiByte int32, i1 int32, i2 int32, l1 uint64, l2 uint64) bool {
	switch hiByte {
	case 34:
		return ((sanyScannerBitVec2[i2] & l2) != 0)
	default:
		return false
	}
}

func (tm *SanyTokenManager) jjCanMove_3(hiByte int32, i1 int32, i2 int32, l1 uint64, l2 uint64) bool {
	switch hiByte {
	case 34:
		return ((sanyScannerBitVec3[i2] & l2) != 0)
	default:
		return false
	}
}

func (tm *SanyTokenManager) jjCanMove_4(hiByte int32, i1 int32, i2 int32, l1 uint64, l2 uint64) bool {
	switch hiByte {
	case 34:
		return ((sanyScannerBitVec4[i2] & l2) != 0)
	default:
		return false
	}
}

func (tm *SanyTokenManager) jjCanMove_5(hiByte int32, i1 int32, i2 int32, l1 uint64, l2 uint64) bool {
	switch hiByte {
	case 39:
		return ((sanyScannerBitVec5[i2] & l2) != 0)
	default:
		return false
	}
}

func (tm *SanyTokenManager) jjCanMove_6(hiByte int32, i1 int32, i2 int32, l1 uint64, l2 uint64) bool {
	switch hiByte {
	case 39:
		return ((sanyScannerBitVec6[i2] & l2) != 0)
	default:
		return false
	}
}

func (tm *SanyTokenManager) jjCanMove_7(hiByte int32, i1 int32, i2 int32, l1 uint64, l2 uint64) bool {
	switch hiByte {
	case 33:
		return ((sanyScannerBitVec7[i2] & l2) != 0)
	default:
		return false
	}
}

func (tm *SanyTokenManager) jjCanMove_8(hiByte int32, i1 int32, i2 int32, l1 uint64, l2 uint64) bool {
	switch hiByte {
	case 33:
		return ((sanyScannerBitVec8[i2] & l2) != 0)
	default:
		return false
	}
}

func (tm *SanyTokenManager) jjCanMove_9(hiByte int32, i1 int32, i2 int32, l1 uint64, l2 uint64) bool {
	switch hiByte {
	case 33:
		return ((sanyScannerBitVec9[i2] & l2) != 0)
	default:
		return false
	}
}

func (tm *SanyTokenManager) jjCanMove_10(hiByte int32, i1 int32, i2 int32, l1 uint64, l2 uint64) bool {
	switch hiByte {
	case 0:
		return ((sanyScannerBitVec12[i2] & l2) != 0)
	default:
		if (sanyScannerBitVec10[i1] & l1) != 0 {
			return true
		}
		return false
	}
}

func (tm *SanyTokenManager) jjCanMove_11(hiByte int32, i1 int32, i2 int32, l1 uint64, l2 uint64) bool {
	switch hiByte {
	case 32:
		return ((sanyScannerBitVec13[i2] & l2) != 0)
	default:
		return false
	}
}

func (tm *SanyTokenManager) jjCanMove_12(hiByte int32, i1 int32, i2 int32, l1 uint64, l2 uint64) bool {
	switch hiByte {
	case 0:
		return ((sanyScannerBitVec14[i2] & l2) != 0)
	default:
		return false
	}
}

func (tm *SanyTokenManager) jjCanMove_13(hiByte int32, i1 int32, i2 int32, l1 uint64, l2 uint64) bool {
	switch hiByte {
	case 37:
		return ((sanyScannerBitVec15[i2] & l2) != 0)
	default:
		return false
	}
}

func (tm *SanyTokenManager) jjCanMove_14(hiByte int32, i1 int32, i2 int32, l1 uint64, l2 uint64) bool {
	switch hiByte {
	case 37:
		return ((sanyScannerBitVec16[i2] & l2) != 0)
	default:
		return false
	}
}

func (tm *SanyTokenManager) jjCanMove_15(hiByte int32, i1 int32, i2 int32, l1 uint64, l2 uint64) bool {
	switch hiByte {
	case 34:
		return ((sanyScannerBitVec17[i2] & l2) != 0)
	default:
		return false
	}
}

func (tm *SanyTokenManager) jjCanMove_16(hiByte int32, i1 int32, i2 int32, l1 uint64, l2 uint64) bool {
	switch hiByte {
	case 34:
		return ((sanyScannerBitVec18[i2] & l2) != 0)
	default:
		return false
	}
}

func (tm *SanyTokenManager) jjCanMove_17(hiByte int32, i1 int32, i2 int32, l1 uint64, l2 uint64) bool {
	switch hiByte {
	case 34:
		return ((sanyScannerBitVec19[i2] & l2) != 0)
	default:
		return false
	}
}

var sanyScannerBitVec0 = []uint64{0x1020200000, 0x0, 0x0, 0x0}

var sanyScannerBitVec1 = []uint64{0x8, 0x0, 0x0, 0x0}

var sanyScannerBitVec2 = []uint64{0x1, 0x0, 0x0, 0x0}

var sanyScannerBitVec3 = []uint64{0x80000000000000, 0x0, 0x0, 0x0}

var sanyScannerBitVec4 = []uint64{0x0, 0x10000000, 0x0, 0x0}

var sanyScannerBitVec5 = []uint64{0x0, 0x0, 0x0, 0x10000000000}

var sanyScannerBitVec6 = []uint64{0x0, 0x0, 0x0, 0x20000000000}

var sanyScannerBitVec7 = []uint64{0x0, 0x0, 0x40000, 0x0}

var sanyScannerBitVec8 = []uint64{0x0, 0x0, 0x10000, 0x0}

var sanyScannerBitVec9 = []uint64{0x0, 0x0, 0x4000000000, 0x0}

var sanyScannerBitVec10 = []uint64{0xfffffffffffffffe, 0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff}

var sanyScannerBitVec12 = []uint64{0x0, 0x0, 0xffffffffffffffff, 0xffffffffffffffff}

var sanyScannerBitVec13 = []uint64{0x0, 0x400000000000000, 0x0, 0x0}

var sanyScannerBitVec14 = []uint64{0x0, 0x0, 0x100000000000, 0x0}

var sanyScannerBitVec15 = []uint64{0x0, 0x0, 0x200000000, 0x0}

var sanyScannerBitVec16 = []uint64{0x0, 0x0, 0x0, 0x80}

var sanyScannerBitVec17 = []uint64{0x8000000000, 0x0, 0x0, 0x0}

var sanyScannerBitVec18 = []uint64{0x10000000000, 0x0, 0x0, 0x0}

var sanyScannerBitVec19 = []uint64{0x100, 0x0, 0x0, 0x0}

var sanyScannerNextStates = []int32{1, 2, 8, 2, 3, 24, 25, 28, 29, 32, 33, 34, 36, 37, 2, 3,
	9, 14, 15, 51, 53, 54, 56, 51, 53, 54, 56, 57, 40, 44, 48, 25,
	28, 29, 32, 33, 34, 165, 166, 169, 170, 173, 174, 175, 177, 178, 222, 225,
	228, 232, 233, 235, 236, 243, 162, 163, 131, 132, 121, 125, 126, 108, 109, 34,
	35, 37, 158, 159, 160, 112, 113, 119, 238, 240, 242, 192, 194, 195, 197, 198,
	181, 185, 189, 200, 206, 207, 213, 217, 218, 220, 249, 255, 264, 268, 140, 148,
	156, 102, 105, 91, 99, 75, 83, 62, 64, 66, 166, 169, 170, 173, 174, 175,
	230, 231,
}

var sanyScannerNewLexState = []int32{-1, -1, 1, 2, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, 2, -1, -1, -1,
	-1, -1, 3, 5, 4, 2, -1, -1, 2, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
	-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
	-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
	-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
	-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
	-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
	-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
	-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
	-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
	-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
	-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
}

var sanyScannerToToken = []uint64{0xfffffff80030000d, 0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff,
	0x7fffffffff,
}

var sanyScannerToSkip = []uint64{0x3e7c00010, 0x0, 0x0, 0x0,
	0x0,
}

var sanyScannerToSpecial = []uint64{0x3e0000000, 0x0, 0x0, 0x0,
	0x0,
}

var sanyScannerToMore = []uint64{0x418000000, 0x0, 0x0, 0x0,
	0x0,
}

func sanyScannerLiteral(s string) *string { return &s }

var sanyScannerLiteralImages = []*string{sanyScannerLiteral(""), nil, sanyScannerLiteral("--->"), nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, sanyScannerLiteral("ASSUME"), sanyScannerLiteral("[]ASSUME"), nil, sanyScannerLiteral("CASE"), sanyScannerLiteral("CHOOSE"), nil, sanyScannerLiteral("ELSE"), sanyScannerLiteral("EXCEPT"), nil, sanyScannerLiteral("EXTENDS"), nil, sanyScannerLiteral("IF"), sanyScannerLiteral("INSTANCE"), sanyScannerLiteral("LET"), sanyScannerLiteral("IN"), sanyScannerLiteral("LOCAL"), sanyScannerLiteral("MODULE"), sanyScannerLiteral("NEW"), sanyScannerLiteral("OTHER"), nil, sanyScannerLiteral("SF_"), sanyScannerLiteral("\\EE"), sanyScannerLiteral("\\AA"), sanyScannerLiteral("THEN"), sanyScannerLiteral("BY"), sanyScannerLiteral("ONLY"), sanyScannerLiteral("DEFINE"), nil, sanyScannerLiteral("THEOREM"), sanyScannerLiteral("USE"), sanyScannerLiteral("HIDE"), sanyScannerLiteral("HAVE"), sanyScannerLiteral("OBVIOUS"), sanyScannerLiteral("OMITTED"), sanyScannerLiteral("LAMBDA"), sanyScannerLiteral("TAKE"), sanyScannerLiteral("PROOF"), sanyScannerLiteral("PROVE"), sanyScannerLiteral("[]PROVE"), sanyScannerLiteral("QED"), sanyScannerLiteral("RECURSIVE"), sanyScannerLiteral("STATE"), nil, sanyScannerLiteral("PICK"), sanyScannerLiteral("WITNESS"), sanyScannerLiteral("SUFFICES"), nil, sanyScannerLiteral("WF_"), sanyScannerLiteral("WITH"), sanyScannerLiteral(","), sanyScannerLiteral(":"), nil, sanyScannerLiteral("."), sanyScannerLiteral("_"), nil, sanyScannerLiteral("("), sanyScannerLiteral(")"), sanyScannerLiteral("-|-"), sanyScannerLiteral("["), sanyScannerLiteral("]_"), sanyScannerLiteral("]"), sanyScannerLiteral("{"), sanyScannerLiteral("}"), nil, nil, nil, sanyScannerLiteral("!"), nil, nil, nil, nil, nil, nil, nil, sanyScannerLiteral("'"), nil, sanyScannerLiteral("^*"), sanyScannerLiteral("^#"), sanyScannerLiteral("-."), nil, sanyScannerLiteral("\\neg"), sanyScannerLiteral("~"), nil, nil, sanyScannerLiteral("ENABLED"), sanyScannerLiteral("UNCHANGED"), sanyScannerLiteral("SUBSET"), sanyScannerLiteral("UNION"), sanyScannerLiteral("DOMAIN"), sanyScannerLiteral("//"), nil, sanyScannerLiteral("/="), sanyScannerLiteral("/"), nil, sanyScannerLiteral("\\approx"), sanyScannerLiteral("\\asymp"), sanyScannerLiteral("\\bigcirc"), sanyScannerLiteral("\\bullet"), sanyScannerLiteral("\\cap"), sanyScannerLiteral("\\cdot"), sanyScannerLiteral("\\circ"), sanyScannerLiteral("\\cong"), sanyScannerLiteral("\\cup"), sanyScannerLiteral("\\div"), sanyScannerLiteral("\\doteq"), sanyScannerLiteral("\\equiv"), sanyScannerLiteral("\\geq"), sanyScannerLiteral("\\gg"), nil, sanyScannerLiteral("\\intersect"), sanyScannerLiteral("\\union"), sanyScannerLiteral("\\land"), sanyScannerLiteral("\\leq"), sanyScannerLiteral("\\ll"), sanyScannerLiteral("\\lor"), sanyScannerLiteral("\\o"), sanyScannerLiteral("\\odot"), sanyScannerLiteral("\\ominus"), sanyScannerLiteral("\\oplus"), sanyScannerLiteral("\\oslash"), sanyScannerLiteral("\\otimes"), sanyScannerLiteral("\\prec"), sanyScannerLiteral("\\preceq"), sanyScannerLiteral("\\propto"), sanyScannerLiteral("\\sim"), sanyScannerLiteral("\\simeq"), sanyScannerLiteral("\\sqcap"), sanyScannerLiteral("\\sqcup"), sanyScannerLiteral("\\sqsubset"), sanyScannerLiteral("\\sqsupset"), sanyScannerLiteral("\\sqsubseteq"), sanyScannerLiteral("\\sqsupseteq"), sanyScannerLiteral("\\star"), sanyScannerLiteral("\\subset"), sanyScannerLiteral("\\subseteq"), sanyScannerLiteral("\\succ"), sanyScannerLiteral("\\succeq"), sanyScannerLiteral("\\supset"), sanyScannerLiteral("\\supseteq"), sanyScannerLiteral("\\uplus"), sanyScannerLiteral("\\wr"), sanyScannerLiteral("\\"), sanyScannerLiteral("~>"), sanyScannerLiteral("=>"), sanyScannerLiteral("=<"), sanyScannerLiteral("=|"), sanyScannerLiteral("="), sanyScannerLiteral("##"), sanyScannerLiteral("#"), sanyScannerLiteral("^^"), sanyScannerLiteral("^"), sanyScannerLiteral("--"), sanyScannerLiteral("-|"), sanyScannerLiteral("-+->"), sanyScannerLiteral("-"), sanyScannerLiteral("**"), sanyScannerLiteral("*"), sanyScannerLiteral("++"), sanyScannerLiteral("+"), sanyScannerLiteral("<=>"), sanyScannerLiteral("<:"), sanyScannerLiteral("<="), sanyScannerLiteral("<"), sanyScannerLiteral(">="), sanyScannerLiteral(">"), sanyScannerLiteral("..."), sanyScannerLiteral(".."), sanyScannerLiteral("||"), sanyScannerLiteral("|"), sanyScannerLiteral("|-"), sanyScannerLiteral("|="), sanyScannerLiteral("&&"), sanyScannerLiteral("&"), sanyScannerLiteral("$$"), sanyScannerLiteral("$"), sanyScannerLiteral("??"), sanyScannerLiteral("%%"), sanyScannerLiteral("%"), sanyScannerLiteral("@@"), sanyScannerLiteral("!!"), sanyScannerLiteral(":>"), sanyScannerLiteral(":="), sanyScannerLiteral("::="), sanyScannerLiteral("(+)"), sanyScannerLiteral("(-)"), sanyScannerLiteral("(.)"), sanyScannerLiteral("(/)"), sanyScannerLiteral("(\\X)"), sanyScannerLiteral("\\notin"), sanyScannerLiteral("≈"), sanyScannerLiteral("≔"), sanyScannerLiteral("≍"), sanyScannerLiteral("◯"), sanyScannerLiteral("⩴"), sanyScannerLiteral("●"), sanyScannerLiteral("∩"), sanyScannerLiteral("⋅"), sanyScannerLiteral("∘"), sanyScannerLiteral("≅"), sanyScannerLiteral("∪"), sanyScannerLiteral("÷"), sanyScannerLiteral("≐"), sanyScannerLiteral("‥"), sanyScannerLiteral("…"), sanyScannerLiteral("≡"), sanyScannerLiteral("‼"), sanyScannerLiteral("≥"), sanyScannerLiteral("≫"), sanyScannerLiteral("⇔"), sanyScannerLiteral("⇒"), sanyScannerLiteral("⫤"), sanyScannerLiteral("↝"), sanyScannerLiteral("≤"), sanyScannerLiteral("≪"), sanyScannerLiteral("⊣"), sanyScannerLiteral("≠"), sanyScannerLiteral("∉"), sanyScannerLiteral("⊙"), sanyScannerLiteral("⊖"), sanyScannerLiteral("⊕"), sanyScannerLiteral("⊘"), sanyScannerLiteral("⊗"), sanyScannerLiteral("⇸"), sanyScannerLiteral("≺"), sanyScannerLiteral("⪯"), sanyScannerLiteral("∝"), sanyScannerLiteral("⁇"), sanyScannerLiteral("⊨"), sanyScannerLiteral("⊢"), sanyScannerLiteral("∼"), sanyScannerLiteral("≃"), sanyScannerLiteral("⊓"), sanyScannerLiteral("⊔"), sanyScannerLiteral("⊏"), sanyScannerLiteral("⊑"), sanyScannerLiteral("⊐"), sanyScannerLiteral("⊒"), sanyScannerLiteral("⋆"), sanyScannerLiteral("⊂"), sanyScannerLiteral("⊆"), sanyScannerLiteral("≻"), sanyScannerLiteral("⪰"), sanyScannerLiteral("⊃"), sanyScannerLiteral("⊇"), sanyScannerLiteral("×"), sanyScannerLiteral("⊎"), sanyScannerLiteral("‖"), sanyScannerLiteral("≀"), sanyScannerLiteral("\\times"), sanyScannerLiteral("\\X"), nil, nil, nil, nil, nil, nil}
