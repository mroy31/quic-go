package congestion

import (
	"time"

	"github.com/lucas-clemente/quic-go/internal/protocol"
)

type CongestionControlType int

const (
	NewRenoControlType CongestionControlType = iota
	CubicControlType                         = iota
	BbrControlType                           = iota
)

type HystartControlType int

const (
	HystartTypeStandard HystartControlType = iota
	HystartTypePlusPlus                    = iota // Hystart++
	HystartTypeNone                        = iota
)

type CongestionOptions struct {
	ControlType CongestionControlType
	Hystart     HystartControlType
}

// A SendAlgorithm performs congestion control
type SendAlgorithm interface {
	TimeUntilSend(bytesInFlight protocol.ByteCount) time.Time
	HasPacingBudget() bool
	OnPacketSent(sentTime time.Time, bytesInFlight protocol.ByteCount, packetNumber protocol.PacketNumber, bytes protocol.ByteCount, isRetransmittable bool)
	CanSend(bytesInFlight protocol.ByteCount) bool
	OnRttUpdated()
	OnPacketAcked(number protocol.PacketNumber, ackedBytes protocol.ByteCount, priorInFlight protocol.ByteCount, eventTime time.Time)
	OnPacketLost(number protocol.PacketNumber, lostBytes protocol.ByteCount, priorInFlight protocol.ByteCount)
	OnRetransmissionTimeout(packetsRetransmitted bool)
	SetMaxDatagramSize(protocol.ByteCount)
}

// A SendAlgorithmWithDebugInfos is a SendAlgorithm that exposes some debug infos
type SendAlgorithmWithDebugInfos interface {
	SendAlgorithm
	InSlowStart() bool
	InRecovery() bool
	GetCongestionWindow() protocol.ByteCount
}
