package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jovian1994/cxh-1207-be-interview/apps/translation/config"
	"github.com/jovian1994/cxh-1207-be-interview/apps/translation/initializer"
	"github.com/jovian1994/cxh-1207-be-interview/pkg/logger"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"
)

func main() {

	currentAbPath := getCurrentAbPath()
	runConfigPath := path.Join(currentAbPath, "etc", "run.yml")
	runConfigItem, err := parseRunConfig(runConfigPath)
	if err != nil {
		fmt.Println(fmt.Sprintf("parse run config failed, err:%s", err.Error()))
		panic(err)
	}
	var ginRunMode string
	var configPath string
	if runConfigItem.RunMode == "dev" {
		ginRunMode = "debug"
		logger.InitLogger(logger.WithDebugLevel())
		configPath = path.Join(currentAbPath, "etc", "app-dev.yaml")

	} else {
		ginRunMode = "release"
		configPath = path.Join(currentAbPath, "etc", "app-prd.yaml")
		logger.InitLogger(
			logger.WithInfoLevel(),
			logger.WithFileP(path.Join(currentAbPath, "logs", "app.log")))
	}
	err = config.ParseConfig(configPath)
	if err != nil {
		fmt.Println(fmt.Sprintf("parse config failed, err:%s", err.Error()))
		panic(err)
	}
	gin.SetMode(ginRunMode)
	engine := gin.Default()
	initializer.ServerInit(engine)
	//创建HTTP服务器
	server := &http.Server{
		Addr:    config.GetConfig().Addr,
		Handler: engine,
	}
	//启动HTTP服务器
	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("listen: %s\n", err)
		}
	}()
	//等待一个INT或TERM信号
	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("收到退出信号 ...")

	//创建超时上下文，Shutdown可以让未处理的连接在这个时间内关闭
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	//停止HTTP服务器
	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown:", err)
	}
	log.Println("Server exiting")
}

func getCurrentAbPath() string {
	dir := getCurrentAbPathByExecutable()
	if strings.Contains(dir, getTmpDir()) {
		return getCurrentAbPathByCaller()
	}
	return dir
}

func getCurrentAbPathByExecutable() string {
	exePath, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}
	res, _ := filepath.EvalSymlinks(filepath.Dir(exePath))
	return res
}

func getCurrentAbPathByCaller() string {
	var abPath string
	_, filename, _, ok := runtime.Caller(0)
	if ok {
		abPath = path.Dir(filename)
	}
	return abPath
}

func getTmpDir() string {
	dir := os.Getenv("TEMP")
	if dir == "" {
		dir = os.Getenv("TMP")
	}
	res, _ := filepath.EvalSymlinks(dir)
	return res
}

type runConfig struct {
	RunMode      string `yaml:"run_mode"`
	ConfigSource string `yaml:"config_source"`
}

func parseRunConfig(dist string) (*runConfig, error) {
	data, err := ioutil.ReadFile(dist)
	if err != nil {
		return nil, err
	}
	var c runConfig
	if err := yaml.Unmarshal(data, &c); err != nil {
		return nil, err
	}
	return &c, nil
}
