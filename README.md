# lean-helix-go


**--Work in progress - not yet functional--**


Go implementation of Lean Helix consensus algorithm.

Spec location: TBD

## TODO
* See [issues on Github](https://github.com/orbs-network/lean-helix-go/issues)

## Design (Lean Helix fully serializes its messages internally)
Uses orbs-spec
* orbs-spec/interfaces/protocol/gossipmessages/lean-helix.proto
* orbs-spec/interfaces/protocol/consensus/lean-helix.proto




## Design (Lean helix only uses interfaces)
This library does not create any Message objects on its own.
It defines interfaces and leaves it to the user of the library to
create structs that implement them. This is because a Message can contain any data the user wishes, and the library needs not be concerned with that,
as long as specific fields of the Message (as required by the interface's methods) are provided.

KeyManager is also provided as interface-only and it is up to the user to implement its methods.

### Interfaces

Only orbs serializes messages because its protocol defines the exact structure of every message on the wire, including the lib's.
The lib does so for testing and can use JSON.

#### Message handling

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

Message types
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
	SenderPublicKey() PublicKey
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
  Signers() []PublicKey
  ConsensusSpecificData() []byte
}
```

#### NetworkCommunication

```
type NetworkCommunication interface {
	SendToMembers(publicKeys []PublicKey, messageType string, message []MessageTransporter)
	RequestOrderedCommittee(seed uint64) []PublicKey
	IsMember(pk PublicKey) bool
	SendPreprepare(publicKeys []PublicKey, message PreprepareMessage)
	SendPrepare(publicKeys []PublicKey, message PrepareMessage)
	SendCommit(publicKeys []PublicKey, message CommitMessage)
	SendViewChange(publicKey PublicKey, message ViewChangeMessage)
	SendNewView(publicKeys []PublicKey, message NewViewMessage)

```


#### KeyManager

```
type KeyManager interface {
	Sign(content []byte) []byte
	Verify(content []byte, sender SenderSignature) bool
	MyPublicKey() PublicKey
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
  * OnReceivePreprepare, OnReceivePrepare, OnReceiveCommit, OnReceiveViewChange, OnReceiveNewView


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

### Lean Helix (object)
The `LeanHelix` object holds the main algo loop that always tries to get new blocks, reach consensus, and broadcast that they should be committed.
There is only a single instance of this type, it is aware of all blocks pending consensus (of all heights). It holds one or more `LeanHelixTerm` instances.

### Term
**Term** is a specific block height - the algo is trying to reach consensus on a block of some height during a term.
A Term handles only a specific height, but the algo can store locally blocks of several heights.
It will associate each block height with its term.

The struct `LeanHelixTerm` handles a specific **term** - that is it handles reaching consensus on a single block of some height. It has no knowledge of blocks of other heights.
There may be multiple `LeanHelixTerm` instances active at the same time.

### View
A view is a single round of consensus with a specific leader. The leader for that view begins with getting a block for which it will try to reach consensus, then it starts spreading it by sending a Preprepare message.
If that leader is unable to reach consensus for any reason (usually timeout), the view is incremented and a new leader is set.
So - each view (modulo the number of nodes in a committee) has a different leader.


### Logging

JSON-based format to make it easily on log capturing tools.

`DEBUG`

`INFO`

By design, there is no `WARN` log level. We believe such messages are no more actionable than regular `INFO` messages.
For example, a view timeout could easily be classified as `WARN` as it is "exceptional" though it is recoverable, it is a normal part of the algo so it is expected to happen occasionally.
If there are too many timeouts, it is the job of a log monitoring tool to trigger some alert over this.

`ERROR`s are reserved for unrecoverable situations, either due to bugs, assertions (again bugs), panics and OS errors - all of which require either node restart or bug fixing.