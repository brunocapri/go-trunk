package trunk

type expiryHeap[T any] []*cacheEntry[T]

func (h *expiryHeap[T]) Len() int { return len(*h) }
func (h *expiryHeap[T]) Less(i, j int) bool {
	return (*h)[i].createdAt.Before((*h)[j].createdAt)
}
func (h *expiryHeap[T]) Swap(i, j int) {
	(*h)[i], (*h)[j] = (*h)[j], (*h)[i]
	(*h)[i].index, (*h)[j].index = i, j
}
func (h *expiryHeap[T]) Push(x interface{}) {
	entry := x.(*cacheEntry[T])
	*h = append(*h, entry)
}
func (h *expiryHeap[T]) Pop() interface{} {
	old := *h
	n := len(old)
	entry := old[n-1]
	*h = old[0 : n-1]
	return entry
}
