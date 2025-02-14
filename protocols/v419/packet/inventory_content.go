package packet

import (
	"github.com/flonja/multiversion/protocols/v419/types"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// InventoryContent is sent by the server to update the full content of a particular inventory. It is usually
// sent for the main inventory of the player, but also works for other inventories that are currently opened
// by the player.
type InventoryContent struct {
	// WindowID is the ID that identifies one of the windows that the client currently has opened, or one of
	// the consistent windows such as the main inventory.
	WindowID uint32
	// Content is the new content of the inventory. The length of this slice must be equal to the full size of
	// the inventory window updated.
	Content []types.ItemInstance
}

// ID ...
func (*InventoryContent) ID() uint32 {
	return packet.IDInventoryContent
}

// Marshal ...
func (pk *InventoryContent) Marshal(w protocol.IO) {
	w.Varuint32(&pk.WindowID)
	protocol.Slice(w, &pk.Content)

}
