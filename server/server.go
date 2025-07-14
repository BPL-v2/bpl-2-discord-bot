package server

import (
	"bpl2-discord/client"
	"bpl2-discord/commands"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/gin-gonic/gin"
)

func SetRoutes(r *gin.Engine, session *discordgo.Session, bplClient *client.ClientWithResponses) {
	group := r.Group("/discord")
	group.Handle("POST", ":guild_id/assign-roles", triggerRoleAssignmentHandler(session, bplClient))
	group.Handle("GET", ":guild_id/members", getAllGuildMembers(session))
}

func triggerRoleAssignmentHandler(session *discordgo.Session, bplClient *client.ClientWithResponses) gin.HandlerFunc {
	return func(c *gin.Context) {
		guildID := c.Param("guild_id")
		if guildID == "" {
			c.JSON(400, gin.H{"message": "missing guild_id"})
			return
		}
		num, err := commands.AssignRoles(session, bplClient, guildID)
		if err != nil {
			c.JSON(500, gin.H{"message": "error assigning roles"})
			return
		}
		c.JSON(200, gin.H{"message": fmt.Sprintf("assigned %d roles", num)})
	}
}

func getAllGuildMembers(session *discordgo.Session) gin.HandlerFunc {
	return func(c *gin.Context) {
		guildID := c.Param("guild_id")
		if guildID == "" {
			c.JSON(400, gin.H{"message": "missing guild_id"})
			return
		}
		members, err := commands.GetAllGuildMembers(session, guildID)
		if err != nil {
			c.JSON(500, gin.H{"message": "error getting members"})
			return
		}
		c.JSON(200, members)
	}

}
