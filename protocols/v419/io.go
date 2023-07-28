package v419

import (
	"fmt"

	"github.com/flonja/multiversion/protocols/v486/types"
	"github.com/sandertv/gophertunnel/minecraft/nbt"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// NewReader creates a new initialised Reader with an underlying protocol.Reader to write to.
func NewReader(r *protocol.Reader) *Reader {
	return &Reader{r}
}

type Reader struct {
	*protocol.Reader
}

func (r *Reader) Reads() bool {
	return true
}

func (r *Reader) StackRequestAction(x *protocol.StackRequestAction) {
	var id uint8
	r.Uint8(&id)
	if !lookupStackRequestAction(id, x) {
		r.UnknownEnumOption(id, "stack request action type")
		return
	}
	(*x).Marshal(r)
}

// lookupStackRequestAction looks up the StackRequestAction matching an ID.
func lookupStackRequestAction(id uint8, x *protocol.StackRequestAction) bool {
	switch id {
	case protocol.StackRequestActionTake:
		*x = &types.TakeStackRequestAction{TakeStackRequestAction: protocol.TakeStackRequestAction{}}
	case protocol.StackRequestActionPlace:
		*x = &types.PlaceStackRequestAction{PlaceStackRequestAction: protocol.PlaceStackRequestAction{}}
	case protocol.StackRequestActionSwap:
		*x = &types.SwapStackRequestAction{SwapStackRequestAction: protocol.SwapStackRequestAction{}}
	case protocol.StackRequestActionDrop:
		*x = &types.DropStackRequestAction{DropStackRequestAction: protocol.DropStackRequestAction{}}
	case protocol.StackRequestActionDestroy:
		*x = &types.DestroyStackRequestAction{DestroyStackRequestAction: protocol.DestroyStackRequestAction{}}
	case protocol.StackRequestActionConsume:
		*x = &types.ConsumeStackRequestAction{DestroyStackRequestAction: protocol.DestroyStackRequestAction{}}
	case protocol.StackRequestActionCreate:
		*x = &protocol.CreateStackRequestAction{}
	case protocol.StackRequestActionPlaceInContainer:
		*x = &types.PlaceInContainerStackRequestAction{PlaceInContainerStackRequestAction: protocol.PlaceInContainerStackRequestAction{}}
	case protocol.StackRequestActionTakeOutContainer:
		*x = &types.TakeOutContainerStackRequestAction{TakeOutContainerStackRequestAction: protocol.TakeOutContainerStackRequestAction{}}
	case protocol.StackRequestActionLabTableCombine:
		*x = &protocol.LabTableCombineStackRequestAction{}
	case protocol.StackRequestActionBeaconPayment:
		*x = &protocol.BeaconPaymentStackRequestAction{}
	case protocol.StackRequestActionMineBlock:
		*x = &protocol.MineBlockStackRequestAction{}
	case protocol.StackRequestActionCraftRecipe:
		*x = &protocol.CraftRecipeStackRequestAction{}
	case protocol.StackRequestActionCraftRecipeAuto:
		*x = &types.AutoCraftRecipeStackRequestAction{AutoCraftRecipeStackRequestAction: protocol.AutoCraftRecipeStackRequestAction{}}
	case protocol.StackRequestActionCraftCreative:
		*x = &protocol.CraftCreativeStackRequestAction{}
	case protocol.StackRequestActionCraftRecipeOptional:
		*x = &protocol.CraftRecipeOptionalStackRequestAction{}
	case protocol.StackRequestActionCraftGrindstone:
		*x = &protocol.CraftGrindstoneRecipeStackRequestAction{}
	case protocol.StackRequestActionCraftLoom:
		*x = &protocol.CraftLoomRecipeStackRequestAction{}
	case protocol.StackRequestActionCraftNonImplementedDeprecated:
		*x = &protocol.CraftNonImplementedStackRequestAction{}
	case protocol.StackRequestActionCraftResultsDeprecated:
		*x = &protocol.CraftResultsDeprecatedStackRequestAction{}
	default:
		return false
	}
	return true
}

// Item reads an item stack from buffer src and stores it into item stack x.
func (r *Reader) Item(x *protocol.ItemStack) {
	x.NBTData = make(map[string]any)
	r.Varint32(&x.NetworkID)
	if x.NetworkID == 0 {
		// The item was air, so there is no more data we should read for the item instance. After all, air
		// items aren't really anything.
		x.MetadataValue, x.Count, x.CanBePlacedOn, x.CanBreak = 0, 0, nil, nil
		return
	}
	var auxValue int32
	r.Varint32(&auxValue)
	x.MetadataValue = uint32(auxValue >> 8)
	x.Count = uint16(auxValue & 0xff)

	var userDataMarker int16
	r.Int16(&userDataMarker)

	if userDataMarker == -1 {
		var userDataVersion uint8
		r.Uint8(&userDataVersion)

		switch userDataVersion {
		case 1:
			r.NBT(&x.NBTData, nbt.NetworkLittleEndian)
		default:
			r.UnknownEnumOption(userDataVersion, "item user data version")
			return
		}
	} else if userDataMarker > 0 {
		r.NBT(&x.NBTData, nbt.LittleEndian)
	}
	var count int32
	r.Varint32(&count)
	r.LimitInt32(count, 0, 1024)

	x.CanBePlacedOn = make([]string, count)
	for i := int32(0); i < count; i++ {
		r.String(&x.CanBePlacedOn[i])
	}

	r.Varint32(&count)
	r.LimitInt32(count, 0, 1024)

	x.CanBreak = make([]string, count)
	for i := int32(0); i < count; i++ {
		r.String(&x.CanBreak[i])
	}
	if x.NetworkID == int32(r.ShieldID()) {
		var blockingTick int64
		r.Varint64(&blockingTick)
	}
}

func (r *Reader) ItemInstance(x *protocol.ItemInstance) {
	r.Varint32(&x.StackNetworkID)
	r.Item(&x.Stack)
	if (x.Stack.Count == 0 || x.Stack.NetworkID == 0) && x.StackNetworkID != 0 {
		r.InvalidValue(x.StackNetworkID, "stack network ID", "stack is empty but network ID is non-zero")
	}
}

// NewWriter creates a new initialised Writer with an underlying protocol.Writer to write to.
func NewWriter(w *protocol.Writer) *Writer {
	return &Writer{w}
}

type Writer struct {
	*protocol.Writer
}

// ItemDescriptorCount writes an ItemDescriptorCount i to the underlying buffer.
func (w *Writer) ItemDescriptorCount(i *protocol.ItemDescriptorCount) {
	var id byte
	switch descriptor := i.Descriptor.(type) {
	case *protocol.InvalidItemDescriptor:
		id = protocol.ItemDescriptorInvalid
	case *protocol.DefaultItemDescriptor:
		id = protocol.ItemDescriptorDefault
	case *protocol.MoLangItemDescriptor:
		id = protocol.ItemDescriptorMoLang
	case *protocol.ItemTagItemDescriptor:
		id = protocol.ItemDescriptorItemTag
	case *protocol.DeferredItemDescriptor:
		id = protocol.ItemDescriptorDeferred
	case *protocol.ComplexAliasItemDescriptor:
		id = protocol.ItemDescriptorComplexAlias
	case *types.DefaultItemDescriptor:
		descriptor.Marshal(w)
		if descriptor.NetworkID != 0 {
			w.Varint32(&i.Count)
		}
		return
	default:
		w.UnknownEnumOption(fmt.Sprintf("%T", i.Descriptor), "item descriptor type")
		return
	}
	w.Uint8(&id)

	i.Descriptor.Marshal(w)
	w.Varint32(&i.Count)
}

// Recipe writes a Recipe to the writer.
func (w *Writer) Recipe(x *protocol.Recipe) {
	var recipeType int32
	if !lookupRecipeType(*x, &recipeType) {
		w.UnknownEnumOption(fmt.Sprintf("%T", *x), "crafting recipe type")
	}
	w.Varint32(&recipeType)
	(*x).Marshal(w)
}

func (w *Writer) Item(x *protocol.ItemStack) {
	w.Varint32(&x.NetworkID)
	if x.NetworkID == 0 {
		// The item was air, so there's no more data to follow. Return immediately.
		return
	}
	aux := int32(x.MetadataValue<<8) | int32(x.Count)
	w.Varint32(&aux)
	if len(x.NBTData) != 0 {
		userDataMarker := int16(-1)
		userDataVer := uint8(1)

		w.Int16(&userDataMarker)
		w.Uint8(&userDataVer)
		w.NBT(&x.NBTData, nbt.NetworkLittleEndian)
	} else {
		userDataMarker := int16(0)

		w.Int16(&userDataMarker)
	}
	placeOnLen := int32(len(x.CanBePlacedOn))
	canBreak := int32(len(x.CanBreak))

	w.Varint32(&placeOnLen)
	for _, block := range x.CanBePlacedOn {
		w.String(&block)
	}
	w.Varint32(&canBreak)
	for _, block := range x.CanBreak {
		w.String(&block)
	}
	if x.NetworkID == int32(w.ShieldID()) {
		var blockingTick int64
		w.Varint64(&blockingTick)
	}
}

// WriteItemInst writes an item instance x to buffer dst.
func (w *Writer) ItemInstance(x *protocol.ItemInstance) {
	w.Varint32(&x.StackNetworkID)
	w.Item(&x.Stack)
	if (x.Stack.Count == 0 || x.Stack.NetworkID == 0) && x.StackNetworkID != 0 {
		w.InvalidValue(x.StackNetworkID, "stack network ID", "stack is empty but network ID is non-zero")
	}
}

// lookupRecipeType looks up the recipe type for a Recipe. False is returned if
// none was found.
func lookupRecipeType(x protocol.Recipe, recipeType *int32) bool {
	switch x.(type) {
	case *protocol.ShapelessRecipe:
		*recipeType = protocol.RecipeShapeless
	case *protocol.ShapedRecipe:
		*recipeType = protocol.RecipeShaped
	case *protocol.FurnaceRecipe:
		*recipeType = protocol.RecipeFurnace
	case *protocol.FurnaceDataRecipe:
		*recipeType = protocol.RecipeFurnaceData
	case *protocol.MultiRecipe:
		*recipeType = protocol.RecipeMulti
	case *protocol.ShulkerBoxRecipe:
		*recipeType = protocol.RecipeShulkerBox
	case *protocol.ShapelessChemistryRecipe:
		*recipeType = protocol.RecipeShapelessChemistry
	case *protocol.ShapedChemistryRecipe:
		*recipeType = protocol.RecipeShapedChemistry
	case *protocol.SmithingTransformRecipe:
		*recipeType = protocol.RecipeSmithingTransform
	case *protocol.SmithingTrimRecipe:
		*recipeType = protocol.RecipeSmithingTrim
	default:
		return false
	}
	return true
}
