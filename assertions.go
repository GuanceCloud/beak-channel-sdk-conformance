package conformance

import (
	"fmt"
	"strings"
	"testing"
)

func AssertMetadata(t *testing.T, platform string, got ConnectorMetadata) {
	t.Helper()
	if strings.TrimSpace(got.ID) == "" {
		t.Fatal("metadata.id is required")
	}
	if strings.TrimSpace(got.Platform) != strings.TrimSpace(platform) {
		t.Fatalf("metadata.platform = %q, want %q", got.Platform, platform)
	}
	if strings.TrimSpace(got.Label) == "" {
		t.Fatal("metadata.label is required")
	}
	if len(got.Capabilities.LoginModes) == 0 {
		t.Fatal("metadata.capabilities.login_modes must not be empty")
	}
	if got.Capabilities.Stream && got.Capabilities.Webhook {
		t.Fatal("metadata capabilities must not declare both stream and webhook")
	}
}

func AssertCredentialSchema(t *testing.T, got CredentialSchema) {
	t.Helper()
	if strings.TrimSpace(got.Type) == "" {
		t.Fatal("credential schema type is required")
	}
	if len(got.LoginModes) == 0 {
		t.Fatal("credential schema login_modes must not be empty")
	}
	if containsString(got.LoginModes, LoginModeCredential) && len(got.Properties) == 0 {
		t.Fatal("credential schema properties must not be empty for credential login")
	}
	for _, required := range got.Required {
		if _, ok := got.Properties[required]; !ok {
			t.Fatalf("credential schema required field %q is not defined in properties", required)
		}
	}
}

func AssertCredentialValidationResult(t *testing.T, req CredentialValidationRequest, got *CredentialValidationResult, err error, expect CredentialValidationExpectation) {
	t.Helper()
	if err != nil {
		t.Fatalf("ValidateCredential returned error: %v", err)
	}
	if got == nil {
		t.Fatal("ValidateCredential returned nil result")
	}
	if got.Valid != expect.Valid {
		t.Fatalf("ValidateCredential valid = %v, want %v; result=%+v", got.Valid, expect.Valid, got)
	}
	if !expect.Valid {
		if strings.TrimSpace(got.Error) == "" {
			t.Fatal("invalid credential result must include error")
		}
		return
	}

	accountKey := strings.TrimSpace(got.AccountKey)
	if accountKey == "" {
		t.Fatal("valid credential result must include account_key")
	}
	if expect.AccountKey != "" && accountKey != expect.AccountKey {
		t.Fatalf("account_key = %q, want %q", accountKey, expect.AccountKey)
	}
	if expect.DisplayName != "" && got.DisplayName != expect.DisplayName {
		t.Fatalf("display_name = %q, want %q", got.DisplayName, expect.DisplayName)
	}
	if expect.RequireAccountID {
		assertAccountID(t, "credential result", accountKey, got.Credential)
	}
	if expect.RequireBotIdentity {
		if !hasBotIdentity(got.State) {
			t.Fatalf("credential result state must include bot_identity or bot_identities: %+v", got.State)
		}
	}

	invalidValues := volatileValues(expect.VolatileKeys, req.Credential, got.Credential, got.State)
	for label, value := range invalidValues {
		if accountKey == value {
			t.Fatalf("account_key must be stable, got volatile %s value %q", label, value)
		}
	}
}

func AssertLoginPollResult(t *testing.T, got *LoginStatus, err error, expect LoginPollExpectation) {
	t.Helper()
	if err != nil {
		t.Fatalf("PollLogin returned error: %v", err)
	}
	if got == nil {
		t.Fatal("PollLogin returned nil result")
	}
	approved := got.Confirmed || strings.EqualFold(got.Status, LoginStatusApproved)
	if approved != expect.Approved {
		t.Fatalf("PollLogin approved = %v, want %v; result=%+v", approved, expect.Approved, got)
	}
	if !expect.Approved {
		return
	}

	accountKey := accountKeyFromLoginStatus(got)
	if accountKey == "" {
		t.Fatalf("approved PollLogin result must include credential.account_id or account credential account_id: %+v", got)
	}
	if expect.AccountKey != "" && accountKey != expect.AccountKey {
		t.Fatalf("PollLogin account key = %q, want %q", accountKey, expect.AccountKey)
	}
	if expect.DisplayName != "" && got.Account.DisplayName != expect.DisplayName {
		t.Fatalf("PollLogin display_name = %q, want %q", got.Account.DisplayName, expect.DisplayName)
	}
	if expect.RequireAccountID {
		assertAccountID(t, "login poll account credential", accountKey, got.Account.Credential)
		if len(got.Credential) > 0 {
			assertAccountID(t, "login poll credential", accountKey, got.Credential)
		}
	}
	if expect.RequireBotIdentity {
		state := mergeMap(got.State, got.Account.State)
		if !hasBotIdentity(state) {
			t.Fatalf("approved PollLogin result must include bot_identity or bot_identities in state: %+v", state)
		}
	}
}

func AssertInboundMessages(t *testing.T, platform string, got []InboundMessage, err error, expect InboundExpectation) {
	t.Helper()
	if err != nil {
		t.Fatalf("ParseInbound returned error: %v", err)
	}
	minMessages := expect.MinMessages
	if minMessages == 0 {
		minMessages = 1
	}
	if len(got) < minMessages {
		t.Fatalf("ParseInbound returned %d messages, want at least %d", len(got), minMessages)
	}
	index := expect.MessageIndex
	if index < 0 || index >= len(got) {
		t.Fatalf("message_index %d out of range for %d messages", index, len(got))
	}
	msg := got[index]
	if strings.TrimSpace(msg.Platform) != "" && strings.TrimSpace(msg.Platform) != platform {
		t.Fatalf("inbound platform = %q, want %q", msg.Platform, platform)
	}
	if strings.TrimSpace(msg.ChatType) == "" {
		t.Fatal("inbound chat_type is required")
	}
	if msg.ChatType != ChatTypeGroup && msg.ChatType != ChatTypeDirect {
		t.Fatalf("inbound chat_type = %q, want %q or %q", msg.ChatType, ChatTypeGroup, ChatTypeDirect)
	}
	if strings.TrimSpace(msg.ChatID) == "" {
		t.Fatal("inbound chat_id is required")
	}
	if strings.TrimSpace(msg.SenderID) == "" {
		t.Fatal("inbound sender_id is required")
	}
	if strings.TrimSpace(msg.MessageID) == "" && strings.TrimSpace(msg.DedupeKey) == "" {
		t.Fatal("inbound message_id or dedupe_key is required")
	}
	if expect.ChatType != "" && msg.ChatType != expect.ChatType {
		t.Fatalf("inbound chat_type = %q, want %q", msg.ChatType, expect.ChatType)
	}
	if expect.ChatID != "" && msg.ChatID != expect.ChatID {
		t.Fatalf("inbound chat_id = %q, want %q", msg.ChatID, expect.ChatID)
	}
	if expect.ChatDisplayName != "" {
		if got := firstString(msg.ChatIdentity.DisplayName, msg.ChatDisplayName); got != expect.ChatDisplayName {
			t.Fatalf("inbound chat display name = %q, want %q", got, expect.ChatDisplayName)
		}
	}
	if expect.ChatIdentityID != "" && msg.ChatIdentity.ID != expect.ChatIdentityID {
		t.Fatalf("inbound chat_identity.id = %q, want %q", msg.ChatIdentity.ID, expect.ChatIdentityID)
	}
	if msg.ChatDisplayName != "" && msg.ChatIdentity.DisplayName != "" && msg.ChatDisplayName != msg.ChatIdentity.DisplayName {
		t.Fatalf("inbound chat_display_name = %q but chat_identity.display_name = %q", msg.ChatDisplayName, msg.ChatIdentity.DisplayName)
	}
	if expect.SenderID != "" && msg.SenderID != expect.SenderID {
		t.Fatalf("inbound sender_id = %q, want %q", msg.SenderID, expect.SenderID)
	}
	if expect.Text != "" && msg.Text != expect.Text {
		t.Fatalf("inbound text = %q, want %q", msg.Text, expect.Text)
	}
	if expect.TextTrimmedEmpty != nil && (strings.TrimSpace(msg.Text) == "") != *expect.TextTrimmedEmpty {
		t.Fatalf("inbound text trimmed empty = %v, want %v; text=%q", strings.TrimSpace(msg.Text) == "", *expect.TextTrimmedEmpty, msg.Text)
	}
	if expect.MentionedMe != nil && msg.MentionedMe != *expect.MentionedMe {
		t.Fatalf("inbound mentioned_me = %v, want %v", msg.MentionedMe, *expect.MentionedMe)
	}
	if expect.MentionAll != nil && msg.MentionAll != *expect.MentionAll {
		t.Fatalf("inbound mention_all = %v, want %v", msg.MentionAll, *expect.MentionAll)
	}
	for _, id := range expect.MentionIDs {
		if !containsMentionID(msg.Mentions, id) {
			t.Fatalf("inbound mentions missing id %q: %+v", id, msg.Mentions)
		}
	}
	if expect.RequireMessageID && strings.TrimSpace(msg.MessageID) == "" {
		t.Fatal("inbound message_id is required by expectation")
	}
	if expect.RequireDedupeKey && strings.TrimSpace(msg.DedupeKey) == "" {
		t.Fatal("inbound dedupe_key is required by expectation")
	}
	if msg.MentionAll && msg.MentionedMe && len(msg.Mentions) == 0 {
		t.Fatal("mention_all must not be the only signal used to set mentioned_me")
	}
}

func assertAccountID(t *testing.T, label, accountKey string, credential map[string]any) {
	t.Helper()
	accountID := strings.TrimSpace(stringValue(credential["account_id"]))
	if accountID == "" {
		t.Fatalf("%s must include credential.account_id", label)
	}
	if accountID != accountKey {
		t.Fatalf("%s credential.account_id = %q, want account_key %q", label, accountID, accountKey)
	}
}

func accountKeyFromLoginStatus(status *LoginStatus) string {
	for _, source := range []map[string]any{status.Credential, status.Account.Credential} {
		if value := strings.TrimSpace(stringValue(source["account_id"])); value != "" {
			return value
		}
		if value := strings.TrimSpace(stringValue(source["account_key"])); value != "" {
			return value
		}
	}
	return ""
}

func hasBotIdentity(state map[string]any) bool {
	if identity, ok := state["bot_identity"].(map[string]any); ok && strings.TrimSpace(stringValue(identity["id"])) != "" {
		return true
	}
	if identities, ok := state["bot_identities"].([]map[string]any); ok {
		for _, identity := range identities {
			if strings.TrimSpace(stringValue(identity["id"])) != "" {
				return true
			}
		}
	}
	if identities, ok := state["bot_identities"].([]any); ok {
		for _, item := range identities {
			identity, _ := item.(map[string]any)
			if strings.TrimSpace(stringValue(identity["id"])) != "" {
				return true
			}
		}
	}
	return false
}

func volatileValues(extraKeys []string, maps ...map[string]any) map[string]string {
	out := map[string]string{}
	for _, values := range maps {
		for key, value := range values {
			normalized := strings.ToLower(strings.TrimSpace(key))
			if isVolatileField(normalized) || containsString(extraKeys, key) {
				if text := strings.TrimSpace(stringValue(value)); text != "" {
					out[key] = text
				}
			}
		}
	}
	return out
}

func containsString(values []string, needle string) bool {
	for _, value := range values {
		if strings.EqualFold(strings.TrimSpace(value), strings.TrimSpace(needle)) {
			return true
		}
	}
	return false
}

func isVolatileField(key string) bool {
	volatileParts := []string{
		"access_token",
		"refresh_token",
		"secret",
		"token",
		"challenge",
		"qrcode",
		"qr_code",
		"message_id",
		"event_id",
		"webhook",
		"cursor",
		"offset",
		"temporary",
		"expires",
		"expire",
	}
	for _, part := range volatileParts {
		if strings.Contains(key, part) {
			return true
		}
	}
	return false
}

func stringValue(value any) string {
	switch typed := value.(type) {
	case string:
		return typed
	case fmt.Stringer:
		return typed.String()
	case nil:
		return ""
	default:
		return fmt.Sprint(typed)
	}
}

func firstString(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func containsMentionID(mentions []MentionIdentity, id string) bool {
	for _, mention := range mentions {
		if mention.ID == id {
			return true
		}
	}
	return false
}

func mergeMap(primary, fallback map[string]any) map[string]any {
	out := map[string]any{}
	for key, value := range fallback {
		out[key] = value
	}
	for key, value := range primary {
		out[key] = value
	}
	return out
}
