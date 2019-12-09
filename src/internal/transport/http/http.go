package http

import (
    "github.com/coreos/etcd/clientv3"
    ginzap "github.com/gin-contrib/zap"
    "github.com/gin-gonic/gin"
    "github.com/google/wire"
    "github.com/opentracing-contrib/go-gin/ginhttp"
    "github.com/opentracing/opentracing-go"
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

func New(opt *Option, logger *zap.Logger, engine *gin.Engine, client *clientv3.Client) (*Server, error)  {
    server := Server{
        logger: logger,
        engine: engine,
        server: &http.Server{
            Handler: engine,
            Addr: net.JoinHostPort(opt.Host, opt.Port),
        },
        client:    client,
    }
    return &server, nil
}

func (h *Server)Start() error {
    go func() {
        if err :=  h.server.ListenAndServe();err != nil {
            h.logger.Error("listen http server fail ", zap.Error(err))
            panic(err)
        }
    }()
    return  nil
}


var ProvideSet = wire.NewSet(NewOption, NewRouter, New)

