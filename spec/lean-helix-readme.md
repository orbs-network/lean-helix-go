# LeanHelix Consensus Algo
> This document describes LeanHelix consensus library interfaces and APIs.
> LeanHelix is a PBFT based algorithm, providing block finality based on committees. A new ordered committee, and its leader, is randomly selected for each block. The committee members actively participate in the block consensus.
> LeanHelix is based on [Helix consensus algorithm paper](https://orbs.com/helix-consensus-whitepaper/ "Helix consensus algorithm paper"). LeanHelix does not implement Helix selection fairness properties.

## Design Notes
* Consensus is performed in an infinte loop triggered at a given context state (a BlockProof which holds all necessary information to start next consensus round). A sync scenario flow, e.g. might shift the consensus loop to a different height.
* This library is dependent on "consumer service" with several context (height based) provided functionalities, detailed below (which could alter its behaviour).
* The proposed design involves another partition into an inner constrained module - "LeanHelixOneHeight" - explicitly devoted to a single round PBFT consensus, further detailed in a seperated file.
  * The "multi-height" library is responsible for looping through the correct term _(height)_, setting the relevant context
  * Including filtering old messages and subsequently relaying future messages at appropriate times.
  * Including, generating the BlockProof and new random_seed
* Configuration related queries are goverened by height - e.g. all known federation members at given height.
* The committee memebers are derived at each block height using an aggregated threshold (set to QuorumSize) signature on previous height's random seed.
* The threshold signatrues are passed as part of the COMMIT messaage.
* COMMIT message is passed to one-height after signature on random seed is verified and signer matches COMMIT signer _(committee member of current height handled at one-height members and discarded + reported )
* COMMIT message holds only one Signer - for both COMMIT BlockRef signature and random seed signature.
* When a block is committed the aggregated signature is comprised matching the QuorumSize COMMIT signed messages _(same members)
* Syncing is perfromed by the consuming service (e.g. BlockStorage), but its validity is justified on BlockProof being verified by LeanHelix library.
* The consensus algo doesn't keep PBFT logs of past block_height (erased on commit). A sync of the blockchain history is perfromed by block sync.
* KeyManager holds a mapping between memberID and its (keyType, publicKey). MemberID = 0 corresponds to master keys _(e.g. in verifying signature aggregation)_.


## Archietcture - components and inetrfaces

### Library API

* `UpdateState(previousBlockProof)`
  Initiates particiaption in a consensus round and terminate particiaption in an on-going round. Called upon block sync upon processing of a block with height higher than the current one.
* `ValidateBlockConsensus(block, blockProof, prevBlockProof)`
  Validates block that the blockProof is valid to the given block. Called as part of the block sync flow upon receiptin of a new block.
* `StopAt(height)`
  Stops the participation in the consensus when the target height is reached.
* `OnConsensusMessage(message)` - called upon reception of a consensus message.

### Dependent Interfaces
> The interfaces used by the Lean Helix library are provided in a `Configuration interface` on creation and procides the necessary functionalities to operate.

#### ConsensusService
* `Commit(block, blockProof)` - Provides a block and a proof upon commit.

#### BlockUtils
* `RequestNewBlock(height, prevBlockHash) : block` - called by the OneHeight logic, returns a block interface with a block proposal _(wait until)_.  
* `ValidateBlock(height, block) : is_valid` - called by the OneHeight logic, valdiates a block proposal.
* `CalcBlockHash(height, block) : block_hash` - called by the OneHeight logic, calculates the hash on a block based on the hashing scheme.

#### Membership
* `MyID(height) : member` - obtain unique identifier for the node, used in consensus process.
* `RequestOrderedCommittee(height, random_seed, Config.commmittee_size) : member_list` -  called at the setup stage of each consensus round (random_seed for round r is determined from the random_seed at round r-1).

#### Communication
* `SendConsensusMessage(height, member_list, message)` - abstraction of sending all consensus related messages [LeanHelix messages](../messages.go)
<!-- I think it should be part fo the SendConsensusMessage, sent to a member list (non-committee)
* `BroadcastPostConsensusMessage(height, message)` - e.g. notify all non committee members of committed block
-->
<!-- moved to API
* `OnConsensusMessage(message)` - relay message to filtering by height.
 -->

#### KeyManager
<!--  * `KeyManager.GetPublicKey(height, KeyType) : PublicKey` - Returnes the node public Public Key. KeyType indicates Consensus / RandomSeed. -->
* `KeyManager.Sign(height, data, KeyType) : signature` - sign using the node's private key. KeyType indicates Consensus / RandomSeed.
* `KeyManager.Verify(height, data, signature, memberID, KeyType) : valid` - verify the validity of a signature.
* `KeyManager.Aggregate(height, signature_list, public_keys_list) : signature` - aggregate the random_seed signatures.

#### Logger and Monitor 
* `Log(data)` - logs an log event. 
* `ReportStatus(data)` - reports monitoring data.
    
<!--
#### ElectionTrigger:
* `ElectionTrigger.RegisterOnTrigger(cb) : uid`
* `ElectionTrigger.unregisterOnTrigger(uid)`
--->

#### Additional configurations and inetrfaces
* Local stroage
  * By default he library uses in memory storage, option to connect a persistent storage.
* Committee size
  * Desired committee size
  
