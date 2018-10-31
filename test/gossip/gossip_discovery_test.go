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
		return primitives.Ed25519PublicKey(strconv.Itoa(int(math.Floor(rand.Float64() * 1000000000))))
	}

	t.Run("create a Discovery instance", func(t *testing.T) {
		instance := NewGossipDiscovery()
		require.NotNil(t, instance, "Discovery instance created")
	})

	t.Run("get Gossip instance by ID", func(t *testing.T) {
		id := genPublicKey()
		gd := NewGossipDiscovery()
		expectedGossip := NewGossip(gd)
		gd.RegisterGossip(id, expectedGossip)
		actualGossip := gd.GetGossipByPK(id)
		require.Equal(t, expectedGossip, actualGossip, "received Gossip instance by ID")
	})

	t.Run("get all Gossip IDs", func(t *testing.T) {
		id1 := genPublicKey()
		id2 := genPublicKey()
		id3 := genPublicKey()
		gd := NewGossipDiscovery()
		g1 := NewGossip(gd)
		g2 := NewGossip(gd)
		g3 := NewGossip(gd)
		gd.RegisterGossip(id1, g1)
		gd.RegisterGossip(id2, g2)
		gd.RegisterGossip(id3, g3)
		expectedPublicKeyStrings := []string{id1.String(), id2.String(), id3.String()}
		actualPublicKeys := gd.AllGossipsPublicKeys()
		actualPublicKeyStrings := make([]string, 0, len(actualPublicKeys))
		for _, pk := range actualPublicKeys {
			actualPublicKeyStrings = append(actualPublicKeyStrings, pk.String())
		}

		require.ElementsMatch(t, actualPublicKeyStrings, expectedPublicKeyStrings)
	})

	t.Run("return gossip=nil if given Id was not registered", func(t *testing.T) {
		id := genPublicKey()
		gd := NewGossipDiscovery()
		gossip := gd.GetGossipByPK(id)

		require.Nil(t, gossip, "GetGossipByPK() returns ok=false if ID not registered")
	})

	t.Run("return a list of all gossips", func(t *testing.T) {
		gd := NewGossipDiscovery()
		id1 := genPublicKey()
		id2 := genPublicKey()
		g1 := NewGossip(gd)
		g2 := NewGossip(gd)
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
		g1 := NewGossip(gd)
		g2 := NewGossip(gd)
		g3 := NewGossip(gd)
		gd.RegisterGossip(id1, g1)
		gd.RegisterGossip(id2, g2)
		gd.RegisterGossip(id3, g3)
		actual := gd.Gossips([]primitives.Ed25519PublicKey{id1, id3})
		expected := []*Gossip{g1, g3}
		require.ElementsMatch(t, actual, expected, "list of requested gossips")
	})

}
