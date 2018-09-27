package leanhelix

type Config struct {
	NetworkCommunication NetworkCommunication
	BlockUtils           BlockUtils
	KeyManager           KeyManager
	Logger               Logger
	ElectionTrigger      ElectionTrigger
	Storage              Storage
}
