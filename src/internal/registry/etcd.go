package registry

import (
    "context"
    "github.com/coreos/etcd/clientv3"
    "github.com/google/wire"
    "github.com/pkg/errors"
    "github.com/spf13/viper"
    "go.uber.org/zap"
    "time"
)

type ServiceReg struct {
	logger   *zap.Logger
    client   *clientv3.Client
    lease     clientv3.Lease
    leaseResponse  *clientv3.LeaseGrantResponse
    cancel         func()
    keepalive     <- chan *clientv3.LeaseKeepAliveResponse
    key            string
}

type Option struct {
   Endpoints []string
   DialTimeout  time.Duration
   UserName   string
   Password   string
   timeNum    int64
}

func NewOption(viper *viper.Viper, logger *zap.Logger) (*Option, error) { var opt  Option
   if err := viper.UnmarshalKey("etcd", &opt);err != nil {
      logger.Error("unmarshal etcd config fail ", zap.Error(err))
      return  nil, err
   }
   return  &opt, nil
}

func NewClient(opt *Option, logger *zap.Logger) (*ServiceReg, error) {
    client ,err := clientv3.New(
        clientv3.Config{
            Endpoints:             opt.Endpoints,
            DialTimeout:           opt.DialTimeout,
            Username:              opt.UserName,
            Password:              opt.Password,
        })
    if err != nil {
        return  nil, errors.Wrap(err, "create etcd client fail")
    }
    serviceReg  :=  ServiceReg{
        client:  client,
    }
    if err := serviceReg.setLease(opt.timeNum) ;err != nil {
        return  nil, err
    }
    go serviceReg.ListenLeaseResponse()
    return &serviceReg,  nil
}

func (s *ServiceReg)setLease(timeNum  int64) error {
    lease := clientv3.NewLease(s.client)
    grantResp , err := lease.Grant(context.TODO(), timeNum)
    if err != nil {
        return err
    }
    ctx, cancel := context.WithCancel(context.TODO())
    leaseRespCh, err := lease.KeepAlive(ctx,  grantResp.ID)
    if err != nil {
        return err
    }
    s.lease = lease
    s.leaseResponse = grantResp
    s.cancel =  cancel
    s.keepalive = leaseRespCh
    return  nil
}

func (s *ServiceReg)ListenLeaseResponse()  {
    for {
        select {
        case resp := <- s.keepalive :
            if resp == nil {
                s.logger.Debug("续租关闭")
                return
            }  else {
                s.logger.Debug("续租成功", zap.String("name", "service registration"))
            }
        }
    }
}

func (s *ServiceReg)PutService(key , val string) error {
    kv := clientv3.NewKV(s.client)
    ctx , cancel := context.WithTimeout(context.Background(), 5 *time.Second)
    defer cancel()
    _, err := kv.Put(ctx, key, val, clientv3.WithLease(s.leaseResponse.ID))
    return err
}

func (s *ServiceReg)RevokeRelease() error {
    s.cancel()
    _, err := s.client.Revoke(context.TODO(), s.leaseResponse.ID)
    return err
}

var ProvideSet = wire.NewSet(NewOption, NewClient)