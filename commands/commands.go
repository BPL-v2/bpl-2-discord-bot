package commands

import (
	"bpl2-discord/client"
	"bpl2-discord/utils"
	"log"
	"os"

	"github.com/bwmarrin/discordgo"
)

// list of all commands available for the discord bot
var commands = []DiscordCommand{
	RoleAssignCommand,
	RoleCreateCommand,

	// GetTimesCommand,
}

type DiscordCommand struct {
	Command *discordgo.ApplicationCommand
	Handler func(s *discordgo.Session, i *discordgo.InteractionCreate, options optionMap, client *client.ClientWithResponses)
}

func (c DiscordCommand) Register(session *discordgo.Session, client *client.ClientWithResponses) {
	session.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if i.Type != discordgo.InteractionApplicationCommand {
			return
		}

		data := i.ApplicationCommandData()
		if data.Name != c.Command.Name {
			return
		}
		c.Handler(s, i, parseOptions(data.Options), client)
	})
}

type optionMap = map[string]*discordgo.ApplicationCommandInteractionDataOption

func parseOptions(options []*discordgo.ApplicationCommandInteractionDataOption) (om optionMap) {
	om = make(optionMap)
	for _, opt := range options {
		om[opt.Name] = opt
	}
	return
}

func commandHandler(commandMap map[string]DiscordCommand, bplClient *client.ClientWithResponses) func(*discordgo.Session, *discordgo.InteractionCreate) {
	return func(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
		if interaction.Type != discordgo.InteractionApplicationCommand {
			return
		}
		data := interaction.ApplicationCommandData()
		if c, ok := commandMap[data.Name]; ok {
			c.Handler(session, interaction, parseOptions(data.Options), bplClient)
		}
	}
}

func cleanUpDeprecatedCommands(session *discordgo.Session, commandMap map[string]DiscordCommand) {
	App := os.Getenv("DISCORD_CLIENT_ID")
	oldCommands, err := session.ApplicationCommands(App, "")
	if err != nil {
		log.Fatalf("could not fetch old commands: %s", err)
		return
	}
	for _, command := range oldCommands {
		if _, ok := commandMap[command.Name]; !ok {
			err := session.ApplicationCommandDelete(App, "", command.ID)
			if err != nil {
				log.Fatalf("could not delete command %s: %s", command.Name, err)
			}
		}
	}
	return
}

func RegisterCommands(session *discordgo.Session, bplClient *client.ClientWithResponses) error {
	App := os.Getenv("DISCORD_CLIENT_ID")
	commandMap := make(map[string]DiscordCommand)
	for _, c := range commands {
		commandMap[c.Command.Name] = c
	}
	session.AddHandler(commandHandler(commandMap, bplClient))
	cleanUpDeprecatedCommands(session, commandMap)

	_, err := session.ApplicationCommandBulkOverwrite(App, "", utils.Map(commands, func(c DiscordCommand) *discordgo.ApplicationCommand {
		return c.Command
	}))
	if err != nil {
		return err
	}
	return nil
}

func DeferResponse(
	s *discordgo.Session,
	i *discordgo.InteractionCreate,
	content string,
) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: content,
		},
	},
	)
}

func EditResponse(
	s *discordgo.Session,
	i *discordgo.InteractionCreate,
	content string,
) {
	s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: &content,
	},
	)
}
