package client

import (
	"context"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
)

func (c *ClientWithResponses) Authenticate() error {
	BackendUrl := os.Getenv("BACKEND_URL_FOR_DISCORD_BOT")
	resp, err := c.LoginDiscordBotWithResponse(context.TODO(), LoginDiscordBotJSONRequestBody{
		Token: os.Getenv("DISCORD_BOT_TOKEN"),
	})

	if err != nil {
		return err
	}
	token := *resp.JSON200
	URL, err := url.Parse(BackendUrl)

	jar, _ := cookiejar.New(nil)
	jar.SetCookies(URL, []*http.Cookie{{Name: "auth", Value: token}})

	clientWithCookie := &http.Client{Jar: jar}
	newClient, err := NewClientWithResponses(BackendUrl, WithHTTPClient(clientWithCookie))
	if err != nil {
		return err
	}
	*c = *newClient
	return nil
}

func AuthenticatedClient() (*ClientWithResponses, error) {
	BackendUrl := os.Getenv("BACKEND_URL_FOR_DISCORD_BOT")
	bplClient, err := NewClientWithResponses(BackendUrl)
	if err != nil {
		log.Fatalf("could not create client: %s", err)
		return nil, err
	}
	err = bplClient.Authenticate()
	if err != nil {
		log.Fatalf("could not authenticate client: %s", err)
		return nil, err
	}
	return bplClient, nil
}
