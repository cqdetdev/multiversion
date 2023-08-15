package v419

import (
	"fmt"
	"image/color"
	"io"

	"github.com/flonja/multiversion/mapping"
	"github.com/flonja/multiversion/protocols/latest"
	legacypacket "github.com/flonja/multiversion/protocols/v419/packet"
	"github.com/flonja/multiversion/protocols/v419/types"
	"github.com/flonja/multiversion/translator"
	"github.com/samber/lo"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/nbt"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"

	_ "embed"
)

var (
	//go:embed required_item_list.json
	itemRuntimeIDData []byte
	//go:embed block_states.nbt
	blockStateData []byte
)

type Protocol struct {
	itemMapping     mapping.Item
	blockMapping    mapping.Block
	itemTranslator  translator.ItemTranslator
	blockTranslator translator.BlockTranslator
}

func New() *Protocol {
	// TODO: add custom block/item replacements (aka make it cool)

	itemMapping := mapping.NewLegacyItemMapping(itemRuntimeIDData, 111)
	blockMapping := mapping.NewBlockMapping(blockStateData).WithBlockActorRemapper(downgradeBlockActorData, upgradeBlockActorData)
	latestBlockMapping := latest.NewBlockMapping()
	return &Protocol{itemMapping: itemMapping, blockMapping: blockMapping,
		itemTranslator:  translator.NewItemTranslator(itemMapping, latest.NewItemMapping(), blockMapping, latestBlockMapping),
		blockTranslator: translator.NewBlockTranslator(blockMapping, latestBlockMapping)}
}

func (p Protocol) ID() int32 {
	return 419
}

func (p Protocol) Ver() string {
	return "1.16.100"
}

// Packets ...
func (Protocol) Packets() packet.Pool {
	pool := packet.NewClientPool()
	for _, pk := range packet.NewServerPool() {
		pool.
	}
	pool[packet.IDActorPickRequest] = func() packet.Packet { return &legacypacket.ActorPickRequest{} }
	pool[packet.IDCommandRequest] = func() packet.Packet { return &legacypacket.CommandRequest{} }
	pool[packet.IDCraftingEvent] = func() packet.Packet { return &legacypacket.CraftingEvent{} }
	pool[packet.IDInventoryTransaction] = func() packet.Packet { return &legacypacket.InventoryTransaction{} }
	pool[packet.IDItemStackRequest] = func() packet.Packet { return &legacypacket.ItemStackRequest{} }
	pool[packet.IDMapInfoRequest] = func() packet.Packet { return &legacypacket.MapInfoRequest{} }
	pool[packet.IDMobArmourEquipment] = func() packet.Packet { return &legacypacket.MobArmourEquipment{} }
	pool[packet.IDMobEquipment] = func() packet.Packet { return &legacypacket.MobEquipment{} }
	pool[packet.IDModalFormResponse] = func() packet.Packet { return &legacypacket.ModalFormResponse{} }
	pool[packet.IDNPCRequest] = func() packet.Packet { return &legacypacket.NPCRequest{} }
	pool[packet.IDPlayerAction] = func() packet.Packet { return &legacypacket.PlayerAction{} }
	pool[packet.IDPlayerAuthInput] = func() packet.Packet { return &legacypacket.PlayerAuthInput{} }
	pool[packet.IDStartGame] = func() packet.Packet { return &legacypacket.StartGame{} }
	pool[packet.IDPlayerSkin] = func() packet.Packet { return &legacypacket.PlayerSkin{} }
	pool[packet.IDRequestChunkRadius] = func() packet.Packet { return &legacypacket.RequestChunkRadius{} }
	pool[packet.IDSetActorData] = func() packet.Packet { return &legacypacket.SetActorData{} }
	pool[packet.IDStructureBlockUpdate] = func() packet.Packet { return &legacypacket.StructureBlockUpdate{} }
	pool[packet.IDStructureTemplateDataRequest] = func() packet.Packet { return &legacypacket.StructureTemplateDataRequest{} }
	return pool
}

// Encryption ...
func (Protocol) Encryption(key [32]byte) packet.Encryption {
	return newCFBEncryption(key[:])
}

func (p Protocol) NewReader(r interface {
	io.Reader
	io.ByteReader
}, shieldID int32, enableLimits bool) protocol.IO {
	return NewReader(protocol.NewReader(r, shieldID, enableLimits))
}

func (p Protocol) NewWriter(w interface {
	io.Writer
	io.ByteWriter
}, shieldID int32) protocol.IO {
	return NewWriter(protocol.NewWriter(w, shieldID))
}

// ConvertToLatest ...
func (p Protocol) ConvertToLatest(pk packet.Packet, conn *minecraft.Conn) []packet.Packet {
	fmt.Printf("1.16.100 -> 1.20.x: %T\n", pk)
	var newPks []packet.Packet
	switch pk := pk.(type) {
	case *legacypacket.ActorPickRequest:
		newPks = append(newPks,
			&packet.ActorPickRequest{
				EntityUniqueID: pk.EntityUniqueID,
				HotBarSlot:     pk.HotBarSlot,
			})
	case *legacypacket.CommandRequest:
		newPks = append(newPks, &packet.CommandRequest{
			CommandLine:   pk.CommandLine,
			CommandOrigin: pk.CommandOrigin,
			Internal:      pk.Internal,
		})
	case *legacypacket.MapInfoRequest:
		newPks = append(newPks,
			&packet.MapInfoRequest{
				MapID: pk.MapID,
			})
	case *legacypacket.ModalFormResponse:
		var response protocol.Optional[[]byte]
		var cancelReason protocol.Optional[uint8]
		if string(pk.ResponseData) == "null" {
			// The response data is not null, so it is a valid response.
			response = protocol.Option(pk.ResponseData)
		} else {
			// We can always default to the user closed reason if the response data doesn't exist.
			cancelReason = protocol.Option[uint8](packet.ModalFormCancelReasonUserClosed)
		}
		newPks = append(newPks,
			&packet.ModalFormResponse{
				FormID:       pk.FormID,
				ResponseData: response,
				CancelReason: cancelReason,
			})
	case *legacypacket.NPCRequest:
		newPks = append(newPks,
			&packet.NPCRequest{
				EntityRuntimeID: pk.EntityRuntimeID,
				RequestType:     pk.RequestType,
				CommandString:   pk.CommandString,
				ActionType:      pk.ActionType,
			})
	case *legacypacket.PlayerAction:
		newPks = append(newPks,
			&packet.PlayerAction{
				EntityRuntimeID: pk.EntityRuntimeID,
				ActionType:      pk.ActionType,
				BlockPosition:   pk.BlockPosition,
				BlockFace:       pk.BlockFace,
			})
	case *legacypacket.PlayerAuthInput:
		newPks = append(newPks,
			&packet.PlayerAuthInput{
				Pitch:            pk.Pitch,
				Yaw:              pk.Yaw,
				Position:         pk.Position,
				MoveVector:       pk.MoveVector,
				HeadYaw:          pk.HeadYaw,
				InputData:        pk.InputData,
				InputMode:        pk.InputMode,
				PlayMode:         pk.PlayMode,
				InteractionModel: packet.InteractionModelCrosshair,
				GazeDirection:    pk.GazeDirection,
				Tick:             pk.Tick,
				Delta:            pk.Delta,
			})
	case *legacypacket.PlayerSkin:
		newPks = append(newPks,
			&packet.PlayerSkin{
				UUID:        pk.UUID,
				Skin:        types.LatestSkin(pk.Skin),
				NewSkinName: pk.NewSkinName,
				OldSkinName: pk.OldSkinName,
			})
	case *legacypacket.SetActorData:
		newPks = append(newPks,
			&packet.SetActorData{
				EntityRuntimeID: pk.EntityRuntimeID,
				EntityMetadata:  types.UpgradeEntityMetadata(pk.EntityMetadata),
				Tick:            pk.Tick,
			})
	case *legacypacket.StructureBlockUpdate:
		newPks = append(newPks,
			&packet.StructureBlockUpdate{
				Position:           pk.Position,
				StructureName:      pk.StructureName,
				DataField:          pk.DataField,
				IncludePlayers:     pk.IncludePlayers,
				ShowBoundingBox:    pk.ShowBoundingBox,
				StructureBlockType: pk.StructureBlockType,
				Settings: protocol.StructureSettings{
					PaletteName:               pk.Settings.PaletteName,
					IgnoreEntities:            pk.Settings.IgnoreEntities,
					IgnoreBlocks:              pk.Settings.IgnoreBlocks,
					Size:                      pk.Settings.Size,
					Offset:                    pk.Settings.Offset,
					LastEditingPlayerUniqueID: pk.Settings.LastEditingPlayerUniqueID,
					Rotation:                  pk.Settings.Rotation,
					Mirror:                    pk.Settings.Mirror,
					Integrity:                 pk.Settings.Integrity,
					Seed:                      pk.Settings.Seed,
					Pivot:                     pk.Settings.Pivot,
				},
				RedstoneSaveMode: pk.RedstoneSaveMode,
				ShouldTrigger:    pk.ShouldTrigger,
			})
	case *legacypacket.RequestChunkRadius:
		newPks = append(newPks,
			&packet.RequestChunkRadius{
				ChunkRadius:    pk.ChunkRadius,
				MaxChunkRadius: pk.ChunkRadius,
			})
	case *legacypacket.StructureTemplateDataRequest:
		newPks = append(newPks,
			&packet.StructureTemplateDataRequest{
				StructureName: pk.StructureName,
				Position:      pk.Position,
				Settings: protocol.StructureSettings{
					PaletteName:               pk.Settings.PaletteName,
					IgnoreEntities:            pk.Settings.IgnoreEntities,
					IgnoreBlocks:              pk.Settings.IgnoreBlocks,
					Size:                      pk.Settings.Size,
					Offset:                    pk.Settings.Offset,
					LastEditingPlayerUniqueID: pk.Settings.LastEditingPlayerUniqueID,
					Rotation:                  pk.Settings.Rotation,
					Mirror:                    pk.Settings.Mirror,
					Integrity:                 pk.Settings.Integrity,
					Seed:                      pk.Settings.Seed,
					Pivot:                     pk.Settings.Pivot,
				},
				RequestType: pk.RequestType,
			})
	case *packet.AdventureSettings:
	case *packet.TickSync:
		return nil
	case *packet.PacketViolationWarning:
		fmt.Println(pk)
	default:
		newPks = append(newPks, pk)
	}
	if pk.ID() == 37 {
		return nil
	}
	return p.blockTranslator.UpgradeBlockPackets(p.itemTranslator.UpgradeItemPackets(newPks, conn), conn)
}

// ConvertFromLatest ...
func (p Protocol) ConvertFromLatest(pk packet.Packet, conn *minecraft.Conn) (result []packet.Packet) {
	result = p.blockTranslator.DowngradeBlockPackets(p.itemTranslator.DowngradeItemPackets([]packet.Packet{pk}, conn), conn)
	for i, pk := range result {
		fmt.Printf("1.20.x -> 1.16.100: %T\n", pk)
		switch pk := pk.(type) {
		case *packet.ActorEvent:
			// if pk.EventType > packet.ActorEvent {
			// 	return nil
			// }
		case *packet.AddActor:
			result[i] = &legacypacket.AddActor{
				EntityUniqueID:  pk.EntityUniqueID,
				EntityRuntimeID: pk.EntityRuntimeID,
				EntityType:      pk.EntityType,
				Position:        pk.Position,
				Velocity:        pk.Velocity,
				Pitch:           pk.Pitch,
				Yaw:             pk.Yaw,
				HeadYaw:         pk.HeadYaw,
				Attributes: lo.Map(pk.Attributes, func(a protocol.AttributeValue, _ int) types.Attribute {
					return types.Attribute{
						Name:  a.Name,
						Value: a.Value,
						Max:   a.Max,
						Min:   a.Min,
					}
				}),
				EntityMetadata: types.DowngradeEntityMetadata(pk.EntityMetadata),
				EntityLinks:    pk.EntityLinks,
			}
		case *packet.AddPlayer:
			result[i] = &legacypacket.AddPlayer{
				UUID:                   pk.UUID,
				Username:               pk.Username,
				EntityUniqueID:         pk.AbilityData.EntityUniqueID,
				EntityRuntimeID:        pk.EntityRuntimeID,
				PlatformChatID:         pk.PlatformChatID,
				Position:               pk.Position,
				Velocity:               pk.Velocity,
				Pitch:                  pk.Pitch,
				Yaw:                    pk.Yaw,
				HeadYaw:                pk.HeadYaw,
				HeldItem:               pk.HeldItem.Stack,
				EntityMetadata:         types.DowngradeEntityMetadata(pk.EntityMetadata),
				CommandPermissionLevel: uint32(pk.AbilityData.CommandPermissions),
				PermissionLevel:        uint32(pk.AbilityData.PlayerPermissions),
				PlayerUniqueID:         pk.AbilityData.EntityUniqueID,
				EntityLinks:            pk.EntityLinks,
				DeviceID:               pk.DeviceID,
				BuildPlatform:          pk.BuildPlatform,
			}
		case *packet.AnimateEntity:
			result[i] = &legacypacket.AnimateEntity{
				Animation:        pk.Animation,
				NextState:        pk.NextState,
				StopCondition:    pk.StopCondition,
				Controller:       pk.Controller,
				BlendOutTime:     pk.BlendOutTime,
				EntityRuntimeIDs: pk.EntityRuntimeIDs,
			}

		case *packet.AvailableCommands:
			result[i] = &legacypacket.AvailableCommands{
				Commands: lo.Map(pk.Commands, func(c protocol.Command, _ int) types.Command {
					return types.Command{
						Name:            c.Name,
						Description:     c.Description,
						Flags:           byte(c.Flags),
						PermissionLevel: c.PermissionLevel,
						//Aliases:         c.AliasesOffset,
						Overloads: lo.Map(c.Overloads, func(o protocol.CommandOverload, i int) types.CommandOverload {
							return types.CommandOverload{Parameters: lo.Map(o.Parameters, func(p protocol.CommandParameter, _ int) types.CommandParameter {
								return types.CommandParameter{
									Name:                p.Name,
									Type:                types.DowngradeParamType(p.Type),
									Optional:            p.Optional,
									CollapseEnumOptions: true,
									//Enum:                types.CommandEnum(p.Enum),
									//Suffix:              p.Suffix,
								}
							})}
						}),
					}
				}),
			}
		case *packet.CameraShake:
			result[i] = &legacypacket.CameraShake{
				Intensity: pk.Intensity,
				Duration:  pk.Duration,
				Type:      pk.Type,
			}
		case *packet.ClientBoundMapItemData:
			result[i] = &legacypacket.ClientBoundMapItemData{
				MapID:          pk.MapID,
				UpdateFlags:    pk.UpdateFlags,
				Dimension:      pk.Dimension,
				LockedMap:      pk.LockedMap,
				Scale:          pk.Scale,
				MapsIncludedIn: pk.MapsIncludedIn,
				TrackedObjects: pk.TrackedObjects,
				Decorations:    pk.Decorations,
				Height:         pk.Height,
				Width:          pk.Width,
				XOffset:        pk.XOffset,
				YOffset:        pk.YOffset,
				Pixels:         [][]color.RGBA{pk.Pixels},
			}
		case *packet.EducationSettings:
			result[i] = &legacypacket.EducationSettings{
				CodeBuilderDefaultURI: pk.CodeBuilderDefaultURI,
				CodeBuilderTitle:      pk.CodeBuilderTitle,
				CanResizeCodeBuilder:  pk.CanResizeCodeBuilder,
				OverrideURI:           pk.OverrideURI,
				HasQuiz:               pk.HasQuiz,
			}
		case *packet.Event:
			// TODO: support
			result[i] = &legacypacket.Event{
				EntityRuntimeID: pk.EntityRuntimeID,
				EventType:       0,
				UsePlayerID:     pk.UsePlayerID,
			}
		case *packet.GameRulesChanged:
			result[i] = &legacypacket.GameRulesChanged{
				GameRules: lo.SliceToMap(pk.GameRules, func(rule protocol.GameRule) (string, any) {
					return rule.Name, rule.Value
				}),
			}
		case *packet.HurtArmour:
			result[i] = &legacypacket.HurtArmour{
				Cause:  pk.Cause,
				Damage: pk.Damage,
			}
		case *packet.NetworkChunkPublisherUpdate:
			result[i] = &legacypacket.NetworkChunkPublisherUpdate{
				Position: pk.Position,
				Radius:   pk.Radius,
			}
		case *packet.NetworkSettings:
			result[i] = &legacypacket.NetworkSettings{
				CompressionThreshold: pk.CompressionThreshold,
			}
		case *packet.PhotoTransfer:
			result[i] = &legacypacket.PhotoTransfer{
				PhotoName: pk.PhotoName,
				PhotoData: pk.PhotoData,
				BookID:    pk.BookID,
			}
		case *packet.PlayerList:
			result[i] = &legacypacket.PlayerList{
				ActionType: pk.ActionType,
				Entries: lo.Map(pk.Entries, func(e protocol.PlayerListEntry, _ int) legacypacket.PlayerListEntry {
					return legacypacket.PlayerListEntry{
						UUID:           e.UUID,
						EntityUniqueID: e.EntityUniqueID,
						Username:       e.Username,
						XUID:           e.XUID,
						PlatformChatID: e.PlatformChatID,
						BuildPlatform:  e.BuildPlatform,
						//Skin:           types.LegacySkin(e.Skin),
						Teacher: e.Teacher,
						Host:    e.Host,
					}
				}),
			}
		case *packet.PlayerSkin:
			result[i] = &legacypacket.PlayerSkin{
				UUID: pk.UUID,
				//Skin:        types.LegacySkin(pk.Skin),
				NewSkinName: pk.NewSkinName,
				OldSkinName: pk.OldSkinName,
			}
		case *packet.PositionTrackingDBServerBroadcast:
			data, _ := nbt.MarshalEncoding(&pk.Payload, nbt.LittleEndian)
			result[i] = &legacypacket.PositionTrackingDBServerBroadcast{
				BroadcastAction: pk.BroadcastAction,
				TrackingID:      pk.TrackingID,
				SerialisedData:  data,
			}
		case *packet.ResourcePacksInfo:
			result[i] = &legacypacket.ResourcePacksInfo{
				TexturePackRequired: pk.TexturePackRequired,
				HasScripts:          pk.HasScripts,
				BehaviourPacks: lo.Map(pk.BehaviourPacks, func(pack protocol.BehaviourPackInfo, _ int) types.ResourcePackInfo {
					return types.ResourcePackInfo{
						UUID:            pack.UUID,
						Version:         pack.Version,
						Size:            pack.Size,
						ContentKey:      pack.ContentKey,
						SubPackName:     pack.SubPackName,
						ContentIdentity: pack.ContentIdentity,
						HasScripts:      pack.HasScripts,
					}
				}),
				TexturePacks: lo.Map(pk.TexturePacks, func(pack protocol.TexturePackInfo, _ int) types.ResourcePackInfo {
					return types.ResourcePackInfo{
						UUID:            pack.UUID,
						Version:         pack.Version,
						Size:            pack.Size,
						ContentKey:      pack.ContentKey,
						SubPackName:     pack.SubPackName,
						ContentIdentity: pack.ContentIdentity,
						HasScripts:      pack.HasScripts,
					}
				}),
			}
		case *packet.SetActorData:
			pk.EntityMetadata = downgradeEntityMetadata(pk.EntityMetadata)
			result[i] = pk
		case *packet.SetTitle:
			result[i] = &legacypacket.SetTitle{
				ActionType:      pk.ActionType,
				Text:            pk.Text,
				FadeInDuration:  pk.FadeInDuration,
				RemainDuration:  pk.RemainDuration,
				FadeOutDuration: pk.FadeOutDuration,
			}
		case *packet.SpawnParticleEffect:
			result[i] = &legacypacket.SpawnParticleEffect{
				Dimension:      pk.Dimension,
				EntityUniqueID: pk.EntityUniqueID,
				Position:       pk.Position,
				ParticleName:   pk.ParticleName,
			}
		case *packet.StartGame:
			// TODO: Adjust our mappings to account for any possible custom blocks.
			force, ok := pk.ForceExperimentalGameplay.Value()
			if !ok {
				force = ok
			}
			result[i] = &legacypacket.StartGame{
				EntityUniqueID:                 pk.EntityUniqueID,
				EntityRuntimeID:                pk.EntityRuntimeID,
				PlayerGameMode:                 pk.PlayerGameMode,
				PlayerPosition:                 pk.PlayerPosition,
				Pitch:                          pk.Pitch,
				Yaw:                            pk.Yaw,
				WorldSeed:                      int32(pk.WorldSeed),
				SpawnBiomeType:                 pk.SpawnBiomeType,
				UserDefinedBiomeName:           pk.UserDefinedBiomeName,
				Dimension:                      pk.Dimension,
				Generator:                      pk.Generator,
				WorldGameMode:                  pk.WorldGameMode,
				Difficulty:                     pk.Difficulty,
				WorldSpawn:                     pk.WorldSpawn,
				AchievementsDisabled:           pk.AchievementsDisabled,
				DayCycleLockTime:               pk.DayCycleLockTime,
				EducationEditionOffer:          pk.EducationEditionOffer,
				EducationFeaturesEnabled:       pk.EducationFeaturesEnabled,
				EducationProductID:             pk.EducationProductID,
				RainLevel:                      pk.RainLevel,
				LightningLevel:                 pk.LightningLevel,
				ConfirmedPlatformLockedContent: pk.ConfirmedPlatformLockedContent,
				MultiPlayerGame:                pk.MultiPlayerGame,
				LANBroadcastEnabled:            pk.LANBroadcastEnabled,
				XBLBroadcastMode:               pk.XBLBroadcastMode,
				PlatformBroadcastMode:          pk.PlatformBroadcastMode,
				CommandsEnabled:                pk.CommandsEnabled,
				TexturePackRequired:            pk.TexturePackRequired,
				GameRules: lo.SliceToMap(pk.GameRules, func(rule protocol.GameRule) (string, any) {
					return rule.Name, rule.Value
				}),
				Experiments:                     pk.Experiments,
				ExperimentsPreviouslyToggled:    pk.ExperimentsPreviouslyToggled,
				BonusChestEnabled:               pk.BonusChestEnabled,
				StartWithMapEnabled:             pk.StartWithMapEnabled,
				PlayerPermissions:               pk.PlayerPermissions,
				ServerChunkTickRadius:           pk.ServerChunkTickRadius,
				HasLockedBehaviourPack:          pk.HasLockedBehaviourPack,
				HasLockedTexturePack:            pk.HasLockedTexturePack,
				FromLockedWorldTemplate:         pk.FromLockedWorldTemplate,
				MSAGamerTagsOnly:                pk.MSAGamerTagsOnly,
				FromWorldTemplate:               pk.FromWorldTemplate,
				WorldTemplateSettingsLocked:     pk.WorldTemplateSettingsLocked,
				OnlySpawnV1Villagers:            pk.OnlySpawnV1Villagers,
				BaseGameVersion:                 pk.BaseGameVersion,
				LimitedWorldWidth:               pk.LimitedWorldWidth,
				LimitedWorldDepth:               pk.LimitedWorldDepth,
				NewNether:                       pk.NewNether,
				ForceExperimentalGameplay:       force,
				LevelID:                         pk.LevelID,
				WorldName:                       pk.WorldName,
				TemplateContentIdentity:         pk.TemplateContentIdentity,
				Trial:                           pk.Trial,
				ServerAuthoritativeMovementMode: uint32(pk.PlayerMovementSettings.MovementType),
				Time:                            pk.Time,
				EnchantmentSeed:                 pk.EnchantmentSeed,
				MultiPlayerCorrelationID:        pk.MultiPlayerCorrelationID,
				Blocks:                          pk.Blocks,
				Items:                           pk.Items,
				ServerAuthoritativeInventory:    pk.ServerAuthoritativeInventory,
			}
		case *packet.UpdateAbilities:
			if len(pk.AbilityData.Layers) == 0 || pk.AbilityData.EntityUniqueID != conn.GameData().EntityUniqueID {
				// We need at least one layer.
				return nil
			}

			base, flags, perms := pk.AbilityData.Layers[0].Values, uint32(0), uint32(0)
			if base&protocol.AbilityMayFly != 0 {
				flags |= packet.AdventureFlagAllowFlight
				if base&protocol.AbilityFlying != 0 {
					flags |= packet.AdventureFlagFlying
				}
			}
			if base&protocol.AbilityNoClip != 0 {
				flags |= packet.AdventureFlagNoClip
			}

			if base&protocol.AbilityBuild != 0 && base&protocol.AbilityMine != 0 {
				flags |= packet.AdventureFlagWorldBuilder
			} else {
				flags |= packet.AdventureFlagWorldImmutable
			}
			if base&protocol.AbilityBuild != 0 {
				perms |= packet.ActionPermissionBuild
			}
			if base&protocol.AbilityMine != 0 {
				perms |= packet.ActionPermissionMine
			}

			if base&protocol.AbilityDoorsAndSwitches != 0 {
				perms |= packet.ActionPermissionDoorsAndSwitches
			}
			if base&protocol.AbilityOpenContainers != 0 {
				perms |= packet.ActionPermissionOpenContainers
			}
			if base&protocol.AbilityAttackPlayers != 0 {
				perms |= packet.ActionPermissionAttackPlayers
			}
			if base&protocol.AbilityAttackMobs != 0 {
				perms |= packet.ActionPermissionAttackMobs
			}
			result[i] = &packet.AdventureSettings{
				Flags:                  flags,
				ActionPermissions:      perms,
				PlayerUniqueID:         pk.AbilityData.EntityUniqueID,
				CommandPermissionLevel: uint32(pk.AbilityData.CommandPermissions),
				PermissionLevel:        uint32(pk.AbilityData.PlayerPermissions),
			}
		case *packet.UpdateAttributes:
			result[i] = &legacypacket.UpdateAttributes{
				EntityRuntimeID: pk.EntityRuntimeID,
				Attributes: lo.Map(pk.Attributes, func(attribute protocol.Attribute, _ int) types.Attribute {
					return types.Attribute{
						Name:    attribute.Name,
						Value:   attribute.Value,
						Min:     attribute.Min,
						Max:     attribute.Max,
						Default: attribute.Default,
					}
				}),
				Tick: pk.Tick,
			}
		case *packet.UpdateBlock:
			pk.NewBlockRuntimeID = p.blockTranslator.DowngradeBlockRuntimeID(pk.NewBlockRuntimeID)
		case *packet.UpdateBlockSynced:
			pk.NewBlockRuntimeID = p.blockTranslator.DowngradeBlockRuntimeID(pk.NewBlockRuntimeID)
		case *packet.UpdateAdventureSettings:
			return nil
		}
	}

	if pk.ID() == 39 {
		return nil
	}

	return result
}
