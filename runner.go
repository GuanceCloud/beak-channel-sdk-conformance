package conformance

import (
	"context"
	"testing"
)

func Run(t *testing.T, cfg Config) {
	t.Helper()
	ctx := context.Background()

	if cfg.Platform == "" {
		t.Fatal("Platform is required")
	}

	if cfg.MetadataProvider != nil {
		t.Run("metadata", func(t *testing.T) {
			AssertMetadata(t, cfg.Platform, cfg.MetadataProvider.Metadata())
		})
	}

	if cfg.CredentialSchemaProvider != nil {
		t.Run("credential_schema", func(t *testing.T) {
			AssertCredentialSchema(t, cfg.CredentialSchemaProvider.CredentialSchema(ctx))
		})
	}

	if len(cfg.CredentialCases) > 0 {
		if cfg.CredentialValidator == nil {
			t.Fatal("CredentialCases require CredentialValidator")
		}
		for _, tc := range cfg.CredentialCases {
			tc := tc
			t.Run(caseName(tc.Name, "credential"), func(t *testing.T) {
				req := tc.Request
				if req.Platform == "" {
					req.Platform = cfg.Platform
				}
				got, err := cfg.CredentialValidator.ValidateCredential(ctx, req)
				AssertCredentialValidationResult(t, req, got, err, tc.Expect)
			})
		}
	}

	if len(cfg.LoginPollCases) > 0 {
		if cfg.LoginPoller == nil {
			t.Fatal("LoginPollCases require LoginPoller")
		}
		for _, tc := range cfg.LoginPollCases {
			tc := tc
			t.Run(caseName(tc.Name, "login_poll"), func(t *testing.T) {
				got, err := cfg.LoginPoller.PollLogin(ctx, tc.Request)
				AssertLoginPollResult(t, got, err, tc.Expect)
			})
		}
	}

	if len(cfg.InboundCases) > 0 {
		if cfg.InboundParser == nil {
			t.Fatal("InboundCases require InboundParser")
		}
		for _, tc := range cfg.InboundCases {
			tc := tc
			t.Run(caseName(tc.Name, "inbound"), func(t *testing.T) {
				fixture := tc.Fixture
				if fixture.Platform == "" {
					fixture.Platform = cfg.Platform
				}
				got, err := cfg.InboundParser.ParseInbound(ctx, fixture)
				AssertInboundMessages(t, cfg.Platform, got, err, tc.Expect)
			})
		}
	}
}

func caseName(name, fallback string) string {
	if name != "" {
		return name
	}
	return fallback
}
