// Code generated by go-swagger; DO NOT EDIT.

package users

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"io"

	"github.com/go-openapi/runtime"

	strfmt "github.com/go-openapi/strfmt"

	models "github.com/giantswarm/gsclientgen/models"
)

// ModifyPasswordReader is a Reader for the ModifyPassword structure.
type ModifyPasswordReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *ModifyPasswordReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {

	case 202:
		result := NewModifyPasswordAccepted()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil

	case 401:
		result := NewModifyPasswordUnauthorized()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result

	case 404:
		result := NewModifyPasswordNotFound()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result

	default:
		result := NewModifyPasswordDefault(response.Code())
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		if response.Code()/100 == 2 {
			return result, nil
		}
		return nil, result
	}
}

// NewModifyPasswordAccepted creates a ModifyPasswordAccepted with default headers values
func NewModifyPasswordAccepted() *ModifyPasswordAccepted {
	return &ModifyPasswordAccepted{}
}

/*ModifyPasswordAccepted handles this case with default header values.

Accepted
*/
type ModifyPasswordAccepted struct {
	Payload *models.V4GenericResponse
}

func (o *ModifyPasswordAccepted) Error() string {
	return fmt.Sprintf("[POST /v4/users/{email}/password/][%d] modifyPasswordAccepted  %+v", 202, o.Payload)
}

func (o *ModifyPasswordAccepted) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.V4GenericResponse)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewModifyPasswordUnauthorized creates a ModifyPasswordUnauthorized with default headers values
func NewModifyPasswordUnauthorized() *ModifyPasswordUnauthorized {
	return &ModifyPasswordUnauthorized{}
}

/*ModifyPasswordUnauthorized handles this case with default header values.

Permission denied
*/
type ModifyPasswordUnauthorized struct {
	Payload *models.V4GenericResponse
}

func (o *ModifyPasswordUnauthorized) Error() string {
	return fmt.Sprintf("[POST /v4/users/{email}/password/][%d] modifyPasswordUnauthorized  %+v", 401, o.Payload)
}

func (o *ModifyPasswordUnauthorized) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.V4GenericResponse)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewModifyPasswordNotFound creates a ModifyPasswordNotFound with default headers values
func NewModifyPasswordNotFound() *ModifyPasswordNotFound {
	return &ModifyPasswordNotFound{}
}

/*ModifyPasswordNotFound handles this case with default header values.

User not found
*/
type ModifyPasswordNotFound struct {
	Payload *models.V4GenericResponse
}

func (o *ModifyPasswordNotFound) Error() string {
	return fmt.Sprintf("[POST /v4/users/{email}/password/][%d] modifyPasswordNotFound  %+v", 404, o.Payload)
}

func (o *ModifyPasswordNotFound) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.V4GenericResponse)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewModifyPasswordDefault creates a ModifyPasswordDefault with default headers values
func NewModifyPasswordDefault(code int) *ModifyPasswordDefault {
	return &ModifyPasswordDefault{
		_statusCode: code,
	}
}

/*ModifyPasswordDefault handles this case with default header values.

Error
*/
type ModifyPasswordDefault struct {
	_statusCode int

	Payload *models.V4GenericResponse
}

// Code gets the status code for the modify password default response
func (o *ModifyPasswordDefault) Code() int {
	return o._statusCode
}

func (o *ModifyPasswordDefault) Error() string {
	return fmt.Sprintf("[POST /v4/users/{email}/password/][%d] modifyPassword default  %+v", o._statusCode, o.Payload)
}

func (o *ModifyPasswordDefault) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.V4GenericResponse)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}