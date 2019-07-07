# Lean Helix Developer Guide

## Updated threading model

The existing single-threaded model using only a main event loop (`Mainloop`) prevents processing `UpdateState` and `Election` events when waiting on a long-running operation.

### Solution

This is a best-effort solution - `mainloop` will cancel the long-running operation's context and once that happens, control will return to `worker`.
This is not as immediate as dumping the existing `worker` and immediately creating a new one and also relies on the cancellability of the long-running operations. 

* The goroutine `mainloop` will no longer process messages directly, rather it is will delegate all messages to a new `worker` goroutine which is allowed to block.
* `mainloop` will process `UpdateState` and `Election` immediately, as it never waits on any long-running operation
* When `mainloop` receives an `UpdateState` or `Election` it cancels the `worker context` and delegates `UpdateState` or `Election` to the `worker` 
* The `worker` goroutine will process `UpdateState`, `Election` and messages. 

#### Known problems
* `ValidateBlockProposal` could still take a long time because its implementation in Orbs (specifically running contract code in the Processor) does not handle context cancellation.
 
 

### Sequence Diagram
**Work in progress**

Paste this into the [Online Sequence Diagram tool](https://sequencediagram.org/)
```
title Lean Helix with Listener
participantgroup #lightgrey **ORBS**

participantgroup #lightgreen **Goroutine**
participant Orbs_Gossip
end

participantgroup #pink **Goroutine**
participant Orbs_NodeSync
end

participantgroup #yellow **Goroutine **
participant Orbs
end

end

participantgroup #lightgrey **LEAN HELIX**

participantgroup #lightblue ** Goroutine \n   (NEW)**
control Listener
end

participantgroup #steelblue **Goroutine**
control Mainloop
actor Term
actor View
end

end

aboxright over Orbs_Gossip: Message
linear
Orbs_Gossip->>Listener: //message
Listener->>Mainloop: //message
linear off
note over Mainloop: filter()\nhandleMessage()


aboxright over Orbs_NodeSync: Node Sync
Orbs_NodeSync->>Listener: //nodeSync
Listener->Term: cancelTerm()
Term->View: cancelView()
Listener->Term: newTerm()
aboxright over Listener: Election
Listener->View: cancelView()
Listener->View: newView()

aboxright over Orbs:Shutdown
linear
Orbs->Listener: //ctx.Done
Listener->Mainloop: //ctx.Done
linear off

```