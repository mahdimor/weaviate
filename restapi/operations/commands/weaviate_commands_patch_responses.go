/*                          _       _
 *__      _____  __ ___   ___  __ _| |_ ___
 *\ \ /\ / / _ \/ _` \ \ / / |/ _` | __/ _ \
 * \ V  V /  __/ (_| |\ V /| | (_| | ||  __/
 *  \_/\_/ \___|\__,_| \_/ |_|\__,_|\__\___|
 *
 * Copyright © 2016 Weaviate. All rights reserved.
 * LICENSE: https://github.com/weaviate/weaviate/blob/master/LICENSE
 * AUTHOR: Bob van Luijt (bob@weaviate.com)
 * See www.weaviate.com for details
 * See package.json for author and maintainer info
 * Contact: @weaviate_iot / yourfriends@weaviate.com
 */
 package commands

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	"github.com/weaviate/weaviate/models"
)

// WeaviateCommandsPatchOKCode is the HTTP code returned for type WeaviateCommandsPatchOK
const WeaviateCommandsPatchOKCode int = 200

/*WeaviateCommandsPatchOK Successful updated.

swagger:response weaviateCommandsPatchOK
*/
type WeaviateCommandsPatchOK struct {

	/*
	  In: Body
	*/
	Payload *models.Command `json:"body,omitempty"`
}

// NewWeaviateCommandsPatchOK creates WeaviateCommandsPatchOK with default headers values
func NewWeaviateCommandsPatchOK() *WeaviateCommandsPatchOK {
	return &WeaviateCommandsPatchOK{}
}

// WithPayload adds the payload to the weaviate commands patch o k response
func (o *WeaviateCommandsPatchOK) WithPayload(payload *models.Command) *WeaviateCommandsPatchOK {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the weaviate commands patch o k response
func (o *WeaviateCommandsPatchOK) SetPayload(payload *models.Command) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *WeaviateCommandsPatchOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(200)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// WeaviateCommandsPatchNotImplementedCode is the HTTP code returned for type WeaviateCommandsPatchNotImplemented
const WeaviateCommandsPatchNotImplementedCode int = 501

/*WeaviateCommandsPatchNotImplemented Not (yet) implemented.

swagger:response weaviateCommandsPatchNotImplemented
*/
type WeaviateCommandsPatchNotImplemented struct {
}

// NewWeaviateCommandsPatchNotImplemented creates WeaviateCommandsPatchNotImplemented with default headers values
func NewWeaviateCommandsPatchNotImplemented() *WeaviateCommandsPatchNotImplemented {
	return &WeaviateCommandsPatchNotImplemented{}
}

// WriteResponse to the client
func (o *WeaviateCommandsPatchNotImplemented) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(501)
}