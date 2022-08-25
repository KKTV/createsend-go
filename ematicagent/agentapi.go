package ematicagent

// Ematic Agent client library to handle api request
// must of file from the third party library
// https://github.com/sourcegraph/createsend-go
// which no longer maintain and support
//

import (
	"errors"
	"log"
	"net/http"
	"sync"
	"time"
)

var (
	// errors
	ErrorData        = errors.New("wrong input data")
	ErrorClient      = errors.New("not validate client")
	ErrorPermission  = errors.New("permission fail")
	EmaticDateFormat = "2006-01-02"
)

// Agent struct for Ematic agent
// membership user role
// signupdate user created date
// lastactivitydate user last token created date
// trialexpireddate trial expired date
// paidexpireddate paid user expired date
// upgradedate upagrde to premium date
// cancellationdate canceld date
type AgentAPI struct {
	ClientID string
	ListID   string
	APIToken string
	API      *APIClient
	Debug    bool
	sync.Mutex
}

// SetListID the List ID might not always the first one
func (a *AgentAPI) SetListID(listID string) (err error) {
	if listID == "" {
		return ErrorData
	}
	a.ListID = listID
	return err
}

// AddSubscriber adds a subscriber.
func (a *AgentAPI) AddSubscriber(email string, kv map[string]interface{}) (err error) {
	sub := NewSubscriber{
		EmailAddress: email,
		Resubscribe:  false,
	}

	fields := []CustomField{}
	for k, v := range kv {
		fields = append(fields, CustomField{Key: k, Value: v})
	}
	sub.CustomFields = fields

	log.Println("[INFO] AddSubscriber", email, sub.CustomFields)
	if a.Debug {
		return
	}

	err = a.API.AddSubscriber(a.ListID, sub)
	return err
}

// UpdateSubscriber updates a subscriber.
func (a *AgentAPI) UpdateSubscriber(email string, kv map[string]interface{}) (err error) {
	sub := NewSubscriber{
		EmailAddress: email,
		Resubscribe:  false,
	}

	fields := []CustomField{}
	for k, v := range kv {
		fields = append(fields, CustomField{Key: k, Value: v})
	}
	sub.CustomFields = fields

	log.Println("[INFO] UpdateSubscriber", email, sub.CustomFields)
	if a.Debug {
		return
	}

	err = a.API.UpdateSubscriber(a.ListID, sub.EmailAddress, sub)

	return err
}

// Unsubscribe changes the status of a subscriber from Active to Unsubscribed.
func (a *AgentAPI) Unsubscribe(email string) (err error) {

	log.Println("[INFO] Unsubscribe", email)
	if a.Debug {
		return
	}

	err = a.API.Unsubscribe(a.ListID, email)
	return err
}

// Signup create a user at ematic
//  membership,  signupdate, lastactivitydate, trialexpireddate
func (a *AgentAPI) Signup(email, trialexpireddate string) (err error) {
	nowDate := time.Now().Format(EmaticDateFormat)
	kv := make(map[string]interface{})
	kv["membership"] = "freetrial"
	kv["trialexpireddate"] = trialexpireddate
	kv["signupdate"] = nowDate
	kv["lastactivitydate"] = nowDate

	log.Println("[INFO] Signup", email, kv)
	if a.Debug {
		return
	}

	return a.AddSubscriber(email, kv)
}

// Signin update user lastactivitydate
func (a *AgentAPI) Signin(email string) (err error) {
	nowDate := time.Now().Format(EmaticDateFormat)
	kv := make(map[string]interface{})
	kv["lastactivitydate"] = nowDate

	log.Println("[INFO] Signin", email, kv)
	if a.Debug {
		return
	}

	return a.UpdateSubscriber(email, kv)
}

// Expired membership expired
func (a *AgentAPI) Expired(email string) (err error) {
	kv := make(map[string]interface{})
	kv["membership"] = "expired"

	log.Println("[INFO] Signin", email, kv)
	if a.Debug {
		return
	}

	return a.UpdateSubscriber(email, kv)
}

// Paid update membership, paidexpireddate, upgradedate
func (a *AgentAPI) Paid(email, paidexpireddate string) (err error) {
	nowDate := time.Now().Format(EmaticDateFormat)
	kv := make(map[string]interface{})
	kv["membership"] = "premium"
	kv["paidexpireddate"] = paidexpireddate
	kv["upgradedate"] = nowDate

	log.Println("[INFO] Paid", email, kv)
	if a.Debug {
		return
	}

	return a.UpdateSubscriber(email, kv)
}

// Cancel update cancellationdate
func (a *AgentAPI) Cancel(email string) (err error) {
	nowDate := time.Now().Format(EmaticDateFormat)
	kv := make(map[string]interface{})
	kv["cancellationdate"] = nowDate

	log.Println("[INFO] Cancel", email, kv)
	if a.Debug {
		return
	}

	return a.UpdateSubscriber(email, kv)
}

// NewEmaticAgent get an ematic agent and validate token
func NewAgentAPI(clientID string, apiToken string) (agent *AgentAPI, err error) {

	if apiToken == "" || clientID == "" {
		return nil, ErrorData
	}

	// basic setup
	agent = new(AgentAPI)
	agent.ClientID = clientID
	agent.APIToken = apiToken

	authClient := &http.Client{
		Timeout:   time.Duration(10 * time.Second),
		Transport: &APIKeyAuthTransport{APIKey: apiToken},
	}

	c := NewAPIClient(authClient)
	agent.API = c

	lists, err := agent.API.ListLists(agent.ClientID)

	if err != nil {
		return nil, err
	}

	if len(lists) == 1 {
		// only one listID, set as default
		agent.ListID = lists[0].ListID
	}

	return agent, nil
}
