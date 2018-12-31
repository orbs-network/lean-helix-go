package mocks

import (
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
)

type Discovery struct {
	communications map[string]*CommunicationMock
}

func NewDiscovery() *Discovery {
	return &Discovery{
		communications: make(map[string]*CommunicationMock),
	}
}

func (d *Discovery) GetCommunicationById(memberId primitives.MemberId) *CommunicationMock {
	return d.getCommunicationByMemberIdStr(memberId.KeyForMap())
}

func (d *Discovery) getCommunicationByMemberIdStr(memberIdStr string) *CommunicationMock {
	result, ok := d.communications[memberIdStr]
	if !ok {
		return nil
	}
	return result
}

func (d *Discovery) RegisterCommunication(memberId primitives.MemberId, communication *CommunicationMock) {
	d.communications[memberId.KeyForMap()] = communication
}

func (d *Discovery) UnregisterCommunication(memberId primitives.MemberId) {
	delete(d.communications, memberId.KeyForMap())
}

func (d *Discovery) Communications(memberIds []primitives.MemberId) []*CommunicationMock {

	if memberIds == nil {
		return d.getAllCommunication()
	}

	res := make([]*CommunicationMock, 0, 1)
	for key := range d.communications {
		if !indexOf(key, memberIds) {
			continue
		}
		if communication := d.getCommunicationByMemberIdStr(key); communication != nil {
			res = append(res, communication)
		}
	}
	return res
}

func indexOf(memberIdStr string, memberId []primitives.MemberId) bool {
	for _, key := range memberId {
		keyStr := key.KeyForMap()
		if keyStr == memberIdStr {
			return true
		}
	}
	return false
}

func (d *Discovery) AllCommunicationsMemberIds() []primitives.MemberId {
	memberIds := make([]primitives.MemberId, 0, len(d.communications))
	for memberId := range d.communications {
		memberIds = append(memberIds, primitives.MemberId(memberId))
	}
	return memberIds
}

func (d *Discovery) getAllCommunication() []*CommunicationMock {
	communications := make([]*CommunicationMock, 0, len(d.communications))
	for _, val := range d.communications {
		communications = append(communications, val)
	}
	return communications
}
