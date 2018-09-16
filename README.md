# lean-helix-go


**--Work in progress - not yet functional--**


Go implementation of Lean Helix consensus algorithm

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

## Design
This library does not create any Message objects on its own.
It defines interfaces and leaves it to the user of the library to
create structs that implement them. This is because a Message can contain any data the user wishes, and the library needs not be concerned with that,
as long as specific fields of the Message (as required by the interface's methods) are provided.

KeyManager is also provided as interface-only and it is up to the user to implement its methods.
