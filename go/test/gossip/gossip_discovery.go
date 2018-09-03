package gossip

import (
	lh "github.com/orbs-network/lean-helix-go/go/leanhelix"
)

type discovery struct {
	gossips map[lh.PublicKey]*Gossip
}

func NewGossipDiscovery() *discovery {
	return &discovery{
		gossips: make(map[lh.PublicKey]*Gossip),
	}
}

func (d *discovery) GetGossipByPK(pk lh.PublicKey) *Gossip {
	return d.gossips[pk]
}

func (d *discovery) RegisterGossip(pk lh.PublicKey, gossip *Gossip) {
	d.gossips[pk] = gossip
}

func (d *discovery) GetGossips(pks []lh.PublicKey) []*Gossip {
	res := make([]*Gossip, 1)
	if pks != nil {
		for _, key := range d.getAllGossipsPKs() {
			if indexOf(key, pks) {
				res = append(res, d.GetGossipByPK(key))
			}
		}
	} else {
		for _, val := range d.gossips {
			res = append(res, val)
		}
	}
	return res
}

func indexOf(pk lh.PublicKey, pks []lh.PublicKey) bool {
	for _, key := range pks {
		if key == pk {
			return true
		}
	}
	return false
}

func (d *discovery) getAllGossipsPKs() []lh.PublicKey {
	keys := make([]lh.PublicKey, 0, len(d.gossips))
	for key := range d.gossips {
		keys = append(keys, key)
	}
	return keys
}
