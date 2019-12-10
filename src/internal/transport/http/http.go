package http

import (
    "context"
    "fmt"
    "github.com/coreos/etcd/clientv3"
    ginzap "github.com/gin-contrib/zap"
    "github.com/gin-gonic/gin"
    "github.com/google/wire"
    "github.com/opentracing-contrib/go-gin/ginhttp"
    "github.com/opentracing/opentracing-go"
    "github.com/slowmoon/go-kits/internal/errors"
    "github.com/slowmoon/go-kits/internal/registry"
    "github.com/slowmoon/go-kits/internal/transport/http/middleware/ginprom"
    "github.com/spf13/viper"
    "go.uber.org/zap"
    "net"
    "net/http"
    "time"
)

//a simple http server

type Option struct {
    Host   string
    Port   string
}

type Server struct {
    name     string
    logger   *zap.Logger
    server   *http.Server
    engine   *gin.Engine
    client   *clientv3.Client
    registry registry.Registry
    option   *Option
}

func NewOption(config *viper.Viper) (*Option, error) {
    var option Option
    if err := config.UnmarshalKey("http", &option);err != nil {
        return  nil, err
    }
    return &option, nil
}

type ControllerInit func(engine *gin.Engine)


func NewRouter(logger *zap.Logger, init ControllerInit, trace opentracing.Tracer) *gin.Engine {
    engine := gin.New()
    //engine.Use(gin.Recovery())
    engine.Use(ginzap.Ginzap(logger, time.RFC3339, true))
    engine.Use(ginzap.RecoveryWithZap(logger, true))
    engine.Use(ginprom.New(engine).Middleware())
    engine.Use(ginhttp.Middleware(trace))
    init(engine)
    return engine
}

func New(opt *Option, logger *zap.Logger, engine *gin.Engine, client *clientv3.Client, registry registry.Registry) (*Server, error)  {
    server := Server{
        logger: logger,
        engine: engine,
        server: &http.Server{
            Handler: engine,
            Addr: net.JoinHostPort(opt.Host, opt.Port),
        },
        registry: registry,
        client:    client,
        option:  opt,
    }
    return &server, nil
}

func (h *Server)Name(name string)  {
    h.name = name
}

func (h *Server)Start() error {
    go func() {
        if err :=  h.server.ListenAndServe();err != nil {
            h.logger.Error("listen http server fail ", zap.Error(err))
            panic(err)
        }
    }()
    key := fmt.Sprintf("%s.%s.%s", h.name, h.option.Host, h.option.Port)
    if h.registry != nil {
        if err := h.registry.Register(key, h.name);err != nil {
            h.logger.Error("server register fail ", zap.Error(err))
            return  err
        }
    }
    return  nil
}

func (h *Server)deregister() error {
    key := fmt.Sprintf("%s.%s.%s", h.name, h.option.Host, h.option.Port)
    if h.registry != nil {
        if err := h.registry.DeRegister(key);err != nil {
            h.logger.Error("server deregister fail ", zap.Error(err))
            return  err
        }
    }
    return  nil
}


func (h *Server)Stop() error {
    ctx , cancel := context.WithTimeout(context.Background(), time.Second *5)
    defer cancel()
    multiError := errors.MultiError{}

    if err := h.server.Shutdown(ctx);err != nil {
        multiError.Add(err)
    }
    if err := h.deregister();err != nil {
        multiError.Add(err)
    }
    if err := h.registry.Close();err != nil {
        h.logger.Error("http registry close fail")
        multiError.Add(err)
    }
    if err := h.client.Close();err != nil {
        h.logger.Error("http registry close fail")
        multiError.Add(err)
    }

    return multiError
}


var ProvideSet = wire.NewSet(NewOption, NewRouter, New)

