package ackhandler

import (
	"github.com/lucas-clemente/quic-go/internal/congestion"
	"github.com/lucas-clemente/quic-go/internal/protocol"
	"github.com/lucas-clemente/quic-go/internal/utils"
	"github.com/lucas-clemente/quic-go/logging"
)

// NewAckHandler creates a new SentPacketHandler and a new ReceivedPacketHandler
func NewAckHandler(
	initialPacketNumber protocol.PacketNumber,
	initialMaxDatagramSize protocol.ByteCount,
	rttStats *utils.RTTStats,
	pers protocol.Perspective,
	tracer logging.ConnectionTracer,
	logger utils.Logger,
	version protocol.VersionNumber,
	congestion congestion.CongestionOptions,
) (SentPacketHandler, ReceivedPacketHandler) {
	sph := newSentPacketHandler(initialPacketNumber, initialMaxDatagramSize, rttStats, pers, congestion, tracer, logger)
	return sph, newReceivedPacketHandler(sph, rttStats, logger, version)
}
