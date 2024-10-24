package stripe

import "github.com/stablecog/sc-go/utils"

// Stripe constants
// Return new instances every time to avoid any potential thread safety issues

func GetPriceIDs() map[int]string {
	return map[int]string{
		// ultimate annual
		6: utils.GetEnv().StripeUltimateAnnualPriceID,
		// pro annual
		5: utils.GetEnv().StripeProAnnualPriceID,
		// starter annual
		4: utils.GetEnv().StripeStarterAnnualPriceID,
		// ultimate
		3: utils.GetEnv().StripeUltimatePriceID,
		// pro
		2: utils.GetEnv().StripeProPriceID,
		// starter
		1: utils.GetEnv().StripeStarterPriceID,
	}
}

func GetPriceIDLevel(priceID string) int {
	switch priceID {
	case utils.GetEnv().StripeUltimateAnnualPriceID:
		return 6
	case utils.GetEnv().StripeProAnnualPriceID:
		return 5
	case utils.GetEnv().StripeStarterAnnualPriceID:
		return 4
	case utils.GetEnv().StripeUltimatePriceID:
		return 3
	case utils.GetEnv().StripeProPriceID:
		return 2
	case utils.GetEnv().StripeStarterPriceID:
		return 1
	}
	return 0
}

func IsAnnualPriceID(priceID string) bool {
	return priceID == utils.GetEnv().StripeUltimateAnnualPriceID || priceID == utils.GetEnv().StripeProAnnualPriceID || priceID == utils.GetEnv().StripeStarterAnnualPriceID
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
