package commands

import (
	"fmt"
	"net/http"
	"os"

	"github.com/giantswarm/microerror"

	"github.com/fatih/color"
	"github.com/giantswarm/gsctl/client/clienterror"
	"github.com/giantswarm/gsctl/config"
	"github.com/spf13/cobra"
)

type deleteClusterArguments struct {
	// API endpoint
	apiEndpoint string
	// cluster ID to delete
	clusterID string
	// don't prompt
	force bool
	// auth scheme
	scheme string
	// auth token
	token string
	// verbosity
	verbose bool
}

func defaultDeleteClusterArguments() deleteClusterArguments {
	endpoint := config.Config.ChooseEndpoint(cmdAPIEndpoint)
	token := config.Config.ChooseToken(endpoint, cmdToken)
	scheme := config.Config.ChooseScheme(endpoint, cmdToken)

	return deleteClusterArguments{
		apiEndpoint: endpoint,
		clusterID:   cmdClusterID,
		force:       cmdForce,
		scheme:      scheme,
		token:       token,
		verbose:     cmdVerbose,
	}
}

const (
	deleteClusterActivityName = "delete-cluster"
)

var (

	// DeleteClusterCommand performs the "delete cluster" function
	DeleteClusterCommand = &cobra.Command{
		Use:   "cluster",
		Short: "Delete cluster",
		Long: `Deletes a Kubernetes cluster.

Caution: This will terminate all workloads on the cluster. Data stored on the
worker nodes will be lost. There is no way to undo this.

Example:

	gsctl delete cluster -c c7t2o`,
		PreRun: deleteClusterValidationOutput,
		Run:    deleteClusterExecutionOutput,
	}
)

func init() {
	DeleteClusterCommand.Flags().StringVarP(&cmdClusterID, "cluster", "c", "", "ID of the cluster to delete")
	DeleteClusterCommand.Flags().BoolVarP(&cmdForce, "force", "", false, "If set, no interactive confirmation will be required (risky!).")

	DeleteClusterCommand.MarkFlagRequired("cluster")

	DeleteCommand.AddCommand(DeleteClusterCommand)
}

// deleteClusterValidationOutput runs our pre-checks.
// If errors occur, error info is printed to STDOUT/STDERR
// and the program will exit with non-zero exit codes.
func deleteClusterValidationOutput(cmd *cobra.Command, args []string) {
	dca := defaultDeleteClusterArguments()

	err := validateDeleteClusterPreConditions(dca)
	if err != nil {
		handleCommonErrors(err)

		var headline = ""
		var subtext = ""

		switch {
		case err.Error() == "":
			return
		case IsCouldNotDeleteClusterError(err):
			headline = "The cluster could not be deleted."
			subtext = "You might try again in a few moments. If that doesn't work, please contact the Giant Swarm support team."
			subtext += " Sorry for the inconvenience!"
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
}

// validateDeleteClusterPreConditions checks preconditions and returns an error in case
func validateDeleteClusterPreConditions(args deleteClusterArguments) error {
	if args.clusterID == "" {
		return microerror.Mask(clusterIDMissingError)
	}
	if config.Config.Token == "" && args.token == "" {
		return microerror.Mask(notLoggedInError)
	}
	return nil
}

// interprets arguments/flags, eventually submits delete request
func deleteClusterExecutionOutput(cmd *cobra.Command, args []string) {
	dca := defaultDeleteClusterArguments()
	deleted, err := deleteCluster(dca)
	if err != nil {
		handleCommonErrors(err)

		fmt.Println(color.RedString(err.Error()))
		os.Exit(1)
	}

	// non-error output
	if deleted {
		fmt.Println(color.GreenString("The cluster with ID '%s' will be deleted as soon as all workloads are terminated.", dca.clusterID))
	} else {
		if dca.verbose {
			fmt.Println(color.GreenString("Aborted."))
		}
	}
}

// deleteCluster performs the cluster deletion API call
//
// The returned tuple contains:
// - bool: true if cluster will reall ybe deleted, false otherwise
// - error: The error that has occurred (or nil)
//
func deleteCluster(args deleteClusterArguments) (bool, error) {
	// confirmation
	if !args.force {
		confirmed := askForConfirmation("Do you really want to delete cluster '" + args.clusterID + "'?")
		if !confirmed {
			return false, nil
		}
	}

	auxParams := ClientV2.DefaultAuxiliaryParams()
	auxParams.ActivityName = deleteClusterActivityName

	// perform API call
	_, err := ClientV2.DeleteCluster(args.clusterID, auxParams)
	if err != nil {
		// create specific error types for cases we care about
		if clientErr, ok := err.(*clienterror.APIError); ok {
			if clientErr.HTTPStatusCode == http.StatusForbidden {
				return false, microerror.Mask(accessForbiddenError)
			}
		}

		return false, microerror.Maskf(couldNotDeleteClusterError, err.Error())
	}

	return true, nil
}
