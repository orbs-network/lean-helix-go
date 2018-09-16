package leanhelix

type Config interface {
	ElectionTrigger() ElectionTrigger
}
