package conf

import (
	"strconv"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	MySQL    MySQLConfig    `mapstructure:"mysql"`
	Redis    RedisConfig    `mapstructure:"redis"`
	Kvrocks  KvrocksConfig  `mapstructure:"kvrocks"`
	RPC      RPCConfig      `mapstructure:"rpc"`
	RocketMQ RocketMQConfig `mapstructure:"rocketmq"`
}

type ServerConfig struct {
	Port string `mapstructure:"port"`
}

type MySQLConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Database string `mapstructure:"database"`
}

func (c *MySQLConfig) DSN() string {
	return c.User + ":" + c.Password + "@tcp(" + c.Host + ":" + strconv.Itoa(c.Port) + ")/" + c.Database + "?charset=utf8&parseTime=True&loc=Local"
}

type RedisConfig struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type KvrocksConfig struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type RPCConfig struct {
	PushAddr   string `mapstructure:"push_addr"`
	ActionAddr string `mapstructure:"action_addr"`
}

type RocketMQConfig struct {
	NameServer string `mapstructure:"nameserver"`
	Retry      int    `mapstructure:"retry"`
}

var C Config

func Init() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./conf")

	if err := viper.ReadInConfig(); err != nil {
		logrus.Fatalf("[conf] read config err: %v", err)
	}

	if err := viper.Unmarshal(&C); err != nil {
		logrus.Fatalf("[conf] unmarshal config err: %v", err)
	}

	logrus.Infof("[conf] config loaded: server=%s, mysql=%s:%d, redis=%s, kvrocks=%s, rpc.push=%s, rpc.action=%s, mq=%s",
		C.Server.Port, C.MySQL.Host, C.MySQL.Port, C.Redis.Addr, C.Kvrocks.Addr, C.RPC.PushAddr, C.RPC.ActionAddr, C.RocketMQ.NameServer)
}