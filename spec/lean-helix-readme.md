# LeanHelix Consensus Algo
> This document describes LeanHelix consensus library interfaces and APIs.\
> LeanHelix is a PBFT based algorithm, providing block finality based on committees. A new ordered committee, and its leader, is randomly selected for each block. The committee members actively participate in the block consensus.
> LeanHelix is based on [Helix consensus algorithm paper](https://orbs.com/helix-consensus-whitepaper/ "Helix consensus algorithm paper"). LeanHelix does not implement Helix selection fairness properties.

## Design Notes
* Consensus is performed in an infinite loop triggered at a given context state (a pair _(Block,BlockProof)_ which holds all necessary information to start next consensus round). For example, a sync scenario flow might shift the consensus loop to a different height.
* The proposed design involves another partition into an inner constrained module - "LeanHelixOneHeight" - explicitly devoted to a single round PBFT consensus, further detailed in a seperate file. The "multi-height" library is responsible for:
  * Looping through the correct height, setting the relevant context
  * Filtering old messages and subsequently relaying future messages at appropriate times
  * Generating the BlockProof and new RandomSeed
* Configuration related queries may be governed by height - e.g. all known federation members at given height.
* The committee members are derived at each block height using an aggregated threshold (set to QuorumSize) signature on previous height's random seed.
* Syncing is perfromed by the consuming service (e.g. BlockStorage), but its validity is justified on BlockProof being verified by LeanHelix library.
* The consensus algo lib doesn't keep Lean Helix messages of past BlockHeight (erased on commit).
* KeyManager holds a mapping between memberID and its (keyType, publicKey). MemberID = 0 corresponds to master keys _(e.g. in verifying signature aggregation)_.
* Block and BlockProof are serialized by the Consumer service.
* Consensus messages _(excluding the Block, BlockProof)_ are serialized by the library. 
* This library is dependent on "consumer service" with several context (height based) provided functionalities, detailed below (which could alter its behaviour).



## Architecture - components and interfaces

### Library API

* `Run()`
Initiates lean-helix library infinite listening loop.
* `SetContext(prevBlock, prevBlockProof)`
  Called upon node sync. Assumes the matching pair _(prevBlock,prevBlockProof)_ are validated!\
  If given prevBlock->height is at least as on-going round, terminate participation in an on-going round and initiate participation in the subsequent consensus round.
* `ValidateBlockConsensus(block, blockProof, prevBlockProof)`
  Validates given block against its BlockProof and its parent BlockProof _(prevBlockProof)_. Called as part of the **block sync** flow upon receiving a new block.
* `StopAt(height)`
  Stops the participation in the consensus when the target height is reached. [Not implemented yet]
* `OnConsensusMessage(message)` - called upon reception of a consensus message.

### Dependent Interfaces
> The interfaces used by the Lean Helix library are provided in a `Configuration interface` on creation and provides the necessary functionalities to operate.

#### ConsensusService
* `Commit(block, blockProof)` - Instructs the service to commit the block (because it successfully passed consensus) together with its BlockProof.
* `NewBlockProof(blockProof_data): BlockProof` - Provides BlockProof serialization.

#### BlockUtils
* `RequestNewBlock(prevBlock) : block` - called by the OneHeight logic, returns a block interface with a block proposal. This block will then go through consensus. 
* `ValidateBlock(height, block, block_hash, prevBlock) : is_valid` - called by the OneHeight logic. Validates the block structure and content _(match it to given block_hash)_. Note: this could include the timestamp - whithin acceptable range of local clock.
* `CalcBlockHash(height, block) : block_hash` - called by the OneHeight logic, the consumer service uses its hashing scheme to calculate the hash on a block (commitment on block content and structure).

#### Membership
* `MyID(height) : member` - obtain unique identifier for the node, used in consensus process.
* `RequestOrderedCommittee(height, random_seed) : member_list` -  called at the setup stage of each consensus round (random_seed for round r is determined from the random_seed at round r-1). Assumes membership holds the both the federation members and the committee size of the given height.

#### KeyManager
* `KeyManager.SignConsensusMessage(height, data) : signature` - sign using the node's private key. 
* `KeyManager.VerifyConsensusMessage(height, data, signature, memberID) : valid` - verify the validity of a signature.
* `KeyManager.SignRandomSeed(height, data) : signature` - sign using the node's private key. 
* `KeyManager.VerifyRandomSeedShare(height, data, signature, memberID) : valid` - verify the validity of a signature.
* `KeyManager.Aggregate(height, signature_and_memberID_list) : signature` - aggregate the RandomSeed signatures.

#### Communication
* `SendConsensusMessage(height, member_list, message)` - abstraction of sending all consensus related messages [LeanHelix messages](../messages.go). Message may include a Block interface, indicating SendMessageWithBlock.

<!-- I think it should be part fo the SendConsensusMessage, sent to a member list (non-committee)
* `BroadcastPostConsensusMessage(height, message)` - e.g. notify all non committee members of committed block
-->




#### Logger and Monitor 
* `Log(data)` - logs an log event. 
* `Monitor(data)` - reports monitoring data.
    
