package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

const (
	InventoryTransactionTypeNormal = iota
	InventoryTransactionTypeMismatch
	InventoryTransactionTypeUseItem
	InventoryTransactionTypeUseItemOnEntity
	InventoryTransactionTypeReleaseItem
)

// InventoryTransaction is a packet sent by the client. It essentially exists out of multiple sub-packets,
// each of which have something to do with the inventory in one way or another. Some of these sub-packets
// directly relate to the inventory, others relate to interaction with the world, that could potentially
// result in a change in the inventory.
type InventoryTransaction struct {
	// LegacyRequestID is an ID that is only non-zero at times when sent by the client. The server should
	// always send 0 for this. When this field is not 0, the LegacySetItemSlots slice below will have values
	// in it.
	// LegacyRequestID ties in with the ItemStackResponse packet. If this field is non-0, the server should
	// respond with an ItemStackResponse packet. Some inventory actions such as dropping an item out of the
	// hotbar are still one using this packet, and the ItemStackResponse packet needs to tie in with it.
	LegacyRequestID int32
	// LegacySetItemSlots are only present if the LegacyRequestID is non-zero. These item slots inform the
	// server of the slots that were changed during the inventory transaction, and the server should send
	// back an ItemStackResponse packet with these slots present in it. (Or false with no slots, if rejected.)
	LegacySetItemSlots []protocol.LegacySetItemSlot
	// HasNetworkIDs specifies if the inventory actions below have network IDs associated with them. It is
	// always set to false when a client sends this packet to the server.
	HasNetworkIDs bool
	// Actions is a list of actions that took place, that form the inventory transaction together. Each of
	// these actions hold one slot in which one item was changed to another. In general, the combination of
	// all of these actions results in a balanced inventory transaction. This should be checked to ensure that
	// no items are cheated into the inventory.
	// Actions []types.InventoryAction
	// // TransactionData is a data object that holds data specific to the type of transaction that the
	// // TransactionPacket held. Its concrete type must be one of NormalTransactionData, MismatchTransactionData
	// // UseItemTransactionData, UseItemOnEntityTransactionData or ReleaseItemTransactionData. If nil is set,
	// // the transaction will be assumed to of type InventoryTransactionTypeNormal.
	// TransactionData types.InventoryTransactionData
}

// ID ...
func (*InventoryTransaction) ID() uint32 {
	return packet.IDInventoryTransaction
}

// Marshal ...
func (pk *InventoryTransaction) Marshal(w protocol.IO) {
	// w.Varint32(&pk.LegacyRequestID)
	// if pk.LegacyRequestID != 0 {
	// 	protocol.FuncSlice(w, &pk.LegacySetItemSlots, func(slot *protocol.LegacySetItemSlot) {
	// 		slot.Marshal(w)
	// 	})
	// }
	// var id uint32
	// switch pk.TransactionData.(type) {
	// case nil, *types.NormalTransactionData:
	// 	id = InventoryTransactionTypeNormal
	// case *types.MismatchTransactionData:
	// 	id = InventoryTransactionTypeMismatch
	// case *types.UseItemTransactionData:
	// 	id = InventoryTransactionTypeUseItem
	// case *types.UseItemOnEntityTransactionData:
	// 	id = InventoryTransactionTypeUseItemOnEntity
	// case *types.ReleaseItemTransactionData:
	// 	id = InventoryTransactionTypeReleaseItem
	// }
	// w.Varuint32(&id)
	// w.Bool(&pk.HasNetworkIDs)
	// protocol.FuncSlice(w, &pk.Actions, func(action *types.InventoryAction) {
	// 	action.Marshal(w, pk.HasNetworkIDs)
	// })
	// if pk.TransactionData != nil {
	// 	pk.TransactionData.Marshal(w)
	// }
}
