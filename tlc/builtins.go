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

var builtInClassInterning classInterning

func init() { resetBuiltInOPNames() }

// Unregistered names preserve the exported operator constants until first use.
func resetBuiltInOPNames() {
	OpLnot = &UniqueString{s: "\\lnot", tok: -1, loc: -1, unregistered: true}
	OpSubset = &UniqueString{s: "SUBSET", tok: -1, loc: -1, unregistered: true}
	OpUnion = &UniqueString{s: "UNION", tok: -1, loc: -1, unregistered: true}
	OpDomain = &UniqueString{s: "DOMAIN", tok: -1, loc: -1, unregistered: true}
	OpBox = &UniqueString{s: "[]", tok: -1, loc: -1, unregistered: true}
	OpDiamond = &UniqueString{s: "<>", tok: -1, loc: -1, unregistered: true}
	OpEnabled = &UniqueString{s: "ENABLED", tok: -1, loc: -1, unregistered: true}
	OpUnchanged = &UniqueString{s: "UNCHANGED", tok: -1, loc: -1, unregistered: true}
	OpEq = &UniqueString{s: "=", tok: -1, loc: -1, unregistered: true}
	OpLand = &UniqueString{s: "\\land", tok: -1, loc: -1, unregistered: true}
	OpLor = &UniqueString{s: "\\lor", tok: -1, loc: -1, unregistered: true}
	OpImplies = &UniqueString{s: "=>", tok: -1, loc: -1, unregistered: true}
	OpCdot = &UniqueString{s: "\\cdot", tok: -1, loc: -1, unregistered: true}
	OpEquiv = &UniqueString{s: "\\equiv", tok: -1, loc: -1, unregistered: true}
	OpLeadsto = &UniqueString{s: "~>", tok: -1, loc: -1, unregistered: true}
	OpArrow = &UniqueString{s: "-+->", tok: -1, loc: -1, unregistered: true}
	OpNoteq = &UniqueString{s: "/=", tok: -1, loc: -1, unregistered: true}
	OpSubseteq = &UniqueString{s: "\\subseteq", tok: -1, loc: -1, unregistered: true}
	OpIn = &UniqueString{s: "\\in", tok: -1, loc: -1, unregistered: true}
	OpNotin = &UniqueString{s: "\\notin", tok: -1, loc: -1, unregistered: true}
	OpSetdiff = &UniqueString{s: "\\", tok: -1, loc: -1, unregistered: true}
	OpCap = &UniqueString{s: "\\intersect", tok: -1, loc: -1, unregistered: true}
	OpCup = &UniqueString{s: "\\union", tok: -1, loc: -1, unregistered: true}
	OpDotdot = &UniqueString{s: "..", tok: -1, loc: -1, unregistered: true}
	OpPlus = &UniqueString{s: "+", tok: -1, loc: -1, unregistered: true}
	OpMinus = &UniqueString{s: "-", tok: -1, loc: -1, unregistered: true}
	OpTimes = &UniqueString{s: "*", tok: -1, loc: -1, unregistered: true}
	OpLT = &UniqueString{s: "<", tok: -1, loc: -1, unregistered: true}
	OpLEQ = &UniqueString{s: "\\leq", tok: -1, loc: -1, unregistered: true}
	OpGT = &UniqueString{s: ">", tok: -1, loc: -1, unregistered: true}
	OpGEQ = &UniqueString{s: "\\geq", tok: -1, loc: -1, unregistered: true}
	OpPrime = &UniqueString{s: "'", tok: -1, loc: -1, unregistered: true}
	OpAA = &UniqueString{s: "$AngleAct", tok: -1, loc: -1, unregistered: true}
	OpBC = &UniqueString{s: "$BoundedChoose", tok: -1, loc: -1, unregistered: true}
	OpBE = &UniqueString{s: "$BoundedExists", tok: -1, loc: -1, unregistered: true}
	OpBF = &UniqueString{s: "$BoundedForall", tok: -1, loc: -1, unregistered: true}
	OpCase = &UniqueString{s: "$Case", tok: -1, loc: -1, unregistered: true}
	OpCP = &UniqueString{s: "$CartesianProd", tok: -1, loc: -1, unregistered: true}
	OpCL = &UniqueString{s: "$ConjList", tok: -1, loc: -1, unregistered: true}
	OpDL = &UniqueString{s: "$DisjList", tok: -1, loc: -1, unregistered: true}
	OpExc = &UniqueString{s: "$Except", tok: -1, loc: -1, unregistered: true}
	OpFA = &UniqueString{s: "$FcnApply", tok: -1, loc: -1, unregistered: true}
	OpFC = &UniqueString{s: "$FcnConstructor", tok: -1, loc: -1, unregistered: true}
	OpITE = &UniqueString{s: "$IfThenElse", tok: -1, loc: -1, unregistered: true}
	OpNRFS = &UniqueString{s: "$NonRecursiveFcnSpec", tok: -1, loc: -1, unregistered: true}
	OpPair = &UniqueString{s: "$Pair", tok: -1, loc: -1, unregistered: true}
	OpRC = &UniqueString{s: "$RcdConstructor", tok: -1, loc: -1, unregistered: true}
	OpRS = &UniqueString{s: "$RcdSelect", tok: -1, loc: -1, unregistered: true}
	OpRFS = &UniqueString{s: "$RecursiveFcnSpec", tok: -1, loc: -1, unregistered: true}
	OpSeq = &UniqueString{s: "$Seq", tok: -1, loc: -1, unregistered: true}
	OpSA = &UniqueString{s: "$SquareAct", tok: -1, loc: -1, unregistered: true}
	OpSE = &UniqueString{s: "$SetEnumerate", tok: -1, loc: -1, unregistered: true}
	OpSF = &UniqueString{s: "$SF", tok: -1, loc: -1, unregistered: true}
	OpSOA = &UniqueString{s: "$SetOfAll", tok: -1, loc: -1, unregistered: true}
	OpSOR = &UniqueString{s: "$SetOfRcds", tok: -1, loc: -1, unregistered: true}
	OpSOF = &UniqueString{s: "$SetOfFcns", tok: -1, loc: -1, unregistered: true}
	OpSSO = &UniqueString{s: "$SubsetOf", tok: -1, loc: -1, unregistered: true}
	OpTup = &UniqueString{s: "$Tuple", tok: -1, loc: -1, unregistered: true}
	OpTE = &UniqueString{s: "$TemporalExists", tok: -1, loc: -1, unregistered: true}
	OpTF = &UniqueString{s: "$TemporalForall", tok: -1, loc: -1, unregistered: true}
	OpUC = &UniqueString{s: "$UnboundedChoose", tok: -1, loc: -1, unregistered: true}
	OpUE = &UniqueString{s: "$UnboundedExists", tok: -1, loc: -1, unregistered: true}
	OpUF = &UniqueString{s: "$UnboundedForall", tok: -1, loc: -1, unregistered: true}
	OpWF = &UniqueString{s: "$WF", tok: -1, loc: -1, unregistered: true}
	OpNop = &UniqueString{s: "$Nop", tok: -1, loc: -1, unregistered: true}
}

func ensureBuiltInOPs() { builtInClassInterning.ensure(initBuiltInOPs) }

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
	ensureBuiltInOPs()
	if op == nil {
		return 0
	}
	return GetOpCodeByToken(UniqueStringOf(op.String()).Token())
}

func GetOpCodeByToken(loc int) int {
	ensureBuiltInOPs()
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
