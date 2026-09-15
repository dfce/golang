package stream

import "sync"

// 全局流管理器（类似单例组件）
// 需要确保全局流管理器（Manager）的读写删除动作全部在互斥锁的绝对安全范围内

type Manager struct {
	mu    sync.RWMutex // 使用读写锁提高高并发性能
	paths map[string]*StreamPath
}

var DefaultManager = &Manager{
	paths: make(map[string]*StreamPath),
}

func (m *Manager) GetOrCreatePath(path string) *StreamPath {
	m.mu.Lock()
	defer m.mu.Unlock()

	sp, exists := m.paths[path]
	if !exists {
		sp = NewStreamPath(path)
		m.paths[path] = sp
	}
	return sp
}

// RemovePath 当发布者断开，流无人观看时，线程安全的销毁通道
func (m *Manager) RemovePath(path string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.paths, path)
}
