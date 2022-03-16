package rfc8888

import (
	"time"

	"github.com/pion/rtcp"
)

type packetReport struct {
	arrivalTime time.Time
	ecn         uint8
}

type Recorder struct {
	ssrc    uint32
	streams map[uint32]*streamLog
}

func NewRecorder() *Recorder {
	return &Recorder{
		streams: map[uint32]*streamLog{},
	}
}

// AddPacket writes a packet to the underlying stream.
func (r *Recorder) AddPacket(ts time.Time, ssrc uint32, seq uint16, ecn uint8) {
	stream, ok := r.streams[ssrc]
	if !ok {
		stream = &streamLog{
			sequence: unwrapper{},
			log:      map[int64]*packetReport{},
		}
		r.streams[ssrc] = stream
	}
	stream.add(ts, seq, ecn)
}

// TODO: Return *rtcp.Packet instead?
func (r *Recorder) BuildReport(now time.Time) *rtcp.CCFeedbackReport {
	// TODO: Implement automatic header generation in pion/rtcp?
	rts := time.Now()
	report := &rtcp.CCFeedbackReport{
		Header:          rtcp.Header{},
		SenderSSRC:      r.ssrc,
		ReportBlocks:    []rtcp.CCFeedbackReportBlock{},
		ReportTimestamp: ntpTime32(rts),
	}

	for ssrc, log := range r.streams {
		block := log.metricsAfter(now.Add(-time.Second), rts)
		report.ReportBlocks = append(report.ReportBlocks, rtcp.CCFeedbackReportBlock{
			MediaSSRC:     ssrc,
			BeginSequence: 0,
			MetricBlocks:  block,
		})
	}

	return report
}

// TODO: Is this conversion correct?
func ntpTime32(t time.Time) uint32 {
	// seconds since 1st January 1900
	s := (float64(t.UnixNano()) / 1000000000) + 2208988800

	// higher 32 bits are the integer part, lower 32 bits are the fractional part
	integerPart := uint16(s)
	fractionalPart := uint16((s - float64(integerPart)) * 0xFFFFFFFF)
	return uint32(integerPart)<<16 | uint32(fractionalPart)
}
