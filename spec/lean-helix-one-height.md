# LeanHelixOneHeight
> This document describes the inner module of LeanHelix plug-in, for one height block consensus.
> Devided into three parts: Happy Flow, Leader Change, Example Flows.


## Design notes
* This module strictly exists _(instantiated and disposed)_ only during a single block_height consensus process - i.e. does not persist to next block_height consensus round.
* The consensus process is based on signed messages, where the signature is verified and assumed to be unforgible _(Counting a Quorum certificate of QuorumSize is based on unique signatures - indicating QuorumSize creators\\signers amongst the known member list)_.
* Assume all incoming messages hold the same block_height as defined in OneHeight state _(Guaranteed by multi-height filter)_.
* PREPARE and COMMIT messages are broadcasted to all member nodes.
* COMMIT messages of earlier\future views are accepted and processed against message Log.
    * A block can be committed (Commit_locally) even if not in Prepared state (The block was received in PRE_PREPARE or NV_PRE_PREPARE)
* The state conditions are checked against MessageLog.
* Message Log contains unique signed valid messages
    * validation at entery point.
    * Criteria checked against messages with same params _(e.g. For Commit Criteria count QuorumSize unique signed messages (Message_type=Commit, block_height, block_hash), where each message signer appears once.)_
* All messages processing according to a basic flow:
    * Validate message
    * Update message Log
    * Check relevant conditions
* Note: Criteria for PreparedLocally and\\or CommittedLocally could be met once an appropriate PRE_PREPARE message is received.
* Dispose is called by multi-height.
* Is Disposed is checked at beginning of validation _(instead of at checking criteria)_ to prevent any further processing of messages.
* PreparedProof in VIEW_CHANGE does not hold block _(only Block_hash)_
* Report conditions which failed validation (e.g., Leader for View sent PREPARE message)
* A node validates a proposed Block only once, on first encounter. Specifically, subsequent NewView message containing the same Block will not trigger a new validation.



## Messages
* PRE_PREPARE - sent by the leader
* PREPARE - sent by the validators
* COMMIT - sent by all nodes, incldues the random seed signature.
* VIEW_CHANGE - sent upon timeout to the next leader candidate
* NEW_VIEW - sent by the newly elected leader (upon view change)



## Configuration:
> Provided on creation by multi-height. Holds all necessary functionalities for one-height to run. \
> TODO: Config properties implicitly embed the Block_height ? could embed and hide.\
> The Config properties are accessed through my_state. In this document it is referenced as Config.Property ((e.g. Config.ElectionTrigger).
  * `CommitBlock(Block, commits_list)` - Callback on committedLocally.
  * BlockUtils (RequestNewBlock, ValidateBlock, CalcBlockHash)
  * Communication (SendConsensusMessage)
  * KeyManager (Sign, verify) - Passed by multi-height with PublicKey mapped as MemberID.
  <!-- * Members (ordered_list of participating members, f is implicitly derived) -->
  * ElectionTrigger (default is timeout based - increasing each view. i.e., trigger after Base_election_timeout*2^View, where Base_election_timeout is embeded by provider)
  * Logger (optional)
  * Monitor (optional, if provided records stats during consensus round)
  * LocalStorage (optional, default in memory - stores messages)


## Databases

#### Messages Log
> Stores the current block_height messages after validation process. i.e., it is assumed the message Log stores valid messages.
* Accessed by (Message_type, Signer, View, Block_hash) or subset of params
* No need to be persistent
* Stores only one valid message per {MessageType, Signer, View}
  _(avoid storing duplciates which may be sent as part of an attack)_

#### State
> Stores the current state variables.\
> Maintain trigger once state variables per view - preparedLocally, NewViewLocally, and committedLocally per current instance (block_height).
* State variables:
  * My_ID - read only! _(E.g., Node public key)_
  * Block_height - read only! _(Term, current round of consensus indicating a single slot in blockchain)_
  * Prev_block_hash - read only! _(ref to previous block at Block_height-1, relayed to stateless BlockUtils in RequestNewBlock)_
  * View _(Derive leader based on members[view mod memebers.length])_
  * PreparedLocally _(Indicating a preparedProof construction is possible at set view. Triggered at most once per view)_
  * NewViewLocally _(Indicating a new leader has been elected and accepted. Triggered at most once per view)_
  * CommittedLocally _(Indicating node obtained members agreement on a Block for current block_height - safe to write to ledger. Triggered once)_
  * Disposed _(If disposed do not process events)_


## Happy Flow
> The document first details flow up to commit _(consensus on Block)_.


&nbsp;
## `NewConsensusRound(block_height, prev_block_hash, id, ordered_members, Config)`
> Performed upon a new consensus round

#### `Init my_state`
* my_state.My_ID = id
* my_state.Block_height = block_height
* my_state.Prev_block_hash = prev_block_hash
* my_state.Members = ordered_members
* my_state.View = -1
* my_state.PreparedLocally = -1
* my_state.NewViewLocally = -1
* my_state.CommittedLocally = False
* Init Messages Log
* Call `InitView(0)` _(Timeout is set)_

#### `Leader Only`
* If not `IsLeader(my_state.View, my_state.My_ID)` Return.
* Block = Request new block proposal by calling `Config.BlockUtils.RequestNewBlock()`.
* Block_hash = Get Block_hash by calling `Config.BlockUtils.CalcBlockHash(Block)`
* Generate PRE_PREPARE message with signature. Store in Log and broadcast to all member nodes.
    * PRE_PREPARE_HEADER:
        * Message_Type = PRE_PREPARE
        * View = my_state.view
        * Block_height = my_state.block_height
        * Block_hash
    * Signer = my_state.My_ID
    * Signature = Get Signature by calling `Config.KeyManager.Sign(PRE_PREPARE_HEADER)`
    * Block _(constructed block proposal)_
* Update the Messages Log
    * Log the PRE_PREPARE message
* Send PRE_PREPARE Message to members
    * `Config.Communication.SendConsensusMessage(my_state.Block_height, my_state.Members, PRE_PREPARE message)`


&nbsp;
## `InitView(View)`
> Trigger once per view.
* If my_state.View >= View Return. _(i.e. view was already intialized in at least as updated state)_
* my_state.View = View
* Config.ElectionTrigger <- Reset(View)  _(set a new ElectionTrigger - leader change, based on view - Call `OnElectionTriggered(View)`)_


&nbsp;
## `QuoromSize()`
> Calculate the quorum size based on number of participating members.
* f = Floor[(my_state.Members.length-1)/3]
* QuorumSize = my_state.Members.length - f
* Return QuorumSize


&nbsp;
## `IsLeader(View, ID)`
> Check ID is leader for View
* LeaderID = `GetLeaderID(View)`
* Return LeaderID == ID

&nbsp;
## `GetLeaderID(View)`
> Deduce leader based on view and memebers
* LeaderIndex = View Modulo my_state.Members.Length
* LeaderID = my_state.Memebers[LeaderIndex]
* Return LeaderID

## `IsMember(ID)`
> Confirm ID is a node in Members
* Return ID in my_state.Members


&nbsp;
## `Dispose()`
> Stop processing events.
* my_state.Disposed = True
* Clear Messages Log
* Clear ElectionTrigger

&nbsp;
## `OnPrePrepareReceived(Message)`
> Process a leader block proposal.
#### Validate message, including Block
* Block_hash = Get Block_hash by calling `Config.BlockUtils.CalcBlockHash(Message.Block)`
* If `ValidatePrePrepare(Message, Block_hash)` Continue
* If Block is not Valid by calling `Config.BlockUtils.ValidateBlock(Message.Block)` Return.
#### Check state still match - Important! state might change during blocking validation process
* If Disposed Return.
* If my_state.View does not match Message.View Return.
#### Continue Process PrePrepare
* Call `ProcessPrePrepare(Message)`  _(Applies for PrePrepare message in New_View as well)_


&nbsp;
## `ValidatePrePrepare(Message, Block_hash)`
> Validate a block proposal message. Make sure state match. Assume block_height was filtered.\
> Report failed validation.\
> Also used upon receiving New_View.
* If Disposed Return False
* If my_state.View does not match Message.View Return False.
* If signature mismatch Return False.
* If signer is not leader of Message.View Return False _(`IsLeader(Message.View, Message.Signer)`)_.
* If Block_hash does not match Message.Block_hash Return False.
* If PRE_PREPARE message already in MessagesLog matching Message(View, Signer, Message_type) Return False.
* Passed validation Return True.



&nbsp;
## `ProcessPrePrepare(Message)`
> Continue Process PrePrepare - Create Prepare message, Log and send.\
> Also used upon receiving New_View.
* Update the Messages Log
  * Log the PRE_PREPARE message
* Generate PREPARE message with signature.
    * PREPARE_HEADER:
        * Message_Type = PREPARE
        * View = my_state.view
        * Block_height = my_state.block_height
        * Block_hash = Message.Block_hash
    * Signer = my_state.My_ID
    * Signature = Get Signature by calling `Config.KeyManager.Sign(PREPARE_HEADER)`
* Update the Messages Log
  * Log the PREPARE message
* Send PREPARE Message to members
    * `Config.Communication.SendConsensusMessage(my_state.Block_height, my_state.Members, PREPARE message)`
#### Continue Process Check Criteria - Check if PreparedLocally and\or CommittedLocally
* If `CheckPreparedLocally(View, Block_hash)`  &nbsp;
    * Call &nbsp; `OnPreparedLocally(View, Block_hash)` _(Check if CommittedLocally inside OnPreparedLocally method)_
* Else _(Node might still be CommittedLocally)_
    * If `CheckCommittedLocally(View, Block_hash)`  &nbsp;
        * Call &nbsp; `OnCommittedLocally(View, Block_hash)`



&nbsp;
## `OnPrepareReceived(Message)`
> Process PREPARE message.
#### Validate and Log message
* If `ValidatePrepare(Message)` Continue
* Update the Messages Log
  * Log the PREPARE message
#### Continue Process Check Criteria - Check if PreparedLocally
* If `CheckPreparedLocally(View, Block_hash)`  &nbsp;
    * Call &nbsp; `OnPreparedLocally(View, Block_hash)` _(Check if CommittedLocally inside OnPreparedLocally method)_


&nbsp;
## `ValidatePrepare(Message)`
> Validate a PREPARE message. Make sure state match. Assume block_height was filtered.\
> Report failed validation.
* If Disposed Return False
* If my_state.View is more advanced (>) Message.View Return False. _(Do not process "Old" PREPARE messages)
* If signer is not valid node member _(`IsMember(Message.Signer)`) Return False.
* If signer is leader of Message.View Return False _(`IsLeader(Message.View, Message.Signer)`)_. _(Leader is not allowed to send PREPARE messages)
* If signature mismatch Return False.
* If PREPARE message already in MessagesLog matching Message(View, Signer, Message_type, Block_hash) Return False.
* Passed validation Return True.


&nbsp;
## `CheckPreparedLocally(Block_height, View, Block_hash)`
> Check if node locked(view based) on current Block proposal. i.e., MessageLog holds a PreparedProof.\
> PreparedProof:= (QuorumSize - 1) PREPARE and 1 PRE_PREPARE _(the QuorumSize signers are unique)_ for matching params _(Block_height, View, Block_hash)_. \
> Note: a node could lock on a different block_hash for the same Block_height at a more advanced View.\
> Assume MessagesLog holds valid messages. \
> Trigger Once per view.
* If my_state.PreparedLocally >= View Return False _(Already preparedProof for up to date state)_
* If no PRE_PREPARE message in MessagesLog(View, Block_hash) Return False.
* If Count PREPARE messages, not signed by leader _(validated on PREPARE entry point)_, in MessagesLog(View, Block_hash) < (`QuorumSize()` - 1) Return False.
* Passed criteria Return True.


&nbsp;
## `OnPreparedLocally(Block_height, View, Block_hash)`
> In possesion of PreparedProof for matching params _(Block_height, View, Block_hash)_, continue flow to commit phase.
> Create Commit message, Log and send, check if committed
* my_state.PreparedLocally = View;
* Generate COMMIT message with signature.
    * COMMIT_HEADER:
        * Message_Type = COMMIT
        * Block_height
        * View
        * Block_hash
    * Signer = my_state.My_ID
    * Signature = Get Signature by calling `Config.KeyManager.Sign(COMMIT_HEADER)`
* Update the MessagesLog
  * Log the COMMIT message
* Send COMMIT Message to members
    * `Config.Communication.SendConsensusMessage(my_state.Block_height, my_state.Members, COMMIT message)`
#### Continue Process Commit - Check if Committed _(skip validate own message)_
* If `CheckCommittedLocally(View, Block_hash)`  &nbsp;
    * Call &nbsp; `OnCommittedLocally(View, Block_hash)`




&nbsp;
## `OnCommitReceived(Message)`
> Process COMMIT message.
#### Validate and Log message
* If `ValidateCommit(Message)` Continue
* Update the Messages Log
  * Log the COMMIT message
#### Continue Process Check Criteria - Check if CommittedLocally
* If `CommittedLocally(View, Block_hash)`  &nbsp;
    * Call &nbsp; `OnCommittedLocally(View, Block_hash)` _(Check if CommittedLocally inside OnPreparedLocally method)_



&nbsp;
## `ValidateCommit(Message)`
> Validate a COMMIT message. Make sure state match. Assume block_height wsas filtered.\
> Report failed validation.\
> Note: node receives COMMIT message even if View does not match _(i.e., future and old)_
* If Disposed Return False
* If signer is not valid node member _(`IsMember(Message.Signer)`) Return False.
* If signature mismatch Return False.
* If COMMIT message already in MessagesLog matching Message(View, Signer, Message_type, Block_hash) Return False.
* Passed validation Return True.



&nbsp;
## `CheckCommittedLocally(View, Block_hash)`
> Check if reached a consensus on Block matching Block_hash. i.e., MessageLog holds
> CommittedProof:= QuorumSize unique COMMIT and 1 PRE_PREPARE for matching params _(Block_height, View, Block_hash)_. \
> Note: safe to write Block to ledger, no other value could be accepted _(under assumptions)_. \
> Assume MessagesLog holds valid messages.\
> Trigger Once per instance.
* If my_state.CommittedLocally Return False _(Already committed Block)_
* If no PRE_PREPARE message in MessagesLog(View, Block_hash) Return False.
* If Count COMMIT messages in MessagesLog(View, Block_hash) < (`QuorumSize()`) Return False.
* Passed criteria Return True.


&nbsp;
## `OnCommittedLocally(View, Block_hash)`
> In possesion of CommittedProof for matching params _(View, Block_hash)_, "End happy flow".
> Note: Dispose is called by outer module.
* my_state.CommittedLocally = True
* Commits_list = Get `QuorumSize()` unique COMMIT messages in MessagesLog(View, Block_hash)
* Block = Get Block matching Block_hash
* Call `Config.CommitBlock(Block, commits_list)`






&nbsp;\
&nbsp;

***
***

&nbsp;\
&nbsp;
## Leader Change
> The second part describes leader change protocol spec.



&nbsp;
## `OnElectionTriggered(View)`
> Propose a leader change.\
> PreparedProof in VIEW_CHANGE does not hold block _(only Block_hash)_
* PreparedProof = Get Prepared Proof by calling `GetPreparedProof()`
* Block = `GetPreparedBlock(PreparedProof)`
* Call `InitView(View + 1)` _(Resets Election trigger, update to next leader View)_
#### Node will send block with VIEW_CHANGE message but only signs the Block_hash ref in PreparedProof
* Generate VIEW_CHANGE message with signature.
    * VIEW_CHANGE_HEADER
        * Message_Type = VIEW_CHANGE
        * View = my_state.View
        * Block_height = my_state.Block_height
        * PreparedProof
    * Block _(Matches the Block_hash in PreparedProof - could be None)_
    * Signer = my_state.My_ID
    * Signature = Get Signature by calling `Config.KeyManager.Sign(VIEW_CHANGE_HEADER)`

#### Leader logs message and checks criteria, others send to leader
* If `IsLeader(my_state.View, my_state.My_ID)`
    * Update the Messages Log
        * Log the VIEW_CHANGE message
        #### Continue Process Check Criteria - Check if ElectedLocally
    * If `CheckElectedLocally(View)` _(skip validation of own message)_
        * Call &nbsp; `OnElectedLocally(View)`
* Else _(not leader)_
    * Send VIEW_CHANGE Message to leader
        * LeaderID = `GetLeaderID(my_state.View)`
        * `Config.Communication.SendConsensusMessage(my_state.Block_height, LeaderID, VIEW_CHANGE message)`


<!-- &nbsp;
## `GetPreparedProofWithBlock()`
> Extract the most recent PreparedProof _(or None)_ from MessageLog with its corresponding Block\
> Assume MessageLog holds valid messages and my_state.PreparedLocally changed after condition met\
> Note: extract based on leader proposal for set view
* If my_state.PreparedLocally == -1 Return None _(No PreparedProof)_
* PrePrepareMessage = MessageLog.GetMessages(PRE_PREPARE, my_state.PreparedLocally)
* PrepareMessages = MessageLog.GetMessages(PREPARE, my_state.PreparedLocally, PrePrepareMessage.Block_hash)[take QuorumSize]
* Generate PreparedProof
  * PrePrepareMessageWithoutBlock =  _(PRE_PREPARE_HEADER, Signer, Signature)_
  * PrepareMessages
* Return PreparedProof, PrePrepareMessage.Block -->

&nbsp;
## `GetPreparedProof()`
> Extract the most recent PreparedProof _(or None)_ from MessageLog with its corresponding Block\
> Assume MessageLog holds valid messages and my_state.PreparedLocally changed after condition met\
> Note: extract based on leader proposal for set view
* If my_state.PreparedLocally == -1 Return None _(No PreparedProof)_
* PrePrepareMessage = MessageLog.GetMessages(PRE_PREPARE, my_state.PreparedLocally) _(Only one PRE_PREPARE message per View)_
* PrepareMessages = MessageLog.GetMessages(PREPARE, my_state.PreparedLocally, PrePrepareMessage.Block_hash)[take (`QuorumSize()` - 1)] _(Block_hash included - might be different for PREPARE with the same View)_
* Generate PreparedProof
  * PrePrepareMessageWithoutBlock =  _(PRE_PREPARE_HEADER, Signer, Signature)_
  * PrepareMessages
* Return PreparedProof

&nbsp;
## `GetPreparedBlock(PreparedProof)`
> Get Block based on the PrePrepare message in PreparedProof
> Assume already in PreparedLocally
> Assume MessageLog holds valid messages, i.e., VIEW_CHANGE with PreparedProof also holds Block
#### Get the block from the PrePrepare message matching PreparedProof
* View = PreparedProof.PrePrepareMessageWithoutBlock.View
* PrePrepareMessage = MessageLog.GetMessages(PRE_PREPARE, View) _(Only one PRE_PREPARE message per View)_
* Block =  PrePrepareMessage.Block
* Return Block



&nbsp;
## `ValidatePreparedProof(PreparedProof, Block_hash)`
> Validate PreparedProof mirror of GetPreparedProofWithBlock. \
> Report if validation failed. \
> Compare all Block_hash in PreparedProof to given - as part of node Vote check _(validating ElectedProof)_ or leader received VIEW_CHANGE. \
> Note: Empty case without PreparedProof and BlockHash passes validation
* If PreparedProof is None and Block_hash is None
    * Return True
* PrePrepareMessage = PreparedProof.PrePrepareMessageWithoutBlock
* PrepareMessages = PreparedProof.PrepareMessages
* If PrepareMessages < (`QuorumSize()` - 1) Return False.
* If not all Params (Block_height, View, Block_hash) in PrePrepareMessage and PrepareMessages match Return False.
* If PrePrepareMessage.Block_hash does not match given Block_hash Return False.
### validate signatures
* If signer is not leader of PrePrepareMessage.View Return False _(`IsLeader(PrePrepareMessage.View, PrePrepareMessage.Signer)`)_.
* If PrePrepareMessage signature mismatch Return False.
* For all PrepareMessages:
    *  If signer is not valid node member _(`IsMember(PrepareMessage.Signer)`) Return False
    *  If signature mismatch Return False.
* All signers are unique - a total of `QuorumSize()` signers _(implicitly Leader is not allowed to send PREPARE messages)_
* Passed validation Return True.


&nbsp;
## `OnViewChangeReceived(Message)`
> Process VIEW_CHANGE message. Only if leader of Message.View
#### Validate and Log message
* If `ValidateViewChange(Message, Mode = MSG)` Continue
* Update the Messages Log
  * Log the VIEW_CHANGE message
#### Continue Process Check Criteria  - Leader of Message.View elected
* If `CheckElectedLocally(Message.View)`
    * Call &nbsp; `OnElectedLocally(View)`


&nbsp;
## `ValidateViewChange(Message, Mode)`
> Validate a VIEW_CHANGE message. Make sure state match. Assume block_height was filtered.\
> Mode indicates whether check as part of ElectedProof i.e., validatePreparedProof without Block. \
> Only if leader of Message.View\
> Report failed validation. \
> Note: VIEW_CHANGE of View + 1 might be common when nodes are out of sync. \
> Note: Could optimize and save ViewChange even without matching block. \
> Note: Accepting ViewChange with Block but no PreparedProof is Valid, this Block will not be accounted in any flow.
* If Disposed Return False
* If node is not leader of Message.View Return False _(`IsLeader(Message.View, my_state.ID)`)_.
* If my_state.View is more advanced (>) Message.View Return False. _(Do not process "Old" VIEW_CHANGE messages)_
* If signer is not valid node member _(`IsMember(Message.Signer)`) Return False.
* If signature mismatch Return False.
* If VIEW_CHANGE message already in MessagesLog matching Message(View, Signer, Message_type, Block_hash) Return False.
* Block_hash = None _(Block_hash sould be None if PreparedProof is None)_
* If Message has PreparedProof:
    * If Mode == MSG _(Leader received VIEW_CHANGE)_
        * Block_hash = Get Block_hash by calling `Config.BlockUtils.CalcBlockHash(Message.Block)`
    * Else If Mode == PROOF _(Check leader Votes)
        * Block_hash = Message.PreparedProof.PrePrepareMessage.Block_hash
* If not `ValidatePreparedProof(PreparedProof, Block_hash)` Return False.
* Passed validation Return True.


&nbsp;
## `CheckElectedLocally(View)`
> Check if node was elected for View. i.e., MessageLog holds a ElectedProof.\
> ElectedProof:= QuorumSize unique VIEW_CHANGE for matching View. \
> Assume MessagesLog holds valid messages. \
> Trigger Once per view.
* If my_state.NewViewLocally >= View Return False _(Already up to date state)_
* If Count VIEW_CHANGE messages in MessagesLog(VIEW_CHANGE,View) < `QuorumSize()` Return False.
* Passed criteria Return True.


&nbsp;
## `OnElectedLocally(View)`
> In possesion of ElectedProof for View, continue flow to notify all members. \
> NEW_VIEW message contains PRE_PREPARE message with Block. \
> Leader proposes its own, new generated Block, only if no node in ElectedProof is PreparedLocally
* my_state.NewViewLocally = View;
* Call `InitView(View)` _(Resets Election trigger - triggers once checked inside InitView)_
* ElectedProof = Get ElectedProof by calling `GetElectedProof(View)` _(ElectedProof sorted by ViewChange.PreparedProof.View Desc)
* Block = Get Block by calling `GetNewViewBlock(ElectedProof)`
* Generate standalone PRE_PREPARE message with signature as part of NEW_VIEW message
    * PRE_PREPARE_HEADER:
        * Message_Type = PRE_PREPARE
        * View = my_state.view
        * Block_height = my_state.block_height
        * Block_hash = Get Block_hash by calling `Config.BlockUtils.CalcBlockHash(Block)`
    * Signer = my_state.My_ID
    * Signature = Get Signature by calling `Config.KeyManager.Sign(PRE_PREPARE_HEADER)`
    * Block
* Update the Messages Log
  * Log the PRE_PREPARE message
* Generate NEW_VIEW message with signature.
    * NEW_VIEW_HEADER:
        * Message_Type = NEW_VIEW
        * Block_height
        * View
        * ElectedProof
    * PrePrepare message
    * Signer = my_state.My_ID
    * Signature = Get Signature by calling `Config.KeyManager.Sign(NEW_VIEW_HEADER)`
* Send NEW_VIEW Message to members
    * `Config.Communication.SendConsensusMessage(my_state.Block_height, my_state.Members, NEW_VIEW message)`


&nbsp;
## `GetElectedProof(View)`
> Generate leader proof for View
> Assume already in OnElectedLocally
* ViewChangeMessages = MessageLog.GetMessages(VIEW_CHANGE, View)
#### take top QuorumSize _(sorted by PreparedProof.View)_
* SortedViewChangeMessages =  Sort ViewChangeMessages by PreparedProof.View Desc
* TopViewChangeMessages = SortedViewChangeMessages[take first `QuorumSize()`]
* ElectedProof = TopViewChangeMessages.WithoutBlock() _(VIEW_CHANGE_HEADER, Signer, Signature)_
* Return ElectedProof


&nbsp;
## `GetNewViewBlock(ElectedProof)`
> Get Block from the VIEW_CHANGE message with highest PreparedProof.View.\
> Assume ElectedProof ordered by PreparedProof.View Desc.\
> Assume MessageLog holds valid VIEW_CHANGE messages, i.e., VIEW_CHANGE with PreparedProof also holds Block.
#### Get the block based on the ViewChangeMessage in ElectedProof with highest PreparedProof.View
* ViewChangeMessage = MessageLog.GetMessages(VIEW_CHANGE, ElectedProof.first().View, ElectedProof.first().Signer) _(highest PreparedProof.View)_
* Block =  ViewChangeMessage.Block
* If Block is None _(No PreparedProof)_
    * Block = Request new block proposal by calling `Config.BlockUtils.RequestNewBlock()`.
* Return Block




<!-- &nbsp;
## `GetNewViewBlock(View)`
> Get Block from the VIEW_CHANGE message with highest PreparedProof.View
> Assume already in OnElectedLocally
* ViewChangeMessages = MessageLog.GetMessages(View)
* FiltteredViewChangeMessages = Filter ViewChangeMessages with PreparedProof
* SortViewChangeMessages =  Sort FiltteredViewChangeMessages by View Desc
* Block =  SortViewChangeMessages.Top().Block _(highest View)_
* If Block is None
    * Block = Request new block proposal by calling `Config.BlockUtils.RequestNewBlock(my_state.Block_height, my_state.Prev_block_hash)`.
* Return Block -->



&nbsp;
## `OnNewViewReceived(Message)`
> Process NEW_VIEW message.
#### Validate and Log message
* If `ValidateNewView(Message)` Continue
* my_state.NewViewLocally = Message.View;
* Call `InitView(Message.View)` _(Resets Election trigger - trigger once is checked inside InitView. i.e., if timedout, the ElecetionTrigger was already Reset)_
#### Continue Process NewViewPrePrepare
* Call `ProcessPrePrepare(Message.PrePrepare)`



&nbsp;
## `ValidateNewView(Message)`
> Validates NewView message and its embeded messages - PrePrepare and ViewChange. \
> Check ElectedProof is Valid - QuorumSize VIEW_CHANGE messages. \
> Check Block proposed matches ElectedProof higest View.
* If Disposed Return False
* If my_state.View is more advanced (>) Message.View Return False. _(Do not process "Old" NEW_VIEW messages)
* If not leader of Message.View Return False _(`IsLeader(Message.NEW_VIEW_HEADER.View, Message.Signer)`)_.
* If NEW_VIEW message already in MessagesLog matching Message(View, Signer, Message_type) Return False.
* If signature mismatch Return False.
* If Message.PrePrepare.View does not match Message.View _(signed by leader in both NewView and PrePrepare but might differ)_
* If not `ValidateElectedProof(Message.ElectedProof)` Return False.
* Block_hash = Get Block_hash by calling `Config.BlockUtils.CalcBlockHash(Message.PrePrepare.Block)`
* If not `ValidateNewViewBlock(Message.ElectedProof, Message.PrePrepare.Block, Block_hash)` Return False.
* If not `ValidatePrePrepare(Message.PrePrepare, Block_hash)` Return False.
* Passed all validation Return True.


&nbsp;
## `ValidateElectedProof(ElectedProof)`
> Validate all votes - VIEW_CHANGE without Block
* If not all Params (Block_height, View) in ViewChangeMessages match Return False.
### validate signatures
* For all ViewChangeMessages:
    *  If Signer is not valid node member _(`IsMember(Message.Signer)`) Return False
    *  If Signature mismatch Return False.
    *  If not `ValidateViewChange(ViewChangeMessage, Mode = PROOF)` Return False.
* All signers are unique - a total of `QuorumSize()` signers
* Passed validation Return True.



&nbsp;
## `ValidateNewViewBlock(ElectedProof, Block, Block_hash)`
> Validate Block was constructed according to rules - based on ElectedProof.\
> Note: If no PreparedProof is found Leader could propose its "own" Block, pass this validation.
#### Get VIEW_CHANGE message in ElectedProof with highest PreparedProof.View or None if no PreparedProof
* ViewChangeMessage = `GetHighestViewChange(ElectedProof)`
* If ViewChangeMessage is not None _(Found PreparedProof in votes: Leader should propose matching Block)_
    * If ViewChangeMessage.PreparedProof.PrePrepare.Block_hash does not match Block_hash
        * Return False. _(Leader proposed a Block which does not match ElectedProof)_
* Else _(The leader proposed aits "own" Block - no preparedProof in ElectedProof - i.e., no locked node. Need to validate content.)_:
   * If Block is not Valid by calling `Config.BlockUtils.ValidateBlock(Block)` Return False.
* Passed validation Return True.


&nbsp;
## `GetHighestViewChange(ElectedProof)`
> Get the VIEW_CHANGE message with highest PreparedProof.View
> Sort ElectedProof ordered by PreparedProof.View Desc.
> Assume validated ElectedProof
> If no PreparedProof return None
* Filter out ElectedProof.ViewChangeMessages without PreparedProof
* Sort ElectedProof by PreparedProof.View Desc
* ViewChangeMessage = ElectedProof.first()   _(highest View, without Block)_
* Return ViewChangeMessage _(None if no PreparedProof)_







## Example Flows
> The last part further elaborates on possible logical flows to accomodate for test driven dev.




