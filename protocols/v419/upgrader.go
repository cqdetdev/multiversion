package v419

import (
	"github.com/flonja/multiversion/protocols/v486/types"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// upgradeBlockActorData upgrades a block actor from a legacy version to the latest version.
func upgradeBlockActorData(data map[string]any) map[string]any {
	switch data["id"] {
	case "Sign":
		textRaw, ok := data["Text"]
		if !ok {
			textRaw = ""
		}
		text, ok := textRaw.(string)
		if !ok {
			text = ""
		}
		data["FrontText"] = map[string]any{"Text": text}
		data["BackText"] = map[string]any{"Text": ""}
	}
	return data
}

// upgradeEntityMetadata upgrades entity metadata from legacy version to latest version.
func upgradeEntityMetadata(data map[uint32]any) map[uint32]any {
	newData := make(map[uint32]any)
	for key, value := range data {
		switch key {
		case 60:
			key = protocol.EntityDataKeyDataRadius
		case 61:
			key = protocol.EntityDataKeyDataWaiting
		case 62:
			key = protocol.EntityDataKeyDataParticle
		case 64:
			key = protocol.EntityDataKeyAttachFace
		case 66:
			key = protocol.EntityDataKeyAttachedPosition
		case 67:
			key = protocol.EntityDataKeyTradeTarget
		case 70:
			key = protocol.EntityDataKeyCommandName
		case 71:
			key = protocol.EntityDataKeyLastCommandOutput
		case 72:
			key = protocol.EntityDataKeyTrackCommandOutput
		case 73:
			key = protocol.EntityDataKeyControllingSeatIndex
		case 74:
			key = protocol.EntityDataKeyStrength
		case 75:
			key = protocol.EntityDataKeyStrengthMax
		case 77:
			key = protocol.EntityDataKeyDataLifetimeTicks
		case 78:
			key = protocol.EntityDataKeyPoseIndex
		case 79:
			key = protocol.EntityDataKeyDataTickOffset
		case 80:
			key = protocol.EntityDataKeyAlwaysShowNameTag
		case 81:
			key = protocol.EntityDataKeyColorTwoIndex
		case 83:
			key = protocol.EntityDataKeyScore
		case 84:
			key = protocol.EntityDataKeyBalloonAnchor
		case 85:
			key = protocol.EntityDataKeyPuffedState
		case 86:
			key = protocol.EntityDataKeyBubbleTime
		case 87:
			key = protocol.EntityDataKeyAgent
		case 90:
			key = protocol.EntityDataKeyEatingCounter
		case 91:
			key = protocol.EntityDataKeyFlagsTwo
		case 94:
			key = protocol.EntityDataKeyDataDuration
		case 95:
			key = protocol.EntityDataKeyDataSpawnTime
		case 96:
			key = protocol.EntityDataKeyDataChangeRate
		case 97:
			key = protocol.EntityDataKeyDataChangeOnPickup
		case 98:
			key = protocol.EntityDataKeyDataPickupCount
		case 99:
			key = protocol.EntityDataKeyInteractText
		case 100:
			key = protocol.EntityDataKeyTradeTier
		case 101:
			key = protocol.EntityDataKeyMaxTradeTier
		case 102:
			key = protocol.EntityDataKeyTradeExperience
		case 104:
			key = protocol.EntityDataKeySkinID
		case 105:
			key = protocol.EntityDataKeyCommandBlockTickDelay
		case 106:
			key = protocol.EntityDataKeyCommandBlockExecuteOnFirstTick
		case 107:
			key = protocol.EntityDataKeyAmbientSoundInterval
		case 108:
			key = protocol.EntityDataKeyAmbientSoundIntervalRange
		case 109:
			key = protocol.EntityDataKeyAmbientSoundEventName
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

	flag2 <<= 1
	flag2 |= (flag1 >> 63) & 1

	newFlag1 := flag1 & ^(^0 << (protocol.EntityDataFlagDash - 1))
	lastHalf := flag1 & (^0<<protocol.EntityDataFlagDash - 1)
	lastHalf <<= 1
	newFlag1 |= lastHalf

	newData[protocol.EntityDataKeyFlagsTwo] = flag2
	newData[protocol.EntityDataKeyFlags] = newFlag1

	return newData
}

func upgradeCraftingDescription(descriptor *types.DefaultItemDescriptor) protocol.ItemDescriptor {
	return &protocol.DefaultItemDescriptor{
		NetworkID:     int16(descriptor.NetworkID),
		MetadataValue: int16(descriptor.MetadataValue),
	}
}

// TODO: add upgrade entity flags
