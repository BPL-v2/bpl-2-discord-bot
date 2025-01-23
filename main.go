package main

import (
	"bpl2-discord/client"
	"bpl2-discord/commands"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

type optionMap = map[string]*discordgo.ApplicationCommandInteractionDataOption

func parseOptions(options []*discordgo.ApplicationCommandInteractionDataOption) (om optionMap) {
	fmt.Println("parseOptions")

	om = make(optionMap)
	for _, opt := range options {
		fmt.Println(opt)
		om[opt.Name] = opt
	}
	return
}

func interactionAuthor(i *discordgo.Interaction) *discordgo.User {
	if i.Member != nil {
		// i.Member.Roles = append(i.Member.Roles, i.Member.User.ID)
		return i.Member.User
	}
	return i.User
}

func handleEcho(s *discordgo.Session, i *discordgo.InteractionCreate, opts optionMap) {
	builder := new(strings.Builder)
	if v, ok := opts["author"]; ok && v.BoolValue() {
		author := interactionAuthor(i.Interaction)
		builder.WriteString("**" + author.Mention() + "** says: ")

	}
	builder.WriteString(opts["message"].StringValue())
	fmt.Println("handleEcho")
	fmt.Println(builder.String())

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: builder.String(),
		},
	})

	if err != nil {
		log.Panicf("could not respond to interaction: %s", err)
	}
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
		return
	}
	session, err := discordgo.New("Bot " + os.Getenv("DISCORD_BOT_TOKEN"))
	bplClient, err := client.AuthenticatedClient()
	if err != nil {
		log.Fatalf("could not create rest client: %s", err)
		return
	}

	commands.RegisterCommands(session, bplClient)
	session.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Printf("Logged in as %s", r.User.String())
	})

	err = session.Open()
	if err != nil {
		log.Fatalf("could not open session: %s", err)
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop
	log.Println("Graceful shutdown")

	err = session.Close()
	if err != nil {
		log.Printf("could not close session gracefully: %s", err)
	}

}
