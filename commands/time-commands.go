package commands

import (
	"bpl2-discord/client"
	"context"
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
)

func toDiscordTimestamp(t string) string {
	layout := time.RFC3339
	parsedTime, err := time.Parse(layout, t)
	if err != nil {
		fmt.Println("Error parsing time:", err)
		return ""
	}
	return fmt.Sprintf("<t:%d:f>", parsedTime.Unix())
}

var GetTimesCommand = DiscordCommand{
	Command: &discordgo.ApplicationCommand{
		Name:        "times",
		Description: "Get the times for the current event",
	},
	Handler: func(s *discordgo.Session, i *discordgo.InteractionCreate, options optionMap, client *client.ClientWithResponses) {
		resp, err := client.GetCurrentEventWithResponse(context.TODO())

		if err != nil {
			EditResponse(s, i, "could not get current event")
			return
		}

		if resp.JSON200 == nil {
			EditResponse(s, i, "no current event")
			return
		}

		event := resp.JSON200
		content := fmt.Sprintf(`
Times for event "%s":
Signups: %s
Start: %s
End: %s
		`, event.Name, toDiscordTimestamp(event.ApplicationStartTime), toDiscordTimestamp(event.EventStartTime), toDiscordTimestamp(event.EventEndTime))

		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: content,
			},
		},
		)

	},
}
