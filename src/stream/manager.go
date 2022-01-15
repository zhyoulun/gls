package stream

import (
	"fmt"
	"github.com/zhyoulun/gls/src/rtmp"
	"sync"
)

var Mgr *Manager

type Manager struct {
	streamsMutex *sync.Mutex
	streams      map[string]*Stream
}

func InitManager() error {
	Mgr = &Manager{
		streamsMutex: &sync.Mutex{},
		streams:      make(map[string]*Stream),
	}
	return nil
}

func (m *Manager) HandlePublish(conn *rtmp.Conn) error {
	var stream *Stream
	var err error
	if stream, err = m.getOrCreateStream(conn.GetConnInfo(), conn.GetStreamName()); err != nil {
		return err
	}
	if err = stream.SetSource(conn); err != nil {
		return err
	}
	return nil

}

func (m *Manager) HandlePlay(conn *rtmp.Conn) error {
	var stream *Stream
	var err error
	if stream, err = m.getOrCreateStream(conn.GetConnInfo(), conn.GetStreamName()); err != nil {
		return err
	}
	if err = stream.AddSink(conn); err != nil {
		return err
	}
	return nil
}

func (m *Manager) getOrCreateStream(connInfo rtmp.ConnectCommentObject, streamName string) (*Stream, error) {
	m.streamsMutex.Lock()
	defer m.streamsMutex.Unlock()

	var err error
	var ok bool
	var stream *Stream

	key := m.genStreamKey(connInfo, streamName)
	if stream, ok = m.streams[key]; !ok {
		if stream, err = newStream(); err != nil {
			return nil, err
		} else {
			m.streams[key] = stream
		}
	}
	return stream, nil
}

//todo 待完善
func (m *Manager) genStreamKey(connInfo rtmp.ConnectCommentObject, streamName string) string {
	return fmt.Sprintf("%s:%s", connInfo.App, streamName)
}
