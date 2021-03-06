// Package cluster implements the "update cluster" command.
package cluster

import (
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/giantswarm/gscliauth/config"
	"github.com/giantswarm/gsclientgen/v2/models"
	"github.com/giantswarm/microerror"
	"github.com/spf13/cobra"

	"github.com/giantswarm/gsctl/capabilities"
	"github.com/giantswarm/gsctl/clustercache"
	"github.com/giantswarm/gsctl/util"

	"github.com/giantswarm/gsctl/client"
	"github.com/giantswarm/gsctl/commands/errors"
	"github.com/giantswarm/gsctl/flags"
)

var (
	// Command is the cobra command for 'gsctl update cluster'
	Command = &cobra.Command{
		Use: "cluster <cluster-name/cluster-id>",
		// Args: cobra.ExactArgs(1) guarantees that cobra will fail if no positional argument is given.
		Args:  cobra.ExactArgs(1),
		Short: "Modify cluster details",
		Long: `Change the details of a cluster

Examples:

  gsctl update cluster f01r4 --name "Precious Production Cluster"
  gsctl update cluster "Cluster name" --name "Precious Production Cluster"

  gsctl update cluster f01r4 --label environment=testing --label labeltodelete=
  gsctl update cluster f01r4 --master-ha=true
`,

		// PreRun checks a few general things, like authentication.
		PreRun: printValidation,

		// Run calls the business function and prints results and errors.
		Run: printResult,
	}

	arguments Arguments
)

const (
	activityName = "update-cluster"
)

func init() {
	initFlags()
}

// initFlags initializes flags in a re-usable way, so we can call it from multiple tests.
func initFlags() {
	Command.ResetFlags()
	Command.Flags().StringVarP(&flags.Name, "name", "n", "", "new cluster name")
	Command.Flags().BoolVar(&flags.MasterHA, "master-ha", false, "switch to high-availability master (AWS only)")
	Command.Flags().StringSliceVar(&flags.Label, "label", nil, "modification of a label in form of 'key=value'. Can be specified multiple times. To delete a label set to 'key='")
}

// Arguments represents all the ways the user can influence the command.
type Arguments struct {
	APIEndpoint       string
	AuthToken         string
	ClusterNameOrID   string
	MasterHA          bool
	Labels            []string
	Name              string
	UserProvidedToken string
	Verbose           bool
}

func collectArguments(positionalArgs []string) Arguments {
	endpoint := config.Config.ChooseEndpoint(flags.APIEndpoint)
	token := config.Config.ChooseToken(endpoint, flags.Token)

	return Arguments{
		APIEndpoint:       endpoint,
		AuthToken:         token,
		ClusterNameOrID:   strings.TrimSpace(positionalArgs[0]),
		MasterHA:          flags.MasterHA,
		Labels:            flags.Label,
		Name:              flags.Name,
		UserProvidedToken: flags.Token,
		Verbose:           flags.Verbose,
	}
}

// result represents all information we get back from modifying a cluster.
type result struct {
	ClusterName string
	HasHAMaster bool
	Labels      map[string]string
}

func verifyPreconditions(cmd *cobra.Command, args Arguments) error {
	if args.APIEndpoint == "" {
		return microerror.Mask(errors.EndpointMissingError)
	} else if args.AuthToken == "" && args.UserProvidedToken == "" {
		return microerror.Mask(errors.NotLoggedInError)
	} else if args.ClusterNameOrID == "" {
		return microerror.Mask(errors.ClusterNameOrIDMissingError)
	} else if cmd.Flag("master-ha").Changed && !args.MasterHA {
		return microerror.Mask(revertHAMasterNotAllowedError)
	} else if (args.Name != "" || args.MasterHA) && len(args.Labels) > 0 {
		return microerror.Mask(errors.ConflictingFlagsError)
	}

	return nil
}

func printValidation(cmd *cobra.Command, positionalArgs []string) {
	arguments = collectArguments(positionalArgs)
	err := verifyPreconditions(cmd, arguments)

	if err == nil {
		return
	}

	client.HandleErrors(err)
	errors.HandleCommonErrors(err)

	headline := ""
	subtext := ""

	switch {
	// If there are specific errors to handle, add them here.
	case IsRevertHAMasterNotAllowed(err):
		headline = "Operation not permitted"
		subtext = "It is not possible to change from multiple master nodes to a single master."

	case errors.IsConflictingFlagsError(err):
		headline = "Conflicting flags used"
		subtext = "--name/-n and --label are exclusive."

	default:
		headline = err.Error()
	}

	// print output
	fmt.Println(color.RedString(headline))
	if subtext != "" {
		fmt.Println(subtext)
	}

	os.Exit(1)
}

func updateCluster(args Arguments) (*result, error) {
	var (
		result *result
		err    error
	)

	if len(args.Labels) > 0 {
		result, err = updateLabels(args)

		return result, microerror.Mask(err)
	}

	if args.Name != "" || args.MasterHA {
		result, err = updateClusterSpec(args)

		return result, microerror.Mask(err)
	}

	return nil, microerror.Mask(errors.NoOpError)
}

// updateClusterSpec updates the cluster.
// It determines whether it is a v5 or v4 cluster and uses the appropriate mechanism.
func updateClusterSpec(args Arguments) (*result, error) {
	clientWrapper, err := client.NewWithConfig(args.APIEndpoint, args.UserProvidedToken)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	clusterID, err := clustercache.GetID(args.APIEndpoint, args.ClusterNameOrID, clientWrapper)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	auxParams := clientWrapper.DefaultAuxiliaryParams()
	auxParams.ActivityName = activityName

	// First try v5.
	if args.Verbose {
		fmt.Println(color.WhiteString("Fetching details for cluster via v5 API endpoint."))
	}
	clusterV5, errV5 := clientWrapper.GetClusterV5(clusterID, auxParams)
	if errV5 == nil {
		requestBody := &models.V5ModifyClusterRequest{}
		{
			if args.Name != "" {
				requestBody.Name = args.Name
			}

			if args.MasterHA {
				capabilityService, err := capabilities.New(config.Config.Provider, clientWrapper)
				if err != nil {
					return nil, microerror.Mask(err)
				}
				haMastersEnabled, _ := capabilityService.HasCapability(clusterV5.Payload.ReleaseVersion, capabilities.HAMasters)

				if !haMastersEnabled {
					return nil, microerror.Mask(haMastersNotSupportedError)
				}

				requestBody.MasterNodes = &models.V5ModifyClusterRequestMasterNodes{
					HighAvailability: true,
				}
			}
		}

		if args.Verbose {
			fmt.Println(color.WhiteString("Sending cluster modification request to v5 endpoint."))
		}
		response, err := clientWrapper.ModifyClusterV5(clusterID, requestBody, auxParams)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		r := &result{}
		{
			if args.Name != "" {
				r.ClusterName = response.Payload.Name
			}
			if args.MasterHA {
				r.HasHAMaster = response.Payload.MasterNodes.HighAvailability
			}
		}

		return r, nil
	} else {
		if args.MasterHA {
			return nil, microerror.Mask(haMastersNotSupportedError)
		}
	}

	// Fallback: try v4.
	if args.Verbose {
		fmt.Println(color.WhiteString("No usable v5 response. Fetching details for cluster via v4 API endpoint."))
	}
	_, errV4 := clientWrapper.GetClusterV4(clusterID, auxParams)
	if errV4 == nil {
		requestBody := &models.V4ModifyClusterRequest{}
		if args.Name != "" {
			requestBody.Name = args.Name
		}

		if args.Verbose {
			fmt.Println(color.WhiteString("Sending cluster modification request to v4 endpoint."))
		}
		response, err := clientWrapper.ModifyClusterV4(clusterID, requestBody, auxParams)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		r := &result{
			ClusterName: response.Payload.Name,
		}

		return r, nil
	}

	// We return the last error here, representatively.
	return nil, microerror.Mask(errV4)
}

// updateLabels applies label changes to v5 clusters. v5 only!
func updateLabels(args Arguments) (*result, error) {
	clientWrapper, err := client.NewWithConfig(args.APIEndpoint, args.UserProvidedToken)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	clusterID, err := clustercache.GetID(args.APIEndpoint, args.ClusterNameOrID, clientWrapper)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	auxParams := clientWrapper.DefaultAuxiliaryParams()
	auxParams.ActivityName = activityName

	// verify this cluster exists and is a v5 cluster
	_, err = clientWrapper.GetClusterV5(clusterID, auxParams)
	if err != nil {
		return nil, microerror.Maskf(errors.ClusterNotFoundError, "cluster with id '%s' not found or not a v5 cluster", clusterID)
	}

	requestBody, err := modifyClusterLabelsRequestFromArguments(args.Labels)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	if args.Verbose {
		fmt.Println(color.WhiteString("Sending cluster modification request to setClusterLabels endpoint."))
	}
	response, err := clientWrapper.UpdateClusterLabels(clusterID, requestBody, auxParams)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r := &result{
		Labels: response.Payload.Labels,
	}

	return r, nil
}

func printResult(cmd *cobra.Command, positionalArgs []string) {
	result, err := updateCluster(arguments)
	if err != nil {
		client.HandleErrors(err)
		errors.HandleCommonErrors(err)

		headline := ""
		subtext := ""

		switch {
		case IsHAMastersNotSupported(err):
			var haMastersRequiredVersion string
			{
				for _, requiredRelease := range capabilities.HAMasters.RequiredReleasePerProvider {
					if requiredRelease.Provider == config.Config.Provider {
						haMastersRequiredVersion = requiredRelease.ReleaseVersion.String()

						break
					}
				}
			}

			headline = "Feature not supported"

			if haMastersRequiredVersion == "" {
				subtext = fmt.Sprintf("Master node high availability is not supported by your provider. (%s)", strings.ToUpper(config.Config.Provider))
			} else {
				subtext = fmt.Sprintf("Master node high availability is only supported by releases %s and higher.", haMastersRequiredVersion)
			}

		case errors.IsNoOpError(err):
			headline = "No flags specified"

		// If there are specific errors to handle, add them here.
		default:
			headline = err.Error()
		}

		// print output
		fmt.Println(color.RedString(headline))
		if subtext != "" {
			fmt.Println(subtext)
		}
		os.Exit(1)
	}

	fmt.Println(color.GreenString("Cluster '%s' has been modified.", arguments.ClusterNameOrID))
	if result.ClusterName != "" {
		fmt.Printf("New cluster name: '%s'\n", result.ClusterName)
	}
	if len(result.Labels) > 0 {
		fmt.Println("New cluster labels:")
		for key, label := range result.Labels {
			if strings.Contains(key, util.LabelFilterKeySubstring) == false {
				fmt.Printf("%s=%s\n", key, label)
			}
		}
	}
}
