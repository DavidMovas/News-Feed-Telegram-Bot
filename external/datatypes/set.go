package set

type Set[T comparable] map[T]struct{}

func New[T comparable](items ...T) Set[T] {
	set := make(Set[T])
	for _, item := range items {
		set[item] = struct{}{}
	}
	return set
}

func (s Set[T]) Has(item T) bool {
	_, ok := s[item]
	return ok
}

func (s Set[T]) Add(item T) {
	s[item] = struct{}{}
}

func (s Set[T]) Remove(item T) {
	delete(s, item)
}

func (s Set[T]) Items() []T {
	items := make([]T, 0, len(s))
	for item := range s {
		items = append(items, item)
	}
	return items
}

func (s Set[T]) Len() int {
	return len(s)
}
