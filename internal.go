package leanhelix

// Sorting View arrays
type ViewCounters []View

func (arr ViewCounters) Len() int           { return len(arr) }
func (arr ViewCounters) Swap(i, j int)      { arr[i], arr[j] = arr[j], arr[i] }
func (arr ViewCounters) Less(i, j int) bool { return arr[i] < arr[j] }
