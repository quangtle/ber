package streaming

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

type StreamSession struct {
	ID        string
	VideoPath string
	StartedAt time.Time
	BytesSent int64
	mu        sync.Mutex
}

type StreamManager struct {
	mu       sync.RWMutex
	sessions map[string]*StreamSession
	maxConcurrent int
}

func NewStreamManager(maxConcurrent int) *StreamManager {
	return &StreamManager{
		sessions:      make(map[string]*StreamSession),
		maxConcurrent: maxConcurrent,
	}
}

func (sm *StreamManager) NewSession(id, videoPath string) (*StreamSession, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if len(sm.sessions) >= sm.maxConcurrent {
		return nil, fmt.Errorf("max concurrent streams reached (%d)", sm.maxConcurrent)
	}

	session := &StreamSession{
		ID:        id,
		VideoPath: videoPath,
		StartedAt: time.Now(),
	}
	sm.sessions[id] = session
	return session, nil
}

func (sm *StreamManager) GetSession(id string) *StreamSession {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.sessions[id]
}

func (sm *StreamManager) CloseSession(id string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	delete(sm.sessions, id)
}

func (sm *StreamManager) ActiveCount() int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return len(sm.sessions)
}

func (ss *StreamSession) StreamRange(w io.Writer, start, end int64) error {
	file, err := os.Open(ss.VideoPath)
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err := file.Seek(start, io.SeekStart); err != nil {
		return err
	}

	n, err := io.CopyN(w, file, end-start+1)
	ss.mu.Lock()
	ss.BytesSent += n
	ss.mu.Unlock()

	return err
}
