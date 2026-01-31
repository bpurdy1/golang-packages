package zerologlogger

func WithLevel(level string) Option {
	return func(c *Config) {
		c.LogLevel = level
	}
}

func WithConsole() Option {
	return func(c *Config) {
		c.ConsoleWriter = true
	}
}

func WithShortCaller() Option {
	return func(c *Config) {
		c.CallerMarshalFunc = ShortCallerMarshalFunc
	}
}

func WithFileBaseCaller() Option {
	return func(c *Config) {
		c.CallerMarshalFunc = FileBaseCallerMarshalFunc
	}
}
