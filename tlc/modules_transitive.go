package tlc

func Warshall(rel Value) (Value, error) {
	enumerable, ok := asEnumerable(rel)
	if !ok {
		return nil, newTLCError(ECGeneral, "TransitiveClosure expected an enumerable set, got %s", rel)
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
	indexOf := func(value Value) (int, error) {
		for i := 0; i < elemList.Len(); i++ {
			eq, err := elemList.At(i).Equal(value)
			if err != nil {
				return 0, err
			}
			if eq {
				return i, nil
			}
		}
		elemList.Add(value)
		return elemList.Len() - 1, nil
	}
	enum := enumerable.Elements()
	for elem := enum.NextElement(); elem != nil; elem = enum.NextElement() {
		tuple := asTupleValue(elem)
		if tuple == nil || len(tuple.Elems) != 2 {
			return nil, newTLCError(ECGeneral, "TransitiveClosure expected ordered pairs, got %s", elem)
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
