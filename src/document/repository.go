package document

import "diploma/src/messages"

type UpdateRepository interface {
	GetUpdates(*messages.Step1SyncMessage) ([][]byte, error)
	StoreUpdate(*messages.UpdateMessage) error
}

type RemovalRepository interface {
	GetRemovals() ([][]byte, error)
	StoreRemoval([]byte) error
}
