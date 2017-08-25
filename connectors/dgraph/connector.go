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
 * Contact: @weaviate_iot / yourfriends@weaviate.com
 */

package dgraph

import (
	"context"
	"encoding/json"
	errors_ "errors"
	"fmt"
	"github.com/go-openapi/strfmt"
	"google.golang.org/grpc"
	"io"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"net/url"
	"os"
	"strings"

	dgraphClient "github.com/dgraph-io/dgraph/client"
	"github.com/dgraph-io/dgraph/protos"
	"github.com/dgraph-io/dgraph/types"
	gouuid "github.com/satori/go.uuid"

	"github.com/weaviate/weaviate/connectors/config"
	"github.com/weaviate/weaviate/connectors/utils"
	"github.com/weaviate/weaviate/models"
	"github.com/weaviate/weaviate/schema"
)

// Dgraph has some basic variables.
type Dgraph struct {
	client *dgraphClient.Dgraph
	kind   string

	actionSchema schemaProperties
	thingSchema  schemaProperties
}

type schemaProperties struct {
	localFile      string
	configLocation string
	schema         schema.Schema
}

// GetName returns a unique connector name
func (f *Dgraph) GetName() string {
	return "dgraph"
}

// SetConfig is used to fill in a struct with config variables
func (f *Dgraph) SetConfig(configInput connectorConfig.Environment) {
	f.thingSchema.configLocation = configInput.Schemas.Thing
	f.actionSchema.configLocation = configInput.Schemas.Action
}

// Connect creates connection and tables if not already available
func (f *Dgraph) Connect() error {
	// Connect with Dgraph host, create connection dail
	// TODO: Put hostname in config using the SetConfig function (configInput.Database.DatabaseConfig and making custom struct)
	dgraphGrpcAddress := "127.0.0.1:9080"

	conn, err := grpc.Dial(dgraphGrpcAddress, grpc.WithInsecure())
	if err != nil {
		log.Println("dail error", err)
	}
	// defer conn.Close()

	// Create temp-folder for caching
	dir, err := ioutil.TempDir("temp", "weaviate_dgraph")

	if err != nil {
		return err
	}
	// defer os.RemoveAll(dir)

	// Set custom options
	var options = dgraphClient.BatchMutationOptions{
		Size:          100,
		Pending:       100,
		PrintCounters: true,
		MaxRetries:    math.MaxUint32,
		Ctx:           f.getContext(),
	}

	f.client = dgraphClient.NewDgraphClient([]*grpc.ClientConn{conn}, options, dir)
	// defer f.client.Close()

	return nil
}

// Init creates a root key, normally this should be validaded, but because it is an indgraph DB it is created always
func (f *Dgraph) Init() error {
	// Generate a basic DB object and print it's key.
	// dbObject := connector_utils.CreateFirstUserObject()

	configFiles := []*schemaProperties{
		&f.actionSchema,
		&f.thingSchema,
	}

	// Get all classes to verify we do not create duplicates
	allClasses, err := f.getAllClasses()

	// Init flush variable
	flushIt := false

	for _, cfv := range configFiles {
		// Validate if given location is URL or local file
		_, err := url.ParseRequestURI(cfv.configLocation)

		// With no error, it is an URL
		if err == nil {
			log.Println("Downloading schema file...")
			cfv.localFile = "temp/schema" + string(connector_utils.GenerateUUID()) + ".json"

			// Create local file
			schemaFile, _ := os.Create(cfv.localFile)
			defer schemaFile.Close()

			// Get the file from online
			resp, err := http.Get(cfv.configLocation)
			if err != nil {
				return err
			}
			defer resp.Body.Close()

			// Write file to local file
			b, _ := io.Copy(schemaFile, resp.Body)
			log.Println("Download complete, file size: ", b)
		} else {
			log.Println("Given schema location is not a valid URL, looking for local file...")

			// Given schema location is not a valid URL, assume it is a local file
			cfv.localFile = cfv.configLocation
		}

		// Read local file which is either just downloaded or given in config.
		log.Println("Read local file...")
		fileContents, err := ioutil.ReadFile(cfv.localFile)

		// Return error when error is given reading file.
		if err != nil {
			return err
		}

		// Merge JSON into Schema objects
		err = json.Unmarshal([]byte(fileContents), &cfv.schema)
		log.Println("File is loaded.")

		// Return error when error is given reading file.
		if err != nil {
			log.Println("Can not parse schema.")
			return err
		}

		// Add schema to database
		for _, class := range cfv.schema.Classes {
			// for _, prop := range class.Properties {
			for _ = range class.Properties {
				// Add Dgraph-schema for every property of individual nodes
				// err = f.client.AddSchema(protos.SchemaUpdate{
				// 	Predicate: "schema." + prop.Name,
				// 	ValueType: uint32(types.UidID),
				// 	Directive: protos.SchemaUpdate_REVERSE,
				// })

				// if err != nil {
				// 	return err
				// }

				// TODO: Add specific schema for datatypes
				// http://schema.org/DataType
			}

			// Create the class name by concatinating context with class
			// TODO: Do not concatinate, split and add classname and context seperatly to node?
			className := cfv.schema.Context + "/" + class.Class

			if _, err := f.getClassFromResult(className, allClasses); err != nil {
				// Add node and edge for every class
				// TODO: Only add if not exists
				node, err := f.client.NodeBlank(class.Class)

				if err != nil {
					return err
				}

				// Add class edge
				edge := node.Edge("class")
				err = edge.SetValueString(className)

				if err != nil {
					return err
				}

				// Add class edge to batch
				err = f.client.BatchSet(edge)

				if err != nil {
					return err
				}

				flushIt = true
			}
		}
	}

	// Add class schema in Dgraph
	if err := f.client.AddSchema(protos.SchemaUpdate{
		Predicate: "class",
		ValueType: uint32(types.StringID),
		Tokenizer: []string{"exact", "term"},
		Directive: protos.SchemaUpdate_INDEX,
		Count:     true,
	}); err != nil {
		return err
	}

	// Add type schema in Dgraph
	if err := f.client.AddSchema(protos.SchemaUpdate{
		Predicate: "type",
		ValueType: uint32(types.UidID),
		Directive: protos.SchemaUpdate_REVERSE,
		Count:     true,
	}); err != nil {
		return err
	}

	// Add ID schema in Dgraph
	if err := f.client.AddSchema(protos.SchemaUpdate{
		Predicate: "id",
		ValueType: uint32(types.UidID),
		Directive: protos.SchemaUpdate_REVERSE,
		Count:     true,
	}); err != nil {
		return err
	}

	// Add UUID schema in Dgraph
	if err := f.client.AddSchema(protos.SchemaUpdate{
		Predicate: "uuid",
		ValueType: uint32(types.StringID),
		Tokenizer: []string{"exact", "term"},
		Directive: protos.SchemaUpdate_INDEX,
		Count:     true,
	}); err != nil {
		return err
	}

	// Add 'action.of' schema in Dgraph
	if err := f.client.AddSchema(protos.SchemaUpdate{
		Predicate: "action.of",
		ValueType: uint32(types.UidID),
		Directive: protos.SchemaUpdate_REVERSE,
		Count:     true,
	}); err != nil {
		return err
	}

	// Add 'action.target' schema in Dgraph
	if err := f.client.AddSchema(protos.SchemaUpdate{
		Predicate: "action.target",
		ValueType: uint32(types.UidID),
		Directive: protos.SchemaUpdate_REVERSE,
		Count:     true,
	}); err != nil {
		return err
	}

	// Add search possibilities for every "timing"
	thingTimings := []string{
		"creationTimeMs",
		"lastSeenTimeMs",
		"lastUpdateTimeMs",
		"lastUseTimeMs",
	}

	// For every timing, add them in the DB
	for _, ms := range thingTimings {
		if err := f.client.AddSchema(protos.SchemaUpdate{
			Predicate: ms,
			ValueType: uint32(types.IntID),
			Tokenizer: []string{"int"},
			Directive: protos.SchemaUpdate_INDEX,
			Count:     true,
		}); err != nil {
			return err
		}
	}

	// Call flush to flush buffers after all mutations are added
	if flushIt {
		err = f.client.BatchFlush()

		if err != nil {
			return err
		}
	}

	return nil
}

// AddThing adds a thing to the Dgraph database with the given UUID
func (f *Dgraph) AddThing(thing *models.ThingCreate, UUID strfmt.UUID) error {
	// TODO: make type interactive
	newNode, err := f.addNewNode(thing.AtContext, "Event", UUID)

	if err != nil {
		return err
	}

	// Add all given information to the new node
	err = f.updateThingNodeEdges(newNode, &thing.Thing)

	if err != nil {
		return err
	}

	return nil

	// TODO: Reset batch before and flush after every function??
}

// GetThing returns the thing in the ThingGetResponse format
func (f *Dgraph) GetThing(UUID strfmt.UUID) (models.ThingGetResponse, error) {
	// Initialize response
	thingResponse := models.ThingGetResponse{}
	thingResponse.Schema = map[string]models.JSONObject{}

	// Do a query to get all node-information based on the given UUID
	variables := make(map[string]string)
	variables["$uuid"] = string(UUID)

	req := dgraphClient.Req{}
	req.SetQueryWithVariables(`{ 
		get(func: eq(uuid, $uuid)) {
			uuid
			~id {
				expand(_all_) {
					expand(_all_)
				}
			}
		}
	}`, variables)

	// Run query created above
	resp, err := f.client.Run(f.getContext(), &req)
	if err != nil {
		return thingResponse, err
	}

	// Merge the results into the model to return
	nodes := resp.GetN()
	for _, node := range nodes {
		f.mergeThingNodeInResponse(node, &thingResponse)
	}

	return thingResponse, nil
}

// ListThings returns the thing in the ThingGetResponse format
func (f *Dgraph) ListThings(limit int, page int) (models.ThingsListResponse, error) {
	// Initialize response
	thingsResponse := models.ThingsListResponse{}
	thingsResponse.Things = make([]*models.ThingGetResponse, limit)

	// Do a query to get all node-information
	req := dgraphClient.Req{}
	req.SetQuery(fmt.Sprintf(`{ 
		things(func: has(type), orderdesc: creationTimeMs, first: %d, offset: %d)  {
			expand(_all_) {
				expand(_all_)
			}
		}
	}	
	`, limit, (page-1)*limit))

	// Run query created above
	resp, err := f.client.Run(f.getContext(), &req)
	if err != nil {
		return thingsResponse, err
	}

	// Merge the results into the model to return
	nodes := resp.GetN()
	resultItems := nodes[0].Children

	// Set the return array length
	thingsResponse.Things = make([]*models.ThingGetResponse, len(resultItems))

	for i, node := range resultItems {
		thingResponse := &models.ThingGetResponse{}
		thingResponse.Schema = map[string]models.JSONObject{}
		f.mergeThingNodeInResponse(node, thingResponse)
		thingsResponse.Things[i] = thingResponse
	}

	// Create query to count total results
	req = dgraphClient.Req{}
	req.SetQuery(`{ 
  		totalResults(func: has(type))  {
    		count()
  		}
	}`)

	// Run query created above
	resp, err = f.client.Run(f.getContext(), &req)
	if err != nil {
		return thingsResponse, err
	}

	// Unmarshal the dgraph response into a struct
	var totalResult TotalResultsResult
	err = dgraphClient.Unmarshal(resp.N, &totalResult)
	if err != nil {
		return thingsResponse, nil
	}

	// Set the total results
	thingsResponse.TotalResults = totalResult.Root.Count

	return thingsResponse, nil
}

// UpdateThing updates the Thing in the DB at the given UUID.
func (f *Dgraph) UpdateThing(thing *models.ThingUpdate, UUID strfmt.UUID) error {
	refThingNode, err := f.getNodeByUUID(UUID)

	if err != nil {
		return err
	}

	err = f.updateThingNodeEdges(refThingNode, &thing.Thing)

	return err
}

// DeleteThing deletes the Thing in the DB at the given UUID.
func (f *Dgraph) DeleteThing(UUID strfmt.UUID) error {
	// Create the query for removing query
	variables := make(map[string]string)
	variables["$uuid"] = string(UUID)

	req := dgraphClient.Req{}
	req.SetQueryWithVariables(`{
		var(func: eq(uuid, $uuid)) {
			node_to_delete as ~id
		}
		
		id_node_to_delete as var(func: eq(uuid, $uuid))
	}
	mutation {
		delete {
			uid(node_to_delete) * * .
			uid(id_node_to_delete) * * .
		}
	}`, variables)

	_, err := f.client.Run(f.getContext(), &req)

	if err != nil {
		return err
	}

	return nil
}

// AddAction adds an Action to the Dgraph database with the given UUID
func (f *Dgraph) AddAction(action *models.Action, UUID strfmt.UUID) error {
	// TODO: make type interactive
	newNode, err := f.addNewNode(action.AtContext, "OnOffAction", UUID)

	if err != nil {
		return err
	}

	// Add all given information to the new node
	err = f.updateActionNodeEdges(newNode, action)

	if err != nil {
		return err
	}

	return nil
}

// GetAction returns an action from the database
func (f *Dgraph) GetAction(UUID strfmt.UUID) (models.ActionGetResponse, error) {
	// Initialize response
	actionResponse := models.ActionGetResponse{}
	actionResponse.Schema = map[string]models.JSONObject{}

	// Do a query to get all node-information based on the given UUID
	variables := make(map[string]string)
	variables["$uuid"] = string(UUID)

	req := dgraphClient.Req{}
	req.SetQueryWithVariables(`{
		get(func: eq(uuid, $uuid)) { 
			uuid
			~id {
				expand(_all_) {
					expand(_all_) {
						expand(_all_) 
					}
				}
				~action.of {
					id {
						uuid
					}
					type {
						class
					}
				}
			}
		}
	}`, variables)

	// Run query created above
	resp, err := f.client.Run(f.getContext(), &req)
	if err != nil {
		return actionResponse, err
	}

	// Merge the results into the model to return
	nodes := resp.GetN()
	for _, node := range nodes {
		f.mergeActionNodeInResponse(node, &actionResponse, "")
	}

	return actionResponse, nil
}

// ListActions lists actions for a specific thing
func (f *Dgraph) ListActions(UUID strfmt.UUID, limit int, page int) (models.ActionsListResponse, error) {
	// Initialize response
	actionsResponse := models.ActionsListResponse{}

	// Do a query to get all node-information
	req := dgraphClient.Req{}
	req.SetQuery(fmt.Sprintf(`{
		actions(func: eq(uuid, "%s")) { 
			uuid
			~id {
				actions: ~action.target (orderdesc: creationTimeUnix) (first: %d, offset: %d) {
					expand(_all_) {
						expand(_all_)
					}
				}
			}
		}
	}`, UUID, limit, (page-1)*limit))

	// Run query created above
	resp, err := f.client.Run(f.getContext(), &req)
	if err != nil {
		return actionsResponse, err
	}

	// Merge the results into the model to return
	nodes := resp.GetN()
	resultItems := nodes[0].Children[0].Children[0].Children

	// Set the return array length
	actionsResponse.Actions = make([]*models.ActionGetResponse, len(resultItems))

	// Loop to add all items in the return object
	for i, node := range resultItems {
		actionResponse := &models.ActionGetResponse{}
		actionResponse.Schema = map[string]models.JSONObject{}
		actionResponse.ThingID = UUID
		f.mergeActionNodeInResponse(node, actionResponse, "")
		actionsResponse.Actions[i] = actionResponse
	}

	// Create query to count total results
	// TODO: Combine the total results code with the code of the 'things
	req = dgraphClient.Req{}
	req.SetQuery(fmt.Sprintf(`{
		totalResults(func: eq(uuid, "%s")) {
			~id {
				~action.target {
					count()
				}
			}
		}
	}`, UUID))

	// Run query created above
	resp, err = f.client.Run(f.getContext(), &req)
	if err != nil {
		return actionsResponse, err
	}

	// Unmarshal the dgraph response into a struct
	var totalResult TotalResultsResult
	err = dgraphClient.Unmarshal(resp.N, &totalResult)
	if err != nil {
		return actionsResponse, nil
	}

	// Set the total results
	actionsResponse.TotalResults = totalResult.Root.Count // TODO: NOT WORKING, MISSING 'totalResults' IN RETURN OBJ, DGRAPH bug?

	return actionsResponse, nil
}

// UpdateAction updates a specific action
func (f *Dgraph) UpdateAction(action *models.Action, UUID strfmt.UUID) error {
	refActionNode, err := f.getNodeByUUID(UUID)

	if err != nil {
		return err
	}

	err = f.updateActionNodeEdges(refActionNode, action)

	return err
}

func (f *Dgraph) addNewNode(nodeContext string, nodeClass string, UUID strfmt.UUID) (dgraphClient.Node, error) {
	// TODO: separate both?
	nodeType := nodeContext + "/" + nodeClass

	// Search for the class to make the connection, create variables
	variables := make(map[string]string)
	variables["$a"] = nodeType

	// Create the query for existing class
	req := dgraphClient.Req{}
	req.SetQueryWithVariables(`{
		class(func: eq(class, $a)) {
			_uid_
			class
		}
	}`, variables)

	// Run the query
	resp, err := f.client.Run(f.getContext(), &req)

	if err != nil {
		return dgraphClient.Node{}, err
	}

	// Unmarshal the result
	var dClass ClassResult
	err = dgraphClient.Unmarshal(resp.N, &dClass)

	if err != nil {
		return dgraphClient.Node{}, err
	}

	// Create the classNode from the result
	classNode := f.client.NodeUid(dClass.Root.ID)

	// Node has been found, create new one and connect it
	newNode, err := f.client.NodeBlank(fmt.Sprintf("%v", gouuid.NewV4()))

	if err != nil {
		return dgraphClient.Node{}, err
	}

	// Add edge between New Node and Class Node
	typeEdge := newNode.ConnectTo("type", classNode)

	// Add class edge to batch
	req = dgraphClient.Req{}
	err = req.Set(typeEdge)

	if err != nil {
		return dgraphClient.Node{}, err
	}

	// Add UUID node
	uuidNode, err := f.client.NodeBlank(string(UUID))

	// Add UUID edge
	uuidEdge := newNode.ConnectTo("id", uuidNode)
	if err = req.Set(uuidEdge); err != nil {
		return dgraphClient.Node{}, err
	}

	// Add UUID to UUID node
	// TODO: Search for uuid edge/node before making new??
	edge := uuidNode.Edge("uuid")
	if err = edge.SetValueString(string(UUID)); err != nil {
		return dgraphClient.Node{}, err
	}
	if err = req.Set(edge); err != nil {
		return dgraphClient.Node{}, err
	}

	// Call run after all mutations are added
	_, err = f.client.Run(f.getContext(), &req)

	return newNode, nil
}

// updateThingNodeEdges updates all the edges of the node, used with a new node or to update/patch a node
func (f *Dgraph) updateThingNodeEdges(node dgraphClient.Node, thing *models.Thing) error {
	// Create update request
	req := dgraphClient.Req{}

	// Init error var
	var err error

	// Add timing nodes
	edge := node.Edge("creationTimeMs")
	if err = edge.SetValueInt(thing.CreationTimeMs); err != nil {
		return err
	}
	if err = req.Set(edge); err != nil {
		return err
	}

	edge = node.Edge("lastSeenTimeMs")
	if err = edge.SetValueInt(thing.LastSeenTimeMs); err != nil {
		return err
	}
	if err = req.Set(edge); err != nil {
		return err
	}

	edge = node.Edge("lastUpdateTimeMs")
	if err = edge.SetValueInt(thing.LastUpdateTimeMs); err != nil {
		return err
	}
	if err = req.Set(edge); err != nil {
		return err
	}

	edge = node.Edge("lastUseTimeMs")
	if err = edge.SetValueInt(thing.LastUseTimeMs); err != nil {
		return err
	}
	if err = req.Set(edge); err != nil {
		return err
	}

	// Add Thing properties
	for propKey, propValue := range thing.Schema {
		err = f.addPropertyEdge(req, node, propKey, propValue)
	}

	// Call run after all mutations are added
	_, err = f.client.Run(f.getContext(), &req)

	return err
}

// updateActionNodeEdges updates all the edges of the node, used with a new node or to update/patch a node
func (f *Dgraph) updateActionNodeEdges(node dgraphClient.Node, action *models.Action) error {
	// Create update request
	req := dgraphClient.Req{}

	// Init error var
	var err error

	// Add timing nodes
	edge := node.Edge("creationTimeUnix")
	if err = edge.SetValueInt(action.CreationTimeUnix); err != nil {
		return err
	}
	if err = req.Set(edge); err != nil {
		return err
	}

	edge = node.Edge("lastUpdateTimeUnix")
	if err = edge.SetValueInt(action.LastUpdateTimeUnix); err != nil {
		return err
	}
	if err = req.Set(edge); err != nil {
		return err
	}

	// Add Action properties
	for propKey, propValue := range action.Schema {
		err = f.addPropertyEdge(req, node, propKey, propValue)
	}

	// Add Thing that gets the action
	objectNode, err := f.getNodeByUUID(action.ThingID)
	if err != nil {
		return err
	}
	err = f.connectRef(req, node, "action.target", objectNode)
	if err != nil {
		return err
	}

	// Add subject Thing by $ref TODO: make interactive
	// subjectNode, err := f.getNodeByUUID(action.Subject.RefValue?)
	subjectNode, err := f.getNodeByUUID("fdcecd93-bb6b-41d1-9635-4596eb98f694")
	if err != nil {
		return err
	}
	err = f.connectRef(req, subjectNode, "action.of", node)
	if err != nil {
		return err
	}

	// Call run after all mutations are added
	_, err = f.client.Run(f.getContext(), &req)

	return err
}

func (f *Dgraph) addPropertyEdge(req dgraphClient.Req, node dgraphClient.Node, propKey string, propValue models.JSONObject) error {
	// TODO: add property: string/int/connection other object, now everything is string
	// TODO: Add 'schema.' to a global var, if this is the nicest way to fix
	var err error

	edgeName := propKey // add "schema." + ??
	if val, ok := propValue["value"]; ok {
		edge := node.Edge(edgeName)
		if err = edge.SetValueString(val.(string)); err != nil {
			return err
		}
		if err = req.Set(edge); err != nil {
			return err
		}
	} else if val, ok = propValue["ref"]; ok {
		refThingNode, err := f.getNodeByUUID(strfmt.UUID(val.(string)))
		if err != nil {
			return err
		}
		err = f.connectRef(req, node, edgeName, refThingNode)
	}

	return nil
}

func (f *Dgraph) connectRef(req dgraphClient.Req, nodeFrom dgraphClient.Node, edgeName string, nodeTo dgraphClient.Node) error {
	relatedEdge := nodeFrom.ConnectTo(edgeName, nodeTo)
	if err := req.Set(relatedEdge); err != nil {
		return err
	}
	return nil
}

// mergeThingNodeInResponse based on https://github.com/dgraph-io/dgraph/blob/release/v0.8.0/wiki/resources/examples/goclient/crawlerRDF/crawler.go#L250-L264
func (f *Dgraph) mergeThingNodeInResponse(node *protos.Node, thingResponse *models.ThingGetResponse) {
	attribute := node.Attribute

	for _, prop := range node.GetProperties() {
		if attribute == "~id" || attribute == "things" {
			if prop.Prop == "creationTimeMs" {
				thingResponse.CreationTimeMs = prop.GetValue().GetIntVal()
			} else if prop.Prop == "lastSeenTimeMs" {
				thingResponse.LastSeenTimeMs = prop.GetValue().GetIntVal()
			} else if prop.Prop == "lastUpdateTimeMs" {
				thingResponse.LastUpdateTimeMs = prop.GetValue().GetIntVal()
			} else if prop.Prop == "lastUseTimeMs" {
				thingResponse.LastUseTimeMs = prop.GetValue().GetIntVal()
			} else {
				thingResponse.Schema[prop.Prop] = map[string]models.JSONValue{
					"value": prop.GetValue().GetStrVal(),
				}
			}
		} else if attribute == "type" {
			if prop.Prop == "context" {
				// thingResponse.AtType = "Person" TODO: FIX?
			} else if prop.Prop == "class" {
				thingResponse.AtContext = prop.GetValue().GetStrVal()
			}
		} else if attribute == "id" {
			if prop.Prop == "uuid" {
				thingResponse.ThingID = strfmt.UUID(prop.GetValue().GetStrVal())
			}
		}
	}

	for _, child := range node.Children {
		f.mergeThingNodeInResponse(child, thingResponse)
	}

}

// mergeActionNodeInResponse based on https://github.com/dgraph-io/dgraph/blob/release/v0.8.0/wiki/resources/examples/goclient/crawlerRDF/crawler.go#L250-L264
func (f *Dgraph) mergeActionNodeInResponse(node *protos.Node, actionResponse *models.ActionGetResponse, parentAttribute string) {
	attribute := node.Attribute

	for _, prop := range node.GetProperties() {
		if attribute == "~id" || attribute == "actions" {
			if prop.Prop == "creationTimeUnix" {
				actionResponse.CreationTimeUnix = prop.GetValue().GetIntVal()
			} else if prop.Prop == "lastUpdateTimeUnix" {
				actionResponse.LastUpdateTimeUnix = prop.GetValue().GetIntVal()
			} else {
				actionResponse.Schema[prop.Prop] = map[string]models.JSONValue{
					"value": prop.GetValue().GetStrVal(),
				}
			}
		} else if attribute == "type" && (parentAttribute == "~id" || parentAttribute == "actions") {
			if prop.Prop == "context" {
				// actionResponse.AtType = "Person" TODO: FIX?
			} else if prop.Prop == "class" {
				actionResponse.AtContext = prop.GetValue().GetStrVal()
			}
		} else if attribute == "id" && (parentAttribute == "~id" || parentAttribute == "actions") {
			if prop.Prop == "uuid" {
				actionResponse.ActionID = strfmt.UUID(prop.GetValue().GetStrVal())
			}
		} else if attribute == "type" && parentAttribute == "action.target" {
			if prop.Prop == "context" {
				// actionResponse.AtType = "Person" TODO: FIX?
			} else if prop.Prop == "class" {
				// actionResponse.ThingID = prop.GetValue().GetStrVal() // TODO: THING ID has to be fixed, it is a REF object
			}
		} else if attribute == "id" && parentAttribute == "action.target" {
			if prop.Prop == "uuid" {
				actionResponse.ThingID = strfmt.UUID(prop.GetValue().GetStrVal())
			}
		} else if attribute == "type" && parentAttribute == "~action.of" {
			if prop.Prop == "context" {
				// actionResponse.AtType = "Person" TODO: FIX?
			} else if prop.Prop == "class" {
				// actionResponse.Subject = prop.GetValue().GetStrVal() // TODO: Subject ID has to be fixed
			}
		} else if attribute == "id" && parentAttribute == "~action.of" {
			if prop.Prop == "uuid" {
				// actionResponse.Subject = strfmt.UUID(prop.GetValue().GetStrVal())
			}
		}
	}

	for _, child := range node.Children {
		f.mergeActionNodeInResponse(child, actionResponse, attribute)
	}

}

func (f *Dgraph) getNodeByUUID(UUID strfmt.UUID) (dgraphClient.Node, error) {
	// Search for the class to make the connection, create variables
	variables := make(map[string]string)
	variables["$uuid"] = string(UUID)

	// Create the query for existing class
	req := dgraphClient.Req{}
	req.SetQueryWithVariables(`{ 
		node(func: eq(uuid, $uuid)) {
			uuid
			~id {
				_uid_
			}
		}
	}`, variables)

	// Run the query
	resp, err := f.client.Run(f.getContext(), &req)

	if err != nil {
		return dgraphClient.Node{}, err
	}

	// Unmarshal the result
	var idResult NodeIDResult
	err = dgraphClient.Unmarshal(resp.N, &idResult)

	if err != nil {
		return dgraphClient.Node{}, err
	}

	// Create the classNode from the result
	node := f.client.NodeUid(idResult.Root.Node.ID)

	return node, err
}

// Add deprecated?
func (f *Dgraph) Add(dbObject connector_utils.DatabaseObject) (string, error) {

	// Return the ID that is used to create.
	return "hoi", nil
}

// Get deprecated?
func (f *Dgraph) Get(UUID string) (connector_utils.DatabaseObject, error) {

	return connector_utils.DatabaseObject{}, nil
}

// List deprecated?
func (f *Dgraph) List(refType string, ownerUUID string, limit int, page int, referenceFilter *connector_utils.ObjectReferences) (connector_utils.DatabaseObjects, int64, error) {
	dataObjs := connector_utils.DatabaseObjects{}

	return dataObjs, 0, nil
}

// ValidateKey Validate if a user has access, returns permissions object
func (f *Dgraph) ValidateKey(token string) ([]connector_utils.DatabaseUsersObject, error) {
	dbUsersObjects := []connector_utils.DatabaseUsersObject{{
		Deleted:        false,
		KeyExpiresUnix: -1,
		KeyToken:       "be9d1928-d004-4b0b-a8c3-2aad06cc67e1",
		Object:         "{}",
		Parent:         "",
		Uuid:           "ce9d1928-d004-4b0b-a8c3-2aad06cc67e1",
	}}

	// keys are found, return true
	return dbUsersObjects, nil
}

// GetKey returns user object by ID
func (f *Dgraph) GetKey(UUID string) (connector_utils.DatabaseUsersObject, error) {
	// Return found object
	return connector_utils.DatabaseUsersObject{}, nil

}

// AddKey to DB
func (f *Dgraph) AddKey(parentUUID string, dbObject connector_utils.DatabaseUsersObject) (connector_utils.DatabaseUsersObject, error) {

	// Return the ID that is used to create.
	return connector_utils.DatabaseUsersObject{}, nil

}

// DeleteKey removes a key from the database
func (f *Dgraph) DeleteKey(UUID string) error {

	return nil
}

// GetChildObjects returns all the child objects
func (f *Dgraph) GetChildObjects(UUID string, filterOutDeleted bool) ([]connector_utils.DatabaseUsersObject, error) {
	// Fill children array
	childUserObjects := []connector_utils.DatabaseUsersObject{}

	return childUserObjects, nil
}

func (f *Dgraph) getAllClasses() (AllClassesResult, error) {
	// Search for all existing classes
	req := dgraphClient.Req{}

	req.SetQuery(`{
		classes(func: has(class)) {
			_uid_
			class
		}
	}`)

	resp, err := f.client.Run(f.getContext(), &req)
	if err != nil {
		return AllClassesResult{}, err
	}

	var allClasses AllClassesResult
	err = dgraphClient.Unmarshal(resp.N, &allClasses)

	if err != nil {
		return AllClassesResult{}, err
	}

	return allClasses, nil
}

func (f *Dgraph) getClassFromResult(className string, allClasses AllClassesResult) (*DgraphClass, error) {
	for _, class := range allClasses.Root {
		if class.Class == className {
			return class, nil
		}
	}
	return &DgraphClass{}, errors_.New("class not found")
}

func (f *Dgraph) getContext() context.Context {
	return context.Background()
}

// TODO REMOVE, JUST FOR TEST
func printNode(depth int, node *protos.Node) {

	fmt.Println(strings.Repeat(" ", depth), "Atrribute : ", node.Attribute)

	// the values at this level
	for _, prop := range node.GetProperties() {
		fmt.Println(strings.Repeat(" ", depth), "Prop : ", prop.Prop, " Value : ", prop.Value, " Type : %T", prop.Value)
	}

	for _, child := range node.Children {
		fmt.Println(strings.Repeat(" ", depth), "+")
		printNode(depth+1, child)
	}

}