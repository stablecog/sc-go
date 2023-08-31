package interactions

import (
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/google/uuid"
	"github.com/olekukonko/tablewriter"
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

				// Get tippable credits
				tippableSum, err := c.Repo.GetTippableSumForUser(u.ID)
				if err != nil {
					log.Errorf("Error getting tippable sum for user %v", err)
					responses.ErrorResponseInitial(s, i, responses.PRIVATE)
					return
				}

				var creditsGT0 []*repository.UserCreditsQueryResult
				totalCredits := 0
				for _, credit := range credits {
					// Only include > 0 and free type
					if credit.RemainingAmount > 0 || credit.CreditTypeID == uuid.MustParse(repository.FREE_CREDIT_TYPE_ID) {
						totalCredits += int(credit.RemainingAmount)
						creditsGT0 = append(creditsGT0, credit)
					}
				}

				// Group credits into groups that expire at the same time
				creditGroups := make(map[time.Time][]*repository.UserCreditsQueryResult)
				for _, credit := range creditsGT0 {
					if credit.ExpiresAt.Before(repository.NEVER_EXPIRE) {
						creditGroups[credit.ExpiresAt] = append(creditGroups[credit.ExpiresAt], credit)
					} else {
						creditGroups[repository.NEVER_EXPIRE] = append(creditGroups[repository.NEVER_EXPIRE], credit)
					}
				}

				var creditGroupArrays [][]*repository.UserCreditsQueryResult
				for _, v := range creditGroups {
					creditGroupArrays = append(creditGroupArrays, v)
				}

				// Localize numbers
				prettyPrinter := message.NewPrinter(language.English)

				// Create table for credits
				creditTable := make([][]string, len(creditsGT0))

				for _, credit := range creditsGT0 {
					creditTable = append(creditTable, []string{fmt.Sprintf("%s:", credit.CreditTypeName), prettyPrinter.Sprintf("%d", int(credit.RemainingAmount))})
				}

				creditDetailString := "**Credit Details**\n"
				for _, creditGroup := range creditGroupArrays {
					if len(creditGroup) > 0 {
						if creditGroup[0].ExpiresAt.Before(repository.NEVER_EXPIRE) {
							expiresAtString := creditGroup[0].ExpiresAt.Format("January 2, 2006")
							creditDetailString += fmt.Sprintf("*Credits Expiring On %s*\n", expiresAtString)
						} else {
							creditDetailString += fmt.Sprintf("*Non-Expiring Credits*\n")
						}
						tableString := &strings.Builder{}
						table := tablewriter.NewWriter(tableString)
						for _, credit := range creditGroup {
							// Build table
							table.Append([]string{credit.CreditTypeName, prettyPrinter.Sprintf("%d", int(credit.RemainingAmount))})
							table.SetBorder(false)
							table.SetCenterSeparator("")
							table.SetColumnSeparator("")
							table.SetRowSeparator("")
							table.SetNoWhiteSpace(true)
							table.SetTablePadding(" ")
						}
						table.Render()
						creditDetailString += fmt.Sprintf("```%s```", tableString.String())
					}
				}

				responses.InitialInteractionResponse(s, i, &responses.InteractionResponseOptions{
					EmbedTitle: "ℹ️ Account Information",
					EmbedContent: prettyPrinter.Sprintf(
						"**Username:** %s\n"+
							"**Email:** %s\n"+
							"**Member since:** %s\n"+
							"**Total Credits:** %d\n"+
							"**Total Tippable Credits:** %d\n\n%s",
						u.Username,
						u.Email,
						u.CreatedAt.Format("January 2, 2006"),
						totalCredits,
						tippableSum,
						creditDetailString,
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
