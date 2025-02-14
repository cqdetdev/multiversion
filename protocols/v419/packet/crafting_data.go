package packet

import (
	"fmt"

	"github.com/flonja/multiversion/protocols/v419/types"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// CraftingData is sent by the server to let the client know all crafting data that the server maintains. This
// includes shapeless crafting, crafting table recipes, furnace recipes etc. Each crafting station's recipes
// are included in it.
type CraftingData struct {
	// Recipes is a list of all recipes available on the server. It includes among others shapeless, shaped
	// and furnace recipes. The client will only be able to craft these recipes.
	Recipes []types.Recipe
	// PotionRecipes is a list of all potion mixing recipes which may be used in the brewing stand.
	PotionRecipes []protocol.PotionRecipe
	// PotionContainerChangeRecipes is a list of all recipes to convert a potion from one type to another,
	// such as from a drinkable potion to a splash potion, or from a splash potion to a lingering potion.
	PotionContainerChangeRecipes []protocol.PotionContainerChangeRecipe
	// ClearRecipes indicates if all recipes currently active on the client should be cleaned. Doing this
	// means that the client will have no recipes active by itself: Any CraftingData packets previously sent
	// will also be discarded, and only the recipes in this CraftingData packet will be used.
	ClearRecipes bool
}

// ID ...
func (*CraftingData) ID() uint32 {
	return packet.IDCraftingData
}

// Marshal ...
func (pk *CraftingData) Marshal(w protocol.IO) {
	l := uint32(len(pk.Recipes))
	w.Varuint32(&l)
	for _, recipe := range pk.Recipes {
		var c int32
		switch recipe.(type) {
		case *types.ShapelessRecipe:
			c = protocol.RecipeShapeless
		case *types.ShapedRecipe:
			c = protocol.RecipeShaped
		case *types.FurnaceRecipe:
			c = protocol.RecipeFurnace
		case *types.FurnaceDataRecipe:
			c = protocol.RecipeFurnaceData
		case *types.MultiRecipe:
			c = protocol.RecipeMulti
		case *types.ShulkerBoxRecipe:
			c = protocol.RecipeShulkerBox
		case *types.ShapelessChemistryRecipe:
			c = protocol.RecipeShapelessChemistry
		case *types.ShapedChemistryRecipe:
			c = protocol.RecipeShapedChemistry
		default:
			w.UnknownEnumOption(fmt.Sprintf("%T", recipe), "crafting recipe type")
		}
		w.Varint32(&c)
		recipe.Marshal(w)
	}
	protocol.Slice(w, &pk.PotionRecipes)
	protocol.Slice(w, &pk.PotionContainerChangeRecipes)
	w.Bool(&pk.ClearRecipes)
}
