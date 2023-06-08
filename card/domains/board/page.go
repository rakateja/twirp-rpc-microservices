package board

type Page[T interface{}] struct {
	Items []T
	Total int
}
