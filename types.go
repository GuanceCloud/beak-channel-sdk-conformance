package conformance

import (
	"context"
	"encoding/json"
)

const (
	ChatTypeGroup  = "group"
	ChatTypeDirect = "direct"

	LoginModeQRCode     = "qr_code"
	LoginModeCredential = "credential"

	LoginStatusApproved = "approved"
	LoginStatusExpired  = "expired"
	LoginStatusFailed   = "failed"
	LoginStatusPending  = "pending"
)

type MetadataProvider interface {
	Metadata() ConnectorMetadata
}

type CredentialSchemaProvider interface {
	CredentialSchema(ctx context.Context) CredentialSchema
}

type CredentialValidator interface {
	ValidateCredential(ctx context.Context, req CredentialValidationRequest) (*CredentialValidationResult, error)
}

type LoginPoller interface {
	PollLogin(ctx context.Context, req LoginPollRequest) (*LoginStatus, error)
}

type InboundParser interface {
	ParseInbound(ctx context.Context, fixture InboundFixture) ([]InboundMessage, error)
}

type ConnectorMetadata struct {
	ID           string       `json:"id"`
	Platform     string       `json:"platform"`
	Label        string       `json:"label"`
	Description  string       `json:"description,omitempty"`
	Capabilities Capabilities `json:"capabilities"`
}

type Capabilities struct {
	LoginModes     []string `json:"login_modes"`
	Text           bool     `json:"text"`
	Media          bool     `json:"media"`
	GroupChat      bool     `json:"group_chat"`
	DirectChat     bool     `json:"direct_chat"`
	Stream         bool     `json:"stream"`
	Webhook        bool     `json:"webhook"`
	BlockStreaming bool     `json:"block_streaming"`
}

type CredentialSchema struct {
	Type                 string                     `json:"type"`
	LoginModes           []string                   `json:"login_modes"`
	Properties           map[string]CredentialField `json:"properties"`
	Required             []string                   `json:"required,omitempty"`
	AdditionalProperties bool                       `json:"additionalProperties"`
}

type CredentialField struct {
	Type        string `json:"type"`
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	Secret      bool   `json:"secret,omitempty"`
}

type CredentialValidationRequest struct {
	WorkspaceUUID string         `json:"workspace_uuid,omitempty"`
	ChannelUUID   string         `json:"channel_uuid,omitempty"`
	Platform      string         `json:"platform,omitempty"`
	Credential    map[string]any `json:"credential,omitempty"`
	State         map[string]any `json:"state,omitempty"`
}

type CredentialValidationResult struct {
	Valid       bool           `json:"valid"`
	AccountKey  string         `json:"account_key,omitempty"`
	DisplayName string         `json:"display_name,omitempty"`
	Credential  map[string]any `json:"credential,omitempty"`
	State       map[string]any `json:"state,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
	Error       string         `json:"error,omitempty"`
}

type ChannelAccount struct {
	UUID          string         `json:"account_uuid"`
	WorkspaceUUID string         `json:"workspace_uuid"`
	ChannelUUID   string         `json:"channel_uuid"`
	Platform      string         `json:"platform"`
	DisplayName   string         `json:"display_name,omitempty"`
	Credential    map[string]any `json:"credential,omitempty"`
	State         map[string]any `json:"state,omitempty"`
	Status        string         `json:"status,omitempty"`
}

type LoginPollRequest struct {
	WorkspaceUUID  string         `json:"workspace_uuid"`
	ChannelUUID    string         `json:"channel_uuid"`
	ChallengeUUID  string         `json:"challenge_uuid,omitempty"`
	ChallengeCode  string         `json:"challenge_code,omitempty"`
	ChallengeState map[string]any `json:"challenge_state,omitempty"`
}

type LoginStatus struct {
	Status     string         `json:"status"`
	Confirmed  bool           `json:"confirmed"`
	Expired    bool           `json:"expired"`
	Account    ChannelAccount `json:"account,omitempty"`
	Credential map[string]any `json:"credential,omitempty"`
	State      map[string]any `json:"state,omitempty"`
}

type InboundFixture struct {
	Name          string          `json:"name,omitempty"`
	WorkspaceUUID string          `json:"workspace_uuid,omitempty"`
	ChannelUUID   string          `json:"channel_uuid,omitempty"`
	AccountUUID   string          `json:"account_uuid,omitempty"`
	Platform      string          `json:"platform,omitempty"`
	Credential    map[string]any  `json:"credential,omitempty"`
	AccountState  map[string]any  `json:"account_state,omitempty"`
	Raw           json.RawMessage `json:"raw,omitempty"`
	Metadata      map[string]any  `json:"metadata,omitempty"`
}

type InboundMessage struct {
	WorkspaceUUID     string            `json:"workspace_uuid"`
	Platform          string            `json:"platform"`
	AccountUUID       string            `json:"account_uuid"`
	ChannelUUID       string            `json:"channel_uuid"`
	ChatType          string            `json:"chat_type"`
	ChatID            string            `json:"chat_id"`
	ThreadID          string            `json:"thread_id,omitempty"`
	ChatDisplayName   string            `json:"chat_display_name,omitempty"`
	ChatAvatarURL     string            `json:"chat_avatar_url,omitempty"`
	ChatIdentity      ChatIdentity      `json:"chat_identity,omitempty"`
	SenderID          string            `json:"sender_id"`
	SenderDisplayName string            `json:"sender_display_name,omitempty"`
	MessageID         string            `json:"message_id,omitempty"`
	Text              string            `json:"text"`
	DedupeKey         string            `json:"dedupe_key,omitempty"`
	Mentions          []MentionIdentity `json:"mentions,omitempty"`
	MentionedMe       bool              `json:"mentioned_me,omitempty"`
	MentionAll        bool              `json:"mention_all,omitempty"`
	Raw               map[string]any    `json:"raw,omitempty"`
}

type MentionIdentity struct {
	ID          string `json:"id"`
	IDType      string `json:"id_type,omitempty"`
	DisplayName string `json:"display_name,omitempty"`
}

type ChatIdentity struct {
	ID          string `json:"id,omitempty"`
	IDType      string `json:"id_type,omitempty"`
	Type        string `json:"type,omitempty"`
	DisplayName string `json:"display_name,omitempty"`
	AvatarURL   string `json:"avatar_url,omitempty"`
}

type CredentialValidationCase struct {
	Name    string                          `json:"name,omitempty"`
	Request CredentialValidationRequest     `json:"request"`
	Expect  CredentialValidationExpectation `json:"expect"`
}

type CredentialValidationExpectation struct {
	Valid              bool     `json:"valid"`
	AccountKey         string   `json:"account_key,omitempty"`
	DisplayName        string   `json:"display_name,omitempty"`
	RequireAccountID   bool     `json:"require_account_id,omitempty"`
	RequireBotIdentity bool     `json:"require_bot_identity,omitempty"`
	VolatileKeys       []string `json:"volatile_keys,omitempty"`
}

type LoginPollCase struct {
	Name    string               `json:"name,omitempty"`
	Request LoginPollRequest     `json:"request"`
	Expect  LoginPollExpectation `json:"expect"`
}

type LoginPollExpectation struct {
	Approved           bool   `json:"approved"`
	AccountKey         string `json:"account_key,omitempty"`
	DisplayName        string `json:"display_name,omitempty"`
	RequireAccountID   bool   `json:"require_account_id,omitempty"`
	RequireBotIdentity bool   `json:"require_bot_identity,omitempty"`
}

type InboundCase struct {
	Name    string             `json:"name,omitempty"`
	Fixture InboundFixture     `json:"fixture"`
	Expect  InboundExpectation `json:"expect"`
}

type InboundExpectation struct {
	MinMessages       int      `json:"min_messages,omitempty"`
	MessageIndex      int      `json:"message_index,omitempty"`
	ChatType          string   `json:"chat_type,omitempty"`
	ChatID            string   `json:"chat_id,omitempty"`
	ChatDisplayName   string   `json:"chat_display_name,omitempty"`
	ChatAvatarURL     string   `json:"chat_avatar_url,omitempty"`
	ChatIdentityID    string   `json:"chat_identity_id,omitempty"`
	SenderID          string   `json:"sender_id,omitempty"`
	SenderDisplayName string   `json:"sender_display_name,omitempty"`
	Text              string   `json:"text,omitempty"`
	TextTrimmedEmpty  *bool    `json:"text_trimmed_empty,omitempty"`
	MentionedMe       *bool    `json:"mentioned_me,omitempty"`
	MentionAll        *bool    `json:"mention_all,omitempty"`
	MentionIDs        []string `json:"mention_ids,omitempty"`
	RequireMessageID  bool     `json:"require_message_id,omitempty"`
	RequireDedupeKey  bool     `json:"require_dedupe_key,omitempty"`
}

type Config struct {
	Platform string

	MetadataProvider         MetadataProvider
	CredentialSchemaProvider CredentialSchemaProvider
	CredentialValidator      CredentialValidator
	LoginPoller              LoginPoller
	InboundParser            InboundParser

	CredentialCases []CredentialValidationCase
	LoginPollCases  []LoginPollCase
	InboundCases    []InboundCase
}
