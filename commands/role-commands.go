package commands

import (
	"bpl2-discord/client"
	"bpl2-discord/utils"
	"context"
	"fmt"
	"strconv"
	"sync"

	"github.com/bwmarrin/discordgo"
)

func getAllGuildMembers(s *discordgo.Session, guildID string) ([]*discordgo.Member, error) {
	lastUserId := ""
	members := make([]*discordgo.Member, 0)
	for {
		newMembers, err := s.GuildMembers(guildID, lastUserId, 1000)
		if err != nil {
			return nil, err
		}
		members = append(members, newMembers...)
		if len(newMembers) < 1000 {
			break
		}
		lastUserId = newMembers[len(newMembers)-1].User.ID
	}
	return members, nil
}

var RoleCreateCommand = DiscordCommand{
	Command: &discordgo.ApplicationCommand{
		Name:                     "role-create",
		Description:              "creates roles for teams",
		DefaultMemberPermissions: &PermissionManageRoles,
	},
	Handler: func(s *discordgo.Session, i *discordgo.InteractionCreate, options optionMap, client *client.ClientWithResponses) {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "role creation started",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		},
		)
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
		allRoles, err := s.GuildRoles(i.GuildID)
		if err != nil {
			EditResponse(s, i, "could not get guild roles")
			return
		}
		for _, team := range event.Teams {
			found := false
			for _, role := range allRoles {
				if role.Name == team.Name {
					found = true
				}
			}
			if !found {
				_, err := s.GuildRoleCreate(i.GuildID, &discordgo.RoleParams{Name: team.Name})
				if err != nil {
					EditResponse(s, i, "could not create role for team "+team.Name)
					return
				}
			}
		}

	},
}

var RoleAssignCommand = DiscordCommand{
	Command: &discordgo.ApplicationCommand{
		Name:                     "role-assign",
		Description:              "assigns roles to sorted users",
		DefaultMemberPermissions: &PermissionManageRoles,
	},
	Handler: func(s *discordgo.Session, i *discordgo.InteractionCreate, options optionMap, client *client.ClientWithResponses) {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "role assignment started",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		},
		)

		resp, err := client.GetCurrentEventWithResponse(context.TODO())

		if err != nil {
			EditResponse(s, i, "could not get current event")
			return
		}
		event := resp.JSON200

		signupResponse, err := client.GetEventSignupsWithResponse(context.TODO(), event.Id)
		if err != nil {
			EditResponse(s, i, "could not get signups")
			return
		}
		discordIdToTeamId := make(map[string]string)
		for teamId, signups := range *signupResponse.JSON200 {
			for _, signup := range signups {
				if signup.User.DiscordId != nil {
					discordIdToTeamId[*signup.User.DiscordId] = teamId
				}
			}
		}

		members, err := getAllGuildMembers(s, i.GuildID)
		if err != nil {
			EditResponse(s, i, "could not get guild members")
			return
		}

		allRoles, err := s.GuildRoles(i.GuildID)
		if err != nil {
			EditResponse(s, i, "could not get guild roles")
			return
		}

		teamRoles := make(map[string]string)
		for _, team := range event.Teams {
			teamId := strconv.Itoa(team.Id)
			for _, role := range allRoles {
				if role.Name == team.Name {
					teamRoles[teamId] = role.ID
				}
			}
			if _, ok := teamRoles[teamId]; !ok {
				EditResponse(s, i, "could not find role for team "+team.Name)
				return
			}
		}
		wg := sync.WaitGroup{}
		mu := sync.Mutex{}
		counter := 0
		for _, member := range members {
			if teamId, ok := discordIdToTeamId[member.User.ID]; ok {
				newRoles := make([]string, 0)
				for _, roleId := range member.Roles {
					if !utils.ValuesContain(teamRoles, roleId) {
						newRoles = append(newRoles, roleId)
					}
				}
				if teamId != "0" {
					newRoles = append(newRoles, teamRoles[teamId])
				}
				if !utils.HaveSameEntries(member.Roles, newRoles) {
					wg.Add(1)
					go func(member *discordgo.Member) {
						defer wg.Done()
						s.GuildMemberEdit(i.GuildID, member.User.ID, &discordgo.GuildMemberParams{Roles: &newRoles})
						mu.Lock()
						counter++
						if counter%10 == 0 {
							EditResponse(s, i, fmt.Sprintf("assigned roles to %d users", counter))
						}
						mu.Unlock()
					}(member)
				}
			}
		}
		wg.Wait()

		EditResponse(s, i, fmt.Sprintf("assigned roles to %d users - finished", counter))

	},
}
