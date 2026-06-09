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
		Platform:    "fake",
		ChatType:    ChatTypeGroup,
		ChatID:      "chat-1",
		SenderID:    "user-1",
		MessageID:   "msg-1",
		Text:        "",
		MentionedMe: true,
		Mentions:    []MentionIdentity{{ID: "bot-open-id", IDType: "open_id"}},
	}}, nil
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
				TextTrimmedEmpty: &trueValue,
				MentionedMe:      &trueValue,
				MentionIDs:       []string{"bot-open-id"},
				RequireMessageID: true,
			},
		}},
	})
}
