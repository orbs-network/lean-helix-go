# New threading model POC

## Problems
### General
Lean Helix (LH) calls external SPIs for some of its required functionality.
Some of those are long operations (`RequestNewBlockProposal`, `ValidateBlockProposal`, `Sign`)
and none of them are under the control of Lean Helix, meaning others can still become long operations.
This leads to the following phenomena:

### NodeSync goroutine blocked
During tests we identified that `RequestNewBlockProposal` which can take up to 9 seconds in the 
current configuration, causes the Orbs BlockSync goroutine to block for that amount of time, 
because the BlockSync goroutine writes to LH's UpdateState channel which cannot be read while LH is
blocking on waiting for  `RequestNewBlockProposal`. 
Similar issue can occur waiting for `ValidateBlockProposal`. 
This also opens LH to a liveness attack by having a byzantine 
leader create a block with infinite-loop transaction. Every node that will try to validate the block
will be blocked indefinitely and will not be unblocked without restart 
(because no node sync, no election, no commit). 

### Gossip
....

##
IDEA: Keep mainloop with AsyncOpChannel, no TermLoop

### Election not triggered
While a long operation is running, LH is kept out of the main loop 
so Election timeout is not handled on time.

## Potential solutions

### Make updateState channel buffered
* BlockSync goroutine will not be blocked, but LH mainloop can still be blocked 
by running `RequestNewBlockProposal` (for example).
 

### Make LH a state machine similar to BlockSync
~~TBD Think if this is relevant to the problems above.~~  

### Add TermLoop goroutine and for long-running ops
MainLoop, TermLoop and some Worker goroutine with AsyncOpChannel to signal end of operation.
* [-] A single AsyncOp channel with some general return values (such as `interface{}`) 
would reduce type safety.
* [-] Multiple AsyncOp channels, one per long op, would open the door to 
a bad practice of mindlessly adding a channel per op. `Normalization of deviance`
* [+] Simpler dev effort as most of the code is under a `Term` struct already. 

### Add goroutines for TermLoop and ViewLoop
* [-] More difficult to develop than previous solution, as need to refactor
a new `View` struct out of existing `Term` struct.
needs to tear out View functionality from Term, and there are naming problems.
* [+] Cleaner than creating the AsyncOpChannel for long running ops. 
It keeps the View synchronous, letting the View simply shutdown if election timeout triggered.
* [+] Easier to test in isolation, Term and View are more coherent semantically 
(i.e. Term and View are different entities, so they get different structs and different goroutines)
* [+] Election timeout is the context cancellation signal for the View -- cleaner design

### Goroutine from mainloop for election and nodesync
* Move channels of NodeSync and Election to a separate goroutine.
* Create separate context for each Term so that long running ops will be cancellable per term.

* [+] 

This will solve the NodeSync
 


## Design Questions
1. Do we intend to change how `Shyness` works?
`Shyness` is the informal name we gave to the feature where a Node that receives 
a block by NodeSync and then starts a new term and also becomes leader 
of first view (V=0), will refuse to be leader and let the next Node in line 
become leader of V=1. This prevents Node that receives 
multiple blocks by NodeSync to become leader every n blocks (where `n` is number of
committee members) and pollute the network with `RequestNewBlockProposal` and 
subsequent PREPREPARE messages.

## Engineering Questions

1. What happens when trying to read from a garbage collected channel
> Not garbage collected till we release all references.
2. What to do with channels waiting for CreateBlock and ValidateBlock when closing the term?
> Nothing, `termloop` goroutine closes when 
3. What to do with the committed channel when closing term
> NOTHING, it just vanishes
4. Race: Election timer pops just after closing its parent term
> Could not reproduce
5. Decide: Who creates the committed channel?
    1. mainloop creates and passes to NewTermLoop() on every new term
	2. mainloop creates ONCE and passes SAME chan to each NewTermLoop() and checks the input value on the channel to know which term it's from.
	3. NewTermLoop() creates and returns it to mainloop

> Currently using option ii - created ONCE on NewLeanHelix(), and passed to a new Term.


6. Decide: Who creates the CreateBlock channel?
> Currently `termloop` creates the channel right before passing it to CreateBlock()

7. POTENTIAL DEADLOCK: (write a test to demonstrate):
With unbuffered channels, mainloop wants to write to termloop.message chan, and termloop wants to write to mainloop.commit_chan
> Could not reproduce

8. Buffered vs unbuffered channels
> Messages channel to be buffered - prevent dangling sender Gossip goroutines

8. Should the Committed channel also be buffered? (in addition to other channels)
> Same as #7
10. Decide which channels are unbuffered (those who must wait for incoming/outgoing message), o/w all should be buffered
> Same as #7
11. The SPI (external service) method `sendMessage()` was only in mainloop but now we have `termloop` that sends messages.
Should `termloop` hold a ref to the SPI?
> Selected solution: lh is holding the SPI ref and passes it to each term.



## Conclusions
### June 5 2019 (w/Shai)
* Get rid of channels for long running processes - let `termloop` wait for long-running completion.
> This is problematic because election will be blocked.
* Write `termloop` component tests (they don't exist today as there is no `termloop` component)
> Good idea, and if ViewLoop goroutine is created, write components tests for it too.
* Consider a Commit/Termination callback instead of passing a commitChannel from `mainloop` into `termloop` 
This is because commitChannel is an implementation detail of `mainloop` and is more difficult to test.
A commit handler can be mocked and can be tested whether it was called or not.
This will wrap a channel, it does not replace a channel.
* Write Acceptance tests: e.g. send PREPREPARE, n * PREPARE, n * COMMIT 
and expect Commit callback to be called.
* Confirm there is no problem with 2 x term alive at the same time (of which only one is the active term)
> No problem, as only one is active, and any other is stale and will be shutdown by context, 
and even it manages to reach consensus on a block, it will be ignored.


## Golang References
* https://go101.org/article/channel-closing.html
* https://blog.golang.org/advanced-go-concurrency-patterns
