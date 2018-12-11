# LeanHelix ConsumerService
> This document describes LeanHelix ConsensusService consumer and its SPI in detail (../lean-helix-readme.md) based on Orbs-spec(github.com/orbs-network/orbs-spec).  
 


## Design Notes
 





## Configuration:
> Provided on creation, holds all necessary dependent functionalities to run LeanHelix consensus service]. \
> TODO: Reconfiguration  - Config properties implicitly embed the Block_height  \
>  
  <!-- * ConsensusContext 



	NodePublicKey() primitives.Ed25519PublicKey
	NodePrivateKey() primitives.Ed25519PrivateKey
	FederationNodes(asOfBlock uint64) map[string]config.FederationNode
	LeanHelixConsensusRoundTimeoutInterval() time.Duration
	ActiveConsensusAlgo() consensus.ConsensusAlgoType


    	ctx context.Context,
	gossip gossiptopics.LeanHelix,
	blockStorage services.BlockStorage,
	consensusContext services.ConsensusContext,
	parentLogger log.BasicLogger,
	config Config,
	metricFactory metric.Factory,


  * `CommitBlock(Block, commits_list)` - Callback on committedLocally.
  * BlockUtils (RequestNewBlock, ValidateBlock, CalcBlockHash)
  * Communication (SendConsensusMessage)
  * KeyManager (Sign, verify) - Passed by multi-height with PublicKey mapped as MemberID.
  <!-- * Members (ordered_list of participating members, f is implicitly derived) -->
  <!-- * ElectionTrigger (default is timeout based - increasing each view. i.e., trigger after Base_election_timeout*2^View, where Base_election_timeout is embeded by provider)
  * Logger (optional)
  * Monitor (optional, if provided records stats during consensus round)
  * LocalStorage (optional, default in memory - stores messages) --> -->


 #### State
> Stores the current state variables.
* State variables:
    * LeanHelix _(current instance of LeanHelix(Config))_
 

#### ConsensusService
> Relay committed block information from the consesnsus library to the system.



## Init(Context)
> Start consensusService consumer - run LeanHelix instance.
* Initialize the [configuration](../config/services.md).
* Load persistent data (if present)
* Subscribe to gossip messages by calling `Gossip.LeanHelix.RegisterLeanHelixHandler`.
* Register to handle transactions and results blocks validation by calling `BlockStorage.ConsensusBlocksHandler`. 
* Wait for `HandleConsensusBlock` from `BlockStorage` to start the consensus algorithm.
   



## `OnCommit(block, blockProof)`  
> Instructs the service to commit the block (successful consensus) together with its BlockProof. \
> BlockProof is opaque - byte_array.
* Set block.TransactionBlockProof (construct based on blockProof)
* Set block.ResultsBlockProof  (construct based on blockProof) 
* Call blockStorage.CommitBlock(block)

 

#### BlockUtils
> Provide block funcionalities including its creation, validation and hashing scheme. 
&nbsp;
## `RequestNewBlockProposal(prevBlock) : block, block_hash` 
> Returns a block interface with a block proposal along with its digest commitment. \
> The block _(block_hash)_ will then go through consensus.
> Block := { TransactionBlock, ResultsBlock }
* Construct the blockProposal = Block, block_hash (block_hash - digest commitment of Block) on top of prevBlock
* Block:  
    * Get new TransactionsBlock  (by calling ConsensusContext.RequestNewTransactionsBlock) 
    * Get new ResultsBlock  (by calling ConsensusContext.RequestNewResultsBlock based on TransactionsBlock) 
* block_hash
    * digest(digest(TransactionsBlock) XOR digest(ResultsBlock))






##`ValidateBlock(height, block, block_hash, prevBlock) : is_valid`
> Validate block proposal against prevBlock and digest commitment.\
> full validation - content and structure _(Note: this includes validating against previous block _(e.g. pointer _(prevBlockHash)_ and timestamp - whithin acceptable range of local clock)_.)_



##`ValidateBlockHash(height, block, block_hash) : is_valid` 
> Validate block proposal against digest commitment (shallow structure validation for composite commitment).




#### Communication
> Abstraction of sending all consensus related messages. \
> [LeanHelix message:= {content, block} content is opaque (byte_array) - Message may include a Block].

## `SendConsensusMessage(height, memberID_list, message)` 
> Multicast abstraction. \
> member_list member_ids matching membershipmemberID_list





#### Membership
> Hold information about federation members at a given height (this could include their reputation).
## `MyID(height) : memberID` 
> Obtain unique identifier for the node, used in consensus process.


## `RequestCommittee(height, random_seed, committee_size) : memberID_list` 
> Called at the setup stage of each consensus round.\
> random_seed for round r is determined from the random_seed at round r-1.\
> Assumes membership holds the federation members of the given height.





#### KeyManager
> Provide signature schemes consumed by LeanHelix. \
> Hold a mapping of height, memberID to publicKey (for consensusMessages and RandomSeed)
* `KeyManager.SignConsensusMessage(height, data) : signature` - sign consensus statements using the node's private key. 
* `KeyManager.VerifyConsensusMessage(height, data, signature, memberID) : valid` - verify the validity of a signature.
* `KeyManager.SignRandomSeed(height, data) : signature` - sign RandomSeed using the node's private key (note: the randomseed and consesnsus keys are independent). 
* `KeyManager.VerifyRandomSeed(height, data, signature, memberID) : valid` - verify the validity of a signature _(also group aggregated against MasterPublicKey)_.
* `KeyManager.AggregateRandomSeed(height, signature_and_memberID_list) : signature` - aggregate the RandomSeed signatures.









#### Validate message, including Block
* Block_hash = Get Block_hash by calling `Config.BlockUtils.CalcBlockHash(Message.Block)`
* If `ValidatePrePrepare(Message, Block_hash)` Continue
* If Block is not Valid by calling `Config.BlockUtils.ValidateBlock(Message.Block)` Return.
#### Check state still match - Important! state might change during blocking validation process
* If Disposed Return.
* If my_state.View does not match Message.View Return.
#### Continue Process PrePrepare
* Call `ProcessPrePrepare(Message)`  _(Applies for PrePrepare message in New_View as well)_

 

## Architecture - components and interfaces

### Library API

* `Run()`
Initiates lean-helix library infinite listening loop.
* `UpdateState(prevBlock, prevBlockProof)`
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



#### Membership
* `MyID(height) : member` - obtain unique identifier for the node, used in consensus process.
* `RequestOrderedCommittee(height, random_seed) : member_list` -  called at the setup stage of each consensus round (random_seed for round r is determined from the random_seed at round r-1). Assumes membership holds the both the federation members and the committee size of the given height.

#### KeyManager
* `KeyManager.SignConsensusMessage(height, data) : signature` - sign using the node's private key. 
* `KeyManager.VerifyConsensusMessage(height, data, signature, memberID) : valid` - verify the validity of a signature.
* `KeyManager.SignRandomSeed(height, data) : signature` - sign using the node's private key. 
* `KeyManager.VerifyRandomSeed(height, data, signature, memberID) : valid` - verify the validity of a signature.
* `KeyManager.AggregateRandomSeed(height, signature_and_memberID_list) : signature` - aggregate the RandomSeed signatures.

#### Communication
* `SendConsensusMessage(height, member_list, message)` - abstraction of sending all consensus related messages [LeanHelix messages](../messages.go). Message may include a Block interface, indicating SendMessageWithBlock.

<!-- I think it should be part fo the SendConsensusMessage, sent to a member list (non-committee)
* `BroadcastPostConsensusMessage(height, message)` - e.g. notify all non committee members of committed block
-->




#### Logger and Monitor 
* `Log(data)` - logs an log event. 
* `Monitor(data)` - reports monitoring data.
    
