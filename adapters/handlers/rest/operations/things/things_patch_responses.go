//                           _       _
// __      _____  __ ___   ___  __ _| |_ ___
// \ \ /\ / / _ \/ _` \ \ / / |/ _` | __/ _ \
//  \ V  V /  __/ (_| |\ V /| | (_| | ||  __/
//   \_/\_/ \___|\__,_| \_/ |_|\__,_|\__\___|
//
//  Copyright © 2016 - 2019 SeMI Holding B.V. (registered @ Dutch Chamber of Commerce no 75221632). All rights reserved.
//  LICENSE WEAVIATE OPEN SOURCE: https://www.semi.technology/playbook/playbook/contract-weaviate-OSS.html
//  LICENSE WEAVIATE ENTERPRISE: https://www.semi.technology/playbook/contract-weaviate-enterprise.html
//  CONCEPT: Bob van Luijt (@bobvanluijt)
//  CONTACT: hello@semi.technology
//

// Code generated by go-swagger; DO NOT EDIT.

package things

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	models "github.com/semi-technologies/weaviate/entities/models"
)

// ThingsPatchNoContentCode is the HTTP code returned for type ThingsPatchNoContent
const ThingsPatchNoContentCode int = 204

/*ThingsPatchNoContent Successfully applied. No content returned

swagger:response thingsPatchNoContent
*/
type ThingsPatchNoContent struct {
}

// NewThingsPatchNoContent creates ThingsPatchNoContent with default headers values
func NewThingsPatchNoContent() *ThingsPatchNoContent {

	return &ThingsPatchNoContent{}
}

// WriteResponse to the client
func (o *ThingsPatchNoContent) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.Header().Del(runtime.HeaderContentType) //Remove Content-Type on empty responses

	rw.WriteHeader(204)
}

// ThingsPatchBadRequestCode is the HTTP code returned for type ThingsPatchBadRequest
const ThingsPatchBadRequestCode int = 400

/*ThingsPatchBadRequest The patch-JSON is malformed.

swagger:response thingsPatchBadRequest
*/
type ThingsPatchBadRequest struct {
}

// NewThingsPatchBadRequest creates ThingsPatchBadRequest with default headers values
func NewThingsPatchBadRequest() *ThingsPatchBadRequest {

	return &ThingsPatchBadRequest{}
}

// WriteResponse to the client
func (o *ThingsPatchBadRequest) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.Header().Del(runtime.HeaderContentType) //Remove Content-Type on empty responses

	rw.WriteHeader(400)
}

// ThingsPatchUnauthorizedCode is the HTTP code returned for type ThingsPatchUnauthorized
const ThingsPatchUnauthorizedCode int = 401

/*ThingsPatchUnauthorized Unauthorized or invalid credentials.

swagger:response thingsPatchUnauthorized
*/
type ThingsPatchUnauthorized struct {
}

// NewThingsPatchUnauthorized creates ThingsPatchUnauthorized with default headers values
func NewThingsPatchUnauthorized() *ThingsPatchUnauthorized {

	return &ThingsPatchUnauthorized{}
}

// WriteResponse to the client
func (o *ThingsPatchUnauthorized) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.Header().Del(runtime.HeaderContentType) //Remove Content-Type on empty responses

	rw.WriteHeader(401)
}

// ThingsPatchForbiddenCode is the HTTP code returned for type ThingsPatchForbidden
const ThingsPatchForbiddenCode int = 403

/*ThingsPatchForbidden Forbidden

swagger:response thingsPatchForbidden
*/
type ThingsPatchForbidden struct {

	/*
	  In: Body
	*/
	Payload *models.ErrorResponse `json:"body,omitempty"`
}

// NewThingsPatchForbidden creates ThingsPatchForbidden with default headers values
func NewThingsPatchForbidden() *ThingsPatchForbidden {

	return &ThingsPatchForbidden{}
}

// WithPayload adds the payload to the things patch forbidden response
func (o *ThingsPatchForbidden) WithPayload(payload *models.ErrorResponse) *ThingsPatchForbidden {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the things patch forbidden response
func (o *ThingsPatchForbidden) SetPayload(payload *models.ErrorResponse) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *ThingsPatchForbidden) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(403)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// ThingsPatchNotFoundCode is the HTTP code returned for type ThingsPatchNotFound
const ThingsPatchNotFoundCode int = 404

/*ThingsPatchNotFound Successful query result but no resource was found.

swagger:response thingsPatchNotFound
*/
type ThingsPatchNotFound struct {
}

// NewThingsPatchNotFound creates ThingsPatchNotFound with default headers values
func NewThingsPatchNotFound() *ThingsPatchNotFound {

	return &ThingsPatchNotFound{}
}

// WriteResponse to the client
func (o *ThingsPatchNotFound) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.Header().Del(runtime.HeaderContentType) //Remove Content-Type on empty responses

	rw.WriteHeader(404)
}

// ThingsPatchUnprocessableEntityCode is the HTTP code returned for type ThingsPatchUnprocessableEntity
const ThingsPatchUnprocessableEntityCode int = 422

/*ThingsPatchUnprocessableEntity The patch-JSON is valid but unprocessable.

swagger:response thingsPatchUnprocessableEntity
*/
type ThingsPatchUnprocessableEntity struct {

	/*
	  In: Body
	*/
	Payload *models.ErrorResponse `json:"body,omitempty"`
}

// NewThingsPatchUnprocessableEntity creates ThingsPatchUnprocessableEntity with default headers values
func NewThingsPatchUnprocessableEntity() *ThingsPatchUnprocessableEntity {

	return &ThingsPatchUnprocessableEntity{}
}

// WithPayload adds the payload to the things patch unprocessable entity response
func (o *ThingsPatchUnprocessableEntity) WithPayload(payload *models.ErrorResponse) *ThingsPatchUnprocessableEntity {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the things patch unprocessable entity response
func (o *ThingsPatchUnprocessableEntity) SetPayload(payload *models.ErrorResponse) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *ThingsPatchUnprocessableEntity) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(422)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// ThingsPatchInternalServerErrorCode is the HTTP code returned for type ThingsPatchInternalServerError
const ThingsPatchInternalServerErrorCode int = 500

/*ThingsPatchInternalServerError An error has occurred while trying to fulfill the request. Most likely the ErrorResponse will contain more information about the error.

swagger:response thingsPatchInternalServerError
*/
type ThingsPatchInternalServerError struct {

	/*
	  In: Body
	*/
	Payload *models.ErrorResponse `json:"body,omitempty"`
}

// NewThingsPatchInternalServerError creates ThingsPatchInternalServerError with default headers values
func NewThingsPatchInternalServerError() *ThingsPatchInternalServerError {

	return &ThingsPatchInternalServerError{}
}

// WithPayload adds the payload to the things patch internal server error response
func (o *ThingsPatchInternalServerError) WithPayload(payload *models.ErrorResponse) *ThingsPatchInternalServerError {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the things patch internal server error response
func (o *ThingsPatchInternalServerError) SetPayload(payload *models.ErrorResponse) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *ThingsPatchInternalServerError) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(500)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}
