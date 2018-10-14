package gossip

import (
	. "github.com/orbs-network/lean-helix-go/primitives"
)

type PublicKeyStr string

type Discovery interface {
	GetGossipByPK(pk Ed25519PublicKey) (*Gossip, bool)
	RegisterGossip(pk Ed25519PublicKey, gossip *Gossip)
	AllGossipsPKs() []Ed25519PublicKey
	Gossips(pks []Ed25519PublicKey) []*Gossip
}

type discovery struct {
	gossips map[string]*Gossip
}

func NewGossipDiscovery() Discovery {
	return &discovery{
		gossips: make(map[string]*Gossip),
	}
}

func (d *discovery) GetGossipByPK(pk Ed25519PublicKey) (*Gossip, bool) {
	result, ok := d.gossips[pk.String()]
	return result, ok
}

func (d *discovery) RegisterGossip(pk Ed25519PublicKey, gossip *Gossip) {
	d.gossips[pk.String()] = gossip
}

func (d *discovery) Gossips(pks []Ed25519PublicKey) []*Gossip {

	if pks == nil {
		return d.getAllGossips()
	}

	res := make([]*Gossip, 0, 1)
	for _, key := range d.AllGossipsPKs() {
		if !indexOf(key, pks) {
			continue
		}
		if gossip, ok := d.GetGossipByPK(key); ok {
			res = append(res, gossip)
		}
	}
	return res
}

func indexOf(pk Ed25519PublicKey, pks []Ed25519PublicKey) bool {
	for _, key := range pks {
		if key.Equal(pk) {
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

func (d *discovery) AllGossipsPKs() []Ed25519PublicKey {
	keys := make([]Ed25519PublicKey, 0, len(d.gossips))
	for key := range d.gossips {
		keys = append(keys, Ed25519PublicKey(key))
	}
	return keys
}
