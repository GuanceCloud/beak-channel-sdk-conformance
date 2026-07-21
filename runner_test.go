package conformance

import (
	"context"
	"errors"
	"strings"
	"testing"
)

type fakeConnector struct{}

func (fakeConnector) Metadata() ConnectorMetadata {
	return ConnectorMetadata{
		ID:       "fake",
		Platform: "fake",
		Label:    "Fake",
		Capabilities: Capabilities{
			LoginModes:          []string{LoginModeCredential, LoginModeQRCode},
			Text:                true,
			GroupChat:           true,
			Webhook:             true,
			WebhookRegistration: WebhookRegistrationManual,
			AckModes:            []string{"reaction"},
		},
	}
}

func (fakeConnector) CredentialSchema(context.Context) CredentialSchema {
	return CredentialSchema{
		Type:       "object",
		LoginModes: []string{LoginModeCredential},
		Properties: map[string]CredentialField{
			"account_id": {Type: "string"},
			"secret":     {Type: "string", Secret: true},
		},
		Required: []string{"account_id", "secret"},
	}
}

func (fakeConnector) ValidateCredential(_ context.Context, req CredentialValidationRequest) (*CredentialValidationResult, error) {
	if req.Credential["secret"] == "timeout" {
		return nil, errors.New("temporary credential validation timeout")
	}
	if req.Credential["secret"] == "bad" {
		return &CredentialValidationResult{
			Valid:      false,
			AccountKey: "stable-account",
			Error:      "invalid secret",
		}, nil
	}
	return &CredentialValidationResult{
		Valid:       true,
		AccountKey:  "stable-account",
		DisplayName: "Stable Bot",
		Credential: map[string]any{
			"account_id": "stable-account",
		},
		State: map[string]any{
			"bot_identity": map[string]any{
				"id":      "bot-open-id",
				"id_type": "open_id",
			},
		},
	}, nil
}

func (fakeConnector) PollLogin(context.Context, LoginPollRequest) (*LoginStatus, error) {
	return &LoginStatus{
		Status:    LoginStatusApproved,
		Confirmed: true,
		Account: ChannelAccount{
			UUID:        "stable-account",
			DisplayName: "Stable Bot",
			Credential:  map[string]any{"account_id": "stable-account"},
			State: map[string]any{
				"bot_identity": map[string]any{"id": "bot-open-id"},
			},
		},
		Credential: map[string]any{"account_id": "stable-account"},
	}, nil
}

func (fakeConnector) ParseInbound(_ context.Context, fixture InboundFixture) ([]InboundMessage, error) {
	if fixture.Name == "ignored" {
		return nil, nil
	}
	return []InboundMessage{{
		Platform:        "fake",
		ChatType:        ChatTypeGroup,
		ChatID:          "chat-1",
		ThreadID:        "thread-1",
		ChatDisplayName: "Ops Room",
		ChatAvatarURL:   "https://example.test/ops.png",
		ChatIdentity: ChatIdentity{
			ID:          "chat-1",
			IDType:      "chat_id",
			Type:        ChatTypeGroup,
			DisplayName: "Ops Room",
			AvatarURL:   "https://example.test/ops.png",
		},
		SenderID:          "user-1",
		SenderDisplayName: "Alice",
		MessageID:         "msg-1",
		Text:              "",
		ReferencedMessage: &ReferencedMessage{
			MessageID:   "quoted-1",
			ThreadID:    "thread-1",
			SenderID:    "user-2",
			MessageType: "text",
			Text:        "quoted text",
		},
		MentionedMe: true,
		Mentions:    []MentionIdentity{{ID: "bot-open-id", IDType: "open_id"}},
	}}, nil
}

func (fakeConnector) Acknowledge(context.Context, OutboundAck) (*AckResult, error) {
	return &AckResult{
		Platform:    "fake",
		AccountUUID: "stable-account",
		Mode:        "reaction",
		Status:      "sent",
		ReactionID:  "reaction-1",
	}, nil
}

func (fakeConnector) Send(_ context.Context, req OutboundMessage) (*SendResult, error) {
	if req.Text == "fail" {
		return nil, errors.New("expected send failure")
	}
	return &SendResult{Platform: "fake", AccountUUID: req.AccountUUID, MessageID: "sent-1", Raw: map[string]any{"delivery_method": "fake"}}, nil
}

func TestRun(t *testing.T) {
	connector := fakeConnector{}
	trueValue := true
	Run(t, Config{
		Platform:                 "fake",
		MetadataProvider:         connector,
		CredentialSchemaProvider: connector,
		CredentialValidator:      connector,
		LoginPoller:              connector,
		InboundParser:            connector,
		Sender:                   connector,
		Acknowledger:             connector,
		CredentialCases: []CredentialValidationCase{{
			Name: "valid credential",
			Request: CredentialValidationRequest{
				Credential: map[string]any{
					"account_id":   "stable-account",
					"secret":       "good",
					"access_token": "volatile-token",
				},
			},
			Expect: CredentialValidationExpectation{
				Valid:                  true,
				AccountKey:             "stable-account",
				DisplayName:            "Stable Bot",
				RequireAccountID:       true,
				RequireBotIdentity:     true,
				RequiredCredentialKeys: []string{"account_id"},
			},
		}, {
			Name: "invalid credential",
			Request: CredentialValidationRequest{
				Credential: map[string]any{
					"account_id": "stable-account",
					"secret":     "bad",
				},
			},
			Expect: CredentialValidationExpectation{Valid: false},
		}, {
			Name: "transient credential validation failure",
			Request: CredentialValidationRequest{
				Credential: map[string]any{
					"account_id": "stable-account",
					"secret":     "timeout",
				},
			},
			Expect: CredentialValidationExpectation{RequireGoError: true},
		}},
		LoginPollCases: []LoginPollCase{{
			Name:    "approved login",
			Request: LoginPollRequest{ChallengeCode: "qr"},
			Expect: LoginPollExpectation{
				Approved:           true,
				AccountKey:         "stable-account",
				DisplayName:        "Stable Bot",
				RequireAccountID:   true,
				RequireBotIdentity: true,
			},
		}},
		InboundCases: []InboundCase{{
			Name:    "follow up only at bot",
			Fixture: InboundFixture{Name: "follow-up"},
			Expect: InboundExpectation{
				ChatDisplayName:   "Ops Room",
				ChatAvatarURL:     "https://example.test/ops.png",
				ChatIdentityID:    "chat-1",
				ThreadID:          "thread-1",
				SenderDisplayName: "Alice",
				TextTrimmedEmpty:  &trueValue,
				MentionedMe:       &trueValue,
				MentionIDs:        []string{"bot-open-id"},
				ReferencedMessage: &ReferencedMessageExpectation{
					MessageID:   "quoted-1",
					ThreadID:    "thread-1",
					SenderID:    "user-2",
					MessageType: "text",
					Text:        "quoted text",
				},
				RequireMessageID: true,
			},
		}, {
			Name:    "unsupported event is explicitly ignored",
			Fixture: InboundFixture{Name: "ignored"},
			Expect:  InboundExpectation{ExpectNoMessages: true},
		}},
		SendCases: []SendCase{{
			Name: "outbound sequence",
			Steps: []SendStep{{
				Name: "success",
				Request: OutboundMessage{
					AccountUUID: "stable-account", ChatType: ChatTypeGroup, ChatID: "chat-1", MessageUUID: "outbound-1", Text: "hello", Format: "markdown",
				},
				Expect: SendExpectation{MessageID: "sent-1", RequireMessageID: true, RequiredRawKeys: []string{"delivery_method"}},
			}, {
				Name:    "expected failure",
				Request: OutboundMessage{AccountUUID: "stable-account", ChatType: ChatTypeGroup, ChatID: "chat-1", Text: "fail"},
				Expect:  SendExpectation{RequireError: true, ErrorContains: "expected send failure"},
			}},
		}},
		AckCases: []AckCase{{
			Name: "processing acknowledgement",
			Request: OutboundAck{
				AccountUUID:     "stable-account",
				ChatType:        ChatTypeGroup,
				ChatID:          "chat-1",
				TargetMessageID: "msg-1",
				Action:          "start",
			},
			Expect: AckExpectation{
				Status:     "sent",
				Mode:       "reaction",
				ReactionID: "reaction-1",
			},
		}},
	})
}

type sdkOwnedConnector struct{ fakeConnector }

func (sdkOwnedConnector) Metadata() ConnectorMetadata {
	metadata := fakeConnector{}.Metadata()
	metadata.Capabilities.RuntimeOwnership = RuntimeOwnershipSDKOwned
	return metadata
}

func TestValidateConfigRequiresCommonContracts(t *testing.T) {
	if _, err := validateConfig(Config{Platform: "fake"}); err == nil || !strings.Contains(err.Error(), "MetadataProvider") {
		t.Fatalf("missing metadata error=%v", err)
	}
	connector := sdkOwnedConnector{}
	cfg := validFakeConfig(connector)
	if _, err := validateConfig(cfg); err == nil || !strings.Contains(err.Error(), SDKOwnedRuntimeScenarioPollRecovery) {
		t.Fatalf("missing sdk-owned runtime error=%v", err)
	}
	cases := []SDKOwnedRuntimeCase{{
		Scenario: SDKOwnedRuntimeScenarioPollRecovery,
		Run:      func(context.Context) (map[string]any, error) { return map[string]any{}, nil },
	}, {
		Scenario: SDKOwnedRuntimeScenarioSessionExpired,
		Run:      func(context.Context) (map[string]any, error) { return map[string]any{}, errors.New("expired") },
	}}
	cfg.SDKOwnedRuntimeCases = cases
	if _, err := validateConfig(cfg); err != nil {
		t.Fatalf("valid sdk-owned config error=%v", err)
	}
}

func TestValidateConfigRequiresCapabilityDrivenAdaptersAndCases(t *testing.T) {
	connector := fakeConnector{}
	tests := []struct {
		name   string
		mutate func(*Config)
		want   string
	}{
		{name: "credential schema adapter", mutate: func(cfg *Config) { cfg.CredentialSchemaProvider = nil }, want: "CredentialSchemaProvider"},
		{name: "credential validator adapter", mutate: func(cfg *Config) { cfg.CredentialValidator = nil }, want: "CredentialValidator"},
		{name: "valid credential case", mutate: func(cfg *Config) {
			cfg.CredentialCases = []CredentialValidationCase{{Expect: CredentialValidationExpectation{Valid: false}}}
		}, want: "valid CredentialCase"},
		{name: "login poller adapter", mutate: func(cfg *Config) { cfg.LoginPoller = nil }, want: "LoginPoller"},
		{name: "approved login poll case", mutate: func(cfg *Config) {
			cfg.LoginPollCases = []LoginPollCase{{Expect: LoginPollExpectation{Approved: false}}}
		}, want: "approved LoginPollCase"},
		{name: "inbound parser adapter", mutate: func(cfg *Config) { cfg.InboundParser = nil }, want: "InboundParser"},
		{name: "delivered inbound case", mutate: func(cfg *Config) {
			cfg.InboundCases = []InboundCase{{Expect: InboundExpectation{ExpectNoMessages: true}}}
		}, want: "returns a message"},
		{name: "acknowledger adapter", mutate: func(cfg *Config) { cfg.Acknowledger = nil }, want: "Acknowledger"},
		{name: "ack cases", mutate: func(cfg *Config) { cfg.AckCases = nil }, want: "AckCases"},
		{name: "successful send case", mutate: func(cfg *Config) { cfg.SendCases = []SendCase{{Expect: SendExpectation{RequireError: true}}} }, want: "successful SendCase"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cfg := validFakeConfig(connector)
			tc.mutate(&cfg)
			if _, err := validateConfig(cfg); err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error=%v, want %q", err, tc.want)
			}
		})
	}
}

func validFakeConfig(connector MetadataProvider) Config {
	fake := fakeConnector{}
	return Config{
		Platform:                 "fake",
		MetadataProvider:         connector,
		CredentialSchemaProvider: fake,
		CredentialValidator:      fake,
		LoginPoller:              fake,
		InboundParser:            fake,
		Sender:                   fake,
		Acknowledger:             fake,
		CredentialCases: []CredentialValidationCase{{
			Expect: CredentialValidationExpectation{Valid: true},
		}},
		LoginPollCases: []LoginPollCase{{
			Expect: LoginPollExpectation{Approved: true},
		}},
		InboundCases: []InboundCase{{
			Expect: InboundExpectation{},
		}},
		SendCases: []SendCase{{
			Expect: SendExpectation{},
		}},
		AckCases: []AckCase{{
			Expect: AckExpectation{Status: "sent"},
		}},
	}
}

type splitPlatformConnector struct{}

func (splitPlatformConnector) Metadata() ConnectorMetadata {
	return ConnectorMetadata{
		ID:       "split",
		Platform: "sdk-platform",
		Label:    "Split",
		Capabilities: Capabilities{
			LoginModes:          []string{LoginModeCredential},
			Text:                true,
			GroupChat:           true,
			Webhook:             true,
			WebhookRegistration: WebhookRegistrationManual,
		},
	}
}

func (splitPlatformConnector) CredentialSchema(context.Context) CredentialSchema {
	return CredentialSchema{Type: "object", LoginModes: []string{LoginModeCredential}, Properties: map[string]CredentialField{"token": {Type: "string", Secret: true}}, Required: []string{"token"}}
}

func (splitPlatformConnector) ValidateCredential(context.Context, CredentialValidationRequest) (*CredentialValidationResult, error) {
	return &CredentialValidationResult{Valid: true, AccountKey: "account-1", Credential: map[string]any{"account_id": "account-1"}}, nil
}

func (splitPlatformConnector) ParseInbound(context.Context, InboundFixture) ([]InboundMessage, error) {
	return []InboundMessage{{
		Platform:  "runtime-platform",
		ChatType:  ChatTypeGroup,
		ChatID:    "chat-1",
		SenderID:  "user-1",
		MessageID: "message-1",
		Text:      "hello",
	}}, nil
}

func (splitPlatformConnector) Acknowledge(context.Context, OutboundAck) (*AckResult, error) {
	return &AckResult{
		Platform:    "runtime-platform",
		AccountUUID: "account-1",
		Status:      "sent",
	}, nil
}

func (splitPlatformConnector) Send(_ context.Context, req OutboundMessage) (*SendResult, error) {
	return &SendResult{Platform: "runtime-platform", AccountUUID: req.AccountUUID, MessageID: "message-2"}, nil
}

func TestRunSupportsMetadataPlatformSeparateFromRuntimePlatform(t *testing.T) {
	connector := splitPlatformConnector{}
	Run(t, Config{
		Platform:                 "runtime-platform",
		MetadataPlatform:         "sdk-platform",
		MetadataProvider:         connector,
		CredentialSchemaProvider: connector,
		CredentialValidator:      connector,
		InboundParser:            connector,
		Sender:                   connector,
		Acknowledger:             connector,
		CredentialCases: []CredentialValidationCase{{
			Expect: CredentialValidationExpectation{Valid: true, AccountKey: "account-1"},
		}},
		InboundCases: []InboundCase{{
			Name:    "runtime inbound platform",
			Fixture: InboundFixture{},
			Expect:  InboundExpectation{ChatID: "chat-1", SenderID: "user-1", Text: "hello", RequireMessageID: true},
		}},
		SendCases: []SendCase{{
			Name:    "runtime send platform",
			Request: OutboundMessage{AccountUUID: "account-1", ChatType: ChatTypeGroup, ChatID: "chat-1", Text: "hello"},
			Expect:  SendExpectation{RequireMessageID: true},
		}},
		AckCases: []AckCase{{
			Name:    "runtime ack platform",
			Request: OutboundAck{AccountUUID: "account-1", ChatType: ChatTypeGroup, ChatID: "chat-1"},
			Expect:  AckExpectation{Status: "sent"},
		}},
	})
}
