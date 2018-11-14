This folder contains Lean Helix's **contract tests**.

Lean Helix ("**library**") depends on implementations of some of its interfaces by the code that uses it (**"service"**).
For example, `KeyManager` interface must be implemented by the service.


The service implements `KeyManager` and so understands its internals.
The library does *not* understand its internals, but it does know how to use it.

To provide the service with assurance that `KeyManager`'s implementation is correct,
the library provides **contract tests** that use the implementation of `KeyManager` is the same way it is used in the
library's regular production code.

You can think of it as integration tests of `KeyManager`'s implementation together with the library.

### What the service should do
In the service, create a test similar to this:
```

// this is a regular test in the service,
// should reside in a file ending with "_test.go"

func TestKeyManager(t *testing.T) {

	// the calling node/machine's public key
	publicKey := ...
	// the calling node/machine's private key
	privateKey := ...

	// mgr implements KeyManager interface
	mgr := MyKeyManagerImpl(publicKey, privateKey)

	// call the lib's contract test with "mgr",
	// the service's implementation of KeyManager
	testSignAndVerify := func(t *testing.T) {
		leanhelixcontracts.TestSignAndVerify(t, mgr)
	}

	t.Run("key manager", testSignAndVerify)

```

Note that these tests do not replace unit tests in the service itself,
because only the service knows the internal implementation details
and how they should be covered by tests.