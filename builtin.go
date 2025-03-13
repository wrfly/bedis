package bedis

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/redis/go-redis/v9"
)

var logs = log.New(os.Stdout, "bedis", log.LstdFlags)

// BuiltinRedis represents the redis runing on localhost,
// connected via a socket
// remember to `StopAndClose` it after the program exit
type BuiltinRedis interface {
	NewClient(cfg redis.Options) (*redis.Client, error)
	DefaultClient() (*redis.Client, error)
	StopAndClose()
}

// Option to config the builtin redis
type Option struct {
	MuteLog bool

	// Set a memory usage limit to the specified amount of bytes.
	// When the memory limit is reached Redis will try to remove keys
	// according to the eviction policy selected (see maxmemory-policy).
	Memory string `default:"3gb"` // redis max memory

	CPUAffinity bool `yaml:"cpu-affinity"`

	Config map[string]string `yaml:"config"` // other configs, kv pair
}

// New will first download the redis and init the config
// if anything goes wrong at the statup stage, return error
// however, if the redis-server exit for 3 times, library will panic,
// means the builtin redis is not avaliable
func New(opt Option) (_ BuiltinRedis, err error) {
	if opt.MuteLog {
		logs.SetOutput(io.Discard)
	}
	// mk root path
	if err := _makeDir(root); err != nil {
		return nil, err
	}
	// x.* doesn't work on go1.13
	runtimePath, err := os.MkdirTemp(root, fmt.Sprintf("%d.", os.Getpid()))
	if err != nil {
		return nil, fmt.Errorf("create temp dir err: %s", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	builtin := &builtinRedis{
		ctx:    ctx,
		cancel: cancel,

		runtimePath: runtimePath,
		confPath:    filepath.Join(runtimePath, "redis.conf"),
		socketPath:  filepath.Join(runtimePath, "redis.sock"),

		opt: opt, // config options
	}

	// set up failed, clean config files and temp dir
	defer func() {
		if err != nil {
			builtin.StopAndClose()
		}
	}()

	if err := builtin.setup(); err != nil {
		return nil, fmt.Errorf("setup redis err: %s", err)
	}

	// keep running in bg
	go func() {
		backoffTime := time.Millisecond * 100
		for i := 0; i < 3; i++ {
			if err := builtin.start(ctx); err != nil {
				logs.Printf("redis exit with error: %s, restart redis", err)
			} else {
				logs.Printf("redis exit with nil error, restart redis")
			}
			time.Sleep(backoffTime)
			backoffTime *= 2

			if ctx.Err() != nil {
				logs.Printf("user cancled, stop redis")
				return
			}
		}
		panic(fmt.Errorf("keep local redis running failed"))
	}()

	time.Sleep(time.Second) // wait for start redis
	if err := builtin.checkStatus(); err != nil {
		return nil, err
	}

	return builtin, nil
}

type builtinRedis struct {
	ctx    context.Context
	cancel context.CancelFunc

	runtimePath string
	confPath    string
	socketPath  string
	memory      uint64

	opt Option

	clients []*redis.Client
}

func (r *builtinRedis) NewClient(cfg redis.Options) (*redis.Client, error) {
	cfg.Network = "unix"
	cfg.Addr = r.socketPath
	client := redis.NewClient(&cfg)
	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}
	r.clients = append(r.clients, client)

	return client, nil
}

func (r *builtinRedis) DefaultClient() (*redis.Client, error) {
	return r.NewClient(redis.Options{})
}

func (r *builtinRedis) StopAndClose() {
	r.cancel()
	for _, client := range r.clients {
		client.Close()
	}
	os.RemoveAll(r.runtimePath)
}
