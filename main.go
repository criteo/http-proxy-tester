package main

import (
	"fmt"
	"log"
	"os"

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

	auth := proxyclient.AuthMethod{
		Type: "basic",
		Params: map[string]string{
			"username": config.Username,
			"password": config.Password,
		},
	}

	var errors []error
	for _, target := range config.Targets {
		for _, proxy := range config.Proxies {
			rc := proxyclient.RequestConfig{
				Target:   target,
				Proxy:    proxy,
				Auth:     &auth,
				Insecure: config.Insecure,
			}
			preq, err := proxyclient.MakeClientAndRequest(rc)

			if err != nil {
				errors = append(errors, fmt.Errorf("could not prepare request: %s", err))
				continue
			}

			resp, err := preq.Client.Do(preq.Request)
			if err != nil {
				errors = append(errors, fmt.Errorf("could not prepare request to %s: %s", target, err))
				continue
			}
			if resp.StatusCode != 200 {
				errors = append(errors, fmt.Errorf("got %v (200 expected) to %s", resp.StatusCode, target))
			}
		}
	}
	// if we have the same amount of errors and target, sys.exit(1)
	measurements := len(config.Targets) * len(config.Proxies)
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
