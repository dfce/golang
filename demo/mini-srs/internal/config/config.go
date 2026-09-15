package config

type Config struct {
	ListenAddr string
	RtmpPort   int
}

func DefaultConfig() *Config {
	return &Config{
		ListenAddr: "0.0.0.0",
		RtmpPort:   1935,
	}
}
