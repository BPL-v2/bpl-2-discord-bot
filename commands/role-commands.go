package commands

import (
	"bpl2-discord/client"
	"log"

	"github.com/bwmarrin/discordgo"
)

func GetAllGuildMembers(s *discordgo.Session, guildID string) ([]*discordgo.Member, error) {
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
		log.Println("RoleCreateCommand called")
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "role creation started",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		},
		)
		event, err := client.GetCurrentEvent()
		if err != nil {
			EditResponse(s, i, "could not get current event")
			return
		}
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
		EditResponse(s, i, "role creation complete")
	},
}

func AssignRoles(s *discordgo.Session, client *client.ClientWithResponses, guildId string) (int, error) {
	// event, err := client.GetCurrentEvent()
	// if err != nil {
	// 	return 0, err
	// }

	// signupResponse, err := client.GetEventSignupsWithResponse(context.TODO(), event.Id)
	// if err != nil || signupResponse.StatusCode() > 299 {
	// 	fmt.Println("could not get signups", err)
	// 	return 0, err
	// }

	// discordIdToTeamId := make(map[string]int)
	// for _, signup := range *signupResponse.JSON200 {
	// 	if signup.User.DiscordId != nil && signup.TeamId != nil {
	// 		discordIdToTeamId[*signup.User.DiscordId] = *signup.TeamId
	// 	}
	// }
	// members, err := GetAllGuildMembers(s, guildId)
	// if err != nil {
	// 	return 0, err
	// }
	// allRoles, err := s.GuildRoles(guildId)
	// if err != nil {
	// 	return 0, err
	// }

	// teamRoles := make(map[int]string)
	// for _, team := range event.Teams {
	// 	for _, role := range allRoles {
	// 		if role.Name == team.Name {
	// 			teamRoles[team.Id] = role.ID
	// 		}
	// 	}
	// 	if _, ok := teamRoles[team.Id]; !ok {
	// 		fmt.Println("could not find role for team", team.Name)
	// 		return 0, fmt.Errorf("could not find role for team %s", team.Name)
	// 	}
	// }
	// wg := sync.WaitGroup{}
	counter := 0
	// for _, member := range members {
	// 	teamId := discordIdToTeamId[member.User.ID]
	// 	newRoles := make([]string, 0)
	// 	for _, roleId := range member.Roles {
	// 		if !utils.ValuesContain(teamRoles, roleId) {
	// 			newRoles = append(newRoles, roleId)
	// 		}
	// 	}
	// 	if teamId != 0 {
	// 		newRoles = append(newRoles, teamRoles[teamId])
	// 	}
	// 	if !utils.HaveSameEntries(member.Roles, newRoles) {
	// 		fmt.Println("assigning role to", member.User.Username)
	// 		counter++
	// 		wg.Add(1)
	// 		go func(member *discordgo.Member) {
	// 			defer wg.Done()
	// 			_, err := s.GuildMemberEdit(guildId, member.User.ID, &discordgo.GuildMemberParams{Roles: &newRoles})
	// 			if err != nil {
	// 				fmt.Println("could not assign roles to", member.User.Username, err)
	// 			}
	// 		}(member)
	// 	}
	// }
	// wg.Wait()
	return counter, nil

}

var RoleAssignCommand = DiscordCommand{
	Command: &discordgo.ApplicationCommand{
		Name:                     "role-assign",
		Description:              "assigns roles to sorted users",
		DefaultMemberPermissions: &PermissionManageRoles,
	},
	Handler: func(s *discordgo.Session, i *discordgo.InteractionCreate, options optionMap, client *client.ClientWithResponses) {
		// log.Println("RoleAssignCommand called")
		// s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		// 	Type: discordgo.InteractionResponseChannelMessageWithSource,
		// 	Data: &discordgo.InteractionResponseData{
		// 		Content: "role assignment started",
		// 		Flags:   discordgo.MessageFlagsEphemeral,
		// 	},
		// },
		// )
		// numAssigned, err := AssignRoles(s, client, i.GuildID)
		// if err != nil {
		// 	fmt.Println(err)
		// 	EditResponse(s, i, "could not assign roles")
		// 	return
		// }
		// EditResponse(s, i, fmt.Sprintf("assigned roles to %d users", numAssigned))
	},
}
