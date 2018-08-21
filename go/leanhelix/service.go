package leanhelix

type Service interface {
	Start()
}

type service struct {
	config     *Config
	network    NetworkCommunication
	blockUtils BlockUtils
	keyManager KeyManager
}

func (s *service) Start() {

}

func NewLeanHelix(config *Config, network NetworkCommunication, blockUtils BlockUtils, keyManager KeyManager) Service {
	return &service{
		config, network, blockUtils, keyManager,
	}
}

func (s *service) Init() {

}
