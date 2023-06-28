package stripe

import "github.com/stablecog/sc-go/utils"

// Stripe constants
// Return new instances every time to avoid any potential thread safety issues

func GetPriceIDs() map[int]string {
	return map[int]string{
		// ultimate
		3: utils.GetEnv("STRIPE_ULTIMATE_PRICE_ID", "price_1Mf591ATa0ehBYTA6ggpEEkA"),
		// pro
		2: utils.GetEnv("STRIPE_PRO_PRICE_ID", "price_1Mf50bATa0ehBYTAPOcfnOjG"),
		// starter
		1: utils.GetEnv("STRIPE_STARTER_PRICE_ID", "price_1Mf56NATa0ehBYTAHkCUablG"),
	}
}

func GetProductIDs() map[int]string {
	return map[int]string{
		// ultimate
		3: utils.GetEnv("STRIPE_ULTIMATE_PRODUCT_ID", "prod_NTzE0C8bEuIv6F"),
		// pro
		2: utils.GetEnv("STRIPE_PRO_PRODUCT_ID", "prod_NTzCojAHPw6tbX"),
		// starter
		1: utils.GetEnv("STRIPE_STARTER_PRODUCT_ID", "prod_NPuwbni7ZNkHDO"),
	}

}

func GetSinglePurchasePriceIDs() map[string]string {
	return map[string]string{
		utils.GetEnv("STRIPE_LARGE_PACK_PRICE_ID", "1"):  utils.GetEnv("STRIPE_LARGE_PACK_PRODUCT_ID", "1"),
		utils.GetEnv("STRIPE_MEDIUM_PACK_PRICE_ID", "2"): utils.GetEnv("STRIPE_MEDIUM_PACK_PRODUCT_ID", "2"),
		utils.GetEnv("STRIPE_MEGA_PACK_PRICE_ID", "3"):   utils.GetEnv("STRIPE_MEGA_PACK_PRODUCT_ID", "3"),
	}
}
