package environment

const (
	// SERVER
	appEnv                     = "APP_ENV"
	HTTPAddress                = "HTTP_ADDRESS"
	HTTPSRedirectPort          = "HTTPS_REDIRECT_PORT"
	HTTPEnableRedirect         = "ENABLE_HTTP_REDIRECT"
	NetworkTestOnStart         = "NETWORK_TEST_ON_START"
	IncludePublicIPInNAT1To1IP = "INCLUDE_PUBLIC_IP_IN_NAT_1_TO_1_IP"
	DisableStatus              = "DISABLE_STATUS"
	EnableProfiling            = "ENABLE_PROFILING"

	// SSL
	SSLKey  = "SSL_KEY"
	SSLCert = "SSL_CERT"

	// AUTHORIZATION
	StreamProfilePath   = "STREAM_PROFILE_PATH"
	StreamProfilePolicy = "STREAM_PROFILE_POLICY"
	WebhookURL          = "WEBHOOK_URL"

	// FRONTEND
	FrontendDisabled   = "DISABLE_FRONTEND"
	frontendPath       = "FRONTEND_PATH"
	FrontendAdminToken = "FRONTEND_ADMIN_TOKEN"

	// CHAT
	ChatMaxHistory      = "CHAT_MAX_HISTORY"
	ChatDefaultTTL      = "CHAT_DEFAULT_TTL"
	ChatCleanupInterval = "CHAT_CLEANUP_INTERVAL"

	// WEBRTC
	IncludeLoopbackCandidate = "INCLUDE_LOOPBACK_CANDIDATE"
	NetworkTypes             = "NETWORK_TYPES"
	TCPMuxForce              = "TCP_MUX_FORCE"
	TCPMuxAddress            = "TCP_MUX_ADDRESS"
	InterfaceFilter          = "INTERFACE_FILTER"
	UDPMuxPort               = "UDP_MUX_PORT"
	UDPMuxPortWHIP           = "UDP_MUX_PORT_WHIP"
	UDPMuxPortWHEP           = "UDP_MUX_PORT_WHEP"
	NAT1To1IP                = "NAT_1_TO_1_IP"
	NATICECandidateType      = "NAT_ICE_CANDIDATE_TYPE"

	// STUN
	STUNServers = "STUN_SERVERS"

	// PEERCONNECTION
	AppendCandidate = "APPEND_CANDIDATE"

	// DEBUGGING
	DebugIncomingAPIRequest = "DEBUG_INCOMING_API_REQUEST"
	DebugPrintAnswer        = "DEBUG_PRINT_ANSWER"
	DebugPrintOffer         = "DEBUG_PRINT_OFFER"
	DebugPrintSSEMessages   = "DEBUG_PRINT_SSE_MESSAGES"

	// LOGGING
	loggingEnabled          = "LOGGING_ENABLED"
	loggingDirectory        = "LOGGING_DIRECTORY"
	loggingSingleFile       = "LOGGING_SINGLEFILE"
	loggingNewFileOnStartup = "LOGGING_NEW_FILE_ON_STARTUP"
	LoggingAPIEnabled       = "LOGGING_API_ENABLED"
	LoggingAPIKey           = "LOGGING_API_KEY"
)
