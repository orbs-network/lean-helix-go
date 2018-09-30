# lean-helix-go


**--Work in progress - not yet functional--**


Go implementation of Lean Helix consensus algorithm

## TODO
* [ ] Decide on types - do we use primitives for height, view, public key, hash (ints, []byte's) which are easier to program, or aliases which are easier to understand but cumbersome to maintain esp with []PublicKey -> []string conversions
* [ ] JSON log



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

## Design
This library does not create any Message objects on its own.
It defines interfaces and leaves it to the user of the library to
create structs that implement them. This is because a Message can contain any data the user wishes, and the library needs not be concerned with that,
as long as specific fields of the Message (as required by the interface's methods) are provided.

KeyManager is also provided as interface-only and it is up to the user to implement its methods.

### Logging

JSON-based format to make it easily on log capturing tools.

`DEBUG`

`INFO`

By design, there is no `WARN` log level. We believe such messages are no more actionable than regular `INFO` messages.
For example, a view timeout could easily be classified as `WARN` as it is "exceptional" though it is recoverable, it is a normal part of the algo so it is expected to happen occasionally.
If there are too many timeouts, it is the job of a log monitoring tool to trigger some alert over this.

`ERROR`s are reserved for unrecoverable situations, either due to bugs, assertions (again bugs), panics and OS errors - all of which require either node restart or bug fixing.