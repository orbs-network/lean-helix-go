# Proof of concepts for alternative threading models
This folder contains proof of concepts for various threading models, it is not part of the product per-se.
It is here for completeness.


## Questions

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

## Possible Models

1. Termloop goroutine with separate goroutine for long-running like CreateBlock
2. Termloop goroutine without separate goroutine for long-running like CreateBlock 
    (send them ctx.cancel() on election / setView)
3. Termloop goroutine and viewloop goroutine. The viewloop calls CreateBlock synchronously.

> Currently using Option 1



## Conclusions
### June 5 2019 (w/Shai)
* Get rid of channels for long running processes - let `termloop` wait for long-running completion.
* Write `termloop` component tests (they don't exist today as there is no `termloop` component)
* Consider a Commit/Termination callback instead of passing a commitChannel from `mainloop` into `termloop` 
This is because commitChannel is an implementation detail of `mainloop` and is more difficult to test.
A commit handler can be mocked and can be tested whether it was called or not.
* Acceptance tests: e.g. send PPM, n * PM, n * CM and expect Commit callback to be called.
* Confirm with OdedW: (i) there is no problem with 2 x term alive at the same time (of which only one is the active term)
* Confirm with OdedW: (ii) check for flaws in general, talk about testability.

OdedW



## References
* https://go101.org/article/channel-closing.html
* https://blog.golang.org/advanced-go-concurrency-patterns
