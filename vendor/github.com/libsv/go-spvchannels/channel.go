package spvchannels

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"
)

func (c *Client) getChanelBaseEndpoint() string {
	u := url.URL{
		Scheme: c.cfg.httpScheme(),
		Host:   c.cfg.baseURL,
		Path:   path.Join("/api", c.cfg.version, "/account"),
	}
	return u.String()
}

func (c *Client) getTokenBaseEndpoint(accountid int64, channelid string) string {
	return fmt.Sprintf("%s/%d/channel/%s/api-token", c.getChanelBaseEndpoint(), accountid, channelid)
}

// ChannelsRequest hold data for get channels request for a particular account
type ChannelsRequest struct {
	AccountID int64 `json:"accountid"`
}

// ChannelsReply hold data for get channels reply. It is a list of channel's detail
type ChannelsReply struct {
	Channels []struct {
		ID           string    `json:"id"`
		Href         string    `json:"href"`
		PublicRead   bool      `json:"public_read"`
		PublicWrite  bool      `json:"public_write"`
		Sequenced    bool      `json:"sequenced"`
		Locked       bool      `json:"locked"`
		Head         int       `json:"head"`
		Retention    Retention `json:"retention"`
		AccessTokens []struct {
			ID          string `json:"id"`
			Token       string `json:"token"`
			Description string `json:"description"`
			CanRead     bool   `json:"can_read"`
			CanWrite    bool   `json:"can_write"`
		} `json:"access_tokens"`
	} `json:"channels"`
}

// ChannelRequest hold data for get channel request
type ChannelRequest struct {
	AccountID int64  `json:"accountid"`
	ChannelID string `json:"channelid"`
}

// ChannelReply hold data for get channel reply
type ChannelReply struct {
	ID          string `json:"id"`
	Href        string `json:"href"`
	PublicRead  bool   `json:"public_read"`
	PublicWrite bool   `json:"public_write"`
	Sequenced   bool   `json:"sequenced"`
	Locked      bool   `json:"locked"`
	Head        int    `json:"head"`
	Retention   struct {
		MinAgeDays int  `json:"min_age_days"`
		MaxAgeDays int  `json:"max_age_days"`
		AutoPrune  bool `json:"auto_prune"`
	} `json:"retention"`
	AccessTokens []struct {
		ID          string `json:"id"`
		Token       string `json:"token"`
		Description string `json:"description"`
		CanRead     bool   `json:"can_read"`
		CanWrite    bool   `json:"can_write"`
	} `json:"access_tokens"`
}

// ChannelUpdateRequest hold data for update channel request.
// The request contains the account and channel identification,
// And the properties values to be updated. These properties defines
// common permission for the channel
type ChannelUpdateRequest struct {
	AccountID   int64  `json:"accountid"`
	ChannelID   string `json:"channelid"`
	PublicRead  bool   `json:"public_read"`
	PublicWrite bool   `json:"public_write"`
	Locked      bool   `json:"locked"`
}

// ChannelUpdateReply hold data for update channel reply
// If successful, the reply contains the confirmed updated properties
type ChannelUpdateReply struct {
	PublicRead  bool `json:"public_read"`
	PublicWrite bool `json:"public_write"`
	Locked      bool `json:"locked"`
}

// ChannelDeleteRequest hold data for delete channel request
// The request contains the account and channel identification
type ChannelDeleteRequest struct {
	AccountID int64  `json:"accountid"`
	ChannelID string `json:"channelid"`
}

// ChannelCreateRequest hold data for create channel request
// The request should contain the account id, and optionally some
// channel properties to be initialised.
type ChannelCreateRequest struct {
	AccountID   int64     `json:"accountid"`
	PublicRead  bool      `json:"public_read"`
	PublicWrite bool      `json:"public_write"`
	Sequenced   bool      `json:"sequenced"`
	Retention   Retention `json:"retention"`
}

// Retention the data retention policy of a channel.
type Retention struct {
	MinAgeDays int  `json:"min_age_days"`
	MaxAgeDays int  `json:"max_age_days"`
	AutoPrune  bool `json:"auto_prune"`
}

// ChannelCreateReply hold data for create channel reply
// It contains the new channel id, it's properties and the first
// default created token to allow authentification the communication
// on this channel
type ChannelCreateReply struct {
	ID           string    `json:"id"`
	Href         string    `json:"href"`
	PublicRead   bool      `json:"public_read"`
	PublicWrite  bool      `json:"public_write"`
	Sequenced    bool      `json:"sequenced"`
	Locked       bool      `json:"locked"`
	Head         int       `json:"head"`
	Retention    Retention `json:"retention"`
	AccessTokens []struct {
		ID          string `json:"id"`
		Token       string `json:"token"`
		Description string `json:"description"`
		CanRead     bool   `json:"can_read"`
		CanWrite    bool   `json:"can_write"`
	} `json:"access_tokens"`
}

// TokenRequest hold data for get token request
// A token belong to a particular channel, which again belong to a particular account.
// To identify a token, it needs to provide account id, channel id, and token id
type TokenRequest struct {
	AccountID int64  `json:"accountid"`
	ChannelID string `json:"channelid"`
	TokenID   string `json:"tokenid"`
}

// TokenReply hold data for get token reply
// The reply contains
//
// - token id (a number which is uniquely identified in the database)
// - token value, which will be used for authentification to read/write messages on the channel
// - some permission properties attached to the token
type TokenReply struct {
	ID          string `json:"id"`
	Token       string `json:"token"`
	Description string `json:"description"`
	CanRead     bool   `json:"can_read"`
	CanWrite    bool   `json:"can_write"`
}

// TokenDeleteRequest hold data for delete token request
// A token belong to a particular channel, which again belong to a particular account.
// To identify a token, it needs to provide account id, channel id, and token id
type TokenDeleteRequest struct {
	AccountID int64  `json:"accountid"`
	ChannelID string `json:"channelid"`
	TokenID   string `json:"tokenid"`
}

// TokensRequest hold data for get tokens request
// The request contains the account id and channel id.
type TokensRequest struct {
	AccountID int64  `json:"accountid"`
	ChannelID string `json:"channelid"`
}

// TokensReply hold data for get tokens reply. It is a list of detail for tokens
type TokensReply []TokenReply

// TokenCreateRequest hold data for create token request
// The request should contains existing account and channel id,
// with optionally some description and permission properties attached to the token
type TokenCreateRequest struct {
	AccountID   int64  `json:"accountid"`
	ChannelID   string `json:"channelid"`
	Description string `json:"description"`
	CanRead     bool   `json:"can_read"`
	CanWrite    bool   `json:"can_write"`
}

// TokenCreateReply hold data for create token reply
// It hold the id and value of the new token, and some of the token's properties
type TokenCreateReply struct {
	ID          string `json:"id"`
	Token       string `json:"token"`
	Description string `json:"description"`
	CanRead     bool   `json:"can_read"`
	CanWrite    bool   `json:"can_write"`
}

// Channels get the list of channels with detail for a particular account
func (c *Client) Channels(ctx context.Context, r ChannelsRequest) (*ChannelsReply, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		fmt.Sprintf("%s/%d/channel/list", c.getChanelBaseEndpoint(), r.AccountID),
		nil,
	)
	if err != nil {
		return nil, err
	}

	res := ChannelsReply{}
	if err := c.sendRequest(req, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// Channel get single channel's detail for a particular account
func (c *Client) Channel(ctx context.Context, r ChannelRequest) (*ChannelReply, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		fmt.Sprintf("%s/%d/channel/%s", c.getChanelBaseEndpoint(), r.AccountID, r.ChannelID),
		nil,
	)

	if err != nil {
		return nil, err
	}

	res := ChannelReply{}
	if err := c.sendRequest(req, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// ChannelUpdate update the channel's properties.
func (c *Client) ChannelUpdate(ctx context.Context, r ChannelUpdateRequest) (*ChannelUpdateReply, error) {
	payload, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		fmt.Sprintf("%s/%d/channel/%s", c.getChanelBaseEndpoint(), r.AccountID, r.ChannelID),
		bytes.NewBuffer(payload),
	)

	if err != nil {
		return nil, err
	}

	res := ChannelUpdateReply{}
	if err := c.sendRequest(req, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// ChannelDelete delete the channel a particular channel. It return nothing, which is a 204 http code (No Content)
func (c *Client) ChannelDelete(ctx context.Context, r ChannelDeleteRequest) error {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodDelete,
		fmt.Sprintf("%s/%d/channel/%s", c.getChanelBaseEndpoint(), r.AccountID, r.ChannelID),
		nil,
	)

	if err != nil {
		return err
	}

	return c.sendRequest(req, nil)
}

// ChannelCreate create a new channel for a particular account
func (c *Client) ChannelCreate(ctx context.Context, r ChannelCreateRequest) (*ChannelCreateReply, error) {
	payload, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		fmt.Sprintf("%s/%d/channel", c.getChanelBaseEndpoint(), r.AccountID),
		bytes.NewBuffer(payload),
	)

	if err != nil {
		return nil, err
	}

	res := ChannelCreateReply{}
	if err := c.sendRequest(req, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// Token get the token's detail.
func (c *Client) Token(ctx context.Context, r TokenRequest) (*TokenReply, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		fmt.Sprintf("%s/%s", c.getTokenBaseEndpoint(r.AccountID, r.ChannelID), r.TokenID),
		nil,
	)

	if err != nil {
		return nil, err
	}

	res := TokenReply{}
	if err := c.sendRequest(req, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// TokenDelete delete a particular token
//
// It is typically used when an admin create a temporary
// communication capability for a particular user on a channel,
//
// After the user has finished he usage of the channel,
// the admin then delete the token
func (c *Client) TokenDelete(ctx context.Context, r TokenDeleteRequest) error {
	req, err := http.NewRequestWithContext(ctx,
		http.MethodDelete,
		fmt.Sprintf("%s/%s", c.getTokenBaseEndpoint(r.AccountID, r.ChannelID), r.TokenID),
		nil,
	)

	if err != nil {
		return err
	}

	return c.sendRequest(req, nil)
}

// Tokens get the list of tokens. It return a full list of tokens
// for a particular account id and channel id
func (c *Client) Tokens(ctx context.Context, r TokensRequest) (*TokensReply, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.getTokenBaseEndpoint(r.AccountID, r.ChannelID), nil)
	if err != nil {
		return nil, err
	}

	res := TokensReply{}
	if err := c.sendRequest(req, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// TokenCreate create a new token for a particular account and channel
//
// It is typically used when and admin/orchestrator need to create
// a communication channel for a group of 2 or more peers.
//
// He then create a channel and a list of token, then give the tokens
// to each peers so they can communicate through the channel
func (c *Client) TokenCreate(ctx context.Context, r TokenCreateRequest) (*TokenCreateReply, error) {
	payload, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.getTokenBaseEndpoint(r.AccountID, r.ChannelID), bytes.NewBuffer(payload),
	)
	if err != nil {
		return nil, err
	}

	res := TokenCreateReply{}
	if err := c.sendRequest(req, &res); err != nil {
		return nil, err
	}

	return &res, nil
}
