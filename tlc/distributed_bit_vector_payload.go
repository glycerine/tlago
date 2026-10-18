package tlc

// BitVector's source fields retain their complete backing array, including
// trailing zero words and aliases. Word storage uses the primitive-array graph.
func (e *distributedPayloadEncoder) bitVector(vector *BitVector) int {
	if vector == nil {
		return 0
	}
	if e.bitVectors == nil {
		e.bitVectors = make(map[*BitVector]int)
	}
	if id := e.bitVectors[vector]; id != 0 {
		return id
	}
	var words DistributedValueNode
	encodeDistributedPrimitiveArray(e, &words, vector.word, "uint64Array", func(v uint64) uint64 { return v })
	e.payload.BitVectors = append(e.payload.BitVectors, words.DataArray)
	id := len(e.payload.BitVectors)
	e.bitVectors[vector] = id
	return id
}
