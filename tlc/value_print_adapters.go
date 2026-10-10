package tlc

// StringWithDelimiter and StringUnchecked mirror Java's final Value string
// overloads. Operator subclasses inherit these through operatorValueBase;
// receiver() retains Java's virtual dispatch to the concrete printer.

func (v *BoolValue) StringWithDelimiter(delimiter string) string {
	return ValueToString(v, delimiter, true)
}

func (v *BoolValue) StringUnchecked(delimiter ...string) string {
	return ValueToStringUnchecked(v, delimiter...)
}

func (v *IntValue) StringWithDelimiter(delimiter string) string {
	return ValueToString(v, delimiter, true)
}

func (v *IntValue) StringUnchecked(delimiter ...string) string {
	return ValueToStringUnchecked(v, delimiter...)
}

func (v *StringValue) StringWithDelimiter(delimiter string) string {
	return ValueToString(v.receiver(), delimiter, true)
}

func (v *StringValue) StringUnchecked(delimiter ...string) string {
	return ValueToStringUnchecked(v.receiver(), delimiter...)
}

func (v *ModelValue) StringWithDelimiter(delimiter string) string {
	return ValueToString(v, delimiter, true)
}

func (v *ModelValue) StringUnchecked(delimiter ...string) string {
	return ValueToStringUnchecked(v, delimiter...)
}

func (v *UndefValue) StringWithDelimiter(delimiter string) string {
	return ValueToString(v, delimiter, true)
}

func (v *UndefValue) StringUnchecked(delimiter ...string) string {
	return ValueToStringUnchecked(v, delimiter...)
}

func (v *UserValue) StringWithDelimiter(delimiter string) string {
	return ValueToString(v, delimiter, true)
}

func (v *UserValue) StringUnchecked(delimiter ...string) string {
	return ValueToStringUnchecked(v, delimiter...)
}

func (v *TupleValue) StringWithDelimiter(delimiter string) string {
	return ValueToString(v, delimiter, true)
}

func (v *TupleValue) StringUnchecked(delimiter ...string) string {
	return ValueToStringUnchecked(v, delimiter...)
}

func (v *SetEnumValue) StringWithDelimiter(delimiter string) string {
	return ValueToString(v, delimiter, true)
}

func (v *SetEnumValue) StringUnchecked(delimiter ...string) string {
	return ValueToStringUnchecked(v, delimiter...)
}

func (v *IntervalValue) StringWithDelimiter(delimiter string) string {
	return ValueToString(v, delimiter, true)
}

func (v *IntervalValue) StringUnchecked(delimiter ...string) string {
	return ValueToStringUnchecked(v, delimiter...)
}

func (v *RecordValue) StringWithDelimiter(delimiter string) string {
	return ValueToString(v, delimiter, true)
}

func (v *RecordValue) StringUnchecked(delimiter ...string) string {
	return ValueToStringUnchecked(v, delimiter...)
}

func (v *FcnRcdValue) StringWithDelimiter(delimiter string) string {
	return ValueToString(v, delimiter, true)
}

func (v *FcnRcdValue) StringUnchecked(delimiter ...string) string {
	return ValueToStringUnchecked(v, delimiter...)
}

func (v *SetCupValue) StringWithDelimiter(delimiter string) string {
	return ValueToString(v, delimiter, true)
}

func (v *SetCupValue) StringUnchecked(delimiter ...string) string {
	return ValueToStringUnchecked(v, delimiter...)
}

func (v *SetCapValue) StringWithDelimiter(delimiter string) string {
	return ValueToString(v, delimiter, true)
}

func (v *SetCapValue) StringUnchecked(delimiter ...string) string {
	return ValueToStringUnchecked(v, delimiter...)
}

func (v *SetDiffValue) StringWithDelimiter(delimiter string) string {
	return ValueToString(v, delimiter, true)
}

func (v *SetDiffValue) StringUnchecked(delimiter ...string) string {
	return ValueToStringUnchecked(v, delimiter...)
}

func (v *UnionValue) StringWithDelimiter(delimiter string) string {
	return ValueToString(v, delimiter, true)
}

func (v *UnionValue) StringUnchecked(delimiter ...string) string {
	return ValueToStringUnchecked(v, delimiter...)
}

func (v *SetOfTuplesValue) StringWithDelimiter(delimiter string) string {
	return ValueToString(v, delimiter, true)
}

func (v *SetOfTuplesValue) StringUnchecked(delimiter ...string) string {
	return ValueToStringUnchecked(v, delimiter...)
}

func (v *SetOfRcdsValue) StringWithDelimiter(delimiter string) string {
	return ValueToString(v, delimiter, true)
}

func (v *SetOfRcdsValue) StringUnchecked(delimiter ...string) string {
	return ValueToStringUnchecked(v, delimiter...)
}

func (v *SetOfFcnsValue) StringWithDelimiter(delimiter string) string {
	return ValueToString(v, delimiter, true)
}

func (v *SetOfFcnsValue) StringUnchecked(delimiter ...string) string {
	return ValueToStringUnchecked(v, delimiter...)
}

func (v *SubsetValue) StringWithDelimiter(delimiter string) string {
	return ValueToString(v, delimiter, true)
}

func (v *SubsetValue) StringUnchecked(delimiter ...string) string {
	return ValueToStringUnchecked(v, delimiter...)
}

func (v *KSubsetValue) StringWithDelimiter(delimiter string) string {
	return ValueToString(v, delimiter, true)
}

func (v *KSubsetValue) StringUnchecked(delimiter ...string) string {
	return ValueToStringUnchecked(v, delimiter...)
}

func (v *LazyValue) StringWithDelimiter(delimiter string) string {
	return ValueToString(v, delimiter, true)
}

func (v *LazyValue) StringUnchecked(delimiter ...string) string {
	return ValueToStringUnchecked(v, delimiter...)
}

func (v *SetPredValue) StringWithDelimiter(delimiter string) string {
	return ValueToString(v, delimiter, true)
}

func (v *SetPredValue) StringUnchecked(delimiter ...string) string {
	return ValueToStringUnchecked(v, delimiter...)
}

func (v *FcnLambdaValue) StringWithDelimiter(delimiter string) string {
	return ValueToString(v, delimiter, true)
}

func (v *FcnLambdaValue) StringUnchecked(delimiter ...string) string {
	return ValueToStringUnchecked(v, delimiter...)
}

func (b *operatorValueBase) StringWithDelimiter(delimiter string) string {
	return ValueToString(b.receiver(), delimiter, true)
}

func (b *operatorValueBase) StringUnchecked(delimiter ...string) string {
	return ValueToStringUnchecked(b.receiver(), delimiter...)
}
