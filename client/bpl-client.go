package client

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt"
)

type AuthTransport struct {
	Transport http.RoundTripper
	Token     string
}

func (a *AuthTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	fmt.Println("Adding auth header")
	req.Header.Set("Authorization", "Bearer "+a.Token)
	return a.Transport.RoundTrip(req)
}

func (c *ClientWithResponses) Authenticate() error {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"user_id":     0,
			"permissions": []string{"admin"},
			"exp":         time.Now().Add(time.Hour * 24 * 1000).Unix(),
		})

	jwt, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	fmt.Println("JWT: ", jwt)
	if err != nil {
		log.Fatalf("could not sign token: %s", err)
	}

	BackendUrl := os.Getenv("BACKEND_URL_FOR_DISCORD_BOT")
	httpClient := &http.Client{
		Transport: &AuthTransport{
			Transport: http.DefaultTransport,
			Token:     jwt,
		},
	}

	// Pass the custom HTTP client to NewClientWithResponses
	newClient, err := NewClientWithResponses(BackendUrl, WithHTTPClient(httpClient))
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
func (c *ClientWithResponses) GetAllEvents() ([]Event, error) {
	resp, err := c.GetEventsWithResponse(context.TODO())
	if err != nil {
		return nil, err
	}
	if resp.JSON200 == nil {
		return nil, fmt.Errorf("no events found")
	}
	events := resp.JSON200
	if len(*events) == 0 {
		return nil, fmt.Errorf("no events found")
	}
	return *events, nil
}

func (c *ClientWithResponses) GetCurrentEvent() (*Event, error) {

	events, err := c.GetAllEvents()
	if err != nil {
		return nil, err
	}
	for _, event := range events {
		if event.IsCurrent {
			return &event, nil
		}
	}
	return nil, nil
}
func (c *ClientWithResponses) GetLatestEvent() (*Event, error) {
	events, err := c.GetAllEvents()
	if err != nil {
		return nil, fmt.Errorf("could not get events: %w", err)
	}
	latestEvent := events[0]
	latestStartTime, err := time.Parse(time.RFC3339, latestEvent.EventStartTime)
	if err != nil {
		return nil, fmt.Errorf("could not parse latest event start time: %w", err)
	}
	for _, event := range events {
		currentStartTime, err := time.Parse(time.RFC3339, event.EventStartTime)
		if err != nil {
			return nil, fmt.Errorf("could not parse event start time: %w", err)
		}
		if currentStartTime.After(latestStartTime) {
			latestStartTime = currentStartTime
			latestEvent = event
		}
	}
	return &latestEvent, nil

}
