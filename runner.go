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
	metadataPlatform := firstString(cfg.MetadataPlatform, cfg.Platform)

	var metadata ConnectorMetadata
	if cfg.MetadataProvider != nil {
		metadata = cfg.MetadataProvider.Metadata()
		t.Run("metadata", func(t *testing.T) {
			AssertMetadata(t, metadataPlatform, metadata)
		})
	}

	if metadata.Capabilities.RuntimeOwnership == RuntimeOwnershipHostStream {
		t.Run("host_stream_contract", func(t *testing.T) {
			if cfg.HostStreamer == nil {
				t.Fatal("host_stream connectors must expose a HostStreamConnector adapter")
			}
			if len(cfg.HostStreamCases) == 0 {
				t.Fatal("host_stream connectors must define HostStreamCases")
			}
			for _, tc := range cfg.HostStreamCases {
				tc := tc
				t.Run(caseName(tc.Name, "host_stream"), func(t *testing.T) {
					req := tc.Request
					fillHostStreamConnectDefaults(&req, cfg.Platform)

					connected, err := cfg.HostStreamer.ConnectStream(ctx, req)
					AssertStreamConnectResult(t, connected, err, tc.Expect)

					serviceID := ""
					var state any
					if connected != nil {
						serviceID = connected.ServiceID
						state = connected.State
					}

					if tc.Ping != nil {
						pingReq := tc.Ping.Request
						if pingReq.ServiceID == "" {
							pingReq.ServiceID = serviceID
						}
						if pingReq.State == nil {
							pingReq.State = state
						}
						frame, err := cfg.HostStreamer.BuildStreamPing(ctx, pingReq)
						AssertStreamPingFrame(t, frame, err, tc.Ping.Expect)
					}

					for _, frameCase := range tc.Frames {
						frameCase := frameCase
						t.Run(caseName(frameCase.Name, "frame"), func(t *testing.T) {
							frameReq := frameCase.Request
							if frameReq.ServiceID == "" {
								frameReq.ServiceID = serviceID
							}
							if frameReq.State == nil {
								frameReq.State = state
							}
							fillHostStreamFrameDefaults(&frameReq, req, cfg.Platform)
							result, err := cfg.HostStreamer.HandleStreamFrame(ctx, frameReq)
							AssertStreamFrameResult(t, cfg.Platform, result, err, frameCase.Expect)
							if result != nil {
								state = result.State
							}
						})
					}
				})
			}
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

	if len(cfg.AckCases) > 0 {
		if cfg.Acknowledger == nil {
			t.Fatal("AckCases require Acknowledger")
		}
		for _, tc := range cfg.AckCases {
			tc := tc
			t.Run(caseName(tc.Name, "ack"), func(t *testing.T) {
				req := tc.Request
				if req.Platform == "" {
					req.Platform = cfg.Platform
				}
				got, err := cfg.Acknowledger.Acknowledge(ctx, req)
				AssertAckResult(t, cfg.Platform, got, err, tc.Expect)
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

func fillHostStreamConnectDefaults(req *HostStreamConnectRequest, platform string) {
	if req.WorkspaceUUID == "" {
		req.WorkspaceUUID = "workspace-1"
	}
	if req.ChannelUUID == "" {
		req.ChannelUUID = "channel-1"
	}
	if req.Account.UUID == "" {
		req.Account.UUID = "account-1"
	}
	if req.Account.WorkspaceUUID == "" {
		req.Account.WorkspaceUUID = req.WorkspaceUUID
	}
	if req.Account.ChannelUUID == "" {
		req.Account.ChannelUUID = req.ChannelUUID
	}
	if req.Account.Platform == "" {
		req.Account.Platform = platform
	}
	if req.Account.Credential == nil {
		req.Account.Credential = req.Credential
	}
	if req.Account.State == nil {
		req.Account.State = req.State
	}
}

func fillHostStreamFrameDefaults(frame *StreamFrameRequest, connect HostStreamConnectRequest, platform string) {
	if frame.WorkspaceUUID == "" {
		frame.WorkspaceUUID = connect.WorkspaceUUID
	}
	if frame.ChannelUUID == "" {
		frame.ChannelUUID = connect.ChannelUUID
	}
	if frame.Account.UUID == "" {
		frame.Account = connect.Account
	}
	if frame.Account.WorkspaceUUID == "" {
		frame.Account.WorkspaceUUID = frame.WorkspaceUUID
	}
	if frame.Account.ChannelUUID == "" {
		frame.Account.ChannelUUID = frame.ChannelUUID
	}
	if frame.Account.Platform == "" {
		frame.Account.Platform = platform
	}
	if frame.Account.Credential == nil {
		if frame.Credential != nil {
			frame.Account.Credential = frame.Credential
		} else {
			frame.Account.Credential = connect.Account.Credential
		}
	}
	if frame.Account.State == nil {
		frame.Account.State = connect.Account.State
	}
}
