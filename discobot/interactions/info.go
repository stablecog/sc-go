package interactions

import (
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/stablecog/sc-go/database/repository"
	"github.com/stablecog/sc-go/discobot/responses"
	"github.com/stablecog/sc-go/log"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

func (c *DiscordInteractionWrapper) NewInfoCommand() *DiscordInteraction {
	return &DiscordInteraction{
		// Command spec
		ApplicationCommand: &discordgo.ApplicationCommand{
			Name:        "info",
			Description: "Display account information such as available credits.",
		},
		// The handler for the command
		Handler: func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			var discordUserId string
			if i.Member != nil {
				discordUserId = i.Member.User.ID
			} else {
				discordUserId = i.User.ID
			}
			if u := c.Disco.CheckAuthorization(s, i); u != nil {
				// Get credits
				credits, err := c.Repo.GetCreditsForUser(u.ID)
				if err != nil {
					log.Errorf("Error getting credits for user %v", err)
					responses.ErrorResponseInitial(s, i, responses.PRIVATE)
					return
				}

				var creditsGT0 []*repository.UserCreditsQueryResult
				totalCredits := 0
				for _, credit := range credits {
					if credit.RemainingAmount > 0 {
						totalCredits += int(credit.RemainingAmount)
						creditsGT0 = append(creditsGT0, credit)
					}
				}

				// Localize numbers
				prettyPrinter := message.NewPrinter(language.English)

				// Format the expiration stuff as a string
				expiryString := ""

				for _, credit := range creditsGT0 {
					if credit.ExpiresAt.Before(repository.NEVER_EXPIRE) {
						expiryString += prettyPrinter.Sprintf("`%s`: %d , expires: %s\n", credit.CreditTypeName, int(credit.RemainingAmount), credit.ExpiresAt.Format("January 2, 2006"))
					} else {
						expiryString += prettyPrinter.Sprintf("`%s`: %d\n", credit.CreditTypeName, int(credit.RemainingAmount))
					}
				}

				responses.InitialInteractionResponse(s, i, &responses.InteractionResponseOptions{
					EmbedTitle: "ℹ️ Account Information",
					EmbedContent: prettyPrinter.Sprintf(
						"Member Since %s\n\n"+
							"`Credits Remaining:` %d\n\n**Credit Details:**\n%s",
						u.CreatedAt.Format("January 2, 2006"),
						totalCredits,
						expiryString,
					),
					Privacy: responses.PRIVATE,
				})
			} else {
				c.LoginInteractionMap.Put(discordUserId, &LoginInteraction{
					Session:     s,
					Interaction: i,
					InsertedAt:  time.Now(),
				})
			}
		},
	}
}
