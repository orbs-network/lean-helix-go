# New threading model POC

## Questions

1. What happens when trying to read from a garbage collected channel
   (maybe irrelevant question if channel is not collected till no readers remain)
2. What to do with channels waiting for CreateBlock and ValidateBlock when closing the term?
3. What to do with the committed channel when closing term
4. Race: Election timer pops just after closing its parent term
5. Decide: Who creates the committed channel?
	i. mainloop creates and passes to NewTermLoop() on every new term
	ii. mainloop creates ONCE and passes SAME chan to each NewTermLoop() and checks the input value on the channel to know which term it's from.
	iii. NewTermLoop() creates and returns it to mainloop
6. Decide: Who creates the CreateBlock channel?
	i. Same i,ii,iii points as previous question
7. Test potential Deadlock with unbuffered channels: message waits on termloop chan, but termloop create_block
8. Should the Committed channel also be buffered? (in addition to other channels)
9. Test Garbage collection of buffered channel (similar to #1)
10. Decide which channels are unbuffered (those who must wait for incoming/outgoing message), o/w all should be buffered

## Conclusions

