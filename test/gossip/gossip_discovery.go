package gossip

import (
	lh "github.com/orbs-network/lean-helix-go"
)

type discovery struct {
	gossips map[lh.PublicKeyStr]*Gossip
}

func NewGossipDiscovery() *discovery {
	return &discovery{
		gossips: make(map[lh.PublicKeyStr]*Gossip),
	}
}

func (d *discovery) GetGossipByPK(pk lh.PublicKey) (*Gossip, bool) {
	result, ok := d.gossips[lh.PublicKeyStr(pk)]
	return result, ok
}

func (d *discovery) RegisterGossip(pk lh.PublicKey, gossip *Gossip) {
	d.gossips[lh.PublicKeyStr(pk)] = gossip
}

func (d *discovery) GetGossips(pks []lh.PublicKey) []*Gossip {

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

func indexOf(pk lh.PublicKey, pks []lh.PublicKey) bool {
	for _, key := range pks {
		if key.Equals(pk) {
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

func (d *discovery) getAllGossipsPKs() []lh.PublicKey {
	keys := make([]lh.PublicKey, 0, len(d.gossips))
	for key := range d.gossips {
		keys = append(keys, lh.PublicKey(key))
	}
	return keys
}
