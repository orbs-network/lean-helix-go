package gossip

import (
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
)

type Discovery struct {
	gossips map[string]*Gossip
}

func NewGossipDiscovery() *Discovery {
	return &Discovery{
		gossips: make(map[string]*Gossip),
	}
}

func (d *Discovery) GetGossipByPK(pk primitives.MemberId) *Gossip {
	return d.getGossipByPKStr(pk.KeyForMap())
}

func (d *Discovery) getGossipByPKStr(pkStr string) *Gossip {
	result, ok := d.gossips[pkStr]
	if !ok {
		return nil
	}
	return result
}

func (d *Discovery) RegisterGossip(pk primitives.MemberId, gossip *Gossip) {
	d.gossips[pk.KeyForMap()] = gossip
}

func (d *Discovery) UnregisterGossip(pk primitives.MemberId) {
	delete(d.gossips, pk.KeyForMap())
}

func (d *Discovery) Gossips(pks []primitives.MemberId) []*Gossip {

	if pks == nil {
		return d.getAllGossips()
	}

	res := make([]*Gossip, 0, 1)
	for key := range d.gossips {
		if !indexOf(key, pks) {
			continue
		}
		if gossip := d.getGossipByPKStr(key); gossip != nil {
			res = append(res, gossip)
		}
	}
	return res
}

func indexOf(pkStr string, memberId []primitives.MemberId) bool {
	for _, key := range memberId {
		keyStr := key.KeyForMap()
		if keyStr == pkStr {
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
