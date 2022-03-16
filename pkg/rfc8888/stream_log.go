package rfc8888

import (
	"time"

	"github.com/pion/rtcp"
)

type streamLog struct {
	sequence           unwrapper
	init               bool
	nextSequenceNumber int64 // next to report
	lastSequenceNumber int64 // highest received
	log                map[int64]*packetReport
}

func newStreamLog() *streamLog {
	return &streamLog{
		sequence:           unwrapper{},
		init:               false,
		nextSequenceNumber: 0,
		lastSequenceNumber: 0,
		log:                map[int64]*packetReport{},
	}
}

func (l *streamLog) add(ts time.Time, sequenceNumber uint16, ecn uint8) {
	unwrappedSequenceNumber := l.sequence.unwrap(sequenceNumber)
	if !l.init {
		l.init = true
		l.nextSequenceNumber = unwrappedSequenceNumber
	}
	l.log[unwrappedSequenceNumber] = &packetReport{
		arrivalTime: ts,
		ecn:         ecn,
	}
	if l.lastSequenceNumber < unwrappedSequenceNumber {
		l.lastSequenceNumber = unwrappedSequenceNumber
	}
}

// metricsAfter iterates over all packets order of their sequence number.
// Packets are removed until the first packet with arrivalTime after t is found.
// All following packets are then converted into a list of
// CCFeedbackMetricBlocks.
func (l *streamLog) metricsAfter(deadline, reference time.Time) []rtcp.CCFeedbackMetricBlock {

	for i := l.nextSequenceNumber; i < l.lastSequenceNumber; i++ {
		if r, ok := l.log[i]; ok {
			if r.arrivalTime.Before(deadline) {
				delete(l.log, i)
				l.nextSequenceNumber = i + 1
			}
			if r.arrivalTime.After(deadline) {
				break
			}
		}
	}

	metricBlocks := make([]rtcp.CCFeedbackMetricBlock, len(l.log))
	for i := l.nextSequenceNumber; i < l.nextSequenceNumber+int64(len(l.log)); i++ {
		received := false
		ecn := uint8(0)
		ato := uint16(0)
		if report, ok := l.log[i]; ok {
			received = true
			ecn = report.ecn
			ato = getArrivalTimeOffset(reference, report.arrivalTime)
		}
		metricBlocks[i-l.nextSequenceNumber] = rtcp.CCFeedbackMetricBlock{
			Received:          received,
			ECN:               rtcp.ECN(ecn),
			ArrivalTimeOffset: ato,
		}
	}
	return metricBlocks
}

func getArrivalTimeOffset(base time.Time, arrival time.Time) uint16 {
	if base.Before(arrival) {
		return 0x1FFF
	}
	ato := uint16(base.Sub(arrival).Seconds() * 1024.0)
	if ato > 0x1FFD {
		return 0x1FFE
	}
	return ato
}
