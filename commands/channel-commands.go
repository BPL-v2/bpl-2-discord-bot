package commands

import (
	"bpl2-discord/client"
	"bpl2-discord/utils"
	"fmt"
	"log"
	"sort"
	"strings"
	"sync"

	"github.com/bwmarrin/discordgo"
)

type ChannelWithChildren struct {
	*discordgo.Channel
	Children []*ChannelWithChildren
}

type ChannelMap map[string]*ChannelWithChildren

func (c ChannelMap) AddThreadChildren(s *discordgo.Session, guildID string) {
	threadList, err := s.GuildThreadsActive(guildID)
	if err != nil {
		fmt.Println("Error getting threads:", err)
		return
	}
	for _, thread := range threadList.Threads {
		if parentChannel, ok := c[thread.ParentID]; ok {
			parentChannel.Children = append(parentChannel.Children, &ChannelWithChildren{Channel: thread})
		} else {
			c[thread.ParentID] = &ChannelWithChildren{
				Channel:  &discordgo.Channel{ID: thread.ParentID},
				Children: []*ChannelWithChildren{{Channel: thread}},
			}
		}
	}
}

func GetAllChannels(s *discordgo.Session, guildID string) (ChannelMap, error) {
	allChannels, err := s.GuildChannels(guildID)
	if err != nil {
		return nil, err
	}
	channelTreeMap := make(map[string]*ChannelWithChildren)
	for _, channel := range allChannels {
		channelTreeMap[channel.ID] = &ChannelWithChildren{
			Channel: channel,
		}
	}
	for _, channel := range allChannels {
		if channel.ParentID != "" {
			channelTreeMap[channel.ParentID].Children = append(channelTreeMap[channel.ParentID].Children, channelTreeMap[channel.ID])
		}
	}
	return channelTreeMap, nil
}

func SplitMessage(message *discordgo.Message) []*discordgo.MessageSend {
	// If a message is too long, split it into multiple messages with the last message containing all the embeds/stickers etc
	contents := make([]string, 0)
	currentContent := ""
	for _, line := range strings.Split(message.Content, "\n") {
		if len(currentContent)+len(line) > 2000 {
			contents = append(contents, currentContent)
			currentContent = ""
		}
		currentContent += line + "\n"
	}
	contents = append(contents, currentContent)

	messageSends := make([]*discordgo.MessageSend, 0)
	messageSends = append(messageSends, &discordgo.MessageSend{
		Content:    contents[0],
		Embeds:     message.Embeds,
		TTS:        message.TTS,
		Components: message.Components,
		StickerIDs: utils.Map(message.StickerItems, func(sticker *discordgo.StickerItem) string {
			return sticker.ID
		}),
		Flags: message.Flags,
	})
	for _, content := range contents[1:] {
		messageSends = append(messageSends, &discordgo.MessageSend{
			Content: content,
		})
	}
	return messageSends

}

func DuplicateMessage(s *discordgo.Session, channelId string, message *discordgo.Message) {
	for _, subMessage := range SplitMessage(message) {
		s.ChannelMessageSendComplex(channelId, subMessage)
	}
}

func DuplicateThread(s *discordgo.Session, channelId string, thread *discordgo.Channel) {
	message := &discordgo.MessageSend{}
	threadMessages, err := s.ChannelMessages(thread.ID, 100, "", "", "")
	sort.Slice(threadMessages, func(i, j int) bool {
		return threadMessages[i].ID < threadMessages[j].ID
	})
	if err == nil {
		message = SplitMessage(threadMessages[0])[0]
	}

	threadCopy, err := s.ForumThreadStartComplex(
		channelId,
		&discordgo.ThreadStart{
			Name: thread.Name,
			Type: thread.Type,
		},
		message,
	)
	for _, message := range threadMessages[1:] {
		DuplicateMessage(s, threadCopy.ID, message)
	}

}

func DuplicateChannel(s *discordgo.Session, guildID string, channel *ChannelWithChildren, parentId string) (*discordgo.Channel, error) {
	copyChannel, err := s.GuildChannelCreateComplex(guildID, discordgo.GuildChannelCreateData{
		Name:                 channel.Name,
		Type:                 channel.Type,
		ParentID:             parentId,
		Position:             channel.Position,
		Topic:                channel.Topic,
		NSFW:                 channel.NSFW,
		Bitrate:              channel.Bitrate,
		UserLimit:            channel.UserLimit,
		PermissionOverwrites: channel.PermissionOverwrites,
	})
	if err != nil {
		return nil, err
	}

	if channel.Type == discordgo.ChannelTypeGuildText {
		messages, err := s.ChannelMessages(channel.ID, 100, "", "", "")
		if err != nil {
			return nil, err
		}
		sort.Slice(messages, func(i, j int) bool {
			return messages[i].ID < messages[j].ID
		})
		for _, message := range messages {
			DuplicateMessage(s, copyChannel.ID, message)
		}
	} else if channel.Type == discordgo.ChannelTypeGuildForum {
		sort.Slice(channel.Children, func(i, j int) bool {
			return channel.Children[i].ID < channel.Children[j].ID
		})
		for _, thread := range channel.Children {
			DuplicateThread(s, copyChannel.ID, thread.Channel)
		}
	}

	return copyChannel, nil
}

var CopyCategoryCommand = DiscordCommand{
	Command: &discordgo.ApplicationCommand{
		Name:        "copy-category",
		Description: "copy an entire category with all its channels",
		Options: []*discordgo.ApplicationCommandOption{{
			Type:         discordgo.ApplicationCommandOptionChannel,
			Name:         "category",
			Description:  "The category to copy",
			Required:     true,
			ChannelTypes: []discordgo.ChannelType{discordgo.ChannelTypeGuildCategory},
		}},
		DefaultMemberPermissions: &PermissionManageChannels,
	},
	Handler: func(s *discordgo.Session, i *discordgo.InteractionCreate, options optionMap, client *client.ClientWithResponses) {
		log.Println("CopyCategoryCommand called")
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "category duplication started",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		},
		)
		categoryId := options["category"].Value.(string)

		channelMap, err := GetAllChannels(s, i.GuildID)
		channelMap.AddThreadChildren(s, i.GuildID)
		if err != nil {
			EditResponse(s, i, "could not get channels")
			return
		}
		copyCategory, err := s.GuildChannelCreateComplex(i.GuildID,
			discordgo.GuildChannelCreateData{
				Name:                 channelMap[categoryId].Name + " (copy)",
				Type:                 discordgo.ChannelTypeGuildCategory,
				PermissionOverwrites: channelMap[categoryId].PermissionOverwrites,
				Topic:                channelMap[categoryId].Topic,
				Bitrate:              channelMap[categoryId].Bitrate,
				UserLimit:            channelMap[categoryId].UserLimit,
				NSFW:                 channelMap[categoryId].NSFW,
				ParentID:             channelMap[categoryId].ParentID,
				RateLimitPerUser:     channelMap[categoryId].RateLimitPerUser,
			})

		if err != nil {
			EditResponse(s, i, "could not create category")
			return
		}
		wg := sync.WaitGroup{}
		for _, child := range channelMap[categoryId].Children {
			wg.Add(1)
			// go DuplicateChannel(s, i.GuildID, child, copyCategory.ID)
			go func() {
				defer wg.Done()
				DuplicateChannel(s, i.GuildID, child, copyCategory.ID)
			}()
		}
		wg.Wait()
		EditResponse(s, i, "category duplication complete")
	},
}
var DeleteCategoryCommand = DiscordCommand{
	Command: &discordgo.ApplicationCommand{
		Name:        "delete-category",
		Description: "delete an entire category and all its channels",
		Options: []*discordgo.ApplicationCommandOption{{
			Type:         discordgo.ApplicationCommandOptionChannel,
			Name:         "category",
			Description:  "The category to delete",
			Required:     true,
			ChannelTypes: []discordgo.ChannelType{discordgo.ChannelTypeGuildCategory},
		}},
		DefaultMemberPermissions: &PermissionManageChannels,
	},
	Handler: func(s *discordgo.Session, i *discordgo.InteractionCreate, options optionMap, client *client.ClientWithResponses) {
		log.Println("DeleteCategoryCommand called")
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "category deletion started",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		},
		)
		categoryId := options["category"].Value.(string)
		// get all channels in the category
		channelTreeMap, err := GetAllChannels(s, i.GuildID)
		if err != nil {
			EditResponse(s, i, "could not get channels")
			return
		}
		for _, child := range channelTreeMap[categoryId].Children {
			go s.ChannelDelete(child.ID)
		}
		go s.ChannelDelete(categoryId)
		EditResponse(s, i, "category deletion complete")
	},
}
