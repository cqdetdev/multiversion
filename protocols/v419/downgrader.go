package v419

import (
	"math"

	"github.com/flonja/multiversion/mapping"
	"github.com/flonja/multiversion/protocols/v486/types"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// / downgradeBlockActorData downgrades a block actor from latest version to legacy version.
func downgradeBlockActorData(data map[string]any) map[string]any {
	switch data["id"] {
	case "Sign":
		delete(data, "BackText")
		frontRaw, ok := data["FrontText"]
		if !ok {
			frontRaw = map[string]any{"Text": ""}
		}
		front, ok := frontRaw.(map[string]any)
		if !ok {
			front = map[string]any{"Text": ""}
		}
		textRaw, ok := front["Text"]
		if !ok {
			textRaw = ""
		}
		text, ok := textRaw.(string)
		if !ok {
			text = ""
		}
		data["Text"] = text
	}
	return data
}

// downgradeEntityMetadata downgrades entity metadata from latest version to legacy version.
func downgradeEntityMetadata(data map[uint32]any) map[uint32]any {
	newData := make(map[uint32]any)
	for key, value := range data {
		switch key {
		case protocol.EntityDataKeyDataRadius:
			key = 60
		case protocol.EntityDataKeyDataWaiting:
			key = 61
		case protocol.EntityDataKeyDataParticle:
			key = 62
		case protocol.EntityDataKeyAttachFace:
			key = 64
		case protocol.EntityDataKeyAttachedPosition:
			key = 66
		case protocol.EntityDataKeyTradeTarget:
			key = 67
		case protocol.EntityDataKeyCommandName:
			key = 70
		case protocol.EntityDataKeyLastCommandOutput:
			key = 71
		case protocol.EntityDataKeyTrackCommandOutput:
			key = 72
		case protocol.EntityDataKeyControllingSeatIndex:
			key = 73
		case protocol.EntityDataKeyStrength:
			key = 74
		case protocol.EntityDataKeyStrengthMax:
			key = 75
		case protocol.EntityDataKeyDataLifetimeTicks:
			key = 77
		case protocol.EntityDataKeyPoseIndex:
			key = 78
		case protocol.EntityDataKeyDataTickOffset:
			key = 79
		case protocol.EntityDataKeyAlwaysShowNameTag:
			key = 80
		case protocol.EntityDataKeyColorTwoIndex:
			key = 81
		case protocol.EntityDataKeyScore:
			key = 83
		case protocol.EntityDataKeyBalloonAnchor:
			key = 84
		case protocol.EntityDataKeyPuffedState:
			key = 85
		case protocol.EntityDataKeyBubbleTime:
			key = 86
		case protocol.EntityDataKeyAgent:
			key = 87
		case protocol.EntityDataKeyEatingCounter:
			key = 90
		case protocol.EntityDataKeyFlagsTwo:
			key = 91
		case protocol.EntityDataKeyDataDuration:
			key = 94
		case protocol.EntityDataKeyDataSpawnTime:
			key = 95
		case protocol.EntityDataKeyDataChangeRate:
			key = 96
		case protocol.EntityDataKeyDataChangeOnPickup:
			key = 97
		case protocol.EntityDataKeyDataPickupCount:
			key = 98
		case protocol.EntityDataKeyInteractText:
			key = 99
		case protocol.EntityDataKeyTradeTier:
			key = 100
		case protocol.EntityDataKeyMaxTradeTier:
			key = 101
		case protocol.EntityDataKeyTradeExperience:
			key = 102
		case protocol.EntityDataKeySkinID:
			key = 104
		case protocol.EntityDataKeyCommandBlockTickDelay:
			key = 105
		case protocol.EntityDataKeyCommandBlockExecuteOnFirstTick:
			key = 106
		case protocol.EntityDataKeyAmbientSoundInterval:
			key = 107
		case protocol.EntityDataKeyAmbientSoundIntervalRange:
			key = 108
		case protocol.EntityDataKeyAmbientSoundEventName:
			key = 109
		}
		newData[key] = value
	}

	var flag1, flag2 int64
	if v, ok := newData[protocol.EntityDataKeyFlags]; ok {
		flag1 = v.(int64)
	}
	if v, ok := newData[protocol.EntityDataKeyFlagsTwo]; ok {
		flag2 = v.(int64)
	}
	if flag1 == 0 && flag2 == 0 {
		return newData
	}

	newFlag1 := flag1 & ^(^0 << (protocol.EntityDataFlagDash - 1))
	lastHalf := flag1 & (^0 << protocol.EntityDataFlagDash)
	lastHalf >>= 1
	lastHalf &= math.MaxInt64

	newFlag1 |= lastHalf

	if flag2 != 0 {
		newFlag1 ^= (flag2 & 1) << 63
		flag2 >>= 1
		flag2 &= math.MaxInt64

		newData[protocol.EntityDataKeyFlagsTwo] = flag2
	}

	newData[protocol.EntityDataKeyFlags] = newFlag1
	return newData
}

func downgradeCraftingDescription(descriptor protocol.ItemDescriptor, m mapping.Item) protocol.ItemDescriptor {
	var networkId int32
	var metadata int32
	switch descriptor := descriptor.(type) {
	case *protocol.DefaultItemDescriptor:
		networkId = int32(descriptor.NetworkID)
		metadata = int32(descriptor.MetadataValue)
	case *protocol.DeferredItemDescriptor:
		if rid, ok := m.ItemNameToRuntimeID(descriptor.Name); ok {
			networkId = rid
			metadata = int32(descriptor.MetadataValue)
		}
	case *protocol.ItemTagItemDescriptor:
		/// ?????
	case *protocol.ComplexAliasItemDescriptor:
		/// ?????
	}
	return &types.DefaultItemDescriptor{
		NetworkID:     networkId,
		MetadataValue: metadata,
	}
}

// TODO: add downgrade entity flags
