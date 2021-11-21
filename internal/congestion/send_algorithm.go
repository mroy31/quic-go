package congestion

import (
	"github.com/lucas-clemente/quic-go/internal/protocol"
	"github.com/lucas-clemente/quic-go/internal/utils"
	"github.com/lucas-clemente/quic-go/logging"
)

func hystartTypeToString(t HystartControlType) string {
	switch t {
	case HystartTypeStandard:
		return "standard"
	case HystartTypePlusPlus:
		return "++"
	case HystartTypeNone:
		return "none"
	}
	return ""
}

func congestionTypeToString(t CongestionControlType) string {
	switch t {
	case CubicControlType:
		return "Cubic"
	case NewRenoControlType:
		return "NewReno"
	case BbrControlType:
		return "Bbr"
	}
	return ""
}

func NewCongestionHandler(
	rttStats *utils.RTTStats,
	initialMaxDatagramSize protocol.ByteCount,
	options CongestionOptions,
	tracer logging.ConnectionTracer,
) SendAlgorithmWithDebugInfos {
	logger := utils.DefaultLogger

	switch options.ControlType {
	case NewRenoControlType:
		logger.Infof("Congestion Control: NewReno with hystart: %s", hystartTypeToString(options.Hystart))
		return NewCubicSender(
			DefaultClock{},
			rttStats,
			initialMaxDatagramSize,
			true, // use Reno
			options.Hystart,
			tracer,
		)
	default:
		logger.Infof("Congestion Control: Cubic with hystart: %s", hystartTypeToString(options.Hystart))
		return NewCubicSender(
			DefaultClock{},
			rttStats,
			initialMaxDatagramSize,
			false, // use Cubic
			options.Hystart,
			tracer,
		)
	}
}
