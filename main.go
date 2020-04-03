package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/spf13/pflag"

	"github.com/criteo/http-proxy-exporter/proxyclient"

	"github.com/spf13/viper"
)

var (
	appName      string
	buildVersion string
	buildNumber  string
	buildTime    string
)

type config struct {
	Username string
	Password string
	Insecure bool
	Timeout  time.Duration
	Proxies  []string
	Targets  []string
}

func main() {
	pflag.StringP("config_path", "c", ".", "path to folder containing config file (config.yml)")
	pflag.BoolP("version", "v", false, "print version and exit")

	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)
	viper.AddConfigPath(viper.GetString("config_path"))
	var config config

	if viper.GetBool("version") {
		fmt.Println(appName)
		fmt.Println("version:", buildVersion)
		fmt.Println("build:", buildNumber)
		fmt.Println("build time:", buildTime)
		os.Exit(0)
	}

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("error reading configuration file: %s", err)
	}
	if err := viper.Unmarshal(&config); err != nil {
		log.Fatalf("error while converting configuration to struct: %s", err)
	}

	measurements := len(config.Targets) * len(config.Proxies)

	ctx := context.Background()
	if config.Timeout > 0 {
		ctx, _ = context.WithTimeout(ctx, config.Timeout)
	}

	if config.Timeout > 0 {
		log.Printf("running %d tests (timeout %s)...", measurements, config.Timeout)
	} else {
		log.Printf("running %d tests...", measurements)
	}

	var errors []error
	var errorsLock sync.Mutex

	var wg sync.WaitGroup
	wg.Add(measurements)

	for _, target := range config.Targets {
		for _, proxy := range config.Proxies {
			go func(target, proxy string) {
				defer wg.Done()

				err := testOne(ctx, config, proxy, target)
				if err != nil {
					errorsLock.Lock()
					errors = append(errors, err)
					errorsLock.Unlock()
				}

			}(target, proxy)
		}
	}

	// wait for tests to finish
	wg.Wait()

	// if we have the same amount of errors and target, sys.exit(1)
	if len(errors) > 0 {
		log.Println("errors happened during check:")
		for _, err := range errors {
			log.Println(err)
		}
		if len(errors) == measurements {
			log.Fatalf("all targets in error")
		}
	}
	success := measurements - len(errors)
	log.Printf("%v/%v targets ok\n", success, measurements)
}

func testOne(ctx context.Context, cfg config, proxy, target string) error {
	auth := proxyclient.AuthMethod{
		Type: "basic",
		Params: map[string]string{
			"username": cfg.Username,
			"password": cfg.Password,
		},
	}

	rc := proxyclient.RequestConfig{
		Target:   target,
		Proxy:    proxy,
		Auth:     &auth,
		Insecure: cfg.Insecure,
	}
	preq, err := proxyclient.MakeClientAndRequest(rc)

	if err != nil {
		return fmt.Errorf("could not prepare request: %s", err)
	}

	req := preq.Request
	req = req.WithContext(ctx)

	log.Printf("testing target %q with proxy %q", target, proxy)
	resp, err := preq.Client.Do(req)
	if err != nil {
		return fmt.Errorf("could not prepare request to %s: %s", target, err)
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("got %v (200 expected) to %s", resp.StatusCode, target)
	}

	return nil
}
