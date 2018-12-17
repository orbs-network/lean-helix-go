# lean-helix-go


**--Work in progress - not yet functional--**


Lean Helix consensus algorithm implementation in Go.

## TODO
* See [issues on Github](https://github.com/orbs-network/lean-helix-go/issues)

## Installation
* Download from Github
* Run `./git-submodule-checkout.sh` - this will install all git submodules under `vendor` folder

### Terminology
* `Library` - this repo
* `Consumer` - the user of this library
* `API` - methods in the library that the consumer can execute (e.g. `NewLeanHelix()`)
* `SPI` - Service Programming Interface - interfaces defined in the library, for which the consumer provides the implementation
  * For example - `Communication` is responsible for transferring messages between nodes in the network. The library cannot assume anything about the consumer's network, therefore it is up to the consumer to provide the actual implementation of message transfer.
* `membuffers` - the 3rd party dependency that the library uses for serializing its messages into and from byte arrays
  * Repo on [github](https://github.com/orbs-network/membuffers)
* `protos` - the *.proto files (in [Google's Protobuf](https://developers.google.com/protocol-buffers/) language) which define the structure of messages passing between Lean Helix and its consumer.
  * The `membuffers` library takes *.proto files and compiles them to *.mb.go files
  * Any change to *.proto files *requires* running `cd types; ./build.sh` to regenerate the respective *.mb.go !

* Lean Helix (object) - runs the infinite loop that with every iteration requests a new block, reaches consensus, and broadcasts the commit message.
  There is only a single instance of this type, it is aware of all blocks pending consensus (of all heights). It holds one or more `LeanHelixTerm` instances.

* `Committee` - the *ordered* list of nodes participating in a single *Term*.
The first node in the committee is the first leader. If that leader is unable to reach consensus on its block proposal, the next node in the committee becomes the new leader and the process repeats.

* `Term` - a.k.a. **consensus round** - the process of reaching consensus for a *specific block height* - the algo is trying to reach consensus on a block of some height during a term.
A Term handles only a specific height, but the algo can locally store blocks for several heights.
It will associate each block height with its term. See struct `LeanHelixTerm`.

* `View` - part of the `Term`, during which a specific leader node proposes a block and tries to reach consensus with its *validators* (the other, non-leader nodes).
If that leader is unable to reach consensus for any reason (usually timeout), the *view* is incremented and a new leader is set.
The next leader is the next node in the committee
So - each view (modulo the number of nodes in a committee) has a different leader.

## Design
Formal spec for this library can be found under the `/spec` folder.

### API
* NewLeanHelix()
* ValidateConsensusOnBlock
* tbd

### SPI
Interfaces provided by the library, implemented by the consumer
* KeyManager (to be renamed to SignerVerifier - tbd)
* BlockUtils (to be split into BlockProvider and BlockValidator)
* NetworkCommunication

### Internal structs
tbd
These are internal, not to be used directly by the consumer of the library

* `LeanHelixTerm` handles a specific **term** - that is it handles reaching consensus on a single block of some height. It has no knowledge of blocks of other heights.
There may be multiple `LeanHelixTerm` instances active at the same time.


TBD

### Messages
The library creates and serializes its own messages into and from byte arrays.
The library does not actually transfer messages over the wire as it does not assume anything about the consumer's network.


#### Message handling

tbd go over make sure still relevant
Message creation
```
type MessageFactory interface {
	// Message creation methods

	CreatePreprepareMessage(blockRef BlockRef, sender SenderSignature, block Block) PreprepareMessage
	CreatePrepareMessage(blockRef BlockRef, sender SenderSignature) PrepareMessage
	CreateCommitMessage(blockRef BlockRef, sender SenderSignature) CommitMessage
	CreateViewChangeMessage(vcHeader ViewChangeHeader, sender SenderSignature, block Block) ViewChangeMessage
	CreateNewViewMessage(preprepareMessage PreprepareMessage, nvHeader NewViewHeader, sender SenderSignature) NewViewMessage

	// Auxiliary methods

	CreateSenderSignature(sender []byte, signature []byte) SenderSignature
	CreateBlockRef(messageType int, blockHeight int, view int, blockHash []byte) BlockRef
	CreateNewViewHeader(messageType int, blockHeight int, view int, confirmations []ViewChangeConfirmation) NewViewHeader
	CreateViewChangeConfirmation(vcHeader ViewChangeHeader, sender SenderSignature) ViewChangeConfirmation
	CreateViewChangeHeader(blockHeight int, view int, proof PreparedProof) ViewChangeHeader
	CreatePreparedProof(ppBlockRef BlockRef, pBlockRef BlockRef, ppSender SenderSignature, pSenders []SenderSignature) PreparedProof
}

```

Message primitives
```
type PreprepareMessage interface {
	Serializable
	SignedHeader() BlockRef
	Sender() SenderSignature
	Block() Block
}

type PrepareMessage interface {
	Serializable
	SignedHeader() BlockRef
	Sender() SenderSignature
}

type CommitMessage interface {
	Serializable
	SignedHeader() BlockRef
	Sender() SenderSignature
}

type ViewChangeMessage interface {
	Serializable
	SignedHeader() ViewChangeHeader
	Sender() SenderSignature
	Block() Block
}

type NewViewMessage interface {
	Serializable
	SignedHeader() NewViewHeader
	PreprepareMessage() PreprepareMessage
	Sender() SenderSignature
}

```

Message parts
```
type BlockRef interface {
	Serializable
	HasMessageType
	BlockHeight() BlockHeight
	View() View
	BlockHash() BlockHash
}

type ViewChangeHeader interface {
	Serializable
	HasMessageType
	BlockHeight() BlockHeight
	View() View
	PreparedProof() PreparedProof
}

type SenderSignature interface {
	Serializable
	SenderMemberId() MemberId
	Signature() Signature
}

type HasMessageType interface {
	MessageType() MessageType
}

type Serializable interface {
	Serialize() []byte
}

// TODO this is different from definition of LeanHelixPreparedProof in lean_helix.mb.go:448 in orbs-spec
type PreparedProof interface {
	Serializable
	PPBlockRef() BlockRef
	PBlockRef() BlockRef
	PPSender() SenderSignature
	PSenders() []SenderSignature
}

type NewViewHeader interface {
	Serializable
	HasMessageType
	BlockHeight() BlockHeight
	View() View
	ViewChangeConfirmations() []ViewChangeConfirmation
}

type ViewChangeConfirmation interface {
	Serializable
	SignedHeader() ViewChangeHeader
	Sender() SenderSignature
}

```

#### Block Proof
```
type BlockProof interface {
  BlockHeight() BlockHeight
  View() View // this is only for display in a future block viewer - it is of no use to Orbs as it is lh-internal
  BlockHash() BlockHash
  Signers() []MemberId
  ConsensusSpecificData() []byte
}
```

#### NetworkCommunication

```
type NetworkCommunication interface {
	SendToMembers(publicKeys []MemberId, messageType string, message []MessageTransporter)
	RequestCommittee(seed uint64) []MemberId
	IsMember(pk MemberId) bool
	SendPreprepare(publicKeys []MemberId, message PreprepareMessage)
	SendPrepare(publicKeys []MemberId, message PrepareMessage)
	SendCommit(publicKeys []MemberId, message CommitMessage)
	SendViewChange(publicKey MemberId, message ViewChangeMessage)
	SendNewView(publicKeys []MemberId, message NewViewMessage)

```


#### KeyManager

```
type KeyManager interface {
	Sign(content []byte) []byte
	Verify(content []byte, sender SenderSignature) bool
	MyMemberId() MemberId
}
```

#### BlockUtils
```
type BlockUtils interface {
	CalculateBlockHash(block Block) BlockHash
	RequestNewBlock(blockHeight BlockHeight) Block
	ValidateBlock(block Block) bool
	RequestCommittee()
}

```

TODO add this:
* MessageReceiver
  * HandleLeanHelixPrePrepare, HandleLeanHelixPrepare, HandleLeanHelixCommit, HandleLeanHelixViewChange, HandleLeanHelixNewView


On creation of the lib instance (of type `LeanHelixLib`) by Orbs, Orbs will pass a service parameter to the lib instance. That `service` will contain implementations of the above interfaces




## Installation

#### Prerequisites

1. Make sure [Go](https://golang.org/doc/install) is installed (version 1.10 or later).

    > Verify with `go version`

2. Make sure [Go workspace bin](https://stackoverflow.com/questions/42965673/cant-run-go-bin-in-terminal) is in your path.

    > Install with ``export PATH=$PATH:`go env GOPATH`/bin``

    > Verify with `echo $PATH`

#### Get and build

1. Get the library into your Go workspace:

     ```sh
     go get github.com/orbs-network/lean-helix-go/go/...
     ```

## Test

1. Test the library (unit tests and end to end tests):

    ```sh
    ./test.sh
    ```

## Terminology



### Logging

JSON-based format to make it easily on log capturing tools.

`DEBUG`

`INFO`

By design, there is no `WARN` log level. We believe such messages are no more actionable than regular `INFO` messages.
For example, a view timeout could easily be classified as `WARN` as it is "exceptional" though it is recoverable, it is a normal part of the algo so it is expected to happen occasionally.
If there are too many timeouts, it is the job of a log monitoring tool to trigger some alert over this.

`ERROR`s are reserved for unrecoverable situations, either due to bugs, assertions (again bugs), panics and OS errors - all of which require either node restart or bug fixing.