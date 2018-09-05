package gossip

import (
	"github.com/orbs-network/lean-helix-go/types"
)

type discovery struct {
	gossips map[types.PublicKey]*Gossip
}

func NewGossipDiscovery() *discovery {
	return &discovery{
		gossips: make(map[types.PublicKey]*Gossip),
	}
}

func (d *discovery) GetGossipByPK(pk types.PublicKey) (*Gossip, bool) {
	result, ok := d.gossips[pk]
	return result, ok
}

func (d *discovery) RegisterGossip(pk types.PublicKey, gossip *Gossip) {
	d.gossips[pk] = gossip
}

func (d *discovery) GetGossips(pks []types.PublicKey) []*Gossip {

	if pks == nil {
		return d.getAllGossips()
	}

	res := make([]*Gossip, 0, 1)
	for _, key := range d.getAllGossipsPKs() {
		if !indexOf(key, pks) {
			continue
		}
		if gossip, ok := d.GetGossipByPK(key); ok {
			res = append(res, gossip)
		}
	}
	return res
}

func indexOf(pk types.PublicKey, pks []types.PublicKey) bool {
	for _, key := range pks {
		if key == pk {
			return true
		}
	}
	return false
}

func (d *discovery) getAllGossips() []*Gossip {
	gossips := make([]*Gossip, 0, len(d.gossips))
	for _, val := range d.gossips {
		gossips = append(gossips, val)
	}
	return gossips
}

func (d *discovery) getAllGossipsPKs() []types.PublicKey {
	keys := make([]types.PublicKey, 0, len(d.gossips))
	for key := range d.gossips {
		keys = append(keys, key)
	}
	return keys
}
