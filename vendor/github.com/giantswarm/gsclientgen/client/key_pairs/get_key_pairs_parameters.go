// Code generated by go-swagger; DO NOT EDIT.

package key_pairs

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"
	"net/http"
	"time"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	cr "github.com/go-openapi/runtime/client"

	strfmt "github.com/go-openapi/strfmt"
)

// NewGetKeyPairsParams creates a new GetKeyPairsParams object
// with the default values initialized.
func NewGetKeyPairsParams() *GetKeyPairsParams {
	var ()
	return &GetKeyPairsParams{

		timeout: cr.DefaultTimeout,
	}
}

// NewGetKeyPairsParamsWithTimeout creates a new GetKeyPairsParams object
// with the default values initialized, and the ability to set a timeout on a request
func NewGetKeyPairsParamsWithTimeout(timeout time.Duration) *GetKeyPairsParams {
	var ()
	return &GetKeyPairsParams{

		timeout: timeout,
	}
}

// NewGetKeyPairsParamsWithContext creates a new GetKeyPairsParams object
// with the default values initialized, and the ability to set a context for a request
func NewGetKeyPairsParamsWithContext(ctx context.Context) *GetKeyPairsParams {
	var ()
	return &GetKeyPairsParams{

		Context: ctx,
	}
}

// NewGetKeyPairsParamsWithHTTPClient creates a new GetKeyPairsParams object
// with the default values initialized, and the ability to set a custom HTTPClient for a request
func NewGetKeyPairsParamsWithHTTPClient(client *http.Client) *GetKeyPairsParams {
	var ()
	return &GetKeyPairsParams{
		HTTPClient: client,
	}
}

/*GetKeyPairsParams contains all the parameters to send to the API endpoint
for the get key pairs operation typically these are written to a http.Request
*/
type GetKeyPairsParams struct {

	/*XGiantSwarmActivity
	  Name of an activity to track, like "list-clusters". This allows to
	analyze several API requests sent in context and gives an idea on
	the purpose.


	*/
	XGiantSwarmActivity *string
	/*XGiantSwarmCmdLine
	  If activity has been issued by a CLI, this header can contain the
	command line


	*/
	XGiantSwarmCmdLine *string
	/*XRequestID
	  A randomly generated key that can be used to track a request throughout
	services of Giant Swarm.


	*/
	XRequestID *string
	/*ClusterID
	  Cluster ID

	*/
	ClusterID string

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithTimeout adds the timeout to the get key pairs params
func (o *GetKeyPairsParams) WithTimeout(timeout time.Duration) *GetKeyPairsParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the get key pairs params
func (o *GetKeyPairsParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the get key pairs params
func (o *GetKeyPairsParams) WithContext(ctx context.Context) *GetKeyPairsParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the get key pairs params
func (o *GetKeyPairsParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the get key pairs params
func (o *GetKeyPairsParams) WithHTTPClient(client *http.Client) *GetKeyPairsParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the get key pairs params
func (o *GetKeyPairsParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithXGiantSwarmActivity adds the xGiantSwarmActivity to the get key pairs params
func (o *GetKeyPairsParams) WithXGiantSwarmActivity(xGiantSwarmActivity *string) *GetKeyPairsParams {
	o.SetXGiantSwarmActivity(xGiantSwarmActivity)
	return o
}

// SetXGiantSwarmActivity adds the xGiantSwarmActivity to the get key pairs params
func (o *GetKeyPairsParams) SetXGiantSwarmActivity(xGiantSwarmActivity *string) {
	o.XGiantSwarmActivity = xGiantSwarmActivity
}

// WithXGiantSwarmCmdLine adds the xGiantSwarmCmdLine to the get key pairs params
func (o *GetKeyPairsParams) WithXGiantSwarmCmdLine(xGiantSwarmCmdLine *string) *GetKeyPairsParams {
	o.SetXGiantSwarmCmdLine(xGiantSwarmCmdLine)
	return o
}

// SetXGiantSwarmCmdLine adds the xGiantSwarmCmdLine to the get key pairs params
func (o *GetKeyPairsParams) SetXGiantSwarmCmdLine(xGiantSwarmCmdLine *string) {
	o.XGiantSwarmCmdLine = xGiantSwarmCmdLine
}

// WithXRequestID adds the xRequestID to the get key pairs params
func (o *GetKeyPairsParams) WithXRequestID(xRequestID *string) *GetKeyPairsParams {
	o.SetXRequestID(xRequestID)
	return o
}

// SetXRequestID adds the xRequestId to the get key pairs params
func (o *GetKeyPairsParams) SetXRequestID(xRequestID *string) {
	o.XRequestID = xRequestID
}

// WithClusterID adds the clusterID to the get key pairs params
func (o *GetKeyPairsParams) WithClusterID(clusterID string) *GetKeyPairsParams {
	o.SetClusterID(clusterID)
	return o
}

// SetClusterID adds the clusterId to the get key pairs params
func (o *GetKeyPairsParams) SetClusterID(clusterID string) {
	o.ClusterID = clusterID
}

// WriteToRequest writes these params to a swagger request
func (o *GetKeyPairsParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

	if err := r.SetTimeout(o.timeout); err != nil {
		return err
	}
	var res []error

	if o.XGiantSwarmActivity != nil {

		// header param X-Giant-Swarm-Activity
		if err := r.SetHeaderParam("X-Giant-Swarm-Activity", *o.XGiantSwarmActivity); err != nil {
			return err
		}

	}

	if o.XGiantSwarmCmdLine != nil {

		// header param X-Giant-Swarm-CmdLine
		if err := r.SetHeaderParam("X-Giant-Swarm-CmdLine", *o.XGiantSwarmCmdLine); err != nil {
			return err
		}

	}

	if o.XRequestID != nil {

		// header param X-Request-ID
		if err := r.SetHeaderParam("X-Request-ID", *o.XRequestID); err != nil {
			return err
		}

	}

	// path param cluster_id
	if err := r.SetPathParam("cluster_id", o.ClusterID); err != nil {
		return err
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
