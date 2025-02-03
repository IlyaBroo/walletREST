package logger

type Option func(*logger)

func WithCfg(cfg *Config) Option {
	return func(l *logger) {
		l.cfg = cfg
	}
}
