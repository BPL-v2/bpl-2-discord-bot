package main

import (
	"bpl2-discord/client"
	"bpl2-discord/commands"
	"bpl2-discord/server"
	"log"
	"os"
	"os/signal"

	"github.com/bwmarrin/discordgo"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

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
	r := gin.Default()
	r.Use(gin.Recovery())
	server.SetRoutes(r, session, bplClient)
	r.Run(":9876")

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop
	log.Println("Graceful shutdown")

	err = session.Close()
	if err != nil {
		log.Printf("could not close session gracefully: %s", err)
	}

}
