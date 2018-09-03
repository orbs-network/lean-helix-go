package gossip

import (
	lh "github.com/orbs-network/lean-helix-go/go/leanhelix"
	"github.com/stretchr/testify/require"
	"math"
	"math/rand"
	"strconv"
	"testing"
)

func TestGossipDiscovery(t *testing.T) {
	genId := func() lh.PublicKey {
		return lh.PublicKey(strconv.Itoa(int(math.Floor(rand.Float64() * 1000))))
	}

	t.Run("create a GossipDiscovery instance", func(t *testing.T) {
		instance := NewGossipDiscovery()
		require.NotNil(t, instance, "GossipDiscovery instance created")
	})

	t.Run("get Gossip instance by ID", func(t *testing.T) {
		id := genId()
		gd := NewGossipDiscovery()
		expectedGossip := NewGossip(gd)
		gd.RegisterGossip(id, expectedGossip)
		actualGossip := gd.GetGossipByPK(id)
		require.Equal(t, expectedGossip, actualGossip, "received Gossip instance by ID")
	})

	t.Run("get all Gossip IDs", func(t *testing.T) {
		id1 := genId()
		id2 := genId()
		id3 := genId()
		gd := NewGossipDiscovery()
		g1 := NewGossip(gd)
		g2 := NewGossip(gd)
		g3 := NewGossip(gd)
		gd.RegisterGossip(id1, g1)
		gd.RegisterGossip(id2, g2)
		gd.RegisterGossip(id3, g3)
		expected := []lh.PublicKey{id1, id2, id3}
		actual := gd.getAllGossipsPKs()

		require.ElementsMatch(t, actual, expected)
	})
}
