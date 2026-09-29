package tlago

type SanyVector[T comparable] struct {
	items []T
}

func NewSanyVector[T comparable]() *SanyVector[T] {
	return &SanyVector[T]{}
}

func (v *SanyVector[T]) AddElement(item T) {
	v.items = append(v.items, item)
}

func (v *SanyVector[T]) Contains(item T) bool {
	for _, existing := range v.items {
		if existing == item {
			return true
		}
	}
	return false
}

func (v *SanyVector[T]) InsertElementAt(item T, index int) {
	if index < 0 || index >= len(v.items) {
		panic("SanyVector.InsertElementAt index out of bounds")
	}
	var zero T
	v.items = append(v.items, zero)
	copy(v.items[index+1:], v.items[index:])
	v.items[index] = item
}

func (v *SanyVector[T]) Elements() []T {
	out := make([]T, len(v.items))
	copy(out, v.items)
	return out
}

func (v *SanyVector[T]) AppendNoRepeats(incoming *SanyVector[T]) {
	if incoming == nil {
		return
	}
	for _, item := range incoming.items {
		if !v.Contains(item) {
			v.AddElement(item)
		}
	}
}

func (v *SanyVector[T]) Size() int {
	return len(v.items)
}

func (v *SanyVector[T]) ElementAt(index int) T {
	return v.items[index]
}
