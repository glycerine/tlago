package tlc

const (
	OpcodeBC = 1 + iota
	OpcodeBE
	OpcodeBF
	OpcodeCase
	OpcodeCP
	OpcodeCL
	OpcodeDL
	OpcodeExc
	OpcodeFA
	OpcodeFC
	OpcodeITE
	OpcodeNRFS
	OpcodePair
	OpcodeRC
	OpcodeRS
	OpcodeRFS
	OpcodeSeq
	OpcodeSE
	OpcodeSOA
	OpcodeSOR
	OpcodeSOF
	OpcodeSSO
	OpcodeTup
	OpcodeUC
	OpcodeUE
	OpcodeUF
)

const (
	OpcodeLnot    = OpcodeUF + 1
	OpcodeNeg     = OpcodeUF + 2
	OpcodeSubset  = OpcodeUF + 3
	OpcodeUnion   = OpcodeUF + 4
	OpcodeDomain  = OpcodeUF + 5
	OpcodeEnabled = OpcodeUF + 8

	OpcodeEq       = OpcodeEnabled + 1
	OpcodeLand     = OpcodeEnabled + 2
	OpcodeLor      = OpcodeEnabled + 3
	OpcodeImplies  = OpcodeEnabled + 4
	OpcodeEquiv    = OpcodeEnabled + 5
	OpcodeNoteq    = OpcodeEnabled + 6
	OpcodeSubseteq = OpcodeEnabled + 7
	OpcodeIn       = OpcodeEnabled + 8
	OpcodeNotin    = OpcodeEnabled + 9
	OpcodeSetdiff  = OpcodeEnabled + 10
	OpcodeCap      = OpcodeEnabled + 11
	OpcodeNop      = OpcodeEnabled + 12
	OpcodeCup      = OpcodeEnabled + 13

	OpcodePrime     = OpcodeCup + 1
	OpcodeUnchanged = OpcodeCup + 2
	OpcodeAA        = OpcodeCup + 3
	OpcodeSA        = OpcodeCup + 4
	OpcodeCdot      = OpcodeCup + 5

	OpcodeSF      = OpcodeCdot + 1
	OpcodeWF      = OpcodeCdot + 2
	OpcodeTE      = OpcodeCdot + 3
	OpcodeTF      = OpcodeCdot + 4
	OpcodeLeadsto = OpcodeCdot + 5
	OpcodeArrow   = OpcodeCdot + 6
	OpcodeBox     = OpcodeCdot + 7
	OpcodeDiamond = OpcodeCdot + 8
)

var (
	OpLnot      *UniqueString
	OpSubset    *UniqueString
	OpUnion     *UniqueString
	OpDomain    *UniqueString
	OpBox       *UniqueString
	OpDiamond   *UniqueString
	OpEnabled   *UniqueString
	OpUnchanged *UniqueString

	OpEq       *UniqueString
	OpLand     *UniqueString
	OpLor      *UniqueString
	OpImplies  *UniqueString
	OpCdot     *UniqueString
	OpEquiv    *UniqueString
	OpLeadsto  *UniqueString
	OpArrow    *UniqueString
	OpNoteq    *UniqueString
	OpSubseteq *UniqueString
	OpIn       *UniqueString
	OpNotin    *UniqueString
	OpSetdiff  *UniqueString
	OpCap      *UniqueString
	OpCup      *UniqueString

	OpDotdot *UniqueString
	OpPlus   *UniqueString
	OpMinus  *UniqueString
	OpTimes  *UniqueString
	OpLT     *UniqueString
	OpLEQ    *UniqueString
	OpGT     *UniqueString
	OpGEQ    *UniqueString

	OpPrime *UniqueString

	OpAA   *UniqueString
	OpBC   *UniqueString
	OpBE   *UniqueString
	OpBF   *UniqueString
	OpCase *UniqueString
	OpCP   *UniqueString
	OpCL   *UniqueString
	OpDL   *UniqueString
	OpExc  *UniqueString
	OpFA   *UniqueString
	OpFC   *UniqueString
	OpITE  *UniqueString
	OpNRFS *UniqueString
	OpPair *UniqueString
	OpRC   *UniqueString
	OpRS   *UniqueString
	OpRFS  *UniqueString
	OpSeq  *UniqueString
	OpSA   *UniqueString
	OpSE   *UniqueString
	OpSF   *UniqueString
	OpSOA  *UniqueString
	OpSOR  *UniqueString
	OpSOF  *UniqueString
	OpSSO  *UniqueString
	OpTup  *UniqueString
	OpTE   *UniqueString
	OpTF   *UniqueString
	OpUC   *UniqueString
	OpUE   *UniqueString
	OpUF   *UniqueString
	OpWF   *UniqueString
	OpNop  *UniqueString

	opCodeTable []int
)

func init() {
	initBuiltInOPs()
}

func initBuiltInOPs() {
	OpLnot = UniqueStringOf("\\lnot")
	OpSubset = UniqueStringOf("SUBSET")
	OpUnion = UniqueStringOf("UNION")
	OpDomain = UniqueStringOf("DOMAIN")
	OpBox = UniqueStringOf("[]")
	OpDiamond = UniqueStringOf("<>")
	OpEnabled = UniqueStringOf("ENABLED")
	OpUnchanged = UniqueStringOf("UNCHANGED")

	OpEq = UniqueStringOf("=")
	OpLand = UniqueStringOf("\\land")
	OpLor = UniqueStringOf("\\lor")
	OpImplies = UniqueStringOf("=>")
	OpCdot = UniqueStringOf("\\cdot")
	OpEquiv = UniqueStringOf("\\equiv")
	OpLeadsto = UniqueStringOf("~>")
	OpArrow = UniqueStringOf("-+->")
	OpNoteq = UniqueStringOf("/=")
	OpSubseteq = UniqueStringOf("\\subseteq")
	OpIn = UniqueStringOf("\\in")
	OpNotin = UniqueStringOf("\\notin")
	OpSetdiff = UniqueStringOf("\\")
	OpCap = UniqueStringOf("\\intersect")
	OpCup = UniqueStringOf("\\union")

	OpDotdot = UniqueStringOf("..")
	OpPlus = UniqueStringOf("+")
	OpMinus = UniqueStringOf("-")
	OpTimes = UniqueStringOf("*")
	OpLT = UniqueStringOf("<")
	OpLEQ = UniqueStringOf("\\leq")
	OpGT = UniqueStringOf(">")
	OpGEQ = UniqueStringOf("\\geq")
	OpPrime = UniqueStringOf("'")

	OpAA = UniqueStringOf("$AngleAct")
	OpBC = UniqueStringOf("$BoundedChoose")
	OpBE = UniqueStringOf("$BoundedExists")
	OpBF = UniqueStringOf("$BoundedForall")
	OpCase = UniqueStringOf("$Case")
	OpCP = UniqueStringOf("$CartesianProd")
	OpCL = UniqueStringOf("$ConjList")
	OpDL = UniqueStringOf("$DisjList")
	OpExc = UniqueStringOf("$Except")
	OpFA = UniqueStringOf("$FcnApply")
	OpFC = UniqueStringOf("$FcnConstructor")
	OpITE = UniqueStringOf("$IfThenElse")
	OpNRFS = UniqueStringOf("$NonRecursiveFcnSpec")
	OpPair = UniqueStringOf("$Pair")
	OpRC = UniqueStringOf("$RcdConstructor")
	OpRS = UniqueStringOf("$RcdSelect")
	OpRFS = UniqueStringOf("$RecursiveFcnSpec")
	OpSeq = UniqueStringOf("$Seq")
	OpSA = UniqueStringOf("$SquareAct")
	OpSE = UniqueStringOf("$SetEnumerate")
	OpSF = UniqueStringOf("$SF")
	OpSOA = UniqueStringOf("$SetOfAll")
	OpSOR = UniqueStringOf("$SetOfRcds")
	OpSOF = UniqueStringOf("$SetOfFcns")
	OpSSO = UniqueStringOf("$SubsetOf")
	OpTup = UniqueStringOf("$Tuple")
	OpTE = UniqueStringOf("$TemporalExists")
	OpTF = UniqueStringOf("$TemporalForall")
	OpUC = UniqueStringOf("$UnboundedChoose")
	OpUE = UniqueStringOf("$UnboundedExists")
	OpUF = UniqueStringOf("$UnboundedForall")
	OpWF = UniqueStringOf("$WF")
	OpNop = UniqueStringOf("$Nop")

	opCodeTable = make([]int, 200)
	putOpCode(OpAA, OpcodeAA)
	putOpCode(OpBC, OpcodeBC)
	putOpCode(OpBE, OpcodeBE)
	putOpCode(OpBF, OpcodeBF)
	putOpCode(OpCase, OpcodeCase)
	putOpCode(OpCP, OpcodeCP)
	putOpCode(OpCL, OpcodeCL)
	putOpCode(OpDL, OpcodeDL)
	putOpCode(OpExc, OpcodeExc)
	putOpCode(OpFA, OpcodeFA)
	putOpCode(OpFC, OpcodeFC)
	putOpCode(OpITE, OpcodeITE)
	putOpCode(OpNRFS, OpcodeNRFS)
	putOpCode(OpPair, OpcodePair)
	putOpCode(OpRC, OpcodeRC)
	putOpCode(OpRS, OpcodeRS)
	putOpCode(OpRFS, OpcodeRFS)
	putOpCode(OpSeq, OpcodeSeq)
	putOpCode(OpSA, OpcodeSA)
	putOpCode(OpSE, OpcodeSE)
	putOpCode(OpSF, OpcodeSF)
	putOpCode(OpSOA, OpcodeSOA)
	putOpCode(OpSOR, OpcodeSOR)
	putOpCode(OpSOF, OpcodeSOF)
	putOpCode(OpSSO, OpcodeSSO)
	putOpCode(OpTup, OpcodeTup)
	putOpCode(OpTE, OpcodeTE)
	putOpCode(OpTF, OpcodeTF)
	putOpCode(OpUC, OpcodeUC)
	putOpCode(OpUE, OpcodeUE)
	putOpCode(OpUF, OpcodeUF)
	putOpCode(OpWF, OpcodeWF)

	putOpCode(OpLnot, OpcodeLnot)
	putOpCode(OpSubset, OpcodeSubset)
	putOpCode(OpUnion, OpcodeUnion)
	putOpCode(OpDomain, OpcodeDomain)
	putOpCode(OpBox, OpcodeBox)
	putOpCode(OpDiamond, OpcodeDiamond)
	putOpCode(OpEnabled, OpcodeEnabled)
	putOpCode(OpUnchanged, OpcodeUnchanged)

	putOpCode(OpEq, OpcodeEq)
	putOpCode(OpLand, OpcodeLand)
	putOpCode(OpLor, OpcodeLor)
	putOpCode(OpImplies, OpcodeImplies)
	putOpCode(OpCdot, OpcodeCdot)
	putOpCode(OpEquiv, OpcodeEquiv)
	putOpCode(OpLeadsto, OpcodeLeadsto)
	putOpCode(OpArrow, OpcodeArrow)
	putOpCode(OpNoteq, OpcodeNoteq)
	putOpCode(OpSubseteq, OpcodeSubseteq)
	putOpCode(OpIn, OpcodeIn)
	putOpCode(OpNotin, OpcodeNotin)
	putOpCode(OpSetdiff, OpcodeSetdiff)
	putOpCode(OpCap, OpcodeCap)
	putOpCode(OpNop, OpcodeNop)
	putOpCode(OpCup, OpcodeCup)
	putOpCode(OpPrime, OpcodePrime)
}

func putOpCode(op *UniqueString, opcode int) {
	loc := op.Token()
	if loc >= len(opCodeTable) {
		next := make([]int, loc+20)
		copy(next, opCodeTable)
		opCodeTable = next
	}
	opCodeTable[loc] = opcode
}

func GetOpCode(op *UniqueString) int {
	if op == nil {
		return 0
	}
	return GetOpCodeByToken(op.Token())
}

func GetOpCodeByToken(loc int) int {
	if loc >= 0 && loc < len(opCodeTable) {
		return opCodeTable[loc]
	}
	return 0
}

func IsTemporalOpcode(opcode int) bool {
	return OpcodeSF <= opcode && opcode <= OpcodeDiamond
}

func IsActionOpcode(opcode int) bool {
	return OpcodePrime <= opcode && opcode <= OpcodeCdot
}
