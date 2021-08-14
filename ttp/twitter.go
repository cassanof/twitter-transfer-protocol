package ttp

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/dghubble/oauth1"
)

const (
	URL_DM_NEW             = "https://api.twitter.com/1.1/direct_messages/events/new.json"
	URL_DM_SHOW            = "https://api.twitter.com/1.1/direct_messages/events/show.json"
	URL_DM_LIST            = "https://api.twitter.com/1.1/direct_messages/events/list.json"
	URL_DM_DESTROY         = "https://api.twitter.com/1.1/direct_messages/events/destroy.json"
	URL_DM_RATELIMIT       = "https://api.twitter.com/1.1/application/rate_limit_status.json?resources=direct_messages"
	URL_USER_FROM_USERNAME = "https://api.twitter.com/2/users/by/username/"
)

type TwitterClient struct {
	config     *oauth1.Config
	token      *oauth1.Token
	httpClient *http.Client
}

type ResponseErrors struct {
	Errors *[]ResponseError `json:"errors"`
}

type ResponseError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type TwitterUser struct {
	Data *TwitterUserData `json:"data"`
}

type TwitterUserData struct {
	Id       string `json:"id"`
	Name     string `json:"name"`
	Username string `json:"username"` // Username is the "handle"
}

type DirectMessageEvent struct {
	Id               string                     `json:"id,omitempty"`
	CreatedTimestamp string                     `json:"created_timestamp,omitempty"`
	Type             string                     `json:"type"` // Defaults to "message_create"
	Message          *DirectMessageEventMessage `json:"message_create"`
}

type DirectMessageEventMessage struct {
	Target      *DirectMessageEventMessageTarget `json:"target"`
	SenderId    string                           `json:"sender_id,omitempty"`
	MessageData *DirectMessageEventMessageData   `json:"message_data"`
}

type DirectMessageEventMessageTarget struct {
	RecipientId string `json:"recipient_id"`
}

type DirectMessageEventMessageData struct {
	Text string `json:"text"`
	// Don't need all the entities for now...
}

type DirectMessageEventSingle struct { // thanks to twitter api for useless json tags...
	Event *DirectMessageEvent `json:"event"`
}

type DirectMessageEvents struct { // again, another thanks for the additional useless "s" to event(s)
	Events     *[]DirectMessageEvent `json:"events"`
	NextCursor string                `json:"next_cursor"`
}

func NewTwitterClient(consumerKey string, consumerSecret string, accessToken string, accessSecret string) TwitterClient {
	config := oauth1.NewConfig(consumerKey, consumerSecret)
	token := oauth1.NewToken(accessToken, accessSecret)
	httpClient := config.Client(oauth1.NoContext, token)
	httpClient.Timeout = 30 * time.Second
	return TwitterClient{
		config,
		token,
		httpClient,
	}
}

func NewSendableDirectMessageEvent(recpt_id string, text string) DirectMessageEventSingle {
	return DirectMessageEventSingle{
		Event: &DirectMessageEvent{
			Type: "message_create",
			Message: &DirectMessageEventMessage{
				Target: &DirectMessageEventMessageTarget{
					RecipientId: recpt_id,
				},
				MessageData: &DirectMessageEventMessageData{
					Text: text,
				},
			},
		},
	}
}

func (tc *TwitterClient) GetUserFromHandle(handle string) (*TwitterUser, error) {
	url := URL_USER_FROM_USERNAME + handle
	user := new(TwitterUser)
	res, err := tc.httpClient.Get(url)
	if err != nil {
		return nil, err
	}

	resBody, err := ioutil.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		return nil, err
	}

	err = checkResForErrors(resBody)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(resBody, &user)

	return user, err
}

// TODO: Make a struct for this
// NOTE: Does the api req even show actual remaining calls? It looks like it doesn't...
func (tc *TwitterClient) GetRateLimitStatus() []byte {
	res, err := tc.httpClient.Get(URL_DM_RATELIMIT)
	if err != nil {
		return nil
	}

	resBody, err := ioutil.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		return nil
	}

	return resBody
}

func (tc *TwitterClient) ListDirectMessages() (*DirectMessageEvents, error) {
	dmEvents := new(DirectMessageEvents)
	res, err := tc.httpClient.Get(URL_DM_LIST)
	if err != nil {
		return nil, err
	}

	resBody, err := ioutil.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		return nil, err
	}

	err = checkResForErrors(resBody)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(resBody, &dmEvents)

	return dmEvents, err
}

func (tc *TwitterClient) ShowDirectMessage(msgId string) (*DirectMessageEventSingle, error) {
	dmEvent := new(DirectMessageEventSingle)
	res, err := tc.httpClient.Get(fmt.Sprintf("%s?id=%s", URL_DM_SHOW, msgId))
	if err != nil {
		return nil, err
	}

	resBody, err := ioutil.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		return nil, err
	}

	err = checkResForErrors(resBody)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(resBody, &dmEvent)

	return dmEvent, err
}

func (tc *TwitterClient) SendDirectMessage(dm *DirectMessageEventSingle) (*DirectMessageEventSingle, error) {
	reqBytes, _ := json.Marshal(dm)
	reqReader := bytes.NewReader(reqBytes)

	res, err := tc.httpClient.Post(URL_DM_NEW, "application/json", reqReader)
	if err != nil {
		return nil, err
	}

	resBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	err = checkResForErrors(resBody)
	if err != nil {
		return nil, err
	}

	resDM := new(DirectMessageEventSingle)
	err = json.Unmarshal(resBody, &resDM)
	if err != nil {
		return nil, err
	}

	return resDM, err
}

func checkResForErrors(resBody []byte) error {
	resErrors := new(ResponseErrors)
	json.Unmarshal(resBody, &resErrors)

	if (ResponseErrors{}) == *resErrors { // Check if resErrors struct is empty
		return nil // all good. there shouldn't be errors
	}

	errCodes := ""
	for i, resError := range *resErrors.Errors {
		errCodes = fmt.Sprintf(
			"%s%s%d",
			errCodes,
			(map[bool]string{true: "", false: ", "})[i == 0], // if i==0 don't put delimiter else put ", ".
			resError.Code,
		)
	}

	return errors.New(errCodes)
}
