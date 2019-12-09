package ginprom

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"strconv"
	"sync"
	"time"
)

const (
	metrics  =  "/metrics"
	faviconPath = "/favicon.ico"
)

var httpHistogram = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Namespace: "http_server",
		Name:        "request_counts",
		Help:        "got the request count",
	},
	[]string{"code", "method", "uri"} ,
)

func init()  {
    prometheus.MustRegister(httpHistogram)
}

type handlePath struct {
	sync.Map
}

func (h handlePath)Get(handle string) string {
	value , ok := h.Load(handle)
	if ok {
		return  value.(string)
	}
	return  ""
}

func (h handlePath)Set(info gin.RouteInfo)  {
	h.Store(info.Handler, info.Path)
}

type Option func( *GinPrometheus)


type GinPrometheus struct {
   engine    *gin.Engine
   pathMap   *handlePath
   ignore    map[string]bool
   updated   bool
}

func (g *GinPrometheus)updatePath()  {
	g.updated = true
	for _, route := range g.engine.Routes() {
		g.pathMap.Set(route)
	}
}

func New(engine *gin.Engine, options...Option) *GinPrometheus {
	prom :=  &GinPrometheus{
		engine:  engine,
		pathMap:  &handlePath{} ,
		ignore: map[string]bool{
			metrics: true,
			faviconPath: true,
		},
		updated: false,
	}
	for _, opt := range options {
		opt(prom)
	}
	return prom
}

func (g *GinPrometheus)Middleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if !g.updated {
			g.updatePath()
		}
		if g.ignore[ctx.Request.URL.Path] {
			ctx.Next()
			return
		}

        start := time.Now()
        ctx.Next()

		httpHistogram.WithLabelValues(
			strconv.Itoa(ctx.Writer.Status()),
			ctx.Request.Method,
			g.pathMap.Get(ctx.HandlerName()),
		).Observe(time.Since(start).Seconds())
	}
}



