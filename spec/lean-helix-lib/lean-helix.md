# LeanHelix
> This document details the LeanHelix plug-in specification, focusing  on switching between consensus rounds. The spec for consensus round is described in [LeanHelixOneHeight](/lean-helix-one-height.md). The public API can be found in [LeanHelix](/lean-helix-readme.md).


## Design Notes
* Consensus is performed in an infinite loop triggered at a given context state. A sync scenario flow, e.g. might shift the consensus loop to a different height.
* This library is dependent on "consumer service" with several context provided functionalities, detailed below (which could alter its behaviour).
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

## Databases

#### Received Messages Cache
> Stores messages of future block_height pending stateful processing until block_height update.\
> Used to reduce the chance for costly syncs.\
> Discard if message.block_height > my_state.block_height + configurable future_block_height_window.
* Accessed by (Block_height, View, Signer)
* Not Persistent _(TBD)_. 
* Stores only one valid message per {Block_height, MessageType, Signer, View}
  _(avoid storing duplciates which may be sent as part of an attack)_

<!-- #### Random Seed Cache
> Stores random seed signatures for current consensus round.
* Accessed by (Block_height, Signer)
* Not presistent
* Stores only one valid info per {Block_height, Signer} -->

<!-- #### Previous block cache
* Stores required data from the previous block_height.
 -->


 #### State
> Stores the current state variables.
* State variables:
    * OneHeightContext:
        * Current_block_height (-1)
        * Prev_block_hash ({})
        * Random_seed (0)
    * OneHeight  _(current instance of LeanHelixOneHeight(Current_block_height))_
    * StopHeight (MAX)





&nbsp;
## `OnConsensusMessage(message)`
> Demux message and filter by height.
#### Filter message by block height
* Current_block_height = my_state.OneHeightContext.Current_block_height
* If Current_block_height >= my_state.StopHeight Return  _(stop process at StopHeight)_.
* If message.Block_height > Current_block_height + configurable future_block_height_window - discard.
* If message.Block_height > Current_block_height - store in Future Messages Cache.
* If message.Block_height < Current_block_height - discard.
#### Demux message and pass to One Height PBFT, strip random_seed info from COMMIT
* Determine the message type
* If message type COMMIT:
    * KeyType = Get KeyType to verify for KeyManager.KEY_TYPES.RandomSeed
    <!--  * PublicKey = `KeyManager.GetPublicKey(message.Signer, message.Block_height, KeyType)`  -->
    * Random_seed = my_state.OneHeightContext.Random_seed
    * Validate the random_seed _(current block_height)_ signature by calling `KeyManager.Verify(message.Block_height, Random_seed, message.Random_seed_share, message.Signer, KeyType)`. If failed validation - discard.
    <!-- * Log info to random_seed_database:
        * random_seed_data.add({COMMIT message.Block_height, COMMIT message.Signer, COMMIT message.Random_seed_share}) -->
* Call the corresponding `my_state.OneHeight.On<XXX>`



&nbsp;
## `Start(previousBlockProof)`
> Start infinite consensus at a given context. \
> Derive next height, prevBlockHash and random_seed from previousBlockProof (called by the ConsensusService). \
> Clear old cache. \
> Derive next committee and instantiate a consensus round. \
> Note: StopHeight has to be higher than next consensus round to run
#### Clear old Cache
* Clear Messages Cache based on Block_height
* my_state.OneHeight.Dispose()
* If previousBlockProof.Block_height + 1 >= my_state.StopHeight Return  _(stop process at StopHeight)_.
#### Reset context for consensus
* Set my_state.OneHeightContext:
    * Current_block_height = previousBlockProof.Block_height + 1
    * Prev_block_hash = previousBlockProof.context.Block_hash
    * Calculate the random seed for the upcoming block:
        * Random_seed = SHA256(previousBlockProof.Random_seed_signature).
* MyID = Get by calling `Membership.MyID(block_height)`
#### Get Committee
* Committee = Get an ordered list of committee members by calling `Membership.RequestOrderedCommittee(Current_block_height, Random_seed, Config.Committee_size(Current_block_height))`
#### setup OneHeight config. Override BlockUtils, Communication and KeyManager - embed context
* Generate OneHeight Config _(Pass\\override functionalities from Config)_:
  * `CommitBlock(Block, commits_list)` - Callback on OneHeight.committedLocally.
  * Config.BlockUtils - (RequestNewBlock, ValidateBlock, CalcBlockHash):
      * Config.BlockUtils.RequestNewBlock := `RequestNewBlock()`
      * Config.BlockUtils.ValidateBlock := `ValidateBlock(block)`
      * Config.BlockUtils.CalcBlockHash := `CalcBlockHash(block)`
  * Config.Communication (SendConsensusMessage)
      * Config.Communication.SendConsensusMessage := `SendConsensusMessage(height, member_list, message)`
  * Config.KeyManager (Sign, Verify)
      * Config.KeyManager.Sign := `Sign(object) : signature`
      * Config.KeyManager.Verify := `Verify(object, signature, memberID) : Valid`
  * Config.ElectionTrigger (optional, default is timeout based )
  * Config.Logger (optional)
  * Config.Monitor (optional, if provided records stats during consensus round)
  * Config.LocalStorage (optional, default in memory - stores messages)
#### Start consensus on a specific height
* Call `my_state.OneHeight.NewConsensusRound(Current_block_height, Prev_block_hash, MyID, Committee, Config)`





&nbsp;
## `ValidateBlock(block)`
> Override BlockUtils.ValidateBlock for OneHeight consensus. 
> Validate against current OneHeightContext _(height, prev_block_hash)_.
* Height = my_state.OneHeightContext.Current_block_height.
* Prev_block_hash = my_state.OneHeightContext.Prev_block_hash
* Return `ValidateBlockLogic(block, Height, Prev_block_hash)`



&nbsp;
## `ValidateBlockLogic(block, height, prev_block_hash)`
> Validate against given params _(height, prev_block_hash)_.
#### Check the hash pointers
* If block.Block_height does not match height Return False.
* If block.prev_block_hash does not match prev_block_hash Return False.
* Return `BlockUtils.ValidateBlock(block.Block_height, block)`

&nbsp;
## `RequestNewBlock()`
> Override BlockUtils.RequestNewBlock for OneHeight consensus.
* Return `BlockUtils.RequestNewBlock(my_state.OneHeightContext.Current_block_height, my_state.OneHeightContext.Prev_block_hash)`

&nbsp;
## `CalcBlockHash(block)`
> Override BlockUtils.CalcBlockHash for OneHeight consensus.
* Return `BlockUtils.ValidateBlock(my_state.OneHeightContext.Current_block_height, block)`




## `Verify(object, signature, memberID)`
> Override KeyManager.Verify - PublicKey mapped as MemberID - for OneHeight consensus.
* Height = my_state.OneHeightContext.Current_block_height
* KeyType = Get KeyType to verify for KeyManager.KEY_TYPES.Consensus
<!-- * PublicKey = Get PublicKey by calling `Config.KeyManager.GetPublicKey(memberID, Height, KeyType)` -->
* Return `Config.KeyManager.Verify(Height, object, signature, memberID, KeyType)`



## `Sign(object)`
> Override KeyManager.Sign for OneHeight consensus sign.
* KeyType = Get KeyType to sign for KeyManager.KEY_TYPES.Consensus
* Height = my_state.OneHeightContext.Current_block_height
* Return `Config.KeyManager.Sign(Height, object, KeyType)`


## `SendConsensusMessage(height, member_list, message)`
> Override Communication.SendConsensusMessage - add random_seed signature to COMMIT message.
* Determine the message type
* If message type COMMIT:
    * KeyType = Get KeyType to sign for KeyManager.KEY_TYPES.RandomSeed
    * RandomSeed = my_state.OneHeightContext.RandomSeed
    * Add Random_seed_share to message
        * message.Random_seed_share = Get signature on random_seed by calling `Config.KeyManager.Sign(height, RandomSeed, KeyType)`
* Call `Communication.SendConsensusMessage(height, member_list, message)`



&nbsp;
## `CommitBlock(block, commits_list)`
> called by the OneHieght. \
> Generates a block_proof, propagates the commit, broadcasts block and starts new consesnus round.
* LeanHelixBlockProof = Get by Calling `GenerateLeanHelixBlockProof(commits_list)`
* Commit the BlockPair to consuming service by calling `Config.ConsensusService.Commit(block, LeanHelixBlockProof)`
* Broadcast to all nodes message with block by calling `Config.Communication.SendConsensusMessage(height, message(block))`
* Trigger next consensus round by Calling `Start(LeanHelixBlockProof)`




&nbsp;
## `GenerateLeanHelixBlockProof(commits_list)`
> Generates a block_proof.
#### Generate PBFT proof
* From first _(any)_ COMMIT message Extract (Block_height, View, Block_hash)
* Signers, Signature_pair_list, RandomSeedShare_pair_list = From COMMIT messages in commits_list Extract COMMIT Signers and list of pairs (Signer, Signature) and list of RandomSeedSignatures
* Generate PBFT_proof:
    <!-- * opaque_message_type = COMMIT -->
    * Block_height
    * View
    * Block_hash
    * SignaturesPairs = Signature_pair_list
#### Generate random seed with proof
<!-- * From random_seed_data extract list of pairs (Signers, Random_seed_share)
    * RandomSeedShare_list, Signers_list = Get from random_seed_data(Block_height, Signers_(use Signers from the commits_list)_) -->
* Aggregate the threshold signatrue
    * RandomSeed_signature = Get by calling `KeyManager.Aggregate(Block_height, RandomSeedShare_pair_list)`
#### Generate LeanHelixBlockProof
* Generate a LeanHelixBlockProof
  * PBFT_proof
  * RandomSeed_signature
 &nbsp;
* Return LeanHelixBlockProof.


&nbsp;
## `QuoromSize(numMemebers)`
> Calculate the quorum size based on number of participating members.
* f = Floor[(numMemebers-1)/3]
* QuorumSize = numMemebers - f
* Return QuorumSize



&nbsp;
## `ValidateBlockConsensus(block, blockProof, prevBlockProof)`
> Called by the ConsensusService - e.g. by blockstorage as part of block sync. \
> Assumes block content valid , verifies the blockProof is valid.
* From blockProof Extract (Block_height, View, Block_hash, SignaturesPairs)
* If block.Block_height does not match Block_height Return False.
* Calculate the random seed from prevBlockProof:
    * Random_seed = SHA256(prevBlockProof.Random_seed_signature).
#### Get Committee
* Committee = Get an ordered list of committee members by calling `Membership.RequestOrderedCommittee(Block_height, Random_seed, Config.Committee_size(BlockHeight))`
* QuorumSize = `QuoromSize(Committee.length)`
#### validate PBFT signatures
* If SignaturesPairs are not QuorumSize of all unique Signers Return False.
* KeyType = Get KeyType to verify for KeyManager.KEY_TYPES.Consensus
*  Generate the signature data - COMMIT_HEADER:
    * Message_Type = COMMIT
    * Block_height
    * View
    * Block_hash
* For each SignaturePair in SignaturesPairs:
    * If SignaturePair.Signer is not in Committee Return False.
    <!-- * PublicKey = Get PublicKey by calling `Config.KeyManager.GetPublicKey(SignaturePair.Signer, Block_height, KeyType)` -->
    * If `Config.KeyManager.Verify(Block_height, COMMIT_HEADER, SignaturePair.Signature, SignaturePair.Signer, KeyType)` fails Return False.
#### validate random seed signature against master publicKey 
* KeyType = Get KeyType to verify for KeyManager.KEY_TYPES.RandomSeed
* Random_seed_signature = blockProof.Random_seed_signature
* Random_seed  _(calc above)_
<!-- * Get master public key:
    * PublicKey = call `KeyManager.GetPublicKey(memberID = 0, Block_height, KeyType)` -->
* If `Config.KeyManager.Verify(Block_height, Random_seed, Random_seed_signature, memberID = 0, KeyType)` fails Return False.
#### validate block
* If `ValidateBlockLogic(block, Block_height, prevBlockProof.Block_hash)` fails Return False.

#### Passed all validation
* Return Valid



&nbsp;
## `StopAt(height)`
> Stops consensus performed on blocks when reaching height.
* my_state.StopHeight = height
* If my_state.OneHeightContext.Current_block_height >= my_state.StopHeight
    * Stop the OneHeight consensus round by calling `my_state.OneHeight.Dispose()`







<!--

    private networkMessagesFilter: NetworkMessagesFilter;




 #### PBFT Verifier
* `Interfaces:
  * `CheckPBFTProof (pbft_proof, block_hash)` - validates a pbft proof for a specific block hash.








#### Generate block proof and commit block
* Aggregate the threshold signatrues of the logged COMMIT messages in (View = message.view) to generate an aggregated threshold signatrue.
* Generate a LeanHelixBlockProof for the TransactionsBlock based on Log(View = message.view):
  * opaque_message_type = COMMIT
  * block_height = my_state.Block_height
  * View = my_state.View
  * block_hash_mask = SHA256(ResultsBlockHeader)
  * block_hash = SHA256(TransactionBlockHeader)
  * For each COMMIT in (View = message.View)
    * block_signatures.add({COMMIT message.Sender, COMMIT message.Signature})
  * random_seed_signature = aggregated threshold signatrue


* Interfaces:

  * `Start(prevBlockHeader)` - start infinite consensus from prevBlockHeader.height + 1 - dependent on blockstorage.
  * `NotifyCommitted(block)` - notify on committed block event - push commit block to blockStorage.
  * `VerifyBlockConsensus (block - headers + proof)` - called by the Block storage as part of block sync (dependent on blockstorage.getBlockHeader(block -> height-1)).
  * `CommitBlock(block, commits_list)` - called by the OneHieght, generates a block_proof, propagates the commit and starts new consesnus round.
  * `Stop(height)` - stops consensus performed on blocks when reaching height.
  *

  *
  * Gossip:
    * `GossipMessageReceived` - triggered by the gossip service upon message received with Conensus topic. Cache future (by height) messages.
    * `MulticastMessage` - called by the OneHeight.
    * `UnicastMessage` - called by the OneHeight.
  *



* Configuration:
  * BlockStorage
  * Gossip
  * ConsensusContext
  * keyManager
  * logger
  * ConsensusStorage
  * Committee_size, f


* Internal methods:
  * `StartNewConsensusRound(height, committee)` - triggered upon block n-1 commit or by Start. OrderedNodes are the committee elected for this height.
  * `ClearOldRound()` - clean log

  * Construct New one-height
    * id (default `Config.KeyManager.MyID()`)
    * Block_height
    * Prev_block_hash
    * Ordered_members := Committee
    * f  _(optional)_
    * Config(Block_height) _(Pass configuration based on height. e.g., Communication.SendConsensusMessage(height, Data)_



#### One Height LeanHelix
> Performs a single round of consensus resulting in a committed block. Created by Helix
>
## Design Notes
* State transitions depend on message log.


* Interfaces:
  * `NewConsensusRound(height, prev_block_hash, committee)` - perfroms a single block height consensus amongst given nodes.
  * PBFT Message processing (Assume messages are filtered by height and node_list):
    * `OnPrePrepareReceived (PrePrepare)`
    * `OnPrepareReceived (Prepare)`
    * `OnCommitReceived (Commit)`
    * `OnViewChangeReceived (ViewChange)`
    * `OnNewViewReceived (NewView)`
  * query message log:
    * `Store/GetPrePrepare(view): PrePrepare, block` - only a single block is possible
    * `Store/GetPrepares(view, block_hash): Prepare_list`
    * `Store/GetCommits(view, block_hash): Commit_list`
    * `Store/GetViewChanges(view): ViewChange_list`

* Internal methods:
  * `GetLeader(view)` - internal, calcualtes the leader for the view based on the ordered node list.
  * `StartNewView(view)` -

  *
  * state transition:
  * `OnPrepared` -
  * `OnCommitted`
  * `OnNewView`


#### PBFT Verifier
* `Interfaces:
  * `CheckPBFTProof (pbft_proof, block_hash)` - validates a pbft proof for a specific block hash. -->
