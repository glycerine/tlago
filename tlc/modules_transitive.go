package tlc

func Warshall(rel Value) (Value, error) {
	enumerable, ok := asEnumerable(rel)
	if !ok {
		return nil, newTLCErrorCode(ECTLCModuleApplyingToWrongValue, "TransitiveClosure", "an enumerable set", ValuesPPR(rel))
	}
	size, err := rel.Size()
	if err != nil {
		return nil, err
	}
	matrix := make([][]bool, 2*size)
	for i := range matrix {
		matrix[i] = make([]bool, 2*size)
	}
	elemList := NewValueVec(0)
	buckets := make(map[int32][]int)
	indexOf := func(value Value) (int, error) {
		hash := ValueJavaHashCode(value)
		for _, idx := range buckets[hash] {
			eq, err := elemList.At(idx).Equal(value)
			if err != nil {
				return 0, err
			}
			if eq {
				return idx, nil
			}
		}
		idx := elemList.Len()
		elemList.Add(value)
		buckets[hash] = append(buckets[hash], idx)
		return idx, nil
	}
	enum := enumerable.Elements()
	for elem := enum.NextElement(); elem != nil; elem = enum.NextElement() {
		tuple := asTupleValue(elem)
		if tuple == nil || len(tuple.Elems) != 2 {
			return nil, newTLCErrorCode(ECTLCModuleTransitiveClosure, ValuesPPR(elem))
		}
		i, err := indexOf(tuple.Elems[0])
		if err != nil {
			return nil, err
		}
		j, err := indexOf(tuple.Elems[1])
		if err != nil {
			return nil, err
		}
		matrix[i][j] = true
	}
	if err := enum.Err(); err != nil {
		return nil, err
	}
	count := elemList.Len()
	for y := 0; y < count; y++ {
		for x := 0; x < count; x++ {
			if matrix[x][y] {
				for z := 0; z < count; z++ {
					if matrix[y][z] {
						matrix[x][z] = true
					}
				}
			}
		}
	}
	out := NewValueVec(0)
	for i := 0; i < count; i++ {
		for j := 0; j < count; j++ {
			if matrix[i][j] {
				out.Add(NewTupleValue([]Value{elemList.At(i), elemList.At(j)}))
			}
		}
	}
	return NewSetEnumValueVec(out, false), nil
}
