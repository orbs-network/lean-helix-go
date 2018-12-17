package gossip

import (
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
)

type Discovery struct {
	gossips map[string]*Gossip
}

func NewDiscovery() *Discovery {
	return &Discovery{
		gossips: make(map[string]*Gossip),
	}
}

func (d *Discovery) GetGossipById(memberId primitives.MemberId) *Gossip {
	return d.getGossipByMemberIdStr(memberId.KeyForMap())
}

func (d *Discovery) getGossipByMemberIdStr(memberIdStr string) *Gossip {
	result, ok := d.gossips[memberIdStr]
	if !ok {
		return nil
	}
	return result
}

func (d *Discovery) RegisterGossip(memberId primitives.MemberId, gossip *Gossip) {
	d.gossips[memberId.KeyForMap()] = gossip
}

func (d *Discovery) UnregisterGossip(memberId primitives.MemberId) {
	delete(d.gossips, memberId.KeyForMap())
}

func (d *Discovery) Gossips(memberIds []primitives.MemberId) []*Gossip {

	if memberIds == nil {
		return d.getAllGossips()
	}

	res := make([]*Gossip, 0, 1)
	for key := range d.gossips {
		if !indexOf(key, memberIds) {
			continue
		}
		if gossip := d.getGossipByMemberIdStr(key); gossip != nil {
			res = append(res, gossip)
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

func (d *Discovery) AllGossipsMemberIds() []primitives.MemberId {
	memberIds := make([]primitives.MemberId, 0, len(d.gossips))
	for memberId := range d.gossips {
		memberIds = append(memberIds, primitives.MemberId(memberId))
	}
	return memberIds
}

func (d *Discovery) getAllGossips() []*Gossip {
	gossips := make([]*Gossip, 0, len(d.gossips))
	for _, val := range d.gossips {
		gossips = append(gossips, val)
	}
	return gossips
}
