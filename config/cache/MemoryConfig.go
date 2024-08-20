package cache

type MemoryConfig struct {
}

func NewMemoryConfig() *MemoryConfig {
	return &MemoryConfig{}
}

func (m *MemoryConfig) GetStorageEngine() string {
	return "memory"
}

func (m *MemoryConfig) GetStoragePath() string {
	return "RAM"
}

func (m *MemoryConfig) GetStorageVersion() string {
	return "v1"
}

func (m *MemoryConfig) Save() error {
	return nil
}

func (m *MemoryConfig) Read() error {
	return nil
}
