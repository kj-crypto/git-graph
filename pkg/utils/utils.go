package utils

type Number interface {
	int | uint | uint8 | uint16 | uint32 | uint64 | int8 | int16 | int32 | int64 | float32 | float64
}

func Max[T Number](x, y T) T {
	if x > y {
		return x
	}
	return y
}

type Set[T comparable] struct {
	items map[T]bool
}

func NewSet[T comparable]() Set[T] {
	return Set[T]{items: make(map[T]bool)}
}

func (s *Set[T]) Add(item T) {
	s.items[item] = true
}

func (s *Set[T]) Delete(item T) {
	delete(s.items, item)
}

func (s *Set[T]) Exists(item T) bool {
	_, exists := s.items[item]
	return exists
}
