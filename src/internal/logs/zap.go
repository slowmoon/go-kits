package logs

import (
   "github.com/google/wire"
   "github.com/spf13/viper"
   "go.uber.org/zap"
   "go.uber.org/zap/zapcore"
   "gopkg.in/natefinch/lumberjack.v2"
   "os"
)

type Option struct {
   Filename   string
   MaxSize    int
   MaxBackups int
   MaxAge     int
   Level      string
   Stdout     bool
}


func NewOption(config *viper.Viper) (*Option, error) {
   var option Option
   if err := config.UnmarshalKey("zap", &option);err != nil {
      return  nil, err
   }
   return  &option, nil
}


func New(op Option) (*zap.Logger, error) {
   var (
      logger  *zap.Logger
      level    = zap.NewAtomicLevel()
   )
   if err := level.UnmarshalText([]byte(op.Level));err != nil {
      return nil, err
   }

    fw := zapcore.AddSync(&lumberjack.Logger{
        Filename:   op.Filename,
        MaxSize:    op.MaxSize, // megabytes
        MaxBackups: op.MaxBackups,
        MaxAge:     op.MaxAge, // days
    })
    ws := zapcore.Lock(os.Stdout)
    cores := make([]zapcore.Core, 0, 2)
    en := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
    cores = append(cores, zapcore.NewCore(en, fw, level))
    if op.Stdout {
        std := zapcore.NewConsoleEncoder(zap.NewProductionEncoderConfig())
        cores = append(cores, zapcore.NewCore(std, ws, level))
    }
   core := zapcore.NewTee(cores...)
   logger = zap.New(core)
   return logger, nil
}


var ProvideSet = wire.NewSet(New, NewOption)