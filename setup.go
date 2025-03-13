package bedis

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

const (
	_repo       = "https://github.com/wrfly/redis-server.tgz"
	_linkLinux  = _repo + "/raw/master/7.4.2/redis-server-linux-libc-x86_64.tgz"
	_linkDarwin = _repo + "/raw/master/7.4.2/redis-server-darwin-libc-arm64.tgz"
)

var (
	root    = filepath.Join(os.TempDir(), "builtin-redis")
	binPath = filepath.Join(root, "redis-server")
)

func _makeDir(path string) error {
	f, err := os.Open(path)
	if err == nil {
		f.Close()
		return nil
	}

	if os.IsNotExist(err) {
		err := os.Mkdir(path, 0755)
		if err != nil {
			return fmt.Errorf("cannot make temp dir `%s`, err: %s", path, err)
		}
		return nil
	}

	return fmt.Errorf("cannot open temp dir `%s`, err: %s", path, err)
}

func _smokeTest() error {
	cmd := exec.Command(binPath, "--version")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	// read command's stdout line by line
	in := bufio.NewScanner(stdout)

	for in.Scan() {
		logs.Println(in.Text())
	}

	return in.Err()
}

// mk all the directories, download binary tarball, write config
// and check the binary downloaded
func (r *builtinRedis) setup() error {
	logs.Printf("setup redis under %s", r.runtimePath)
	// download binary is not exist
	if !fileExists(binPath) {
		if err := downloadRedisBinary(); err != nil {
			return err
		}
	}

	// write config
	if err := r.buildConfig(); err != nil {
		return fmt.Errorf("build config err: %s", err)
	}

	// check whether redis can be executed
	return _smokeTest()
}

func downloadRedisBinary() error {
	if f, err := os.Open(binPath); err == nil {
		f.Close()
		return nil
	}

	var link string
	GOOS := runtime.GOOS
	if runtime.GOOS == "darwin" {
		link = _linkDarwin
	} else if runtime.GOOS == "linux" {
		link = _linkLinux
	} else {
		return fmt.Errorf("unsupported platform %s", GOOS)
	}

	logs.Printf("download redis-server from %s", link)
	resp, err := http.Get(link)
	if err != nil {
		return fmt.Errorf("cannot download, err: %s", err)
	}
	defer resp.Body.Close()

	if err := unTarball(root, resp.Body); err != nil {
		return fmt.Errorf("extract tarball err: %s", err)
	}

	return nil
}
