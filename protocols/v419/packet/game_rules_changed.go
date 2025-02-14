package packet

import (
	"github.com/flonja/multiversion/protocols/v419/types"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// GameRulesChanged is sent by the server to the client to update client-side game rules, such as game rules
// like the 'showCoordinates' game rule.
type GameRulesChanged struct {
	// GameRules defines game rules currently active with their respective values. The value of these game
	// rules may be either 'bool', 'int32' or 'float32'. Some game rules are server side only, and don't
	// necessarily need to be sent to the client.
	GameRules map[string]any
}

// ID ...
func (*GameRulesChanged) ID() uint32 {
	return packet.IDGameRulesChanged
}

// Marshal ...
func (pk *GameRulesChanged) Marshal(w protocol.IO) {
	types.WriteGameRules(w, &pk.GameRules)
}
