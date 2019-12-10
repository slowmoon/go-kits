package registry

//注册接口类

type Registry interface {
    Register(key , value string)error
    DeRegister(key string) error
    Close() error
}


