package gossip

import (
	. "github.com/orbs-network/lean-helix-go/primitives"
)

type PublicKeyStr string

type Discovery interface {
	GetGossipByPK(pk Ed25519PublicKey) *Gossip
	RegisterGossip(pk Ed25519PublicKey, gossip *Gossip)
	UnregisterGossip(pk Ed25519PublicKey)
	AllGossipsPublicKeys() []Ed25519PublicKey
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

func (d *discovery) GetGossipByPK(pk Ed25519PublicKey) *Gossip {
	return d.GetGossipByPKStr(pk.String())
}

func (d *discovery) GetGossipByPKStr(pkStr string) *Gossip {
	result, ok := d.gossips[pkStr]
	if !ok {
		return nil
	}
	return result
}

func (d *discovery) RegisterGossip(pk Ed25519PublicKey, gossip *Gossip) {
	d.gossips[pk.String()] = gossip
}
func (d *discovery) UnregisterGossip(pk Ed25519PublicKey) {
	delete(d.gossips, pk.String())
}

func (d *discovery) Gossips(pks []Ed25519PublicKey) []*Gossip {

	if pks == nil {
		return d.getAllGossips()
	}

	res := make([]*Gossip, 0, 1)
	for key := range d.gossips {
		if !indexOf(key, pks) {
			continue
		}
		if gossip := d.GetGossipByPKStr(key); gossip != nil {
			res = append(res, gossip)
		}
	}
	return res
}

func indexOf(pkStr string, publicKeys []Ed25519PublicKey) bool {
	for _, key := range publicKeys {
		keyStr := key.String()
		if keyStr == pkStr {
			return true
		}
	}
	return false
}

func (d *discovery) AllGossipsPublicKeys() []Ed25519PublicKey {
	publicKeys := make([]Ed25519PublicKey, 0, len(d.gossips))
	for _, val := range d.gossips {
		publicKeys = append(publicKeys, val.publicKey)
	}
	return publicKeys
}

func (d *discovery) getAllGossips() []*Gossip {
	gossips := make([]*Gossip, 0, len(d.gossips))
	for _, val := range d.gossips {
		gossips = append(gossips, val)
	}
	return gossips
}
