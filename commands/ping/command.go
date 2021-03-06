// Packe ping implements the ping command.
package ping

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/fatih/color"
	"github.com/giantswarm/gscliauth/config"
	"github.com/giantswarm/microerror"
	rootcerts "github.com/hashicorp/go-rootcerts"
	"github.com/spf13/cobra"

	"github.com/giantswarm/gsctl/commands/errors"
	"github.com/giantswarm/gsctl/flags"
)

var (
	// Command is the "ping" CLI command
	Command = &cobra.Command{
		Use:    "ping",
		Short:  "Check API connection",
		Long:   `Tests the connection to the API`,
		PreRun: printValidation,
		Run:    runCommand,
	}
)

// Arguments specifies all the arguments to be used for our business function.
type Arguments struct {
	apiEndpoint string
	verbose     bool
}

// collectArguments fills arguments from user input, config, and environment.
func collectArguments() Arguments {
	endpoint := config.Config.ChooseEndpoint(flags.APIEndpoint)

	return Arguments{
		apiEndpoint: endpoint,
		verbose:     flags.Verbose,
	}
}

func printValidation(cmd *cobra.Command, cmdLineArgs []string) {
	args := collectArguments()
	err := verifyPreconditions(args, cmdLineArgs)

	if err == nil {
		return
	}

	errors.HandleCommonErrors(err)

	// handle non-common errors
	fmt.Println(color.RedString(err.Error()))
	os.Exit(1)
}

func verifyPreconditions(args Arguments, cmdLineArgs []string) error {
	if args.apiEndpoint == "" {
		return microerror.Mask(errors.EndpointMissingError)
	}

	return nil
}

// runCommand executes the ping() function
// and prints output in a user-friendly way
func runCommand(cmd *cobra.Command, cmdLineArgs []string) {
	args := collectArguments()

	duration, err := ping(args.apiEndpoint)
	if err != nil {

		errors.HandleCommonErrors(err)

		fmt.Println(color.RedString("Could not reach API"))
		fmt.Println(err.Error())
		os.Exit(1)
	}

	fmt.Println(color.GreenString("API connection is fine"))
	fmt.Printf("Ping took %d Milliseconds\n", duration/time.Millisecond)
}

// ping checks the API connection and returns
// duration (in case of success) and error (in case of failure)
func ping(endpointURL string) (time.Duration, error) {
	var duration time.Duration

	// create root URI for the endpoint
	u, err := url.Parse(endpointURL)
	if err != nil {
		return duration, microerror.Mask(err)
	}
	u, err = u.Parse("/")
	if err != nil {
		return duration, microerror.Mask(err)
	}

	// create client and request
	request, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return duration, microerror.Mask(err)
	}
	request.Header.Set("User-Agent", config.UserAgent())

	// create client
	tlsConfig := &tls.Config{}
	err = rootcerts.ConfigureTLS(tlsConfig, &rootcerts.Config{
		CAFile: os.Getenv("GSCTL_CAFILE"),
		CAPath: os.Getenv("GSCTL_CAPATH"),
	})
	if err != nil {
		return duration, microerror.Mask(err)
	}
	t := &http.Transport{
		Proxy:           http.ProxyFromEnvironment,
		TLSClientConfig: tlsConfig,
	}
	pingClient := &http.Client{
		Timeout:   5 * time.Second,
		Transport: t,
	}

	start := time.Now()
	resp, err := pingClient.Do(request)
	if err != nil {
		return duration, microerror.Mask(err)
	}
	defer resp.Body.Close()

	duration = time.Since(start)
	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusForbidden {
			return duration, microerror.Mask(errors.AccessForbiddenError)
		}
		return duration, microerror.Mask(fmt.Errorf("bad status code %d", resp.StatusCode))
	}

	return duration, nil
}
