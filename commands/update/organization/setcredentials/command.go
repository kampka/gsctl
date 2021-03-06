package setcredentials

import (
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/giantswarm/gscliauth/config"
	"github.com/giantswarm/gsclientgen/v2/models"
	"github.com/giantswarm/microerror"
	"github.com/spf13/cobra"

	"github.com/giantswarm/gsctl/client"
	"github.com/giantswarm/gsctl/client/clienterror"
	"github.com/giantswarm/gsctl/commands/errors"
	"github.com/giantswarm/gsctl/flags"
)

const (
	activityName = "set-org-credentials"
)

var (
	// Command performs the "update organization set-credentials" function
	Command = &cobra.Command{
		Use:     "set-credentials",
		Aliases: []string{"sc"},
		Short:   "Set credentials of an organization",
		Long: `Set the credentials used to create and operate the clusters of an organization.

Setting credentials of an organization will result in all future clusters
being run in the account/subscription referenced by the credentials. Once
credentials are set for an organization, this currently cannot be undone.

For details on how to prepare the account/subscription, consult the documentation at

  - https://docs.giantswarm.io/getting-started/cloud-provider-accounts/aws/ (AWS)
  - https://docs.giantswarm.io/getting-started/cloud-provider-accounts/azure/ (Azure)

`,
		Example: `
  gsctl update organization set-credentials -o acme \
    --aws-operator-role arn:aws:iam::<AWS-ACCOUNT-ID>:role/GiantSwarmAWSOperator \
    --aws-admin-role arn:aws:iam::<AWS-ACCOUNT-ID>:role/GiantSwarmAdmin

  gsctl update organization set-credentials -o acme \
    --azure-subscription-id <AZURE-SUBSCRIPTION-ID> \
    --azure-tenant-id <AZURE-TENANT-ID> \
    --azure-client-id <AZURE-CLIENT-ID> \
    --azure-secret-key <AZURE-SECRET-KEY>
`,

		// PreRun checks a few general things, like authentication and flags
		// compatibility.
		PreRun: printValidation,

		// Run calls the business function and prints results and errors.
		Run: printResult,
	}

	// AWS role ARN flags
	cmdAWSOperatorRoleARN string
	cmdAWSAdminRoleARN    string

	// Azure-related flags
	cmdAzureSubscriptionID string
	cmdAzureTenantID       string
	cmdAzureClientID       string
	cmdAzureSecretKey      string

	// Here we briefly store the info which provider we are dealing with
	provider string

	arguments Arguments
)

type Arguments struct {
	apiEndpoint         string
	authToken           string
	awsAdminRole        string
	awsOperatorRole     string
	azureClientID       string
	azureSecretKey      string
	azureSubscriptionID string
	azureTenantID       string
	organizationID      string
	scheme              string
	userProvidedToken   string
	verbose             bool
}

type setOrgCredentialsResult struct {
	credentialID string
}

func init() {
	Command.Flags().StringVarP(&flags.OrganizationID, "organization", "o", "", "ID of the organization to set credentials for")
	Command.Flags().StringVarP(&cmdAWSOperatorRoleARN, "aws-operator-role", "", "", "AWS ARN of the role to use for operating clusters")
	Command.Flags().StringVarP(&cmdAWSAdminRoleARN, "aws-admin-role", "", "", "AWS ARN of the role to be used by Giant Swarm staff")
	Command.Flags().StringVarP(&cmdAzureSubscriptionID, "azure-subscription-id", "", "", "ID of the Azure subscription to run clusters in")
	Command.Flags().StringVarP(&cmdAzureTenantID, "azure-tenant-id", "", "", "ID of the Azure tenant to run clusters in")
	Command.Flags().StringVarP(&cmdAzureClientID, "azure-client-id", "", "", "ID of the Azure service principal to use for operating clusters")
	Command.Flags().StringVarP(&cmdAzureSecretKey, "azure-secret-key", "", "", "Secret key for the Azure service principal to use for operating clusters")
}

func collectArguments() Arguments {
	endpoint := config.Config.ChooseEndpoint(flags.APIEndpoint)
	token := config.Config.ChooseToken(endpoint, flags.Token)
	scheme := config.Config.ChooseScheme(endpoint, flags.Token)

	return Arguments{
		apiEndpoint:         endpoint,
		authToken:           token,
		awsAdminRole:        cmdAWSAdminRoleARN,
		awsOperatorRole:     cmdAWSOperatorRoleARN,
		azureClientID:       cmdAzureClientID,
		azureSecretKey:      cmdAzureSecretKey,
		azureSubscriptionID: cmdAzureSubscriptionID,
		azureTenantID:       cmdAzureTenantID,
		organizationID:      flags.OrganizationID,
		scheme:              scheme,
		userProvidedToken:   flags.Token,
		verbose:             flags.Verbose,
	}
}

func printValidation(cmd *cobra.Command, cmdLineArgs []string) {
	arguments = collectArguments()
	err := verifyPreconditions(arguments)

	if err == nil {
		return
	}

	client.HandleErrors(err)
	errors.HandleCommonErrors(err)

	// From here on we handle errors that can only occur in this command
	headline := ""
	subtext := ""

	switch {
	case errors.IsOrganizationNotSpecifiedError(err):
		headline = "No organization given"
		subtext = "Please specify the organization to set credentials for using the -o|--organization flag."
	case errors.IsProviderNotSupportedError(err):
		headline = "Unsupported provider"
		subtext = "Setting credentials is only supported on AWS and Azure installations."
	case errors.IsRequiredFlagMissingError(err):
		headline = "Missing flag: " + err.Error()
		subtext = "Please use --help to see details regarding the command's usage."
	case errors.IsConflictingFlagsError(err):
		headline = "Conflicting flags"
		subtext = "Please use only AWS or Azure related flags with this installation. See --help for details."
	case errors.IsOrganizationNotFoundError(err):
		headline = fmt.Sprintf("Organization '%s' not found", arguments.organizationID)
		subtext = "The specified organization does not exist, or you are not a member. Please check the exact upper/lower case spelling."
		subtext += "\nUse 'gsctl list organizations' to list all organizations."
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

func verifyPreconditions(args Arguments) error {
	if args.apiEndpoint == "" {
		return microerror.Mask(errors.EndpointMissingError)
	}
	if args.organizationID == "" {
		return microerror.Mask(errors.OrganizationNotSpecifiedError)
	}
	if config.Config.Token == "" && args.authToken == "" {
		return microerror.Mask(errors.NotLoggedInError)
	}

	// get installation's provider (supported: aws, azure)
	if args.verbose {
		fmt.Println(color.WhiteString("Determining which provider this installation uses"))
	}

	clientWrapper, err := client.NewWithConfig(args.apiEndpoint, args.userProvidedToken)
	if err != nil {
		return microerror.Mask(err)
	}

	auxParams := clientWrapper.DefaultAuxiliaryParams()
	auxParams.ActivityName = activityName

	response, err := clientWrapper.GetInfo(auxParams)
	if err != nil {
		if clienterror.IsUnauthorizedError(err) {
			return microerror.Mask(errors.NotAuthorizedError)
		}
		if clienterror.IsAccessForbiddenError(err) {
			return microerror.Mask(errors.AccessForbiddenError)
		}

		return microerror.Mask(err)
	}

	provider = response.Payload.General.Provider

	if provider != "aws" && provider != "azure" {
		return microerror.Mask(errors.ProviderNotSupportedError)
	}

	// check flags based on provider
	{
		if provider == "aws" {
			if args.awsAdminRole == "" {
				return microerror.Maskf(errors.RequiredFlagMissingError, "--aws-admin-role")
			}
			if args.awsOperatorRole == "" {
				return microerror.Maskf(errors.RequiredFlagMissingError, "--aws-operator-role")
			}

			// conflicts
			if args.azureClientID != "" || args.azureSecretKey != "" || args.azureSubscriptionID != "" || args.azureTenantID != "" {
				return microerror.Maskf(errors.ConflictingFlagsError, "Azure-related flags not allowed here")
			}
		}
		if provider == "azure" {
			if args.azureClientID == "" {
				return microerror.Maskf(errors.RequiredFlagMissingError, "--azure-client-id")
			}
			if args.azureSecretKey == "" {
				return microerror.Maskf(errors.RequiredFlagMissingError, "--azure-secret-key")
			}
			if args.azureSubscriptionID == "" {
				return microerror.Maskf(errors.RequiredFlagMissingError, "--azure-subscription-id")
			}
			if args.azureTenantID == "" {
				return microerror.Maskf(errors.RequiredFlagMissingError, "--azure-tenant-id")
			}

			// conflicts
			if args.awsAdminRole != "" || args.awsOperatorRole != "" {
				return microerror.Maskf(errors.ConflictingFlagsError, "AWS-related flags not allowed here")
			}
		}
	}

	// check organization membership and existence
	if args.verbose {
		fmt.Println(color.WhiteString("Verify organization membership"))
	}
	orgsResponse, err := clientWrapper.GetOrganizations(auxParams)
	{
		if err != nil {
			if clienterror.IsUnauthorizedError(err) {
				return microerror.Mask(errors.NotAuthorizedError)
			}
			if clienterror.IsAccessForbiddenError(err) {
				return microerror.Mask(errors.AccessForbiddenError)
			}

			return microerror.Mask(err)
		}

		foundOrg := false
		for _, org := range orgsResponse.Payload {
			if org.ID == args.organizationID {
				foundOrg = true
			}
		}
		if !foundOrg {
			return microerror.Mask(errors.OrganizationNotFoundError)
		}
	}

	return nil
}

// printResult calls the busniness function and produces
// meanigful terminal output.
func printResult(cmd *cobra.Command, cmdLineArgs []string) {
	result, err := setOrgCredentials(arguments)

	if err != nil {
		client.HandleErrors(err)
		errors.HandleCommonErrors(err)

		// From here on we handle errors that can only occur in this command
		headline := ""
		subtext := ""

		switch {
		case errors.IsCredentialsAlreadySetError(err):
			headline = "Credentials already set"
			subtext = fmt.Sprintf("Organization '%s' has credentials already. These cannot be overwritten.", arguments.organizationID)
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

	// success
	fmt.Println(color.GreenString("Credentials set successfully"))
	fmt.Printf("The credentials are stored with the unique ID '%s'.\n", result.credentialID)
}

// setOrgCredentials performs the API call and provides a result.
func setOrgCredentials(args Arguments) (*setOrgCredentialsResult, error) {
	// build request body based on provider
	requestBody := &models.V4AddCredentialsRequest{Provider: &provider}
	if provider == "aws" {
		requestBody.Aws = &models.V4AddCredentialsRequestAws{
			Roles: &models.V4AddCredentialsRequestAwsRoles{
				Admin:       &args.awsAdminRole,
				Awsoperator: &args.awsOperatorRole,
			},
		}
	} else if provider == "azure" {
		requestBody.Azure = &models.V4AddCredentialsRequestAzure{
			Credential: &models.V4AddCredentialsRequestAzureCredential{
				SubscriptionID: &args.azureSubscriptionID,
				TenantID:       &args.azureTenantID,
				ClientID:       &args.azureClientID,
				SecretKey:      &args.azureSecretKey,
			},
		}
	}

	if args.verbose {
		fmt.Println(color.WhiteString("Sending API request to set credentials"))
	}

	clientWrapper, err := client.NewWithConfig(args.apiEndpoint, args.userProvidedToken)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	auxParams := clientWrapper.DefaultAuxiliaryParams()
	auxParams.ActivityName = activityName

	response, err := clientWrapper.SetCredentials(args.organizationID, requestBody, auxParams)
	if err != nil {
		if clienterror.IsConflictError(err) {
			return nil, microerror.Mask(errors.CredentialsAlreadySetError)
		}

		return nil, microerror.Mask(err)
	}

	// Location header returned is in the format
	// /v4/organizations/myorg/credentials/{credential_id}/
	segments := strings.Split(response.Location, "/")
	result := &setOrgCredentialsResult{
		credentialID: segments[len(segments)-2],
	}

	return result, nil
}
