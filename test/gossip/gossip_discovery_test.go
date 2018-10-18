package gossip

import (
	"github.com/orbs-network/lean-helix-go/primitives"
	"github.com/stretchr/testify/require"
	"math"
	"math/rand"
	"strconv"
	"testing"
)

func TestGossipDiscovery(t *testing.T) {
	genPublicKey := func() primitives.Ed25519PublicKey {
		return primitives.Ed25519PublicKey(strconv.Itoa(int(math.Floor(rand.Float64() * 1000))))
	}

	t.Run("create a GossipDiscovery instance", func(t *testing.T) {
		instance := NewGossipDiscovery()
		require.NotNil(t, instance, "GossipDiscovery instance created")
	})

	t.Run("get Gossip instance by ID", func(t *testing.T) {
		id := genPublicKey()
		gd := NewGossipDiscovery()
		expectedGossip := NewGossip(gd, id)
		gd.RegisterGossip(id, expectedGossip)
		actualGossip, _ := gd.GetGossipByPK(id)
		require.Equal(t, expectedGossip, actualGossip, "received Gossip instance by ID")
	})

	t.Run("get all Gossip IDs", func(t *testing.T) {
		id1 := genPublicKey()
		id2 := genPublicKey()
		id3 := genPublicKey()
		gd := NewGossipDiscovery()
		g1 := NewGossip(gd, id1)
		g2 := NewGossip(gd, id2)
		g3 := NewGossip(gd, id3)
		gd.RegisterGossip(id1, g1)
		gd.RegisterGossip(id2, g2)
		gd.RegisterGossip(id3, g3)
		expectedPublicKeyStrings := []string{id1.String(), id2.String(), id3.String()}
		actualPublicKeyStrings := gd.AllGossipsPKs()

		require.ElementsMatch(t, actualPublicKeyStrings, expectedPublicKeyStrings)
	})

	t.Run("return ok=false if given Id was not registered", func(t *testing.T) {
		id := genPublicKey()
		gd := NewGossipDiscovery()
		_, ok := gd.GetGossipByPK(id)

		require.False(t, ok, "GetGossipByPK() returns ok=false if ID not registered")
	})

	t.Run("return a list of all gossips", func(t *testing.T) {
		gd := NewGossipDiscovery()
		id1 := genPublicKey()
		id2 := genPublicKey()
		g1 := NewGossip(gd, id1)
		g2 := NewGossip(gd, id2)
		gd.RegisterGossip(id1, g1)
		gd.RegisterGossip(id2, g2)
		actual := gd.Gossips(nil)
		expected := []*Gossip{g1, g2}
		require.ElementsMatch(t, actual, expected, "list of all gossips")
	})

	t.Run("return a list of requested gossips", func(t *testing.T) {
		gd := NewGossipDiscovery()
		id1 := genPublicKey()
		id2 := genPublicKey()
		id3 := genPublicKey()
		g1 := NewGossip(gd, id1)
		g2 := NewGossip(gd, id2)
		g3 := NewGossip(gd, id3)
		gd.RegisterGossip(id1, g1)
		gd.RegisterGossip(id2, g2)
		gd.RegisterGossip(id3, g3)
		actual := gd.Gossips([]primitives.Ed25519PublicKey{id1, id3})
		expected := []*Gossip{g1, g3}
		require.ElementsMatch(t, actual, expected, "list of requested gossips")
	})

}
