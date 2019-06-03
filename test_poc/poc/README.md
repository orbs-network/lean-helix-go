# New threading model POC

## Questions

1. What happens when trying to read from a garbage collected channel
   NOT GARBAGE COLLECTED TILL WE RELEASE ALL REFERENCES.
2. What to do with channels waiting for CreateBlock and ValidateBlock when closing the term?
DONE
3. What to do with the committed channel when closing term
NOTHING, it just vanishes
4. Race: Election timer pops just after closing its parent term

5. Decide: Who creates the committed channel?
Currently created once on NewLeanHelix()

	i. mainloop creates and passes to NewTermLoop() on every new term
	ii. mainloop creates ONCE and passes SAME chan to each NewTermLoop() and checks the input value on the channel to know which term it's from.
	iii. NewTermLoop() creates and returns it to mainloop
6. Decide: Who creates the CreateBlock channel?
	i. Same i,ii,iii points as previous question
7. Test potential Deadlock with unbuffered channels: mainloop wants to write to termloop.message chan, and termloop wants to write to mainloop.commit_chan

8. Should the Committed channel also be buffered? (in addition to other channels)
Same as #7
10. Decide which channels are unbuffered (those who must wait for incoming/outgoing message), o/w all should be buffered
Same as #7
11. Where is the method sendMessage()? it was only in mainloop but now we have termloop so do we need one sendMessage() for mainloop
and another for termloop?
    i. Selected solution: lh is holding the ref to SPI and passes it to each term.


## Models
1. Termloop goroutine with separate goroutine for long-running like CreateBlock
2. Termloop goroutine without separate goroutine for long-running like CreateBlock 
    (send them ctx.cancel() on election / setView)
3. Termloop goroutine and viewloop goroutine. The viewloop calls CreateBlock synchronously.

## Conclusions

## References
* https://go101.org/article/channel-closing.html
