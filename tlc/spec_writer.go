package tlc

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync/atomic"
	"time"
)

const (
	tlaModuleTLC             = "TLC"
	tlaModuleToolbox         = "Toolbox"
	tlaModuleNaturals        = "Naturals"
	tlaModuleSequences       = "Sequences"
	tlaModuleTLCExt          = "TLCExt"
	tlaModuleIOUtils         = "IOUtils"
	tlaModuleJson            = "Json"
	tlaKeywordActionConst    = "ACTION_CONSTRAINT"
	tlaKeywordAssume         = "ASSUME"
	tlaKeywordConstant       = "CONSTANT"
	tlaKeywordConstants      = "CONSTANTS"
	tlaKeywordExtends        = "EXTENDS"
	tlaKeywordInit           = "INIT"
	tlaKeywordInvariant      = "INVARIANT"
	tlaKeywordModule         = "MODULE"
	tlaKeywordNext           = "NEXT"
	tlaKeywordProperty       = "PROPERTY"
	tlaKeywordSpec           = "SPECIFICATION"
	tlaKeywordSymmetry       = "SYMMETRY"
	tlaKeywordUnion          = "\\union"
	tlaKeywordVariable       = "VARIABLE"
	tlaKeywordView           = "VIEW"
	tlaKeywordTrue           = "TRUE"
	tlaKeywordFalse          = "FALSE"
	tlaSchemeSpec            = "spec"
	tlaSchemeInit            = "init"
	tlaSchemeNext            = "next"
	tlaSchemeConstant        = "const"
	tlaSchemeSymmetry        = "symm"
	tlaSchemeDefOverride     = "def_ov"
	tlaSchemeConstraint      = "constr"
	tlaSchemeActionConst     = "action_constr"
	tlaSchemeInvariant       = "inv"
	tlaSchemeProperty        = "prop"
	tlaSchemeView            = "view"
	tlaSchemePostCondition   = "postcondition"
	tlaSchemeAlias           = "alias"
	tlaSchemeConstantExpr    = "const_expr"
	tlaSchemeTraceExprVar    = "__trace_var"
	tlaSchemeTraceExprDef    = "trace_def"
	tlaTraceExprModuleName   = "TTrace"
	tlaTraceExploreModule    = "TEExpression"
	tlaTraceExpressionVar    = "TraceExp"
	tlaTraceExploreInit      = "traceExploreInit"
	tlaTraceExploreNext      = "traceExploreNext"
	tlaTraceExploreActions   = "traceExploreActionConstraint"
	tlaSpecTEExpression      = "expression"
	tlaSpecTETTraceExpr      = "_expression"
	tlaSpecTETraceDef        = "trace"
	tlaSpecTETrace           = "_trace"
	tlaSpecTETraceModuleName = "TETrace"
	tlaSpecTEConstantsName   = "TEConstants"
	tlaSpecTEInit            = "_init"
	tlaSpecTENext            = "_next"
	tlaSpecTELassoStart      = "_TTraceLassoStart"
	tlaSpecTELassoEnd        = "_TTraceLassoEnd"
	tlaSpecTEView            = "_view"
	tlaSpecTEInitID          = "_SpecTEInit"
	tlaSpecTENextID          = "_SpecTENext"
	tlaTEInitID              = "_TEInit"
	tlaTENextID              = "_TENext"
	tlaTEInitAttr            = "teBehaviorInit"
	tlaTENextAttr            = "teBehaviorNext"
	tlaTraceExprPrefix       = "_traceExpression_"
	tlaTraceExprComment      = "\\* TRACE EXPRESSION: "
	tlaGenerationPrefix      = "\\* Generated on "
	tlaIndent                = "    "
	tlaSpace                 = " "
	tlaSep                   = "----"
	tlaEqSp                  = " = "
	tlaAssignArrow           = " <- "
	tlaRecordArrowSp         = " |-> "
	tlaDefines               = " == "
	tlaDefinesCR             = " ==\n"
	tlaComment               = "\\* "
	tlaCommentNS             = "\\*"
	tlaAttribute             = "@"
	tlaColon                 = ":"
	tlaBeginTupleSp          = "<<"
	tlaEndTupleSp            = ">>"
	tlaPrime                 = "'"
	tlaOr                    = "\\/"
	tlaNot                   = "~"
	tlaEventuallyAlways      = "<>[]"
	tlaInfOften              = "[]<>"
	tlaTraceNA               = "\"--\""
	tlaLParen                = "("
	tlaRParen                = ")"
	tlaLSquare               = "["
	tlaRSquare               = "]"
	tlaIndentedConj          = tlaIndent + tlaAnd + tlaSpace
	tlaIndentedDisj          = tlaIndent + tlaOr + tlaSpace
	tlaConstantEvalMarker    = "\"$!@$!@$!@$!@$!\""
)

var specWriterCounter atomic.Int64

func init() {
	specWriterCounter.Store(0)
}

type SpecWriter struct {
	tlaBuffer strings.Builder
	cfgBuffer *strings.Builder
}

func NewSpecWriter(generateConfigurationContent bool) *SpecWriter {
	w := &SpecWriter{}
	if generateConfigurationContent {
		w.cfgBuffer = &strings.Builder{}
	}
	return w
}

func SpecWriterModuleClosingTag(rightMarginWidth int, addModificationHistory bool) string {
	if rightMarginWidth <= 0 {
		rightMarginWidth = 77
	}
	var b strings.Builder
	b.WriteString(strings.Repeat("=", rightMarginWidth))
	b.WriteString(tlaCR)
	if addModificationHistory {
		b.WriteString("\\* Modification History")
		b.WriteString(tlaCR)
		b.WriteString("\\* Created ")
		b.WriteString(specWriterJavaDateString(time.Now()))
		b.WriteString(" by ")
		b.WriteString(os.Getenv("USER"))
		b.WriteString(tlaCR)
	}
	return b.String()
}

func SpecWriterGeneratedTimestampLine() string {
	return tlaGenerationPrefix + specWriterJavaDateString(time.Now())
}

func specWriterJavaDateString(t time.Time) string {
	return t.Format("Mon Jan 02 15:04:05 MST 2006")
}

func SpecWriterValidIdentifier(scheme string) string {
	return fmt.Sprintf("%s_%d%d", scheme, time.Now().UnixMilli(), 1000*specWriterCounter.Add(1))
}

func SpecWriterValidIdentifierNoTimestamp(scheme string) string {
	return "_" + scheme
}

func SpecWriterModuleNameChecked(moduleFilename string, checkExistence bool) string {
	if checkExistence {
		if _, err := os.Stat(moduleFilename); err != nil {
			return ""
		}
	}
	base := filepath.Base(moduleFilename)
	ext := filepath.Ext(base)
	if ext == "" {
		return base
	}
	return strings.TrimSuffix(base, ext)
}

func SpecWriterExtendingModuleContent(moduleFilename string, extendedModuleNames ...string) string {
	var b strings.Builder
	b.WriteString(tlaSep)
	b.WriteString(" ")
	b.WriteString(tlaKeywordModule)
	b.WriteString(" ")
	b.WriteString(SpecWriterModuleNameChecked(moduleFilename, false))
	b.WriteString(" ")
	b.WriteString(tlaSep)
	b.WriteString(tlaCR)
	if len(extendedModuleNames) != 0 {
		b.WriteString(tlaKeywordExtends)
		b.WriteString(" ")
		b.WriteString(strings.Join(extendedModuleNames, ", "))
		b.WriteString(tlaCR)
		b.WriteString(tlaCR)
	}
	return b.String()
}

func SpecWriterCreateListContent(formulas []*Formula, labelingScheme string) [][]string {
	out := make([][]string, 0, len(formulas))
	for i, formula := range formulas {
		label := SpecWriterValidIdentifier(labelingScheme)
		text := ""
		if formula != nil {
			text = formula.GetFormula()
		}
		out = append(out, []string{label, label + tlaDefinesCR + text, fmt.Sprint(i)})
	}
	return out
}

func SpecWriterCreateFormulaListContent(serialized []string, labelingScheme string) [][]string {
	return SpecWriterCreateListContent(DeserializeFormulaList(serialized), labelingScheme)
}

func SpecWriterCreateSourceContent(value string, identifierOrScheme string, isScheme ...bool) [][]string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	useScheme := true
	if len(isScheme) > 0 {
		useScheme = isScheme[0]
	}
	identifier := identifierOrScheme
	if useScheme {
		identifier = SpecWriterValidIdentifier(identifierOrScheme)
	}
	return [][]string{{identifier, identifier + tlaDefinesCR + value}}
}

func (w *SpecWriter) TLAString() string {
	if w == nil {
		return ""
	}
	return w.tlaBuffer.String()
}

func (w *SpecWriter) CFGString() string {
	if w == nil || w.cfgBuffer == nil {
		return ""
	}
	return w.cfgBuffer.String()
}

func (w *SpecWriter) WriteStreams(tlaOut io.Writer, cfgOut io.Writer) error {
	if w == nil {
		return nil
	}
	w.tlaBuffer.WriteString(w.GetTLAModuleClosingTag())
	if tlaOut != nil {
		if _, err := io.Copy(tlaOut, bytes.NewBufferString(w.tlaBuffer.String())); err != nil {
			return err
		}
	}
	if w.cfgBuffer != nil {
		w.cfgBuffer.WriteString(SpecWriterGeneratedTimestampLine())
		if cfgOut != nil {
			if _, err := io.Copy(cfgOut, bytes.NewBufferString(w.cfgBuffer.String())); err != nil {
				return err
			}
		}
	}
	return nil
}

func (w *SpecWriter) WriteFiles(tlaFile string, cfgFile string) error {
	var tlaOut *os.File
	var cfgOut *os.File
	var err error
	if tlaFile != "" {
		tlaOut, err = os.Create(tlaFile)
		if err != nil {
			return err
		}
		defer tlaOut.Close()
	}
	if cfgFile != "" {
		cfgOut, err = os.Create(cfgFile)
		if err != nil {
			return err
		}
		defer cfgOut.Close()
	}
	return w.WriteStreams(tlaOut, cfgOut)
}

func (w *SpecWriter) GetTLAModuleClosingTag() string {
	return SpecWriterModuleClosingTag(77, false)
}

func (w *SpecWriter) AddPrimer(moduleFilename string, extendedModuleName string) {
	if w == nil {
		return
	}
	w.tlaBuffer.WriteString(SpecWriterExtendingModuleContent(moduleFilename, extendedModuleName, tlaModuleTLC))
}

func (w *SpecWriter) AddSpecDefinition(specDefinition []string, attributeName string) {
	if w == nil || len(specDefinition) < 2 {
		return
	}
	if w.cfgBuffer != nil {
		w.cfgBuffer.WriteString(tlaKeywordSpec)
		w.cfgBuffer.WriteString(tlaSpace)
		w.cfgBuffer.WriteString(specDefinition[0])
		w.cfgBuffer.WriteString(tlaCR)
	}
	w.tlaBuffer.WriteString(tlaComment)
	w.tlaBuffer.WriteString(tlaKeywordSpec)
	w.tlaBuffer.WriteString(tlaSpace)
	w.tlaBuffer.WriteString(tlaAttribute)
	w.tlaBuffer.WriteString(attributeName)
	w.tlaBuffer.WriteString(tlaCR)
	w.tlaBuffer.WriteString(specDefinition[1])
	w.tlaBuffer.WriteString(specWriterClosingSep())
}

func (w *SpecWriter) AddInitNextDefinitions(initDefinition []string, nextDefinition []string, initAttributeName string, nextAttributeName string) {
	if w == nil || len(initDefinition) < 2 || len(nextDefinition) < 2 {
		return
	}
	if w.cfgBuffer != nil {
		w.cfgBuffer.WriteString(tlaComment)
		w.cfgBuffer.WriteString(tlaKeywordInit)
		w.cfgBuffer.WriteString(" definition")
		w.cfgBuffer.WriteString(tlaCR)
		w.cfgBuffer.WriteString(tlaKeywordInit)
		w.cfgBuffer.WriteString(tlaCR)
		w.cfgBuffer.WriteString(initDefinition[0])
		w.cfgBuffer.WriteString(tlaCR)
	}
	w.tlaBuffer.WriteString(tlaComment)
	w.tlaBuffer.WriteString(tlaKeywordInit)
	w.tlaBuffer.WriteString(" definition ")
	w.tlaBuffer.WriteString(tlaAttribute)
	w.tlaBuffer.WriteString(initAttributeName)
	w.tlaBuffer.WriteString(tlaCR)
	w.tlaBuffer.WriteString(initDefinition[1])
	w.tlaBuffer.WriteString(tlaCR)

	if w.cfgBuffer != nil {
		w.cfgBuffer.WriteString(tlaComment)
		w.cfgBuffer.WriteString(tlaKeywordNext)
		w.cfgBuffer.WriteString(" definition")
		w.cfgBuffer.WriteString(tlaCR)
		w.cfgBuffer.WriteString(tlaKeywordNext)
		w.cfgBuffer.WriteString(tlaCR)
		w.cfgBuffer.WriteString(nextDefinition[0])
		w.cfgBuffer.WriteString(tlaCR)
	}
	w.tlaBuffer.WriteString(tlaComment)
	w.tlaBuffer.WriteString(tlaKeywordNext)
	w.tlaBuffer.WriteString(" definition ")
	w.tlaBuffer.WriteString(tlaAttribute)
	w.tlaBuffer.WriteString(nextAttributeName)
	w.tlaBuffer.WriteString(tlaCR)
	w.tlaBuffer.WriteString(nextDefinition[1])
	w.tlaBuffer.WriteString(tlaCR)
}

func (w *SpecWriter) AddConstantsRaw(rawConstants []string) {
	if w == nil || w.cfgBuffer == nil || len(rawConstants) == 0 {
		return
	}
	w.cfgBuffer.WriteString(tlaKeywordConstants)
	w.cfgBuffer.WriteString(tlaCR)
	for _, constant := range rawConstants {
		w.cfgBuffer.WriteString(constant)
		w.cfgBuffer.WriteString(tlaCR)
	}
}

func (w *SpecWriter) AddConstants(constants []*Assignment, modelValues *TypedSet, attributeConstants string, attributeMVs string) {
	if w == nil {
		return
	}
	w.AddMVTypedSet(modelValues, "MV CONSTANT declarations ", attributeMVs)
	var symmetrySets []string
	for _, constant := range constants {
		if constant != nil && constant.IsModelValue() && constant.IsSetOfModelValues() {
			w.AddMVTypedSet(constant.GetSetOfModelValues(), "MV CONSTANT declarations", attributeConstants)
		}
	}
	for _, constant := range constants {
		if constant == nil || !constant.IsModelValue() {
			continue
		}
		if constant.IsSetOfModelValues() {
			if w.cfgBuffer != nil {
				w.cfgBuffer.WriteString(tlaComment)
				w.cfgBuffer.WriteString("MV CONSTANT definitions")
				w.cfgBuffer.WriteString(tlaCR)
			}
			w.tlaBuffer.WriteString(tlaComment)
			w.tlaBuffer.WriteString("MV CONSTANT definitions ")
			w.tlaBuffer.WriteString(constant.GetLeft())
			w.tlaBuffer.WriteString(tlaCR)
			id := w.AddArrowAssignment(constant, tlaSchemeConstant)
			if constant.IsSymmetricalSet() {
				symmetrySets = append(symmetrySets, id)
			}
			w.tlaBuffer.WriteString(tlaSep)
			w.tlaBuffer.WriteString(tlaCR)
			w.tlaBuffer.WriteString(tlaCR)
		} else if w.cfgBuffer != nil {
			w.cfgBuffer.WriteString(tlaComment)
			w.cfgBuffer.WriteString(tlaKeywordConstant)
			w.cfgBuffer.WriteString(" declarations")
			w.cfgBuffer.WriteString(tlaCR)
			w.cfgBuffer.WriteString(tlaKeywordConstant)
			w.cfgBuffer.WriteString(tlaSpace)
			w.cfgBuffer.WriteString(constant.GetLabel())
			w.cfgBuffer.WriteString(tlaEqSp)
			w.cfgBuffer.WriteString(constant.GetRight())
			w.cfgBuffer.WriteString(tlaCR)
		}
	}
	if len(symmetrySets) != 0 {
		label := SpecWriterValidIdentifier(tlaSchemeSymmetry)
		w.tlaBuffer.WriteString(tlaComment)
		w.tlaBuffer.WriteString(tlaKeywordSymmetry)
		w.tlaBuffer.WriteString(" definition")
		w.tlaBuffer.WriteString(tlaCR)
		if w.cfgBuffer != nil {
			w.cfgBuffer.WriteString(tlaComment)
			w.cfgBuffer.WriteString(tlaKeywordSymmetry)
			w.cfgBuffer.WriteString(" definition")
			w.cfgBuffer.WriteString(tlaCR)
		}
		w.tlaBuffer.WriteString(label)
		w.tlaBuffer.WriteString(tlaDefines)
		w.tlaBuffer.WriteString(tlaCR)
		for i, set := range symmetrySets {
			w.tlaBuffer.WriteString("Permutations(")
			w.tlaBuffer.WriteString(set)
			w.tlaBuffer.WriteString(")")
			if i != len(symmetrySets)-1 {
				w.tlaBuffer.WriteString(" ")
				w.tlaBuffer.WriteString(tlaKeywordUnion)
				w.tlaBuffer.WriteString(" ")
			}
		}
		w.tlaBuffer.WriteString(specWriterClosingSep())
		if w.cfgBuffer != nil {
			w.cfgBuffer.WriteString(tlaKeywordSymmetry)
			w.cfgBuffer.WriteString(tlaSpace)
			w.cfgBuffer.WriteString(label)
			w.cfgBuffer.WriteString(tlaCR)
		}
	}
}

func (w *SpecWriter) AddConstantsBis(constants []*Assignment, attributeConstants string) {
	if w == nil {
		return
	}
	for i, constant := range constants {
		if constant == nil || constant.IsModelValue() {
			continue
		}
		if w.cfgBuffer != nil {
			w.cfgBuffer.WriteString(tlaComment)
			w.cfgBuffer.WriteString(tlaKeywordConstant)
			w.cfgBuffer.WriteString(" definitions")
			w.cfgBuffer.WriteString(tlaCR)
		}
		w.tlaBuffer.WriteString(tlaComment)
		w.tlaBuffer.WriteString(tlaKeywordConstant)
		w.tlaBuffer.WriteString(" definitions ")
		w.tlaBuffer.WriteString(tlaAttribute)
		w.tlaBuffer.WriteString(attributeConstants)
		w.tlaBuffer.WriteString(tlaColon)
		w.tlaBuffer.WriteString(fmt.Sprint(i))
		w.tlaBuffer.WriteString(constant.GetLeft())
		w.tlaBuffer.WriteString(tlaCR)
		w.AddArrowAssignment(constant, tlaSchemeConstant)
		w.tlaBuffer.WriteString(tlaSep)
		w.tlaBuffer.WriteString(tlaCR)
		w.tlaBuffer.WriteString(tlaCR)
	}
}

func (w *SpecWriter) AddConstantExpressionEvaluation(expression string, attributeName string) {
	if w == nil || strings.TrimSpace(expression) == "" {
		return
	}
	id := SpecWriterValidIdentifier(tlaSchemeConstantExpr)
	w.tlaBuffer.WriteString(tlaComment)
	w.tlaBuffer.WriteString("Constant expression definition ")
	w.tlaBuffer.WriteString(tlaAttribute)
	w.tlaBuffer.WriteString(attributeName)
	w.tlaBuffer.WriteString(tlaCR)
	w.tlaBuffer.WriteString(id)
	w.tlaBuffer.WriteString(tlaDefines)
	w.tlaBuffer.WriteString(tlaCR)
	w.tlaBuffer.WriteString(expression)
	w.tlaBuffer.WriteString(specWriterClosingSep())
	w.tlaBuffer.WriteString(tlaComment)
	w.tlaBuffer.WriteString("Constant expression ASSUME statement ")
	w.tlaBuffer.WriteString(tlaAttribute)
	w.tlaBuffer.WriteString(attributeName)
	w.tlaBuffer.WriteString(tlaCR)
	w.tlaBuffer.WriteString("ASSUME PrintT(")
	w.tlaBuffer.WriteString(tlaBeginTupleSp)
	w.tlaBuffer.WriteString(tlaConstantEvalMarker)
	w.tlaBuffer.WriteString(tlaComma)
	w.tlaBuffer.WriteString(id)
	w.tlaBuffer.WriteString(tlaEndTupleSp)
	w.tlaBuffer.WriteString(")")
	w.tlaBuffer.WriteString(specWriterClosingSep())
}

func (w *SpecWriter) AddNewDefinitions(definitions string, attributeName string) {
	if w == nil || strings.TrimSpace(definitions) == "" {
		return
	}
	w.tlaBuffer.WriteString(tlaComment)
	w.tlaBuffer.WriteString("New definitions ")
	w.tlaBuffer.WriteString(tlaAttribute)
	w.tlaBuffer.WriteString(attributeName)
	w.tlaBuffer.WriteString(tlaCR)
	w.tlaBuffer.WriteString(definitions)
	w.tlaBuffer.WriteString(specWriterClosingSep())
}

func (w *SpecWriter) AddFormulaList(elements [][]string, keyword string, attributeName string) {
	if w == nil || len(elements) == 0 {
		return
	}
	if w.cfgBuffer != nil {
		w.cfgBuffer.WriteString(tlaComment)
		w.cfgBuffer.WriteString(keyword)
		w.cfgBuffer.WriteString(" definition")
		w.cfgBuffer.WriteString(tlaCR)
		w.cfgBuffer.WriteString(keyword)
		w.cfgBuffer.WriteString(tlaCR)
	}
	for i, element := range elements {
		if len(element) == 0 {
			continue
		}
		if w.cfgBuffer != nil {
			w.cfgBuffer.WriteString(element[0])
			w.cfgBuffer.WriteString(tlaCR)
		}
		if len(element) > 1 && element[1] != "" {
			index := fmt.Sprint(i)
			if len(element) > 2 {
				index = element[2]
			}
			w.tlaBuffer.WriteString(tlaComment)
			w.tlaBuffer.WriteString(keyword)
			w.tlaBuffer.WriteString(" definition ")
			w.tlaBuffer.WriteString(tlaAttribute)
			w.tlaBuffer.WriteString(attributeName)
			w.tlaBuffer.WriteString(tlaColon)
			w.tlaBuffer.WriteString(index)
			w.tlaBuffer.WriteString(tlaCR)
			w.tlaBuffer.WriteString(element[1])
			w.tlaBuffer.WriteString(specWriterClosingSep())
		}
	}
}

func (w *SpecWriter) AddFormulaElement(element string, keyword string, attributeName string) {
	w.AddFormulaList([][]string{{element, ""}}, keyword, attributeName)
}

func (w *SpecWriter) AddView(viewString string, attributeName string) {
	if w == nil || strings.TrimSpace(viewString) == "" {
		return
	}
	id := SpecWriterValidIdentifier(tlaSchemeView)
	if w.cfgBuffer != nil {
		w.cfgBuffer.WriteString(tlaComment)
		w.cfgBuffer.WriteString("VIEW definition")
		w.cfgBuffer.WriteString(tlaCR)
		w.cfgBuffer.WriteString(tlaKeywordView)
		w.cfgBuffer.WriteString(tlaCR)
		w.cfgBuffer.WriteString(id)
		w.cfgBuffer.WriteString(tlaCR)
	}
	w.tlaBuffer.WriteString(tlaComment)
	w.tlaBuffer.WriteString("VIEW definition ")
	w.tlaBuffer.WriteString(tlaAttribute)
	w.tlaBuffer.WriteString(attributeName)
	w.tlaBuffer.WriteString(tlaCR)
	w.tlaBuffer.WriteString(id)
	w.tlaBuffer.WriteString(tlaDefines)
	w.tlaBuffer.WriteString(tlaCR)
	w.tlaBuffer.WriteString(viewString)
	w.tlaBuffer.WriteString(specWriterClosingSep())
}

func (w *SpecWriter) AddPostCondition(postConditionString string, attributeName string) {
	if w == nil || strings.TrimSpace(postConditionString) == "" {
		return
	}
	id := SpecWriterValidIdentifier(tlaSchemePostCondition)
	if w.cfgBuffer != nil {
		w.cfgBuffer.WriteString(tlaComment)
		w.cfgBuffer.WriteString("POSTCONDITION definition")
		w.cfgBuffer.WriteString(tlaCR)
		w.cfgBuffer.WriteString("POSTCONDITION")
		w.cfgBuffer.WriteString(tlaCR)
		w.cfgBuffer.WriteString(id)
		w.cfgBuffer.WriteString(tlaCR)
	}
	w.tlaBuffer.WriteString(tlaComment)
	w.tlaBuffer.WriteString("POSTCONDITION definition ")
	w.tlaBuffer.WriteString(tlaAttribute)
	w.tlaBuffer.WriteString(attributeName)
	w.tlaBuffer.WriteString(tlaCR)
	w.tlaBuffer.WriteString(id)
	w.tlaBuffer.WriteString(tlaDefines)
	w.tlaBuffer.WriteString(tlaCR)
	w.tlaBuffer.WriteString(postConditionString)
	w.tlaBuffer.WriteString(specWriterClosingSep())
}

func (w *SpecWriter) AddAlias(aliasString string, attributeName string) {
	if w == nil || strings.TrimSpace(aliasString) == "" {
		return
	}
	id := SpecWriterValidIdentifier(tlaSchemeAlias)
	w.AddAliasToCfg(id)
	w.tlaBuffer.WriteString(tlaComment)
	w.tlaBuffer.WriteString("ALIAS definition ")
	w.tlaBuffer.WriteString(tlaAttribute)
	w.tlaBuffer.WriteString(attributeName)
	w.tlaBuffer.WriteString(tlaCR)
	w.tlaBuffer.WriteString(id)
	w.tlaBuffer.WriteString(tlaDefines)
	w.tlaBuffer.WriteString(tlaCR)
	w.tlaBuffer.WriteString(aliasString)
	w.tlaBuffer.WriteString(specWriterClosingSep())
}

func (w *SpecWriter) AddAliasToCfg(aliasName string) {
	if w == nil || w.cfgBuffer == nil {
		return
	}
	w.cfgBuffer.WriteString(tlaCR)
	w.cfgBuffer.WriteString("ALIAS")
	w.cfgBuffer.WriteString(tlaCR)
	w.cfgBuffer.WriteString(tlaIndent)
	w.cfgBuffer.WriteString(aliasName)
	w.cfgBuffer.WriteString(tlaCR)
}

func (w *SpecWriter) AddArrowAssignment(constant *Assignment, schema string) string {
	if w == nil || constant == nil {
		return ""
	}
	identifier := SpecWriterValidIdentifier(schema)
	return SpecWriterAddArrowAssignmentIDToBuffers(&w.tlaBuffer, w.cfgBuffer, constant, identifier, identifier)
}

func SpecWriterAddArrowAssignmentToBuffers(tlaBuffer *strings.Builder, cfgBuffer *strings.Builder, constant *Assignment, schema string) string {
	identifier := SpecWriterValidIdentifier(schema)
	return SpecWriterAddArrowAssignmentIDToBuffers(tlaBuffer, cfgBuffer, constant, identifier, identifier)
}

func SpecWriterAddArrowAssignmentIDToBuffers(tlaBuffer *strings.Builder, cfgBuffer *strings.Builder, constant *Assignment, id string, configID string) string {
	if tlaBuffer == nil || constant == nil {
		return id
	}
	tlaBuffer.WriteString(constant.GetParametrizedLabel(id))
	tlaBuffer.WriteString(tlaDefines)
	tlaBuffer.WriteString(tlaCR)
	tlaBuffer.WriteString(constant.GetRight())
	tlaBuffer.WriteString(tlaCR)
	if cfgBuffer != nil {
		cfgBuffer.WriteString(tlaCR)
		cfgBuffer.WriteString(tlaKeywordConstant)
		cfgBuffer.WriteString(tlaCR)
		cfgBuffer.WriteString(tlaIndent)
		cfgBuffer.WriteString(constant.GetLabel())
		cfgBuffer.WriteString(tlaAssignArrow)
		cfgBuffer.WriteString(configID)
		cfgBuffer.WriteString(tlaCR)
	}
	return id
}

func (w *SpecWriter) AddMVTypedSet(mvSet *TypedSet, comment string, attributeName string) {
	if w == nil || mvSet == nil || mvSet.ValueCount() == 0 {
		return
	}
	if comment != "" {
		w.tlaBuffer.WriteString(tlaComment)
		w.tlaBuffer.WriteString(comment)
		w.tlaBuffer.WriteString(tlaAttribute)
		w.tlaBuffer.WriteString(attributeName)
		w.tlaBuffer.WriteString(tlaCR)
	}
	w.tlaBuffer.WriteString(tlaKeywordConstants)
	w.tlaBuffer.WriteString(tlaCR)
	w.tlaBuffer.WriteString(mvSet.StringWithoutBraces())
	w.tlaBuffer.WriteString(specWriterClosingSep())
	if w.cfgBuffer != nil {
		if comment != "" {
			w.cfgBuffer.WriteString(tlaComment)
			w.cfgBuffer.WriteString(comment)
			w.cfgBuffer.WriteString(tlaCR)
		}
		w.cfgBuffer.WriteString(tlaKeywordConstants)
		w.cfgBuffer.WriteString(tlaCR)
		for i := 0; i < mvSet.ValueCount(); i++ {
			mv := mvSet.Value(i)
			w.cfgBuffer.WriteString(mv)
			w.cfgBuffer.WriteString(tlaEqSp)
			w.cfgBuffer.WriteString(mv)
			w.cfgBuffer.WriteString(tlaCR)
		}
	}
}

type TraceExpressionInformation struct {
	Expression   string
	Identifier   string
	VariableName string
	Level        int
}

func NewTraceExpressionInformation(expression string, identifier string, variableName string) *TraceExpressionInformation {
	return &TraceExpressionInformation{Expression: expression, Identifier: identifier, VariableName: variableName}
}

func NewTraceExpressionInformationHolders(expressions []*Formula, attributeName string) []*TraceExpressionInformation {
	_ = attributeName
	holders := make([]*TraceExpressionInformation, len(expressions))
	for i, formula := range expressions {
		if formula == nil || formula.GetFormula() == "" {
			continue
		}
		identifier := SpecWriterValidIdentifier(tlaSchemeTraceExprDef)
		if formula.IsNamed() {
			holders[i] = NewTraceExpressionInformation(formula.GetRightHandSide(), identifier, formula.GetLeftHandSide())
		} else {
			varName := SpecWriterValidIdentifier(tlaSchemeTraceExprVar)
			holders[i] = NewTraceExpressionInformation(formula.GetFormula(), identifier, varName)
		}
	}
	return holders
}

type SpecTraceExpressionWriter struct {
	*SpecWriter
}

func NewSpecTraceExpressionWriter() *SpecTraceExpressionWriter {
	return &SpecTraceExpressionWriter{SpecWriter: NewSpecWriter(true)}
}

func SpecTraceExpressionAddInitNextToBuffers(tlaBuffer *strings.Builder, cfgBuffer *strings.Builder, trace []*MCState, expressionData []*TraceExpressionInformation) []string {
	initID := SpecWriterValidIdentifier(tlaSchemeInit)
	nextID := SpecWriterValidIdentifier(tlaSchemeNext)
	actionConstraintID := SpecWriterValidIdentifier(tlaSchemeActionConst)
	SpecTraceExpressionAddInitNextIDToBuffers(tlaBuffer, cfgBuffer, trace, expressionData, initID, nextID, actionConstraintID)
	return []string{initID, nextID, actionConstraintID}
}

func SpecTraceExpressionAddInitNextIDToBuffers(tlaBuffer *strings.Builder, cfgBuffer *strings.Builder, trace []*MCState, expressionData []*TraceExpressionInformation, initID string, nextID string, actionConstraintID string) {
	SpecTraceExpressionAddInitNextFullToBuffers(tlaBuffer, cfgBuffer, trace, expressionData, initID, nextID, actionConstraintID, tlaSchemeNext, false)
}

func SpecTraceExpressionAddInitNextFullToBuffers(tlaBuffer *strings.Builder, cfgBuffer *strings.Builder, trace []*MCState, expressionData []*TraceExpressionInformation, initID string, nextID string, actionConstraintID string, nextSubActionBasename string, leaveStubsForTraceExpression bool) {
	parts := SpecTraceExpressionBuildInitNextBuffers(cfgBuffer, trace, expressionData, initID, nextID, actionConstraintID, nextSubActionBasename, leaveStubsForTraceExpression)
	if tlaBuffer != nil {
		tlaBuffer.WriteString(parts[0].String())
		tlaBuffer.WriteString(parts[1].String())
	}
}

func SpecTraceExpressionBuildInitNextBuffers(cfgBuffer *strings.Builder, trace []*MCState, expressionData []*TraceExpressionInformation, initID string, nextID string, actionConstraintID string, nextSubActionBasename string, leaveStubsForTraceExpression bool) [2]strings.Builder {
	var subActionsAndConstraint strings.Builder
	var initAndNext strings.Builder
	if len(trace) == 0 {
		return [2]strings.Builder{subActionsAndConstraint, initAndNext}
	}
	currentState := trace[0]
	if cfgBuffer != nil {
		cfgBuffer.WriteString(tlaCR)
		cfgBuffer.WriteString(tlaCR)
		cfgBuffer.WriteString(tlaKeywordInit)
		cfgBuffer.WriteString(tlaCR)
		cfgBuffer.WriteString(initID)
		cfgBuffer.WriteString(tlaCR)
	}
	if leaveStubsForTraceExpression {
		initAndNext.WriteString(tlaComment)
		initAndNext.WriteString(tlaKeywordVariable)
		initAndNext.WriteString(" ")
		initAndNext.WriteString(tlaTraceExpressionVar)
		initAndNext.WriteString(tlaCR)
		initAndNext.WriteString(tlaCR)
	}
	initAndNext.WriteString(tlaComment)
	initAndNext.WriteString("TRACE INIT definition ")
	initAndNext.WriteString(tlaTraceExploreInit)
	initAndNext.WriteString(tlaCR)
	initAndNext.WriteString(initID)
	initAndNext.WriteString(tlaDefinesCR)
	for _, variable := range currentState.Variables {
		initAndNext.WriteString(tlaIndentedConj)
		initAndNext.WriteString(variable.Name)
		initAndNext.WriteString(tlaEqSp)
		initAndNext.WriteString(tlaLParen)
		initAndNext.WriteString(tlaCR)
		initAndNext.WriteString(variable.ValueAsStringReIndentedAs(tlaTraceTripleIndent()))
		initAndNext.WriteString(tlaCR)
		initAndNext.WriteString(tlaIndent)
		initAndNext.WriteString(tlaIndent)
		initAndNext.WriteString(tlaRParen)
		initAndNext.WriteString(tlaCR)
	}
	for _, info := range expressionData {
		initAndNext.WriteString(tlaIndentedConj)
		initAndNext.WriteString(info.VariableName)
		initAndNext.WriteString(tlaEqSp)
		initAndNext.WriteString(tlaLParen)
		initAndNext.WriteString(tlaCR)
		initAndNext.WriteString(tlaTraceTripleIndent())
		if info.Level == 2 {
			initAndNext.WriteString(tlaTraceNA)
		} else {
			initAndNext.WriteString(info.Identifier)
		}
		initAndNext.WriteString(tlaCR)
		initAndNext.WriteString(tlaIndent)
		initAndNext.WriteString(tlaIndent)
		initAndNext.WriteString(tlaRParen)
		initAndNext.WriteString(tlaCR)
	}
	if leaveStubsForTraceExpression {
		initAndNext.WriteString(tlaComment)
		initAndNext.WriteString(tlaIndentedConj)
		initAndNext.WriteString(tlaTraceExpressionVar)
		initAndNext.WriteString(tlaEqSp)
		initAndNext.WriteString(tlaKeywordTrue)
		initAndNext.WriteString(tlaCR)
	}
	initAndNext.WriteString(specWriterClosingSep())
	initAndNext.WriteString(tlaCR)

	if cfgBuffer != nil {
		cfgBuffer.WriteString(tlaComment)
		cfgBuffer.WriteString(tlaKeywordNext)
		cfgBuffer.WriteString(" definition")
		cfgBuffer.WriteString(tlaCR)
		cfgBuffer.WriteString(tlaKeywordNext)
		cfgBuffer.WriteString(tlaCR)
		cfgBuffer.WriteString(nextID)
		cfgBuffer.WriteString(tlaCR)
	}
	nextState := currentState
	isSingleState := true
	if len(trace) > 1 {
		nextState = trace[1]
		isSingleState = false
	}
	var nextDisjunctBuffer strings.Builder
	nextDisjunctBuffer.WriteString(nextID)
	nextDisjunctBuffer.WriteString(tlaDefinesCR)
	firstIndent := tlaIndent
	if leaveStubsForTraceExpression {
		nextDisjunctBuffer.WriteString(tlaAnd)
		nextDisjunctBuffer.WriteString(" ")
		firstIndent = " "
	}
	var actionConstraintBuffer strings.Builder
	actionConstraintBuffer.WriteString(actionConstraintID)
	actionConstraintBuffer.WriteString(tlaDefinesCR)
	actionConstraintBuffer.WriteString(tlaBeginTupleSp)
	actionConstraintBuffer.WriteString(tlaCR)
	if cfgBuffer != nil {
		cfgBuffer.WriteString(tlaComment)
		cfgBuffer.WriteString("Action Constraint definition")
		cfgBuffer.WriteString(tlaCR)
		cfgBuffer.WriteString(tlaComment)
		cfgBuffer.WriteString(tlaKeywordActionConst)
		cfgBuffer.WriteString(tlaCR)
		cfgBuffer.WriteString(tlaComment)
		cfgBuffer.WriteString(actionConstraintID)
		cfgBuffer.WriteString(tlaCR)
	}
	subActionIndex := 0
	traceIndex := 1
	for nextState != nil {
		nextDisjunct := fmt.Sprintf("%s_sa_%d", nextSubActionBasename, subActionIndex)
		if subActionIndex == 0 {
			nextDisjunctBuffer.WriteString(firstIndent)
		} else {
			nextDisjunctBuffer.WriteString(tlaIndent)
		}
		nextDisjunctBuffer.WriteString(tlaOr)
		nextDisjunctBuffer.WriteString(" ")
		nextDisjunctBuffer.WriteString(nextDisjunct)
		nextDisjunctBuffer.WriteString(tlaCR)
		actionConstraintBuffer.WriteString(nextDisjunct)
		subActionsAndConstraint.WriteString(tlaComment)
		subActionsAndConstraint.WriteString("TRACE Sub-Action definition ")
		subActionsAndConstraint.WriteString(fmt.Sprint(subActionIndex))
		subActionsAndConstraint.WriteString(tlaCR)
		subActionsAndConstraint.WriteString(nextDisjunct)
		subActionsAndConstraint.WriteString(tlaDefinesCR)
		subActionIndex++
		if nextState.BackToState {
			nextState = trace[nextState.StateNumber-1]
		} else if nextState.Stuttering {
			nextState = currentState
		}
		subActionsAndConstraint.WriteString(tlaIndent)
		subActionsAndConstraint.WriteString(tlaLParen)
		subActionsAndConstraint.WriteString(tlaCR)
		for _, variable := range currentState.Variables {
			subActionsAndConstraint.WriteString(tlaIndent)
			subActionsAndConstraint.WriteString(tlaIndentedConj)
			subActionsAndConstraint.WriteString(variable.Name)
			subActionsAndConstraint.WriteString(tlaEqSp)
			subActionsAndConstraint.WriteString(tlaLParen)
			subActionsAndConstraint.WriteString(tlaCR)
			subActionsAndConstraint.WriteString(variable.ValueAsStringReIndentedAs(tlaTraceTripleIndent() + tlaIndent))
			subActionsAndConstraint.WriteString(tlaCR)
			subActionsAndConstraint.WriteString(tlaTraceTripleIndent())
			subActionsAndConstraint.WriteString(tlaRParen)
			subActionsAndConstraint.WriteString(tlaCR)
		}
		if isSingleState {
			subActionsAndConstraint.WriteString(tlaIndent)
			subActionsAndConstraint.WriteString(tlaIndentedConj)
			subActionsAndConstraint.WriteString(tlaKeywordFalse)
			subActionsAndConstraint.WriteString(tlaCR)
		}
		for i := range currentState.Variables {
			variable := nextState.Variables[i]
			subActionsAndConstraint.WriteString(tlaIndent)
			subActionsAndConstraint.WriteString(tlaIndentedConj)
			subActionsAndConstraint.WriteString(variable.Name)
			subActionsAndConstraint.WriteString(tlaPrime)
			subActionsAndConstraint.WriteString(tlaEqSp)
			subActionsAndConstraint.WriteString(tlaLParen)
			subActionsAndConstraint.WriteString(tlaCR)
			subActionsAndConstraint.WriteString(variable.ValueAsStringReIndentedAs(tlaTraceTripleIndent() + tlaIndent))
			subActionsAndConstraint.WriteString(tlaCR)
			subActionsAndConstraint.WriteString(tlaTraceTripleIndent())
			subActionsAndConstraint.WriteString(tlaRParen)
			subActionsAndConstraint.WriteString(tlaCR)
		}
		for _, info := range expressionData {
			subActionsAndConstraint.WriteString(tlaIndent)
			subActionsAndConstraint.WriteString(tlaIndentedConj)
			subActionsAndConstraint.WriteString(info.VariableName)
			subActionsAndConstraint.WriteString(tlaPrime)
			subActionsAndConstraint.WriteString(tlaEqSp)
			subActionsAndConstraint.WriteString(tlaLParen)
			subActionsAndConstraint.WriteString(tlaCR)
			subActionsAndConstraint.WriteString(tlaTraceTripleIndent())
			subActionsAndConstraint.WriteString(info.Identifier)
			subActionsAndConstraint.WriteString(tlaCR)
			subActionsAndConstraint.WriteString(tlaTraceTripleIndent())
			subActionsAndConstraint.WriteString(tlaRParen)
			if info.Level < 2 {
				subActionsAndConstraint.WriteString(tlaPrime)
			}
			subActionsAndConstraint.WriteString(tlaCR)
		}
		subActionsAndConstraint.WriteString(tlaIndent)
		subActionsAndConstraint.WriteString(tlaRParen)
		subActionsAndConstraint.WriteString(tlaCR)
		subActionsAndConstraint.WriteString(tlaCR)
		if traceIndex < len(trace)-1 {
			actionConstraintBuffer.WriteString(tlaComma)
		}
		actionConstraintBuffer.WriteString(tlaCR)
		currentState = nextState
		if traceIndex < len(trace)-1 {
			traceIndex++
			nextState = trace[traceIndex]
		} else {
			nextState = nil
		}
	}
	initAndNext.WriteString(tlaComment)
	initAndNext.WriteString("TRACE NEXT definition ")
	initAndNext.WriteString(tlaTraceExploreNext)
	initAndNext.WriteString(tlaCR)
	initAndNext.WriteString(nextDisjunctBuffer.String())
	if leaveStubsForTraceExpression {
		initAndNext.WriteString(tlaComment)
		initAndNext.WriteString(tlaAnd)
		initAndNext.WriteString(" ")
		initAndNext.WriteString(tlaTraceExpressionVar)
		initAndNext.WriteString(tlaPrime)
		initAndNext.WriteString(tlaEqSp)
		initAndNext.WriteString(tlaTraceExpressionVar)
		initAndNext.WriteString(tlaCR)
	}
	initAndNext.WriteString(tlaCR)
	initAndNext.WriteString(tlaCR)
	subActionsAndConstraint.WriteString(tlaComment)
	subActionsAndConstraint.WriteString("TRACE Action Constraint definition ")
	subActionsAndConstraint.WriteString(tlaTraceExploreActions)
	subActionsAndConstraint.WriteString(tlaCR)
	subActionsAndConstraint.WriteString(actionConstraintBuffer.String())
	subActionsAndConstraint.WriteString(tlaEndTupleSp)
	subActionsAndConstraint.WriteString("[TLCGet(\"level\")]")
	subActionsAndConstraint.WriteString(specWriterClosingSep())
	subActionsAndConstraint.WriteString(tlaCR)
	return [2]strings.Builder{subActionsAndConstraint, initAndNext}
}

func SpecTraceExpressionAddTraceFunctionToBuffers(tlaBuffer *strings.Builder, cfgBuffer *strings.Builder, input []*MCState, id string, configID string) string {
	trace := make([]*MCState, 0, len(input))
	for _, state := range input {
		if state != nil && !state.BackToState && !state.Stuttering {
			trace = append(trace, state)
		}
	}
	if len(trace) == 0 {
		return SpecWriterAddArrowAssignmentIDToBuffers(tlaBuffer, cfgBuffer, NewAssignment(tlaTraceExploreTrace(), nil, tlaBeginTupleSp+tlaEndTupleSp), id, configID)
	}
	var traceFunctionDef strings.Builder
	traceFunctionDef.WriteString(tlaIndent)
	traceFunctionDef.WriteString(tlaBeginTupleSp)
	traceFunctionDef.WriteString(tlaCR)
	for i, state := range trace {
		traceFunctionDef.WriteString(tlaIndent)
		traceFunctionDef.WriteString(tlaLParen)
		traceFunctionDef.WriteString(state.AsSimpleRecord())
		traceFunctionDef.WriteString(tlaRParen)
		if i < len(trace)-1 {
			traceFunctionDef.WriteString(tlaComma)
			traceFunctionDef.WriteString(tlaCR)
		}
	}
	traceFunctionDef.WriteString(tlaCR)
	traceFunctionDef.WriteString(tlaIndent)
	traceFunctionDef.WriteString(tlaEndTupleSp)
	traceFunctionDef.WriteString(specWriterClosingSep())
	traceFunctionDef.WriteString(tlaCR)
	return SpecWriterAddArrowAssignmentIDToBuffers(tlaBuffer, cfgBuffer, NewAssignment(tlaTraceExploreTrace(), nil, traceFunctionDef.String()), id, configID)
}

func (w *SpecTraceExpressionWriter) AddPrimer(moduleFilename string, extendedModuleName string, extraExtendedModules ...[]string) {
	if w == nil {
		return
	}
	seen := NewInsMap[string, struct{}]()
	if extendedModuleName != "" {
		seen.Set(extendedModuleName, struct{}{})
	}
	if len(extraExtendedModules) > 0 {
		for _, module := range extraExtendedModules[0] {
			seen.Set(module, struct{}{})
		}
	}
	modules := make([]string, 0, seen.Len())
	for module := range seen.All() {
		modules = append(modules, module)
	}
	w.tlaBuffer.WriteString(SpecWriterExtendingModuleContent(moduleFilename, modules...))
}

func (w *SpecTraceExpressionWriter) CreateAndAddVariablesAndDefinitions(expressions []*Formula, attributeName string) []*TraceExpressionInformation {
	data := NewTraceExpressionInformationHolders(expressions, attributeName)
	w.AddVariablesAndDefinitions(data, attributeName, true)
	return data
}

func (w *SpecTraceExpressionWriter) AddTraceExpressionStub(moduleName string, teName string, variables []string) {
	if w == nil {
		return
	}
	w.tlaBuffer.WriteString(teName)
	w.tlaBuffer.WriteString(tlaDefines)
	w.tlaBuffer.WriteString(tlaCR)
	w.tlaBuffer.WriteString(tlaIndent)
	w.tlaBuffer.WriteString(tlaLSquare)
	w.tlaBuffer.WriteString(tlaCR)
	var local strings.Builder
	local.WriteString(tlaComment)
	local.WriteString(fmt.Sprintf("To hide variables of the `%s` spec from the error trace,", moduleName))
	local.WriteString(tlaCR)
	local.WriteString(tlaComment)
	local.WriteString("remove the variables below.  The trace will be written in the order")
	local.WriteString(tlaCR)
	local.WriteString(tlaComment)
	local.WriteString("of the fields of this record.")
	local.WriteString(tlaCR)
	fields := make([]string, 0, len(variables))
	for _, variable := range variables {
		fields = append(fields, variable+tlaRecordArrowSp+variable)
	}
	local.WriteString(strings.Join(fields, tlaCR+tlaComma))
	local.WriteString(tlaCR)
	local.WriteString(tlaCR)
	local.WriteString(tlaComment)
	local.WriteString("Put additional constant-, state-, and action-level expressions here:")
	local.WriteString(tlaCR)
	local.WriteString(tlaComment)
	local.WriteString(",_stateNumber |-> _TEPosition")
	local.WriteString(tlaCR)
	if len(variables) > 0 {
		someVar := variables[0]
		local.WriteString(tlaComment)
		local.WriteString(fmt.Sprintf(",_%sUnchanged |-> %s = %s'", someVar, someVar, someVar))
		local.WriteString(tlaCR)
		local.WriteString(tlaCR)
		local.WriteString(tlaComment)
		local.WriteString(fmt.Sprintf("Format the `%s` variable as Json value.", someVar))
		local.WriteString(tlaCR)
		local.WriteString(tlaComment)
		local.WriteString(fmt.Sprintf(",_%sJson |->", someVar))
		local.WriteString(tlaCR)
		local.WriteString(tlaComment)
		local.WriteString(tlaIndent)
		local.WriteString("LET J == INSTANCE Json")
		local.WriteString(tlaCR)
		local.WriteString(tlaComment)
		local.WriteString(tlaIndent)
		local.WriteString(fmt.Sprintf("IN J!ToJson(%s)", someVar))
		local.WriteString(tlaCR)
		local.WriteString(tlaCR)
		local.WriteString(tlaComment)
		local.WriteString("Lastly, you may build expressions over arbitrary sets of states by")
		local.WriteString(tlaCR)
		local.WriteString(tlaComment)
		local.WriteString("leveraging the _TETrace operator.  For example, this is how to")
		local.WriteString(tlaCR)
		local.WriteString(tlaComment)
		local.WriteString("count the number of times a spec variable changed up to the current")
		local.WriteString(tlaCR)
		local.WriteString(tlaComment)
		local.WriteString("state in the trace.")
		local.WriteString(tlaCR)
		local.WriteString(tlaComment)
		local.WriteString(fmt.Sprintf(",_%sModCount |->", someVar))
		local.WriteString(tlaCR)
		local.WriteString(tlaComment)
		local.WriteString(tlaIndent)
		local.WriteString("LET F[s \\in DOMAIN _TETrace] ==")
		local.WriteString(tlaCR)
		local.WriteString(tlaComment)
		local.WriteString(tlaIndent)
		local.WriteString(tlaIndent)
		local.WriteString("IF s = 1 THEN 0")
		local.WriteString(tlaCR)
		local.WriteString(tlaComment)
		local.WriteString(tlaIndent)
		local.WriteString(tlaIndent)
		local.WriteString(fmt.Sprintf("ELSE IF _TETrace[s].%s # _TETrace[s-1].%s", someVar, someVar))
		local.WriteString(tlaCR)
		local.WriteString(tlaComment)
		local.WriteString(tlaIndent)
		local.WriteString(tlaIndent)
		local.WriteString(tlaIndent)
		local.WriteString("THEN 1 + F[s-1] ELSE F[s-1]")
		local.WriteString(tlaCR)
		local.WriteString(tlaComment)
		local.WriteString(tlaIndent)
		local.WriteString("IN F[_TEPosition - 1]")
		local.WriteString(tlaCR)
	}
	w.tlaBuffer.WriteString(SpecTraceExpressionIndentString(local.String(), 2))
	w.tlaBuffer.WriteString(tlaCR)
	w.tlaBuffer.WriteString(tlaIndent)
	w.tlaBuffer.WriteString(tlaRSquare)
	w.tlaBuffer.WriteString(tlaCR)
	w.tlaBuffer.WriteString(tlaCR)
}

func (w *SpecTraceExpressionWriter) AddViewConfig(view string) {
	if w == nil || w.cfgBuffer == nil {
		return
	}
	w.cfgBuffer.WriteString(tlaCR)
	w.cfgBuffer.WriteString(tlaKeywordView)
	w.cfgBuffer.WriteString(tlaCR)
	w.cfgBuffer.WriteString(tlaIndent)
	w.cfgBuffer.WriteString(view)
	w.cfgBuffer.WriteString(tlaCR)
}

func (w *SpecTraceExpressionWriter) AddTraceView(vars string) {
	if w == nil {
		return
	}
	w.AddViewConfig(tlaSpecTEView)
	w.tlaBuffer.WriteString(tlaCR)
	w.tlaBuffer.WriteString(tlaSpecTEView)
	w.tlaBuffer.WriteString(tlaDefinesCR)
	w.tlaBuffer.WriteString(tlaIndent)
	w.tlaBuffer.WriteString(tlaBeginTupleSp)
	w.tlaBuffer.WriteString(vars)
	w.tlaBuffer.WriteString(", IF TLCGet(\"level\") = ")
	w.tlaBuffer.WriteString(tlaSpecTELassoEnd)
	w.tlaBuffer.WriteString(" + 1 THEN ")
	w.tlaBuffer.WriteString(tlaSpecTELassoStart)
	w.tlaBuffer.WriteString(" ELSE TLCGet(\"level\")")
	w.tlaBuffer.WriteString(tlaEndTupleSp)
	w.tlaBuffer.WriteString(tlaCR)
}

func (w *SpecTraceExpressionWriter) AddFooter() {
	if w != nil {
		w.tlaBuffer.WriteString(w.GetTLAModuleClosingTag())
	}
}

func (w *SpecTraceExpressionWriter) AddVariablesAndDefinitions(data []*TraceExpressionInformation, attributeName string, addDefinitions bool) {
	if w == nil || len(data) == 0 {
		return
	}
	var variableDecls strings.Builder
	var definitions strings.Builder
	for i, info := range data {
		variableDecls.WriteString(info.VariableName)
		if i != len(data)-1 {
			variableDecls.WriteString(tlaComma)
		}
		if addDefinitions {
			definitions.WriteString(tlaComment)
			definitions.WriteString("TRACE EXPLORER identifier definition ")
			definitions.WriteString(tlaAttribute)
			definitions.WriteString(attributeName)
			definitions.WriteString(tlaColon)
			definitions.WriteString(fmt.Sprint(i))
			definitions.WriteString(tlaCR)
			definitions.WriteString(info.Identifier)
			definitions.WriteString(tlaDefinesCR)
			definitions.WriteString(info.Expression)
			definitions.WriteString(specWriterClosingSep())
			definitions.WriteString(tlaCR)
		}
	}
	w.tlaBuffer.WriteString(tlaComment)
	w.tlaBuffer.WriteString("TRACE EXPLORER variable declaration ")
	w.tlaBuffer.WriteString(tlaAttribute)
	w.tlaBuffer.WriteString(attributeName)
	w.tlaBuffer.WriteString(tlaCR)
	w.tlaBuffer.WriteString("VARIABLES ")
	w.tlaBuffer.WriteString(variableDecls.String())
	w.tlaBuffer.WriteString(specWriterClosingSep())
	w.tlaBuffer.WriteString(tlaCR)
	if addDefinitions {
		w.tlaBuffer.WriteString(definitions.String())
	}
}

func (w *SpecTraceExpressionWriter) AddProperties(trace []*MCState, specName ...string) {
	if w == nil || len(trace) == 0 {
		return
	}
	finalState := trace[len(trace)-1]
	if finalState != nil && finalState.Stuttering && len(trace) >= 2 {
		w.addStutteringProperty(trace[len(trace)-2])
	} else if finalState != nil && finalState.BackToState && len(trace) >= 2 {
		targetIndex := finalState.StateNumber - 1
		if targetIndex >= 0 && targetIndex < len(trace) {
			w.addBackToStateProperty(trace[len(trace)-2], trace[targetIndex])
		}
	} else {
		w.addInvariant(finalState)
	}
	w.tlaBuffer.WriteString(specWriterClosingSep())
	w.tlaBuffer.WriteString(tlaCR)
	if w.cfgBuffer != nil {
		w.cfgBuffer.WriteString(tlaCR)
		w.cfgBuffer.WriteString(ConfigKeywordCheckDeadlock)
		w.cfgBuffer.WriteString(tlaCR)
		w.cfgBuffer.WriteString(tlaIndent)
		w.cfgBuffer.WriteString(tlaComment)
		w.cfgBuffer.WriteString(ConfigKeywordCheckDeadlock)
		w.cfgBuffer.WriteString(" off because of PROPERTY or INVARIANT above.")
		w.cfgBuffer.WriteString(tlaCR)
		w.cfgBuffer.WriteString(tlaIndent)
		w.cfgBuffer.WriteString(tlaKeywordFalse)
	}
}

func (w *SpecTraceExpressionWriter) addInvariant(finalState *MCState) {
	id := SpecWriterValidIdentifierNoTimestamp(tlaSchemeInvariant)
	if w.cfgBuffer != nil {
		w.cfgBuffer.WriteString(tlaCR)
		w.cfgBuffer.WriteString(tlaKeywordInvariant)
		w.cfgBuffer.WriteString(tlaCR)
		w.cfgBuffer.WriteString(tlaIndent)
		w.cfgBuffer.WriteString(id)
		w.cfgBuffer.WriteString(tlaCR)
	}
	w.tlaBuffer.WriteString(tlaCR)
	w.tlaBuffer.WriteString(id)
	w.tlaBuffer.WriteString(tlaDefinesCR)
	w.tlaBuffer.WriteString(tlaIndent)
	w.tlaBuffer.WriteString(tlaNot)
	w.tlaBuffer.WriteString(tlaLParen)
	w.tlaBuffer.WriteString(tlaCR)
	w.tlaBuffer.WriteString(specTraceExpressionStateConjunction(finalState, "TLCGet(\"level\") = Len(_TETrace)"))
	w.tlaBuffer.WriteString(tlaCR)
	w.tlaBuffer.WriteString(tlaIndent)
	w.tlaBuffer.WriteString(tlaRParen)
}

func (w *SpecTraceExpressionWriter) addStutteringProperty(finalState *MCState) {
	id := SpecWriterValidIdentifierNoTimestamp(tlaSchemeProperty)
	if w.cfgBuffer != nil {
		w.cfgBuffer.WriteString(tlaCR)
		w.cfgBuffer.WriteString(tlaKeywordProperty)
		w.cfgBuffer.WriteString(tlaCR)
		w.cfgBuffer.WriteString(tlaIndent)
		w.cfgBuffer.WriteString(id)
		w.cfgBuffer.WriteString(tlaCR)
	}
	w.tlaBuffer.WriteString(tlaCR)
	w.tlaBuffer.WriteString(id)
	w.tlaBuffer.WriteString(tlaDefinesCR)
	w.tlaBuffer.WriteString(tlaIndent)
	w.tlaBuffer.WriteString(tlaNot)
	w.tlaBuffer.WriteString(tlaEventuallyAlways)
	w.tlaBuffer.WriteString(tlaLParen)
	w.tlaBuffer.WriteString(tlaCR)
	w.tlaBuffer.WriteString(specTraceExpressionStateConjunction(finalState, ""))
	w.tlaBuffer.WriteString(tlaCR)
	w.tlaBuffer.WriteString(tlaIndent)
	w.tlaBuffer.WriteString(tlaRParen)
}

func (w *SpecTraceExpressionWriter) addBackToStateProperty(finalState *MCState, backToState *MCState) {
	id := SpecWriterValidIdentifierNoTimestamp(tlaSchemeProperty)
	if w.cfgBuffer != nil {
		w.cfgBuffer.WriteString(tlaCR)
		w.cfgBuffer.WriteString(tlaKeywordProperty)
		w.cfgBuffer.WriteString(tlaCR)
		w.cfgBuffer.WriteString(tlaIndent)
		w.cfgBuffer.WriteString(id)
		w.cfgBuffer.WriteString(tlaCR)
	}
	w.tlaBuffer.WriteString(tlaCR)
	w.tlaBuffer.WriteString(id)
	w.tlaBuffer.WriteString(tlaDefinesCR)
	var local strings.Builder
	local.WriteString(tlaNot)
	local.WriteString(tlaLParen)
	local.WriteString(tlaLParen)
	local.WriteString(tlaInfOften)
	local.WriteString(tlaLParen)
	local.WriteString(tlaCR)
	local.WriteString(specTraceExpressionStateConjunction(finalState, ""))
	local.WriteString(tlaCR)
	local.WriteString(tlaRParen)
	local.WriteString(tlaRParen)
	local.WriteString(tlaAnd)
	local.WriteString(tlaLParen)
	local.WriteString(tlaInfOften)
	local.WriteString(tlaLParen)
	local.WriteString(tlaCR)
	local.WriteString(specTraceExpressionStateConjunction(backToState, ""))
	local.WriteString(tlaCR)
	local.WriteString(tlaRParen)
	local.WriteString(tlaRParen)
	local.WriteString(tlaRParen)
	w.tlaBuffer.WriteString(SpecTraceExpressionIndentString(local.String(), 1))
}

func (w *SpecTraceExpressionWriter) AddInfoComments(data []*TraceExpressionInformation) {
	if w == nil {
		return
	}
	for _, info := range data {
		w.tlaBuffer.WriteString(tlaComment)
		w.tlaBuffer.WriteString(tlaColon)
		w.tlaBuffer.WriteString(fmt.Sprint(info.Level))
		w.tlaBuffer.WriteString(tlaColon)
		w.tlaBuffer.WriteString(info.VariableName)
		w.tlaBuffer.WriteString(tlaColon)
		w.tlaBuffer.WriteString(info.Expression)
		w.tlaBuffer.WriteString(tlaConstantEvalMarker)
		w.tlaBuffer.WriteString(tlaCR)
	}
}

func (w *SpecTraceExpressionWriter) AddInitNext(trace []*MCState, args ...any) []string {
	if w == nil {
		return nil
	}
	if len(args) == 0 {
		return SpecTraceExpressionAddInitNextToBuffers(&w.tlaBuffer, w.cfgBuffer, trace, nil)
	}
	var data []*TraceExpressionInformation
	argIndex := 0
	if typed, ok := args[0].([]*TraceExpressionInformation); ok {
		data = typed
		argIndex++
	}
	if len(args)-argIndex >= 3 {
		initID, _ := args[argIndex].(string)
		nextID, _ := args[argIndex+1].(string)
		actionID, _ := args[argIndex+2].(string)
		if len(args)-argIndex >= 4 {
			base, _ := args[argIndex+3].(string)
			SpecTraceExpressionAddInitNextFullToBuffers(&w.tlaBuffer, w.cfgBuffer, trace, data, initID, nextID, actionID, base, true)
		} else {
			SpecTraceExpressionAddInitNextIDToBuffers(&w.tlaBuffer, w.cfgBuffer, trace, data, initID, nextID, actionID)
		}
	}
	return nil
}

func (w *SpecTraceExpressionWriter) AddInitNextTraceFunction(trace []*MCState, teSpecModuleName string, vars []string, modelConfig *ModelConfig) {
	_ = modelConfig
	if w == nil || len(trace) == 0 {
		return
	}
	if w.cfgBuffer != nil {
		w.cfgBuffer.WriteString(tlaCR)
		w.cfgBuffer.WriteString(tlaCR)
		w.cfgBuffer.WriteString(tlaKeywordInit)
		w.cfgBuffer.WriteString(tlaCR)
		w.cfgBuffer.WriteString(tlaIndent)
		w.cfgBuffer.WriteString(tlaSpecTEInit)
		w.cfgBuffer.WriteString(tlaCR)
	}
	w.tlaBuffer.WriteString(tlaSpecTEInit)
	w.tlaBuffer.WriteString(tlaDefinesCR)
	for _, variable := range vars {
		w.tlaBuffer.WriteString(tlaIndentedConj)
		w.tlaBuffer.WriteString(variable)
		w.tlaBuffer.WriteString(tlaEqSp)
		w.tlaBuffer.WriteString("_TETrace[1].")
		w.tlaBuffer.WriteString(variable)
		w.tlaBuffer.WriteString(tlaCR)
	}
	w.tlaBuffer.WriteString(tlaSep)
	w.tlaBuffer.WriteString(tlaCR)
	w.tlaBuffer.WriteString(tlaCR)
	finalState := trace[len(trace)-1]
	nextID := tlaSpecTENext
	if w.cfgBuffer != nil {
		w.cfgBuffer.WriteString(tlaCR)
		w.cfgBuffer.WriteString(tlaKeywordNext)
		w.cfgBuffer.WriteString(tlaCR)
		w.cfgBuffer.WriteString(tlaIndent)
		w.cfgBuffer.WriteString(nextID)
		w.cfgBuffer.WriteString(tlaCR)
	}
	w.tlaBuffer.WriteString(nextID)
	w.tlaBuffer.WriteString(tlaDefinesCR)
	if len(trace) == 1 {
		w.tlaBuffer.WriteString(tlaIndent)
		w.tlaBuffer.WriteString(tlaIndentedConj)
		w.tlaBuffer.WriteString(tlaKeywordFalse)
		w.tlaBuffer.WriteString(tlaCR)
	} else {
		w.tlaBuffer.WriteString(tlaIndentedConj)
		w.tlaBuffer.WriteString("\\E i,j \\in DOMAIN _TETrace:")
		w.tlaBuffer.WriteString(tlaCR)
		w.tlaBuffer.WriteString(tlaIndent)
		w.tlaBuffer.WriteString(tlaIndentedConj)
		w.tlaBuffer.WriteString(tlaOr)
		w.tlaBuffer.WriteString(tlaSpace)
		w.tlaBuffer.WriteString(tlaAnd)
		w.tlaBuffer.WriteString(" j = i + 1")
		w.tlaBuffer.WriteString(tlaCR)
		w.tlaBuffer.WriteString(tlaIndent)
		w.tlaBuffer.WriteString(tlaIndent)
		w.tlaBuffer.WriteString(tlaSpace)
		w.tlaBuffer.WriteString(tlaSpace)
		w.tlaBuffer.WriteString(tlaIndentedConj)
		w.tlaBuffer.WriteString("i = TLCGet(\"level\")")
		w.tlaBuffer.WriteString(tlaCR)
		if finalState != nil && finalState.BackToState {
			w.tlaBuffer.WriteString(tlaIndent)
			w.tlaBuffer.WriteString(tlaIndent)
			w.tlaBuffer.WriteString("   ")
			w.tlaBuffer.WriteString(tlaOr)
			w.tlaBuffer.WriteString(tlaSpace)
			w.tlaBuffer.WriteString(tlaAnd)
			w.tlaBuffer.WriteString(" i = ")
			w.tlaBuffer.WriteString(tlaSpecTELassoEnd)
			w.tlaBuffer.WriteString(tlaCR)
			w.tlaBuffer.WriteString(tlaIndent)
			w.tlaBuffer.WriteString(tlaIndent)
			w.tlaBuffer.WriteString("  ")
			w.tlaBuffer.WriteString(tlaIndentedConj)
			w.tlaBuffer.WriteString("j = ")
			w.tlaBuffer.WriteString(tlaSpecTELassoStart)
			w.tlaBuffer.WriteString(tlaCR)
		}
		for _, variable := range vars {
			w.tlaBuffer.WriteString(tlaIndent)
			w.tlaBuffer.WriteString(tlaIndentedConj)
			w.tlaBuffer.WriteString(variable)
			w.tlaBuffer.WriteString(" ")
			w.tlaBuffer.WriteString(tlaEqSp)
			w.tlaBuffer.WriteString("_TETrace[i].")
			w.tlaBuffer.WriteString(variable)
			w.tlaBuffer.WriteString(tlaCR)
			w.tlaBuffer.WriteString(tlaIndent)
			w.tlaBuffer.WriteString(tlaIndentedConj)
			w.tlaBuffer.WriteString(variable)
			w.tlaBuffer.WriteString(tlaPrime)
			w.tlaBuffer.WriteString(tlaEqSp)
			w.tlaBuffer.WriteString("_TETrace[j].")
			w.tlaBuffer.WriteString(variable)
			w.tlaBuffer.WriteString(tlaCR)
		}
	}
	jsonComment := tlaIndent + tlaComment
	if _, ok := tlcLookupSystemProperty("TLC_TRACE_EXPLORER_JSON_UNCOMMENTED"); ok {
		jsonComment = ""
	}
	w.tlaBuffer.WriteString(tlaCR)
	w.tlaBuffer.WriteString(tlaComment)
	w.tlaBuffer.WriteString("Uncomment the ASSUME below to write the states of the error trace")
	w.tlaBuffer.WriteString(tlaCR)
	w.tlaBuffer.WriteString(tlaComment)
	w.tlaBuffer.WriteString("to the given file in Json format. Note that you can pass any tuple")
	w.tlaBuffer.WriteString(tlaCR)
	w.tlaBuffer.WriteString(tlaComment)
	w.tlaBuffer.WriteString("to `JsonSerialize`. For example, a sub-sequence of _TETrace.")
	w.tlaBuffer.WriteString(tlaCR)
	w.tlaBuffer.WriteString(jsonComment)
	w.tlaBuffer.WriteString(tlaKeywordAssume)
	w.tlaBuffer.WriteString(tlaCR)
	w.tlaBuffer.WriteString(jsonComment)
	w.tlaBuffer.WriteString(tlaIndent)
	w.tlaBuffer.WriteString("LET J == INSTANCE Json")
	w.tlaBuffer.WriteString(tlaCR)
	w.tlaBuffer.WriteString(jsonComment)
	w.tlaBuffer.WriteString(tlaIndent)
	w.tlaBuffer.WriteString(tlaIndent)
	w.tlaBuffer.WriteString(fmt.Sprintf("IN J!JsonSerialize(\"%s.json\", _TETrace)", teSpecModuleName))
	w.tlaBuffer.WriteString(tlaCR)
	w.tlaBuffer.WriteString(tlaCR)
}

func (w *SpecTraceExpressionWriter) AddTraceFunction(input []*MCState, ids ...string) string {
	if w == nil {
		return ""
	}
	id := ""
	configID := ""
	if len(ids) >= 2 {
		id = ids[0]
		configID = ids[1]
	} else {
		id = SpecWriterValidIdentifier(tlaSchemeDefOverride)
		configID = id
	}
	return SpecTraceExpressionAddTraceFunctionToBuffers(&w.tlaBuffer, w.cfgBuffer, input, id, configID)
}

func (w *SpecTraceExpressionWriter) AddTraceFunctionInstance(moduleName string) {
	if w == nil {
		return
	}
	w.tlaBuffer.WriteString(fmt.Sprintf("%s ==%s%sLET %s == INSTANCE %s%s%sIN %s!%s",
		tlaSpecTETrace, tlaCR, tlaIndent, moduleName, moduleName, tlaCR, tlaIndent, moduleName, tlaSpecTETraceDef))
	w.tlaBuffer.WriteString(tlaCR)
	w.tlaBuffer.WriteString(tlaSep)
	w.tlaBuffer.WriteString(tlaCR)
}

func (w *SpecTraceExpressionWriter) AddTraceExpressionInstance(moduleName string) {
	if w == nil {
		return
	}
	w.tlaBuffer.WriteString(fmt.Sprintf("%s ==%s%sLET %s == INSTANCE %s%s%sIN %s!%s",
		tlaSpecTETTraceExpr, tlaCR, tlaIndent, moduleName, moduleName, tlaCR, tlaIndent, moduleName, tlaSpecTEExpression))
	w.tlaBuffer.WriteString(tlaCR)
	w.tlaBuffer.WriteString(tlaSep)
	w.tlaBuffer.WriteString(tlaCR)
	w.tlaBuffer.WriteString(tlaCR)
}

func (w *SpecTraceExpressionWriter) Append(text string) *strings.Builder {
	if w == nil {
		return nil
	}
	w.tlaBuffer.WriteString(text)
	return &w.tlaBuffer
}

func (w *SpecTraceExpressionWriter) String() string {
	if w == nil {
		return ""
	}
	return w.tlaBuffer.String()
}

func (w *SpecTraceExpressionWriter) GetComment() string {
	if w == nil {
		return ""
	}
	return tlaCommentNS + strings.ReplaceAll(w.tlaBuffer.String(), tlaCR, tlaCR+tlaCommentNS)
}

func SpecTraceExpressionIndentString(text string, n int) string {
	indent := strings.Repeat(tlaIndent, n)
	lines := strings.Split(text, tlaCR)
	for len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return indent + strings.Join(lines, tlaCR+indent)
}

func (w *SpecTraceExpressionWriter) WrapConfig(moduleFilename string) {
	if w == nil || w.cfgBuffer == nil {
		return
	}
	original := w.cfgBuffer.String()
	w.cfgBuffer.Reset()
	w.cfgBuffer.WriteString(tlaCR)
	w.cfgBuffer.WriteString(tlaSep)
	w.cfgBuffer.WriteString(" CONFIG ")
	w.cfgBuffer.WriteString(moduleFilename)
	w.cfgBuffer.WriteString(" ")
	w.cfgBuffer.WriteString(tlaSep)
	w.cfgBuffer.WriteString(tlaCR)
	w.cfgBuffer.WriteString(original)
	w.cfgBuffer.WriteString(w.GetTLAModuleClosingTag())
}

type TraceExpressionExplorerSpecWriter struct {
	*SpecWriter
	VariableExpressionMap *InsMap[string, string]
}

func NewTraceExpressionExplorerSpecWriter(expressions []string) *TraceExpressionExplorerSpecWriter {
	w := &TraceExpressionExplorerSpecWriter{
		SpecWriter:            NewSpecWriter(true),
		VariableExpressionMap: NewInsMap[string, string](),
	}
	for i, expression := range expressions {
		key := fmt.Sprintf("%s%d", tlaTraceExprPrefix, i+1)
		w.VariableExpressionMap.Set(key, expression)
		w.tlaBuffer.WriteString(tlaTraceExprComment)
		w.tlaBuffer.WriteString(key)
		w.tlaBuffer.WriteString(tlaEqSp)
		w.tlaBuffer.WriteString(expression)
		w.tlaBuffer.WriteString(tlaCR)
	}
	w.AddPrimer(tlaTraceExploreModule, tlaTraceExprModuleName)
	w.declareExpressionVariables()
	w.createInitNextWithExpressions()
	w.tlaBuffer.WriteString(SpecWriterGeneratedTimestampLine())
	w.tlaBuffer.WriteString(tlaCR)
	return w
}

func TraceExpressionVariableExpressionMapFromTLA(content string) map[string]string {
	out := make(map[string]string)
	lines := strings.Split(content, tlaCR)
	for _, line := range lines {
		if !strings.HasPrefix(line, tlaTraceExprComment) {
			break
		}
		decl := strings.TrimPrefix(line, tlaTraceExprComment)
		index := strings.Index(decl, tlaEqSp)
		if index < 0 {
			continue
		}
		out[decl[:index]] = decl[index+len(tlaEqSp):]
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func (w *TraceExpressionExplorerSpecWriter) ConfigBufferString() string {
	if w == nil || w.cfgBuffer == nil {
		return ""
	}
	return w.cfgBuffer.String()
}

func (w *TraceExpressionExplorerSpecWriter) declareExpressionVariables() {
	if w == nil {
		return
	}
	w.tlaBuffer.WriteString("VARIABLES ")
	first := true
	for _, entry := range w.sortedVariableExpressionEntries() {
		if !first {
			w.tlaBuffer.WriteString(", ")
		}
		first = false
		w.tlaBuffer.WriteString(entry.variable)
	}
	w.tlaBuffer.WriteString(tlaCR)
	w.tlaBuffer.WriteString(tlaCR)
}

func (w *TraceExpressionExplorerSpecWriter) createInitNextWithExpressions() {
	if w == nil {
		return
	}
	var initConj strings.Builder
	initConj.WriteString(tlaIndentedConj)
	initConj.WriteString(tlaSpecTEInitID)
	initConj.WriteString(tlaCR)
	w.addExpressionsToBuffer(&initConj, false)
	initContent := SpecWriterCreateSourceContent(initConj.String(), tlaTEInitID, false)
	w.AddFormulaList(initContent, tlaKeywordInit, tlaTEInitAttr)

	var nextConj strings.Builder
	nextConj.WriteString(tlaIndentedConj)
	nextConj.WriteString(tlaSpecTENextID)
	nextConj.WriteString(tlaCR)
	w.addExpressionsToBuffer(&nextConj, true)
	nextContent := SpecWriterCreateSourceContent(nextConj.String(), tlaTENextID, false)
	w.AddFormulaList(nextContent, tlaKeywordNext, tlaTENextAttr)
}

func (w *TraceExpressionExplorerSpecWriter) addExpressionsToBuffer(buffer *strings.Builder, primed bool) {
	if w == nil || buffer == nil {
		return
	}
	for _, entry := range w.sortedVariableExpressionEntries() {
		buffer.WriteString(tlaIndentedConj)
		buffer.WriteString(entry.variable)
		if primed {
			buffer.WriteString(tlaPrime)
		}
		buffer.WriteString(tlaEqSp)
		buffer.WriteString(tlaLParen)
		buffer.WriteString(entry.expression)
		buffer.WriteString(tlaRParen)
		buffer.WriteString(tlaCR)
	}
}

type traceExpressionEntry struct {
	variable   string
	expression string
}

func (w *TraceExpressionExplorerSpecWriter) sortedVariableExpressionEntries() []traceExpressionEntry {
	if w == nil || w.VariableExpressionMap == nil {
		return nil
	}
	entries := make([]traceExpressionEntry, 0, w.VariableExpressionMap.Len())
	for variable, expression := range w.VariableExpressionMap.All() {
		entries = append(entries, traceExpressionEntry{variable: variable, expression: expression})
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].variable < entries[j].variable
	})
	return entries
}

type TraceExplorationSpec struct {
	OutputPath        string
	OriginalModule    string
	TESpecModuleName  string
	PostConditionBins []PostCondition
}

type PostCondition struct {
	Module string
	Op     string
	Key    string
	Value  string
}

func NewTraceExplorationSpec(outputPath string, timestamp time.Time, originalModuleName string) *TraceExplorationSpec {
	return NewTraceExplorationSpecNamed(outputPath, DeriveTESpecModuleName(originalModuleName, timestamp), originalModuleName)
}

func NewTraceExplorationSpecNamed(outputPath string, teModuleName string, originalModuleName string) *TraceExplorationSpec {
	spec := &TraceExplorationSpec{OutputPath: outputPath, TESpecModuleName: teModuleName, OriginalModule: originalModuleName}
	spec.PostConditionBins = append(spec.PostConditionBins, PostCondition{
		Module: "_TLCTrace",
		Op:     "_TLCTraceSilent",
		Key:    "_TLCTraceFile",
		Value:  filepath.Join(outputPath, teModuleName+".bin"),
	})
	return spec
}

func (s *TraceExplorationSpec) Generate(tool *Tool, errTrace *MCError) (string, error) {
	if s == nil || errTrace == nil || len(errTrace.States) <= 1 {
		return "", nil
	}
	if s.OutputPath == "" {
		s.OutputPath = "."
	}
	if s.OriginalModule == "" && tool != nil {
		s.OriginalModule = tool.GetRootName()
	}
	if s.OriginalModule == "" {
		s.OriginalModule = "Spec"
	}
	if s.TESpecModuleName == "" {
		s.TESpecModuleName = DeriveTESpecModuleName(s.OriginalModule, time.Now())
	}
	if err := os.MkdirAll(s.OutputPath, 0o755); err != nil {
		PrintMessage(ECTLCTESpecGenerationError, err.Error())
		return "", err
	}
	path := filepath.Join(s.OutputPath, s.TESpecModuleName+".tla")
	file, err := os.Create(path)
	if err != nil {
		PrintMessage(ECTLCTESpecGenerationError, err.Error())
		return "", err
	}
	defer file.Close()
	if err := s.WriteSpecTE(tool, errTrace, file); err != nil {
		PrintMessage(ECTLCTESpecGenerationError, err.Error())
		return "", err
	}
	PrintMessage(ECTLCTESpecGenerationComplete, path)
	return path, nil
}

func (s *TraceExplorationSpec) WriteSpecTE(tool *Tool, errTrace *MCError, out io.Writer) error {
	if s == nil || errTrace == nil || out == nil || len(errTrace.States) == 0 {
		return nil
	}
	if s.OriginalModule == "" && tool != nil {
		s.OriginalModule = tool.GetRootName()
	}
	if s.OriginalModule == "" {
		s.OriginalModule = "Spec"
	}
	if s.TESpecModuleName == "" {
		s.TESpecModuleName = DeriveTESpecModuleName(s.OriginalModule, time.Now())
	}
	modelConfig := newModelConfig("", false)
	if tool != nil && tool.GetModelConfig() != nil {
		modelConfig = tool.GetModelConfig()
	}
	variables := traceExplorationVariables(errTrace)
	writer := NewSpecTraceExpressionWriter()

	constants := modelConfig.GetConstantsAsList()
	declaredConstantNames := make(map[string]struct{}, len(constants))
	var indentedConstants []string
	for _, entry := range constants {
		if len(entry) == 0 {
			continue
		}
		declaredConstantNames[entry[0]] = struct{}{}
		line := entry[0]
		if len(entry) > 1 {
			line = entry[0] + tlaEqSp + entry[1]
		}
		indentedConstants = append(indentedConstants, SpecTraceExpressionIndentString(line, 1))
	}
	SetModelValues()
	for _, mv := range ModelValues() {
		if mv == nil {
			continue
		}
		name := mv.String()
		indentedConstants = append(indentedConstants, SpecTraceExpressionIndentString(name+tlaEqSp+name, 1))
	}
	if errTrace.IsLasso() {
		last := errTrace.States[len(errTrace.States)-1]
		indentedConstants = append(indentedConstants, tlaSpecTELassoStart+tlaEqSp+fmt.Sprint(last.StateNumber))
		indentedConstants = append(indentedConstants, tlaSpecTELassoEnd+tlaEqSp+fmt.Sprint(len(errTrace.States)-1))
	}
	writer.AddConstantsRaw(indentedConstants)

	var modConstants []string
	for _, mv := range ModelValues() {
		if mv == nil {
			continue
		}
		name := mv.String()
		if _, declared := declaredConstantNames[name]; !declared {
			modConstants = append(modConstants, name)
		}
	}
	if errTrace.IsLasso() {
		modConstants = append(modConstants, tlaSpecTELassoStart, tlaSpecTELassoEnd)
	}
	teConstantSpecName := fmt.Sprintf("%s_%s", s.OriginalModule, tlaSpecTEConstantsName)
	teConstantModules := []string{tlaModuleTLC}
	modelValuesAsConstants := ""
	if len(modConstants) != 0 {
		teConstantModules = append(teConstantModules, teConstantSpecName)
		modelValuesAsConstants = "CONSTANTS " + strings.Join(modConstants, ", ") + tlaCR
	}

	specTEExtendedModules := traceExploreSortedUnique(tlaModuleToolbox, tlaModuleTLCExt, tlaModuleNaturals, tlaModuleSequences)
	specTEExtendedModules = traceExploreSortedUnique(append(specTEExtendedModules, teConstantModules...)...)
	writer.AddPrimer(s.TESpecModuleName, s.OriginalModule, specTEExtendedModules)
	writer.AddTraceExpressionInstance(fmt.Sprintf("%s_%s", s.OriginalModule, tlaTraceExploreModule))
	teTraceName := fmt.Sprintf("%s_%s", s.OriginalModule, tlaSpecTETraceModuleName)
	writer.AddTraceFunctionInstance(teTraceName)
	writer.AddProperties(errTrace.States, s.OriginalModule)
	writer.AddInitNextTraceFunction(errTrace.States, s.TESpecModuleName, variables, modelConfig)
	if errTrace.IsLassoWithDuplicates() {
		writer.AddTraceView(strings.Join(variables, ", "))
	}
	writer.AddFooter()

	writer.Append(tlaCR)
	te := NewSpecTraceExpressionWriter()
	teModuleName := fmt.Sprintf("%s_%s", s.OriginalModule, tlaTraceExploreModule)
	te.Append(tlaCR)
	writer.Append(fmt.Sprintf(" Note that you can extract this module `%s`", teModuleName)).WriteString(tlaCR)
	writer.Append("  to a dedicated file to reuse `expression` (the module in the ").WriteString(tlaCR)
	writer.Append(fmt.Sprintf("  dedicated `%s.tla` file takes precedence ", teModuleName)).WriteString(tlaCR)
	writer.Append(fmt.Sprintf("  over the module `%s` below).", teModuleName))
	te.AddPrimer(teModuleName, s.OriginalModule, specTEExtendedModules)
	te.AddTraceExpressionStub(s.OriginalModule, tlaSpecTEExpression, variables)
	te.AddFooter()
	writer.Append(tlaCR + te.String() + tlaCR + tlaCR)

	writer.Append(tlaCR)
	writer.Append("Parsing and semantic processing can take forever if the trace below is long.").WriteString(tlaCR)
	writer.Append(" In this case, it is advised to uncomment the module below to deserialize the").WriteString(tlaCR)
	writer.Append(" trace from a generated binary file.").WriteString(tlaCR)
	overrideModules := traceExploreSortedUnique(tlaModuleIOUtils)
	overrideModules = traceExploreSortedUnique(append(overrideModules, teConstantModules...)...)
	overrideWriter := NewSpecTraceExpressionWriter()
	overrideWriter.Append(tlaCR)
	overrideWriter.AddPrimer(teTraceName, s.OriginalModule, overrideModules)
	overrideWriter.Append(tlaSpecTETraceDef).WriteString(tlaDefines)
	overrideWriter.Append(fmt.Sprintf("IODeserialize(\"%s%s\", TRUE)\n\n", s.TESpecModuleName, ".bin"))
	overrideWriter.AddFooter()
	writer.Append(tlaCR + overrideWriter.GetComment() + tlaCR + tlaCR)

	writer.AddPrimer(teTraceName, s.OriginalModule, traceExploreSortedUnique(teConstantModules...))
	writer.AddTraceFunction(errTrace.States, tlaSpecTETraceDef, tlaSpecTETrace)
	writer.AddAliasToCfg(tlaSpecTETTraceExpr)
	if modelValuesAsConstants != "" {
		writer.AddFooter()
		writer.Append(tlaCR)
		writer.AddPrimer(teConstantSpecName, s.OriginalModule)
		writer.Append(modelValuesAsConstants).WriteString(tlaCR)
	}
	writer.WrapConfig(s.TESpecModuleName)
	return writer.WriteStreams(out, out)
}

func TraceExplorationModuleID(timestamp time.Time) string {
	return fmt.Sprint(timestamp.Unix())
}

func DeriveTESpecModuleName(originalModuleName string, timestamp time.Time) string {
	return fmt.Sprintf("%s_%s_%s", originalModuleName, tlaTraceExprModuleName, TraceExplorationModuleID(timestamp))
}

func IsTESpecFile(tlaFilePath string) bool {
	if tlaFilePath == "" {
		return false
	}
	base := filepath.Base(tlaFilePath)
	ok, _ := regexp.MatchString("^.*_"+tlaTraceExprModuleName+".*(.tla)?$", base)
	return ok
}

func traceExplorationVariables(errTrace *MCError) []string {
	vars := StateVariables()
	if len(vars) != 0 {
		out := make([]string, 0, len(vars))
		for _, variable := range vars {
			if variable.Name != nil {
				out = append(out, variable.Name.String())
			}
		}
		return out
	}
	if errTrace == nil {
		return nil
	}
	for _, state := range errTrace.States {
		if state == nil || state.BackToState || state.Stuttering {
			continue
		}
		out := make([]string, 0, len(state.Variables))
		for _, variable := range state.Variables {
			if variable != nil {
				out = append(out, variable.Name)
			}
		}
		return out
	}
	return nil
}

func traceExploreSortedUnique(names ...string) []string {
	seen := make(map[string]struct{}, len(names))
	out := make([]string, 0, len(names))
	for _, name := range names {
		if name == "" {
			continue
		}
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		out = append(out, name)
	}
	sort.Strings(out)
	return out
}

func specTraceExpressionStateConjunction(state *MCState, prefixConjunct string) string {
	if state == nil || state.BackToState || state.Stuttering {
		return ""
	}
	var formula strings.Builder
	if prefixConjunct != "" {
		formula.WriteString(prefixConjunct)
		formula.WriteString(tlaCR)
		formula.WriteString(tlaAnd)
		formula.WriteString(tlaCR)
	}
	for i, variable := range state.Variables {
		formula.WriteString(variable.Name)
		formula.WriteString(tlaEqSp)
		formula.WriteString(tlaLParen)
		formula.WriteString(variable.ValueAsString)
		formula.WriteString(tlaRParen)
		if i != len(state.Variables)-1 {
			formula.WriteString(tlaCR)
			formula.WriteString(tlaAnd)
			formula.WriteString(tlaCR)
		}
	}
	return SpecTraceExpressionIndentString(formula.String(), 2)
}

func specWriterClosingSep() string {
	return tlaCR + tlaSep + tlaCR
}

func tlaTraceTripleIndent() string {
	return tlaIndent + tlaIndent + tlaIndent
}

func tlaTraceExploreTrace() string {
	return "_TETrace"
}
