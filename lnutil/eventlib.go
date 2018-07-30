package lnutil

// EventType is an enumeration containing the various events that
// can happen, that are forwarded to the RPC surface
type LitEventType int

const (
	LitEventTypeChannelPushReceived LitEventType = 0
)

type LitEvent struct {
	Type LitEventType
}
