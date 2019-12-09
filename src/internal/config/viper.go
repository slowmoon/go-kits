package config

import (
     "github.com/google/wire"
     "github.com/spf13/viper"
)

func New(name string) (*viper.Viper, error) {
     viper := viper.New()
     viper.AddConfigPath(".")
     viper.SetConfigName(name)
     if err := viper.ReadInConfig();err != nil {
          return  nil, err
     }
     return  viper, nil
}

var ProvideSet =  wire.NewSet(New)
