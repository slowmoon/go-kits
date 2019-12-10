package application

import (
	"github.com/google/wire"
	"github.com/slowmoon/go-kits/internal/transport/http"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"os"
	"os/signal"
	"syscall"
)

type Application struct {
	name    string
	http   *http.Server
	logger *zap.Logger
}


type AppOption   func(a *Application)

func WithHttpServer (http *http.Server) AppOption {
	return func(a *Application) {
		http.Name(a.name)
		a.http = http
	}
}

type Option struct {
    Name   string
}

func NewOption(config *viper.Viper) (*Option, error) {
	var opt Option
	if err := config.UnmarshalKey("app", &opt)	;err != nil {
		return  nil, err
	}
	return &opt, nil
}


func NewApplication(config * Option, logger *zap.Logger, option... AppOption) (*Application, error) {
	app := Application{
		name: config.Name,
		logger: logger,
	}
	for _, opt := range  option {
		opt(&app)
	}
	return &app, nil
}


func (a *Application)Start() error {
	if a.http!= nil {
		if err := a.http.Start();err != nil {
			a.logger.Error("http server start fail ", zap.Error(err))
			return  err
		}
	}
	return  nil
}

func (a *Application)Await()  {
	wait := make(chan os.Signal, 1)
	signal.Notify(wait, syscall.SIGINT, syscall.SIGTERM)
	select {
	case t := <- wait:
		a.logger.Info("receive stop signal from outside", zap.String("signal", t.String()))
		if err := a.Stop();err != nil {
			a.logger.Error("stop application error", zap.Error(err))
			os.Exit(1)
		}
	}
}

func (a *Application)Stop() error {
	return nil
}

var ProvideSet = wire.NewSet(NewOption, NewApplication)
