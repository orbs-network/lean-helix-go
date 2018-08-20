package leanhelix

type MessageType string

const (
	MESSAGE_TYPE_PREPREPARE  MessageType = "preprepare"
	MESSAGE_TYPE_PREPARE     MessageType = "prepare"
	MESSAGE_TYPE_COMMIT      MessageType = "commit"
	MESSAGE_TYPE_VIEW_CHANGE MessageType = "view-change"
	MESSAGE_TYPE_NEW_VIEW    MessageType = "new-view"
)
