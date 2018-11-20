This folder contains Lean Helix's **contract tests**.

Lean Helix ("**library**") depends on implementations of some
of its interfaces by the code that uses it (**"consumer"**).
For example, `KeyManager` interface must be implemented by the consumer.
See the Design section of the main README.md for more details.

The consumer implements `KeyManager` and so understands its internals.
The library does *not* understand its internals,
but it *does* know how to use it.

To provide the consumer with assurance that `KeyManager`'s implementation
is correct, the library provides **contract tests** that use
the implementation of `KeyManager` is the same way it is used in the
library's regular production code.

You can think of it as integration tests of `KeyManager`'s
implementation together with the library.

### What the library does
The library contains functions that accept `*testing.T`
and an instance of implemented interface. For example:
```
func TestSignAndVerify(t *testing.T, mgr leanhelix.KeyManager)
```
where:
 * **t** is propagated from the consumer
 * **mgr** is an instance of `KeyManager`


### What the consumer should do
In the consumer, create a test similar to this:
```

// this is a regular test in the consumer,
// should reside in a file ending with "_test.go"

func TestKeyManager(t *testing.T) {

	// the calling node/machine's public key
	publicKey := ...
	// the calling node/machine's private key
	privateKey := ...

	// mgr implements KeyManager interface
	mgr := MyKeyManagerImpl(publicKey, privateKey)

	// call the lib's contract test with "mgr",
	// the consumer's implementation of KeyManager
	testSignAndVerify := func(t *testing.T) {
		leanhelixcontracts.TestSignAndVerify(t, mgr)
	}

	t.Run("key manager", testSignAndVerify)

```

Note that these tests do not replace unit tests in the consumer itself,
because only the consumer knows the internal implementation details
and how they should be covered by tests.

### Terminology note
KeyManager is an example of an SPI (Service Programming Interface) -
it is an interface the consumer must implement so the library can use it.
This is opposed to the more common API (Application Programming Interface)
which is an interface the library implements and is used by the consumer
(e.g. `LeanHelix.ValidateBlockConsensus` which the library implements
and the consumer calls).
