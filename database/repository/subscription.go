package repository

import (
	"encoding/json"
	"io"
	"os"
	"path"

	"github.com/stablecog/go-apps/database/ent"
	"github.com/stablecog/go-apps/utils"
	"k8s.io/klog/v2"
)

// Load the subscription tiers in the database
func (r *Repository) CreateSubscriptionTiers() error {
	// Read the subscription tiers from json
	tierPath := path.Join(utils.RootDir(), "subscription_tiers.json")
	jsonFile, err := os.Open(tierPath)
	if err != nil {
		return err
	}
	defer jsonFile.Close()
	byteValue, err := io.ReadAll(jsonFile)
	if err != nil {
		return err
	}

	var tiers []SubscriptionTierInput
	err = json.Unmarshal(byteValue, &tiers)
	if err != nil {
		return err
	}

	// Create the subscription tiers in the database
	for _, tier := range tiers {
		_, err = r.DB.SubscriptionTier.Create().SetName(tier.Name).SetBaseCredits(tier.BaseCredits).Save(r.Ctx)
		if err != nil {
			if ent.IsConstraintError(err) {
				// Update instead
				_, err = r.DB.SubscriptionTier.Update().SetName(tier.Name).SetBaseCredits(tier.BaseCredits).Save(r.Ctx)
				if err != nil {
					klog.Errorf("Failed to update subscription tier %s: %v", tier.Name, err)
					return err
				}
			}
			klog.Errorf("Failed to create subscription tier %s: %v", tier.Name, err)
			return err
		}
	}
	return nil
}

// The subscription_tiers.json
type SubscriptionTierInput struct {
	Name        string `json:"name"`
	BaseCredits int32  `json:"base_credits"`
}
