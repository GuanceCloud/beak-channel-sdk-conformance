package conformance

import (
	"context"
	"testing"
)

type fakeConnector struct{}

func (fakeConnector) Metadata() ConnectorMetadata {
	return ConnectorMetadata{
		ID:       "fake",
		Platform: "fake",
		Label:    "Fake",
		Capabilities: Capabilities{
			LoginModes: []string{LoginModeCredential, LoginModeQRCode},
			Text:       true,
			GroupChat:  true,
			Webhook:    true,
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

func (fakeConnector) ParseInbound(context.Context, InboundFixture) ([]InboundMessage, error) {
	return []InboundMessage{{
		Platform:        "fake",
		ChatType:        ChatTypeGroup,
		ChatID:          "chat-1",
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
		MentionedMe:       true,
		Mentions:          []MentionIdentity{{ID: "bot-open-id", IDType: "open_id"}},
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
				Valid:              true,
				AccountKey:         "stable-account",
				DisplayName:        "Stable Bot",
				RequireAccountID:   true,
				RequireBotIdentity: true,
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
				SenderDisplayName: "Alice",
				TextTrimmedEmpty:  &trueValue,
				MentionedMe:       &trueValue,
				MentionIDs:        []string{"bot-open-id"},
				RequireMessageID:  true,
			},
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

type splitPlatformConnector struct{}

func (splitPlatformConnector) Metadata() ConnectorMetadata {
	return ConnectorMetadata{
		ID:       "split",
		Platform: "sdk-platform",
		Label:    "Split",
		Capabilities: Capabilities{
			LoginModes: []string{LoginModeCredential},
			Text:       true,
			GroupChat:  true,
			Webhook:    true,
		},
	}
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

func TestRunSupportsMetadataPlatformSeparateFromRuntimePlatform(t *testing.T) {
	connector := splitPlatformConnector{}
	Run(t, Config{
		Platform:         "runtime-platform",
		MetadataPlatform: "sdk-platform",
		MetadataProvider: connector,
		InboundParser:    connector,
		Acknowledger:     connector,
		InboundCases: []InboundCase{{
			Name:    "runtime inbound platform",
			Fixture: InboundFixture{},
			Expect:  InboundExpectation{ChatID: "chat-1", SenderID: "user-1", Text: "hello", RequireMessageID: true},
		}},
		AckCases: []AckCase{{
			Name:    "runtime ack platform",
			Request: OutboundAck{AccountUUID: "account-1", ChatType: ChatTypeGroup, ChatID: "chat-1"},
			Expect:  AckExpectation{Status: "sent"},
		}},
	})
}
