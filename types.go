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

	RuntimeOwnershipHostStream = "host_stream"
	RuntimeOwnershipSDKOwned   = "sdk_owned"

	StreamMessageTypeText   = 1
	StreamMessageTypeBinary = 2
	StreamMessageTypePing   = 9

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

type Acknowledger interface {
	Acknowledge(ctx context.Context, req OutboundAck) (*AckResult, error)
}

type HostStreamer interface {
	ConnectStream(ctx context.Context, req HostStreamConnectRequest) (*StreamConnectResult, error)
	BuildStreamPing(ctx context.Context, req StreamPingRequest) (*StreamFrame, error)
	HandleStreamFrame(ctx context.Context, req StreamFrameRequest) (*StreamFrameResult, error)
}

type ConnectorMetadata struct {
	ID           string       `json:"id"`
	Platform     string       `json:"platform"`
	Label        string       `json:"label"`
	Description  string       `json:"description,omitempty"`
	Capabilities Capabilities `json:"capabilities"`
}

type Capabilities struct {
	LoginModes       []string `json:"login_modes"`
	Text             bool     `json:"text"`
	Media            bool     `json:"media"`
	GroupChat        bool     `json:"group_chat"`
	DirectChat       bool     `json:"direct_chat"`
	Stream           bool     `json:"stream"`
	Webhook          bool     `json:"webhook"`
	BlockStreaming   bool     `json:"block_streaming"`
	AckModes         []string `json:"ack_modes,omitempty"`
	RuntimeOwnership string   `json:"runtime_ownership,omitempty"`
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

type OutboundAck struct {
	WorkspaceUUID     string         `json:"workspace_uuid"`
	Platform          string         `json:"platform"`
	AccountUUID       string         `json:"account_uuid"`
	ChannelUUID       string         `json:"channel_uuid"`
	SessionUUID       string         `json:"session_uuid"`
	SourceMessageUUID string         `json:"source_message_uuid,omitempty"`
	ChatType          string         `json:"chat_type"`
	ChatID            string         `json:"chat_id"`
	TargetMessageID   string         `json:"target_message_id,omitempty"`
	Intent            string         `json:"intent,omitempty"`
	Action            string         `json:"action,omitempty"`
	Mode              string         `json:"mode,omitempty"`
	Emoji             string         `json:"emoji,omitempty"`
	Raw               map[string]any `json:"raw,omitempty"`
}

type AckResult struct {
	Platform    string         `json:"platform"`
	AccountUUID string         `json:"account_uuid"`
	Mode        string         `json:"mode,omitempty"`
	Status      string         `json:"status"`
	ReactionID  string         `json:"reaction_id,omitempty"`
	Raw         map[string]any `json:"raw,omitempty"`
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
	MetadataPlatform   string   `json:"metadata_platform,omitempty"`
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

type AckCase struct {
	Name    string         `json:"name,omitempty"`
	Request OutboundAck    `json:"request"`
	Expect  AckExpectation `json:"expect"`
}

type AckExpectation struct {
	Status     string `json:"status,omitempty"`
	Mode       string `json:"mode,omitempty"`
	ReactionID string `json:"reaction_id,omitempty"`
}

type RuntimeHealthExpectation struct {
	ConnectionState             string `json:"connection_state,omitempty"`
	RequireConnectedAt          bool   `json:"require_connected_at,omitempty"`
	RequireDisconnectedAt       bool   `json:"require_disconnected_at,omitempty"`
	RequireLastActivityAt       bool   `json:"require_last_activity_at,omitempty"`
	RequireLastPingAt           bool   `json:"require_last_ping_at,omitempty"`
	RequireLastPongAt           bool   `json:"require_last_pong_at,omitempty"`
	RequireLastEventAt          bool   `json:"require_last_event_at,omitempty"`
	RequireLastError            bool   `json:"require_last_error,omitempty"`
	RequireLastErrorAt          bool   `json:"require_last_error_at,omitempty"`
	RequireReconnectRequestedAt bool   `json:"require_reconnect_requested_at,omitempty"`
	RequireReconnectError       bool   `json:"require_reconnect_error,omitempty"`
	RequireReconnectErrorAt     bool   `json:"require_reconnect_error_at,omitempty"`
	SessionExpired              *bool  `json:"session_expired,omitempty"`
}

type HostStreamConnectRequest struct {
	WorkspaceUUID string         `json:"workspace_uuid,omitempty"`
	ChannelUUID   string         `json:"channel_uuid,omitempty"`
	Account       ChannelAccount `json:"account,omitempty"`
	Credential    map[string]any `json:"credential,omitempty"`
	State         map[string]any `json:"state,omitempty"`
}

type StreamConnectResult struct {
	URL             string            `json:"url,omitempty"`
	Headers         map[string]string `json:"headers,omitempty"`
	ServiceID       string            `json:"service_id,omitempty"`
	ReadMessageType int               `json:"read_message_type,omitempty"`
	PingInterval    any               `json:"ping_interval,omitempty"`
	PongTimeout     any               `json:"pong_timeout,omitempty"`
	State           any               `json:"state,omitempty"`
	HealthUpdates   map[string]any    `json:"health_updates,omitempty"`
}

type StreamPingRequest struct {
	ServiceID string `json:"service_id,omitempty"`
	State     any    `json:"state,omitempty"`
}

type StreamFrameRequest struct {
	WorkspaceUUID string         `json:"workspace_uuid,omitempty"`
	ChannelUUID   string         `json:"channel_uuid,omitempty"`
	Account       ChannelAccount `json:"account,omitempty"`
	Credential    map[string]any `json:"credential,omitempty"`
	MessageType   int            `json:"message_type,omitempty"`
	Data          []byte         `json:"data,omitempty"`
	ServiceID     string         `json:"service_id,omitempty"`
	State         any            `json:"state,omitempty"`
}

type StreamFrame struct {
	MessageType int    `json:"message_type,omitempty"`
	Data        []byte `json:"data,omitempty"`
}

type StreamFrameResult struct {
	ResponseFrames []StreamFrame      `json:"response_frames,omitempty"`
	HealthUpdates  map[string]any     `json:"health_updates,omitempty"`
	EventResult    *StreamEventResult `json:"event_result,omitempty"`
	CloseReason    string             `json:"close_reason,omitempty"`
	State          any                `json:"state,omitempty"`
}

type StreamEventResult struct {
	Type        string          `json:"type"`
	Ignored     bool            `json:"ignored,omitempty"`
	Reason      string          `json:"reason,omitempty"`
	SessionUUID string          `json:"session_uuid,omitempty"`
	MessageUUID string          `json:"message_uuid,omitempty"`
	Inbound     *InboundMessage `json:"inbound,omitempty"`
}

type HostStreamCase struct {
	Name    string                       `json:"name,omitempty"`
	Request HostStreamConnectRequest     `json:"request,omitempty"`
	Expect  HostStreamConnectExpectation `json:"expect,omitempty"`
	Ping    *HostStreamPingCase          `json:"ping,omitempty"`
	Frames  []HostStreamFrameCase        `json:"frames,omitempty"`
}

type HostStreamConnectExpectation struct {
	URLContains            string                   `json:"url_contains,omitempty"`
	ReadMessageType        int                      `json:"read_message_type,omitempty"`
	RequireServiceID       bool                     `json:"require_service_id,omitempty"`
	RequirePingInterval    bool                     `json:"require_ping_interval,omitempty"`
	RequirePongTimeout     bool                     `json:"require_pong_timeout,omitempty"`
	RequireState           bool                     `json:"require_state,omitempty"`
	RequireConnectedHealth bool                     `json:"require_connected_health,omitempty"`
	RuntimeHealth          RuntimeHealthExpectation `json:"runtime_health,omitempty"`
}

type HostStreamPingCase struct {
	Request StreamPingRequest         `json:"request,omitempty"`
	Expect  HostStreamPingExpectation `json:"expect,omitempty"`
}

type HostStreamPingExpectation struct {
	MessageType int  `json:"message_type,omitempty"`
	RequireData bool `json:"require_data,omitempty"`
}

type HostStreamFrameCase struct {
	Name    string                     `json:"name,omitempty"`
	Request StreamFrameRequest         `json:"request"`
	Expect  HostStreamFrameExpectation `json:"expect,omitempty"`
}

type HostStreamFrameExpectation struct {
	MinResponseFrames   int                      `json:"min_response_frames,omitempty"`
	ResponseMessageType int                      `json:"response_message_type,omitempty"`
	CloseReason         string                   `json:"close_reason,omitempty"`
	RuntimeHealth       RuntimeHealthExpectation `json:"runtime_health,omitempty"`
	EventType           string                   `json:"event_type,omitempty"`
	EventIgnored        *bool                    `json:"event_ignored,omitempty"`
	RequireEventResult  bool                     `json:"require_event_result,omitempty"`
	RequireFrameState   bool                     `json:"require_frame_state,omitempty"`
}

type Config struct {
	// Platform is the runtime channel/account platform expected in Beak-facing
	// messages and acknowledgements.
	Platform string
	// MetadataPlatform is the SDK connector identity reported by Metadata().
	// It defaults to Platform. Set it when one connector package supports
	// multiple runtime brands, for example lark metadata with feishu runtime.
	MetadataPlatform string

	MetadataProvider         MetadataProvider
	CredentialSchemaProvider CredentialSchemaProvider
	CredentialValidator      CredentialValidator
	LoginPoller              LoginPoller
	InboundParser            InboundParser
	Acknowledger             Acknowledger
	HostStreamer             HostStreamer

	CredentialCases []CredentialValidationCase
	LoginPollCases  []LoginPollCase
	InboundCases    []InboundCase
	AckCases        []AckCase
	HostStreamCases []HostStreamCase
}
