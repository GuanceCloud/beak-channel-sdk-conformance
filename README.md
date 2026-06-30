# Beak Channel SDK Conformance

`beak-channel-sdk-conformance` is a small Go test helper for channel SDK repositories. It lets
SDK teams verify that their connector output matches Beak's shared IM contract
before publishing a SDK release.

The package is intentionally independent from Beak server internals and from
platform SDKs. SDK tests should provide a thin adapter that converts platform SDK
types into the conformance package types.

## Usage

Keep conformance tests in a separate test module so the publishable connector
module does not expose a test-only dependency to downstream Beak hosts. In this
repository, `beak-channel-sdk-conformance-tests` imports the real connector modules and this
sibling helper through local `replace` directives.

Add or update a SDK conformance test:

```go
func TestBeakSDKConformance(t *testing.T) {
    adapter := newLarkConformanceAdapter()
    conformance.Run(t, conformance.Config{
        Platform:                 "lark",
        MetadataProvider:         adapter,
        CredentialSchemaProvider: adapter,
        CredentialValidator:      adapter,
        LoginPoller:              adapter,
        InboundParser:            adapter,
        CredentialCases: conformance.MustLoadJSON[[]conformance.CredentialValidationCase](
            t,
            "testdata/beak-conformance/credential_cases.json",
        ),
        LoginPollCases: conformance.MustLoadJSON[[]conformance.LoginPollCase](
            t,
            "testdata/beak-conformance/login_poll_cases.json",
        ),
        InboundCases: conformance.MustLoadJSON[[]conformance.InboundCase](
            t,
            "testdata/beak-conformance/inbound_cases.json",
        ),
    })
}
```

`Platform` is the runtime channel/account platform expected in Beak-facing
messages, acks, and stream frames. `MetadataPlatform` is the SDK package identity
reported by `Metadata()`, and defaults to `Platform`. Set both when one SDK
package serves multiple runtime brands:

```go
conformance.Run(t, conformance.Config{
    Platform:         "feishu",
    MetadataPlatform: "lark",
    MetadataProvider: adapter,
    InboundParser:    adapter,
    Acknowledger:     adapter,
    HostStreamer:     adapter,
})
```

The adapter should only convert types. It must not add business logic that the
real SDK connector does not have.

For SDKs developed outside this repository, use a Go workspace or a separate
test module that requires `beak-channel-sdk-conformance`. Do not add `beak-channel-sdk-conformance` as a
normal dependency of the publishable connector module.

## First-pass Gate

The first conformance gate focuses on problems that have caused Beak integration
regressions:

- `ValidateCredential` returns a stable `account_key`.
- normalized `credential.account_id` matches `account_key` when required.
- volatile values such as access tokens, QR challenges, webhook URLs, event IDs,
  message IDs, cursors, offsets, and expiring tokens are not used as account keys.
- valid credential and approved QR/OAuth login results expose standard
  `bot_identity` or `bot_identities` in state.
- approved `PollLogin` results expose the same stable account identity in
  credential/account credential.
- inbound messages include standard `chat_type`, `chat_id`, `sender_id`, and
  `message_id` or `dedupe_key`.
- inbound display fields use the shared contract: `chat_display_name`,
  `chat_avatar_url`, `chat_identity.display_name`, `chat_identity.avatar_url`,
  and `sender_display_name` must be consistent when a platform provides them.
- `mention_all=true` is not treated as the only signal for `mentioned_me=true`.
- follow-up messages that only mention the bot are not dropped and set
  `mentioned_me=true`.
- runtime health uses the standard account state keys:
  `stream_connection_state`, `stream_connected_at`,
  `stream_disconnected_at`, `stream_last_activity_at`,
  `stream_last_ping_at`, `stream_last_pong_at`, `stream_last_event_at`,
  `stream_last_error`, `stream_last_error_at`,
  `stream_reconnect_requested_at`, `stream_reconnect_error`,
  `stream_reconnect_error_at`, and `stream_session_expired`.
- host-owned stream SDKs do not create their own main reconnect loop; SDK-owned
  polling SDKs write `connected`, `reconnecting`, `reconnect_failed`, `stopped`,
  or `expired` through the standard `stream_connection_state` key.

## Suggested Fixture Layout

Each SDK owns platform-specific fixtures:

```text
testdata/beak-conformance/
  credential_cases.json
  login_poll_cases.json
  inbound_cases.json
```

Example `credential_cases.json`:

```json
[
  {
    "name": "valid credential",
    "request": {
      "workspace_uuid": "wksp_test",
      "channel_uuid": "chan_test",
      "credential": {
        "account_id": "stable-account",
        "client_secret": "secret"
      }
    },
  "expect": {
    "valid": true,
    "account_key": "stable-account",
    "metadata_platform": "feishu",
    "require_account_id": true,
    "require_bot_identity": true
  }
  }
]
```

Run:

```bash
cd beak-channel-sdk-conformance-tests && go test ./...
```

Connector modules should keep `beak-channel-sdk-conformance` in a separate test module. Do
not add this helper as a normal dependency of a publishable connector module,
otherwise downstream Beak integrations may inherit a test-only module
requirement.

Runtime-facing outputs must always set `platform` to `Config.Platform`.
Conformance fails inbound messages, acknowledgements, and host-stream inbound
event results that omit it or return a different value. For credential
validation, set `expect.metadata_platform` when the SDK should expose the
runtime platform in `CredentialValidationResult.Metadata["platform"]`.
