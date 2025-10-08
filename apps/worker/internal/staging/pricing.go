package staging

import "github.com/virtual-staging-ai/worker/internal/staging/model"

// ModelPricing defines the cost per prediction for each model.
// Prices are based on Replicate's pricing as of 2025-01.
// See: https://replicate.com/pricing
type ModelPricing struct {
	ModelID      model.ModelID
	CostPerImage float64 // USD per image
}

var modelPricingTable = []ModelPricing{
	{
		ModelID:      model.ModelQwenImageEdit,
		CostPerImage: 0.03, // $0.03 per image (estimated)
	},
	{
		ModelID:      model.ModelFluxKontextMax,
		CostPerImage: 0.08, // $0.08 per image (estimated)
	},
}

// GetModelCost returns the cost per image for a given model.
// If the model is not found, returns 0.
func GetModelCost(modelID model.ModelID) float64 {
	for _, pricing := range modelPricingTable {
		if pricing.ModelID == modelID {
			return pricing.CostPerImage
		}
	}
	return 0.0
}
