package leanhelix

type ElectionTrigger interface {
	RegisterOnTrigger(view View, cb func(view View))
	UnregisterOnTrigger()
}
