package stripe

import "github.com/stablecog/sc-go/utils"

// Stripe constants
// Return new instances every time to avoid any potential thread safety issues

func GetPriceIDs() map[int]string {
	return map[int]string{
		// ultimate
		3: utils.GetEnv().StripeUltimatePriceID,
		// pro
		2: utils.GetEnv().StripeProPriceID,
		// starter
		1: utils.GetEnv().StripeStarterPriceID,
	}
}

func GetProductIDs() map[int]string {
	return map[int]string{
		// ultimate
		3: utils.GetEnv().StripeUltimateProductID,
		// pro
		2: utils.GetEnv().StripeProProductID,
		// starter
		1: utils.GetEnv().StripeStarterProductID,
	}

}

func GetSinglePurchasePriceIDs() map[string]string {
	return map[string]string{
		utils.GetEnv().StripeLargePackPriceID:  utils.GetEnv().StripeLargePackProductID,
		utils.GetEnv().StripeMediumPackPriceID: utils.GetEnv().StripeMediumPackProductID,
		utils.GetEnv().StripeMegaPackPriceID:   utils.GetEnv().StripeMegaPackProductID,
	}
}
