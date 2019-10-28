# Lean Helix Developer Guide

## Go Modules
As of October 2019 Lean Helix switched to Go Modules for its dependency management.
This goes in line with corresponding update to other Orbs repos, and
replaces the previous solution of using a `vendor` folder.

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

## Leader Election
Leader election timeout is based on the current view. The timeout is `base * 2^V`.
For example, if `base = 4 seconds` then in the first view `V=0`, the timeout is `4s*2^0=4s`.
In the second view `V=1` the timeout is `4s*2^1=8s` and so on.

### Architecture
Leader election is governed by an implementation of the interface `ElectionTrigger` under the `interfaces` package.

The Lean Helix implementation is `electiontrigger.TimerBasedElectionTrigger`.

* `MainLoop` accepts in configuration an instance of `ElectionTrigger`.
* `MainLoop` listens on the channel provided by `ElectionTrigger.ElectionChannel()`
* `RegisterOnElection` resets the `height`, `view`, `electionHandler` and resets the timer (whose timeout is based on provided `view`).
** It is called when initializing a new view. For Example, V=2 just started, so `RegisterOnElection` is called with view=2 and the timer is reset to trigger in 16 seconds (`4s*2^2`)
** Internally, the timer is set to invoke `sendTrigger()` when it expires. `sendTrigger()` writes a func called `trigger` to the election channel.

  