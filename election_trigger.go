package leanhelix

type ElectionTrigger interface {
	RegisterOnTrigger(view ViewCounter, cb func(view ViewCounter))
	UnregisterOnTrigger()
}
