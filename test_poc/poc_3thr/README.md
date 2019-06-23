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


