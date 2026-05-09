package conf

import (
	"fmt"
	"net"
	"strconv"
	"strings"

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

	resolvedNameServer, err := ResolveNameServerAddr(C.RocketMQ.NameServer)
	if err != nil {
		logrus.Fatalf("[conf] resolve rocketmq nameserver err: %v", err)
	}

	if resolvedNameServer != C.RocketMQ.NameServer {
		logrus.Infof("[conf] rocketmq nameserver resolved: %s -> %s", C.RocketMQ.NameServer, resolvedNameServer)
		C.RocketMQ.NameServer = resolvedNameServer
	}

	logrus.Infof("[conf] config loaded: server=%s, mysql=%s:%d, redis=%s, kvrocks=%s, rpc.push=%s, rpc.action=%s, mq=%s",
		C.Server.Port,
		C.MySQL.Host,
		C.MySQL.Port,
		C.Redis.Addr,
		C.Kvrocks.Addr,
		C.RPC.PushAddr,
		C.RPC.ActionAddr,
		C.RocketMQ.NameServer,
	)
}

func ResolveNameServerAddr(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", fmt.Errorf("rocketmq nameserver is empty")
	}

	separators := func(r rune) bool {
		return r == ',' || r == ';'
	}

	items := strings.FieldsFunc(raw, separators)
	resolved := make([]string, 0, len(items))

	for _, item := range items {
		addr := strings.TrimSpace(item)
		if addr == "" {
			continue
		}

		resolvedAddr, err := resolveHostPortToIPv4(addr)
		if err != nil {
			return "", err
		}

		resolved = append(resolved, resolvedAddr)
	}

	if len(resolved) == 0 {
		return "", fmt.Errorf("rocketmq nameserver is empty after parsing: %q", raw)
	}

	return strings.Join(resolved, ";"), nil
}

func resolveHostPortToIPv4(addr string) (string, error) {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return "", fmt.Errorf("invalid nameserver address %q, expected host:port: %w", addr, err)
	}

	if ip := net.ParseIP(host); ip != nil {
		if v4 := ip.To4(); v4 != nil {
			return net.JoinHostPort(v4.String(), port), nil
		}
		return "", fmt.Errorf("nameserver address %q is IPv6, expected IPv4", addr)
	}

	ips, err := net.LookupIP(host)
	if err != nil {
		return "", fmt.Errorf("lookup nameserver host %q failed: %w", host, err)
	}

	for _, ip := range ips {
		if v4 := ip.To4(); v4 != nil {
			return net.JoinHostPort(v4.String(), port), nil
		}
	}

	return "", fmt.Errorf("no IPv4 found for nameserver host %q", host)
}
