package server

import (
	"bpl2-discord/client"
	"bpl2-discord/commands"
	"fmt"
	"os"

	"github.com/bwmarrin/discordgo"
	"github.com/gin-gonic/gin"
)

func SetRoutes(r *gin.Engine, session *discordgo.Session, bplClient *client.ClientWithResponses) {
	group := r.Group("/discord")
	group.Handle("POST", "/assign-roles", triggerRoleAssignmentHandler(session, bplClient))
}

func triggerRoleAssignmentHandler(session *discordgo.Session, bplClient *client.ClientWithResponses) gin.HandlerFunc {
	return func(c *gin.Context) {
		num, err := commands.AssignRoles(session, bplClient, os.Getenv("GUILD_ID"))
		if err != nil {
			c.JSON(500, gin.H{"message": "error assigning roles"})
			return
		}
		c.JSON(200, gin.H{"message": fmt.Sprintf("assigned %d roles", num)})
	}
}
