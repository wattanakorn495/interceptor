package rfc8888

import (
	"sync"
	"time"

	"github.com/pion/interceptor"
	"github.com/pion/logging"
	"github.com/pion/rtcp"
)

type SenderInterceptorFactory struct {
	opts []Option
}

func (s *SenderInterceptorFactory) NewInterceptor(id string) (interceptor.Interceptor, error) {
	i := &SenderInterceptor{}
	return i, nil
}

type SenderInterceptor struct {
	interceptor.NoOp
	log      logging.LeveledLogger
	lock     sync.Mutex
	wg       sync.WaitGroup
	recorder *Recorder
	interval time.Duration
	close    chan struct{}
}

func (s *SenderInterceptor) BindRTCPWriter(writer interceptor.RTCPWriter) interceptor.RTCPWriter {
	if s.isClosed() {
		return writer
	}

	s.lock.Lock()
	defer s.lock.Unlock()

	s.recorder = NewRecorder()
	s.wg.Add(1)
	go s.loop(writer)
	return writer
}

func (s *SenderInterceptor) BindRemoteStream(info *interceptor.StreamInfo, reader interceptor.RTPReader) interceptor.RTPReader {
	return interceptor.RTPReaderFunc(func(b []byte, a interceptor.Attributes) (int, interceptor.Attributes, error) {

		i, attr, err := reader.Read(b, a)
		if err != nil {
			return 0, nil, err
		}

		if attr == nil {
			attr = make(interceptor.Attributes)
		}
		header, err := attr.GetRTPHeader(b[:i])
		if err != nil {
			return 0, nil, err
		}

		s.lock.Lock()
		// TODO: Get ECN?
		s.recorder.AddPacket(time.Now(), header.SSRC, header.SequenceNumber, 0)
		s.lock.Unlock()

		return i, attr, nil
	})
}

// Close closes the interceptor.
func (s *SenderInterceptor) Close() error {
	defer s.wg.Wait()

	if !s.isClosed() {
		close(s.close)
	}

	return nil
}

func (s *SenderInterceptor) isClosed() bool {
	select {
	case <-s.close:
		return true
	default:
		return false
	}
}

func (s *SenderInterceptor) loop(writer interceptor.RTCPWriter) {

	ticker := time.NewTicker(20 * time.Millisecond)
	for {
		select {
		case <-s.close:
			ticker.Stop()
			return

		case now := <-ticker.C:
			s.lock.Lock()
			pkts := s.recorder.BuildReport(now)
			s.lock.Unlock()
			if pkts == nil {
				continue
			}
			if _, err := writer.Write([]rtcp.Packet{pkts}, nil); err != nil {
				s.log.Error(err.Error())
			}
		}
	}
}
