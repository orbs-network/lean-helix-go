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
* Signature is opaque on library side.
* Consensus messages _(excluding the Block, BlockProof)_ are serialized by the library. 
* This library is dependent on "consumer service" with several context (height based) provided functionalities, detailed below (which could alter its behaviour).
* BlockProof is serialized by the library _(passed as byte_array)_ 
* Block is serialized by the Consumer service. 



## Architecture - components and interfaces

### Library API

* `NewLeanHelix(config, onCommitCallback)`
* `Run(ctx)`
Initiates lean-helix library infinite listening loop.
* `UpdateState(block, blockProof)`
  Called upon node sync.  Assumes the matching pair _(block,blockProof)_ are validated!\
  Conditional update: If given block->height is at least as on-going round, terminate participation in an on-going round and initiate participation in the subsequent consensus round.
* `ValidateBlockConsensus(block, blockProof, prevBlockProof): is_valid`
  Validates given block against its BlockProof and its parent BlockProof _(prevBlockProof)_. Called as part of the **block sync** flow upon receiving a new block.
* `HandleConsensusMessage(message)` - called upon reception of a consensus message.

### LeanHelix SPI
> The interfaces used by the Lean Helix library are provided in a `Configuration interface` on creation and provides the necessary functionalities to operate.\
> Described below in a suggested separated modules. 

#### BlockUtils
> Provide block funcionalities including its creation, validation and hashing scheme. 
* `RequestNewBlockProposal(ctx, blockHeight, prevBlock) : block, block_hash` - Returns a block interface with a block proposal along with its digest commitment. The block _(block_hash)_ will then go through consensus. 
* `ValidateBlockProposal(ctx, blockHeight, block, block_hash, prevBlock) : is_valid` - Validate block proposal against prevBlock and digest commitment - full validation - content and structure _(Note: this includes validating against previous block _(e.g. pointer _(prevBlockHash)_ and timestamp - whithin acceptable range of local clock)_.)_
* `ValidateBlockCommitment(blockHeight, block, block_hash) : is_valid` - Validate block proposal against digest commitment (shallow structure validation for composite commitment). 

#### Membership
> Hold information about federation members at a given height (this could include their reputation).
* `MyMemberId() : memberID` - obtain unique identifier for the node, used in consensus process.
* `RequestOrderedCommittee(ctx, blockHeight, randomSeed, committeeSize) : memberID_list` -  called at the setup stage of each consensus round (random_seed for round r is determined from the random_seed at round r-1). Assumes membership holds the federation members of the given height.

#### KeyManager
> Provide signature schemes consumed by LeanHelix. \
> Hold a mapping of height, memberID to publicKey (for consensusMessages and RandomSeed)
* `KeyManager.SignConsensusMessage(blockHeight, content) : signature` - sign consensus statements using the node's private key. 
* `KeyManager.VerifyConsensusMessage(blockHeight, content, signature, memberId) : is_valid` - verify the validity of a signature.
* `KeyManager.SignRandomSeed(blockHeight, content) : signature` - sign RandomSeed using the node's private key (note: the randomseed and consesnsus keys are independent). 
* `KeyManager.VerifyRandomSeed(blockHeight, content, signature, memberId) : is_valid` - verify the validity of a signature _(also group aggregated against MasterPublicKey)_.
* `KeyManager.AggregateRandomSeed(blockHeight, randomSeedShares) : signature` - aggregate the RandomSeed signatures.

#### Communication
* `SendConsensusMessage(ctx, blockHeight, recipients, message)` - abstraction of sending all consensus related messages [LeanHelix messages:= {content, block} content is opaque (byte_array) - Message may include a Block].

<!-- I think it should be part fo the SendConsensusMessage, sent to a member list (non-committee)
* `BroadcastPostConsensusMessage(height, message)` - e.g. notify all non committee members of committed block
-->

#### Logger and Monitor 
* `Log(data)` - logs an log event. 
* `Monitor(data)` - reports monitoring data.
    
 
