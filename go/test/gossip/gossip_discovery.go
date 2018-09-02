package gossip

import (
	lh "github.com/orbs-network/lean-helix-go/go/leanhelix"
)

type GossipDiscovery struct {
	gossips map[lh.PublicKey]Gossip
}

func (gd *GossipDiscovery) getGossipByPk(pk lh.PublicKey) Gossip {
	return gd.gossips[pk]
}

func (gd *GossipDiscovery) registerGossip(pk lh.PublicKey, gossip Gossip) {
	gd.gossips[pk] = gossip
}

func (gd *GossipDiscovery) getGossips(pks []lh.PublicKey) []Gossip {
	res := make([]Gossip, 1)
	if pks != nil {
		for _, key := range gd.getAllGossipsPks() {
			if indexOf(key, pks) {
				res = append(res, gd.getGossipByPk(key))
			}
		}
	} else {
		for _, val := range gd.gossips {
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

func (gd *GossipDiscovery) getAllGossipsPks() []lh.PublicKey {
	keys := make([]lh.PublicKey, len(gd.gossips))
	for key, _ := range gd.gossips {
		keys = append(keys, key)
	}
	return keys
}
