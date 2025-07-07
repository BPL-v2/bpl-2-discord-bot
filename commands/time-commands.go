package commands

import (
	"bpl2-discord/client"
	"fmt"
	"log"
	"math/rand"
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
		log.Println("GetTimesCommand called")
		event, err := client.GetLatestEvent()
		if err != nil {
			EditResponse(s, i, "could not get current event")
			return
		}
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

var sortWhenMessages = []string{
	"Don't know... why are you asking me?",
	"I'm not sure, but I'm sure it's gonna be soon.",
	"Some time between now and the end of the world.",
	"When the sorting algorithm finally gets its act together.",
	"After I finish my coffee break, which might be never.",
	"When pigs fly and hell freezes over.",
	"Soon™ - trademark pending.",
	"When the stars align and the gods smile upon us.",
	"Probably when you least expect it, which is never.",
	"After I solve world peace and cure the common cold.",
	"When the algorithm stops being indecisive.",
	"After I count all the grains of sand on the beach.",
	"Soon, but not soon enough for your liking.",
	"Right after I finish this incredibly long list of excuses.",
	"After I figure out why my socks keep disappearing in the dryer.",
	"Soon, but time is relative, so who really knows?",
	"When my code stops having a mind of its own.",
	"After I stop making up new excuses as to why I can't tell you the time.",
	"After boat league is over.",
	"How would I know? Nobody is telling me anything.",
	"That's a great question, let me ask my friend Google.",
}

var SortWhenCommand = DiscordCommand{
	Command: &discordgo.ApplicationCommand{
		Name:        "sortwhen",
		Description: "Returns time for the first sort",
	},
	Handler: func(s *discordgo.Session, i *discordgo.InteractionCreate, options optionMap, client *client.ClientWithResponses) {
		log.Println("SortWhenCommand called")
		randomIndex := rand.Intn(len(sortWhenMessages))
		content := sortWhenMessages[randomIndex]
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: content,
			},
		},
		)
	},
}
var SignupswhenCommand = DiscordCommand{
	Command: &discordgo.ApplicationCommand{
		Name:        "signupswhen",
		Description: "Returns time when signups begin",
	},
	Handler: func(s *discordgo.Session, i *discordgo.InteractionCreate, options optionMap, client *client.ClientWithResponses) {
		log.Println("SignupswhenCommand called")
		event, err := client.GetLatestEvent()
		if err != nil {
			EditResponse(s, i, "could not get current event")
			return
		}
		content := fmt.Sprintf(`%s. Just kidding, signups begin at %s
		`, sortWhenMessages[rand.Intn(len(sortWhenMessages))], toDiscordTimestamp(event.ApplicationStartTime))
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: content,
			},
		},
		)
	},
}
