package bedis

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

const _configURL = "https://raw.githubusercontent.com/wrfly/redis-server.tgz/master/redis.conf"

var _defaultConfigs = `
# do not save to disk
save ""
# one database only
databases 1
# default
appendonly no
# double the default size
maxclients 1024

# default memory 1g
maxmemory 3gb

# disable TCP
port 0
# enable unixsocket
unixsocket /tmp/redis.sock
# socket permission
unixsocketperm 700

# LFU
maxmemory-policy allkeys-lfu

# disable commands
rename-command APPEND ""
rename-command BGSAVE ""
rename-command RENAME ""
rename-command SAVE ""
rename-command SPOP ""
rename-command SREM ""

`

func (r *builtinRedis) buildConfig() error {
	// mk config path
	if err := _makeDir(r.runtimePath); err != nil {
		return err
	}

	// download config
	resp, err := http.Get(_configURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	configBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// build config file
	f, err := os.Create(r.confPath)
	if err != nil {
		return err
	}
	defer f.Close()

	// write main config
	_, err = f.Write(configBytes)
	if err != nil {
		return err
	}

	// local configs
	writeLine(f, "######### LOCAL CONFIG #########")
	writeLine(f, _defaultConfigs)
	writeLine(f, "unixsocket %s", r.socketPath) // write unix socket file

	// other configs
	for k, v := range r.opt.Config {
		fmt.Fprintf(f, "%s %s\n", k, v)
	}

	return nil
}
