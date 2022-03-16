package rfc8888

import (
	"testing"
	"time"

	"github.com/pion/rtcp"
	"github.com/stretchr/testify/assert"
)

type input struct {
	ts  time.Time
	nr  uint16
	ecn uint8
}

func TestStreamLogAdd(t *testing.T) {
	tests := []struct {
		name         string
		inputs       []input
		expectedNext int64
		expectedLast int64
		expectedLog  map[int64]*packetReport
	}{
		{
			name:         "emptyLog",
			inputs:       []input{},
			expectedNext: 0,
			expectedLast: 0,
			expectedLog:  map[int64]*packetReport{},
		},
		{
			name: "addInOrderSequence",
			inputs: []input{
				{
					ts:  time.Time{},
					nr:  0,
					ecn: 0,
				},
				{
					ts:  time.Time{}.Add(10 * time.Millisecond),
					nr:  1,
					ecn: 0,
				},
				{
					ts:  time.Time{}.Add(20 * time.Millisecond),
					nr:  2,
					ecn: 0,
				},
				{
					ts:  time.Time{}.Add(30 * time.Millisecond),
					nr:  3,
					ecn: 0,
				},
			},
			expectedNext: 0,
			expectedLast: 3,
			expectedLog: map[int64]*packetReport{
				0: {
					arrivalTime: time.Time{},
					ecn:         0,
				},
				1: {
					arrivalTime: time.Time{}.Add(10 * time.Millisecond),
					ecn:         0,
				},
				2: {
					arrivalTime: time.Time{}.Add(20 * time.Millisecond),
					ecn:         0,
				},
				3: {
					arrivalTime: time.Time{}.Add(30 * time.Millisecond),
					ecn:         0,
				},
			},
		},
		{
			name: "reorderedSequence",
			inputs: []input{
				{
					ts:  time.Time{},
					nr:  0,
					ecn: 0,
				},
				{
					ts:  time.Time{}.Add(10 * time.Millisecond),
					nr:  2,
					ecn: 0,
				},
				{
					ts:  time.Time{}.Add(20 * time.Millisecond),
					nr:  1,
					ecn: 0,
				},
				{
					ts:  time.Time{}.Add(30 * time.Millisecond),
					nr:  3,
					ecn: 0,
				},
			},
			expectedNext: 0,
			expectedLast: 3,
			expectedLog: map[int64]*packetReport{
				0: {
					arrivalTime: time.Time{},
					ecn:         0,
				},
				1: {
					arrivalTime: time.Time{}.Add(20 * time.Millisecond),
					ecn:         0,
				},
				2: {
					arrivalTime: time.Time{}.Add(10 * time.Millisecond),
					ecn:         0,
				},
				3: {
					arrivalTime: time.Time{}.Add(30 * time.Millisecond),
					ecn:         0,
				},
			},
		},
		{
			name: "reorderedWrappingSequence",
			inputs: []input{
				{
					ts:  time.Time{},
					nr:  65534,
					ecn: 0,
				},
				{
					ts:  time.Time{}.Add(10 * time.Millisecond),
					nr:  0,
					ecn: 0,
				},
				{
					ts:  time.Time{}.Add(20 * time.Millisecond),
					nr:  65535,
					ecn: 0,
				},
				{
					ts:  time.Time{}.Add(30 * time.Millisecond),
					nr:  2,
					ecn: 0,
				},
				{
					ts:  time.Time{}.Add(40 * time.Millisecond),
					nr:  1,
					ecn: 0,
				},
				{
					ts:  time.Time{}.Add(50 * time.Millisecond),
					nr:  3,
					ecn: 0,
				},
			},
			expectedNext: 65534,
			expectedLast: 65539,
			expectedLog: map[int64]*packetReport{
				65534: {
					arrivalTime: time.Time{},
					ecn:         0,
				},
				65535: {
					arrivalTime: time.Time{}.Add(20 * time.Millisecond),
					ecn:         0,
				},
				65536: {
					arrivalTime: time.Time{}.Add(10 * time.Millisecond),
					ecn:         0,
				},
				65537: {
					arrivalTime: time.Time{}.Add(40 * time.Millisecond),
					ecn:         0,
				},
				65538: {
					arrivalTime: time.Time{}.Add(30 * time.Millisecond),
					ecn:         0,
				},
				65539: {
					arrivalTime: time.Time{}.Add(50 * time.Millisecond),
					ecn:         0,
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			sl := newStreamLog()
			for _, input := range test.inputs {
				sl.add(input.ts, input.nr, input.ecn)
			}
			assert.Equal(t, test.expectedNext, sl.nextSequenceNumber)
			assert.Equal(t, test.expectedLast, sl.lastSequenceNumber)
			assert.Equal(t, test.expectedLog, sl.log)
		})
	}
}

func TestStreamLogMetricsAfter(t *testing.T) {
	tests := []struct {
		name                string
		inputs              []input
		timestamp           time.Time
		expectedNext        int64
		expectedLast        int64
		expectedLog         map[int64]*packetReport
		expectedBeginNumber uint16
		expectedMetrics     []rtcp.CCFeedbackMetricBlock
	}{
		{
			name:                "emptyLog",
			inputs:              []input{},
			timestamp:           time.Time{},
			expectedNext:        0,
			expectedLast:        0,
			expectedLog:         map[int64]*packetReport{},
			expectedBeginNumber: 0,
			expectedMetrics:     []rtcp.CCFeedbackMetricBlock{},
		},
		{
			name: "addInOrderSequence",
			inputs: []input{
				{
					ts:  time.Time{},
					nr:  0,
					ecn: 0,
				},
				{
					ts:  time.Time{}.Add(10 * time.Millisecond),
					nr:  1,
					ecn: 0,
				},
				{
					ts:  time.Time{}.Add(20 * time.Millisecond),
					nr:  2,
					ecn: 0,
				},
				{
					ts:  time.Time{}.Add(30 * time.Millisecond),
					nr:  3,
					ecn: 0,
				},
			},
			timestamp:    time.Time{}.Add(15 * time.Millisecond),
			expectedNext: 2,
			expectedLast: 3,
			expectedLog: map[int64]*packetReport{
				2: {
					arrivalTime: time.Time{}.Add(20 * time.Millisecond),
					ecn:         0,
				},
				3: {
					arrivalTime: time.Time{}.Add(30 * time.Millisecond),
					ecn:         0,
				},
			},
			expectedBeginNumber: 0,
			expectedMetrics: []rtcp.CCFeedbackMetricBlock{
				{
					Received:          true,
					ECN:               0,
					ArrivalTimeOffset: 1003,
				},
				{
					Received:          true,
					ECN:               0,
					ArrivalTimeOffset: 993,
				},
			},
		},
		{
			name: "reorderedSequence",
			inputs: []input{
				{
					ts:  time.Time{},
					nr:  0,
					ecn: 0,
				},
				{
					ts:  time.Time{}.Add(10 * time.Millisecond),
					nr:  2,
					ecn: 0,
				},
				{
					ts:  time.Time{}.Add(20 * time.Millisecond),
					nr:  1,
					ecn: 0,
				},
				{
					ts:  time.Time{}.Add(30 * time.Millisecond),
					nr:  3,
					ecn: 0,
				},
			},
			timestamp:    time.Time{}.Add(15 * time.Millisecond),
			expectedNext: 1,
			expectedLast: 3,
			expectedLog: map[int64]*packetReport{
				1: {
					arrivalTime: time.Time{}.Add(20 * time.Millisecond),
					ecn:         0,
				},
				2: {
					arrivalTime: time.Time{}.Add(10 * time.Millisecond),
					ecn:         0,
				},
				3: {
					arrivalTime: time.Time{}.Add(30 * time.Millisecond),
					ecn:         0,
				},
			},
			expectedBeginNumber: 1,
			expectedMetrics: []rtcp.CCFeedbackMetricBlock{
				{
					Received:          true,
					ECN:               0,
					ArrivalTimeOffset: 1003,
				},
				{
					Received:          true,
					ECN:               0,
					ArrivalTimeOffset: 1013,
				},
				{
					Received:          true,
					ECN:               0,
					ArrivalTimeOffset: 993,
				},
			},
		},
		{
			name: "reorderedWrappingSequence",
			inputs: []input{
				{
					ts:  time.Time{},
					nr:  65534,
					ecn: 0,
				},
				{
					ts:  time.Time{}.Add(10 * time.Millisecond),
					nr:  0,
					ecn: 0,
				},
				{
					ts:  time.Time{}.Add(20 * time.Millisecond),
					nr:  65535,
					ecn: 0,
				},
				{
					ts:  time.Time{}.Add(30 * time.Millisecond),
					nr:  2,
					ecn: 0,
				},
				{
					ts:  time.Time{}.Add(40 * time.Millisecond),
					nr:  1,
					ecn: 0,
				},
				{
					ts:  time.Time{}.Add(50 * time.Millisecond),
					nr:  3,
					ecn: 0,
				},
			},
			timestamp:    time.Time{}.Add(15 * time.Millisecond),
			expectedNext: 65535,
			expectedLast: 65539,
			expectedLog: map[int64]*packetReport{
				65535: {
					arrivalTime: time.Time{}.Add(20 * time.Millisecond),
					ecn:         0,
				},
				65536: {
					arrivalTime: time.Time{}.Add(10 * time.Millisecond),
					ecn:         0,
				},
				65537: {
					arrivalTime: time.Time{}.Add(40 * time.Millisecond),
					ecn:         0,
				},
				65538: {
					arrivalTime: time.Time{}.Add(30 * time.Millisecond),
					ecn:         0,
				},
				65539: {
					arrivalTime: time.Time{}.Add(50 * time.Millisecond),
					ecn:         0,
				},
			},
			expectedBeginNumber: 0,
			expectedMetrics: []rtcp.CCFeedbackMetricBlock{
				{
					Received:          true,
					ECN:               0,
					ArrivalTimeOffset: 1003,
				},
				{
					Received:          true,
					ECN:               0,
					ArrivalTimeOffset: 1013,
				},
				{
					Received:          true,
					ECN:               0,
					ArrivalTimeOffset: 983,
				},
				{
					Received:          true,
					ECN:               0,
					ArrivalTimeOffset: 993,
				},
				{
					Received:          true,
					ECN:               0,
					ArrivalTimeOffset: 972,
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			sl := newStreamLog()
			for _, input := range test.inputs {
				sl.add(input.ts, input.nr, input.ecn)
			}
			metrics := sl.metricsAfter(test.timestamp, time.Time{}.Add(time.Second))
			assert.Equal(t, test.expectedNext, sl.nextSequenceNumber)
			assert.Equal(t, test.expectedLast, sl.lastSequenceNumber)
			assert.Equal(t, test.expectedLog, sl.log)
			assert.Equal(t, test.expectedMetrics, metrics)
		})
	}
}
