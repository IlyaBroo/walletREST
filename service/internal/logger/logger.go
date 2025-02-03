package logger

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/Graylog2/go-gelf/gelf"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type logLevel string

const (
	localLogLevel = "local"
	stageLogLevel = "stage"
	prodLogLevel  = "prod"
)

type Logger interface {
	DebugCtx(ctx context.Context, msg string)
	InfoCtx(ctx context.Context, msg string)
	WarnCtx(ctx context.Context, msg string)
	ErrorCtx(ctx context.Context, msg string)
	FatalCtx(ctx context.Context, msg string, err error)
}

type logger struct {
	log        *logrus.Logger
	cfg        *Config
	gelfWriter *gelf.Writer
}

type Config struct {
	Path         string    `yaml:"path"`
	Level        logLevel  `yaml:"level"`
	Source       bool      `yaml:"source"`
	Service_name string    `yaml:"service_name"`
	Writer       io.Writer `yaml:"writer"`
}

func NewLogger(options ...Option) (Logger, error) {
	l := new(logger)
	l.log = logrus.New()

	for _, opt := range options {
		opt(l)
	}
	if err := l.setLogLevel(); err != nil {
		return nil, err
	}

	if err := l.setupOutput(); err != nil {
		return nil, err
	}
	return l, nil
}

func (l *logger) setLogLevel() error {
	switch l.cfg.Level {
	case localLogLevel:
		l.log.SetLevel(logrus.DebugLevel)
	case stageLogLevel:
		l.log.SetLevel(logrus.InfoLevel)
	case prodLogLevel:
		l.log.SetLevel(logrus.ErrorLevel)
	default:
		return fmt.Errorf("неизвестный уровень логирования: %s", l.cfg.Level)
	}
	return nil
}

func (l *logger) setupOutput() error {
	if l.cfg.Path != "" {
		file, err := os.OpenFile(l.cfg.Path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return err
		}
		l.log.SetOutput(file)
	} else {
		l.log.SetOutput(l.cfg.Writer)
	}
	return nil
}

func (l *logger) DebugCtx(ctx context.Context, msg string) {
	requestID, ok := ctx.Value("requestID").(string)
	if !ok {
		requestID = "unknown"
	}
	l.log.WithFields(logrus.Fields{
		"request_ID":   requestID,
		"service_name": l.cfg.Service_name,
	}).Debug(msg)

}

func (l *logger) InfoCtx(ctx context.Context, msg string) {

	requestID, ok := ctx.Value("requestID").(string)
	if !ok {
		requestID = "unknown"
	}
	l.log.WithFields(logrus.Fields{
		"request_ID":   requestID,
		"service_name": l.cfg.Service_name,
	}).Info(msg)

}
func (l *logger) WarnCtx(ctx context.Context, msg string) {
	requestID, ok := ctx.Value("requestID").(string)
	if !ok {
		requestID = "unknown"
	}
	l.log.WithFields(logrus.Fields{
		"request_ID":   requestID,
		"service_name": l.cfg.Service_name,
	}).Warn(msg)

}

func (l *logger) ErrorCtx(ctx context.Context, msg string) {

	requestID, ok := ctx.Value("requestID").(string)
	if !ok {
		requestID = "unknown"
	}
	l.log.WithFields(logrus.Fields{
		"request_ID":   requestID,
		"service_name": l.cfg.Service_name,
	}).Error(msg)
}
func (l *logger) FatalCtx(ctx context.Context, msg string, err error) {
	requestID, ok := ctx.Value("requestID").(string)
	if !ok {
		requestID = "unknown"
	}
	l.log.WithFields(logrus.Fields{
		"request_ID":   requestID,
		"service_name": l.cfg.Service_name,
		"stack_trace":  errors.WithStack(err),
	}).Fatal(msg)
}
