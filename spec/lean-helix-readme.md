# LeanHelix Consensus Algo
> This document describes LeanHelix consensus library interfaces and APIs.
> LeanHelix is a PBFT based algorithm, providing block finality based on committees. A new ordered committee, and its leader, is randomly selected for each block. The committee members actively participate in the block consensus.
> LeanHelix is based on [Helix consensus algorithm paper](https://orbs.com/helix-consensus-whitepaper/ "Helix consensus algorithm paper"). LeanHelix does not implement Helix selection fairness properties.

## Design Notes
* Consensus is performed in an infinite loop triggered at a given context state (a BlockProof which holds all necessary information to start next consensus round). For example, a sync scenario flow might shift the consensus loop to a different height.
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

* `UpdateState(previousBlockProof)`
  Initiates participation in a consensus round and terminate participation in an on-going round. Called upon block sync upon processing of a block with height higher than the current one.
* `ValidateBlockConsensus(block, blockProof, prevBlockProof)`
  Validates given block against its BlockProof and its parent BlockProof _(prevBlockProof)_. Called as part of the **block sync** flow upon receiving a new block.
* `StopAt(height)`
  Stops the participation in the consensus when the target height is reached.
* `OnConsensusMessage(message)` - called upon reception of a consensus message.

### Dependent Interfaces
> The interfaces used by the Lean Helix library are provided in a `Configuration interface` on creation and provides the necessary functionalities to operate.

#### ConsensusService
* `Commit(block, blockProof)` - Instructs the service to commit the block (because it successfully passed consensus) together with its BlockProof.
* `NewBlockProof(blockProof_data): BlockProof` - Provides BlockProof serialization.

#### BlockUtils
* `RequestNewBlock(height, prevBlockHash) : block` - called by the OneHeight logic, returns a block interface with a block proposal. This block will then go through consensus.
* `ValidateBlock(height, block) : is_valid` - called by the OneHeight logic.
* `CalcBlockHash(height, block) : block_hash` - called by the OneHeight logic, the consumer service uses its hashing scheme to calculate the hash on a block.

#### Membership
* `MyID(height) : member` - obtain unique identifier for the node, used in consensus process.
* `RequestOrderedCommittee(height, random_seed, Config.commmittee_size) : member_list` -  called at the setup stage of each consensus round (random_seed for round r is determined from the random_seed at round r-1).

#### Communication
* `SendConsensusMessage(height, member_list, message)` - abstraction of sending all consensus related messages [LeanHelix messages](../messages.go). Message may include a Block interface, indicating SendMessageWithBlock.

<!-- I think it should be part fo the SendConsensusMessage, sent to a member list (non-committee)
* `BroadcastPostConsensusMessage(height, message)` - e.g. notify all non committee members of committed block
-->
<!-- moved to API
* `OnConsensusMessage(message)` - relay message to filtering by height.
 -->

#### KeyManager
<!--  * `KeyManager.GetPublicKey(height, SignatureScheme) : PublicKey` - Returnes the node public Public Key. KeyType indicates Consensus / RandomSeed. -->
* `KeyManager.Sign(height, data, SignatureScheme) : signature` - sign using the node's private key. SignatureScheme is an enum with options: Consensus / RandomSeed.
* `KeyManager.Verify(height, data, signature, memberID, SignatureScheme) : valid` - verify the validity of a signature.
* `KeyManager.Aggregate(height, signature_list, memberIDs_list) : signature` - aggregate the RandomSeed signatures.

#### Logger and Monitor 
* `Log(data)` - logs an log event. 
* `Monitor(data)` - reports monitoring data.
    
<!--
#### ElectionTrigger:
* `ElectionTrigger.RegisterOnTrigger(cb) : uid`
* `ElectionTrigger.unregisterOnTrigger(uid)`
--->

#### Additional configurations and interfaces
* Committee size
  * Desired committee size
  
