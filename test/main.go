// Copyright 2018-2020 opcua authors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.

package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/gopcua/opcua"
	"github.com/gopcua/opcua/debug"
	"github.com/gopcua/opcua/id"
	"github.com/gopcua/opcua/ua"
)

// SubOpts holds the subscription options
type SubOpts struct {
	HandleID int
	Endpoint string
	Policy   string
	Mode     string
	CertFile string
	KeyFile  string
	NodeID   string
	Event    bool
	Interval time.Duration
	Debug    bool
}

func main() {
	opts := SubOpts{
		HandleID: 1,
		Endpoint: "opc.tcp://localhost:4840",
		Policy:   "",
		Mode:     "",
		CertFile: "",
		KeyFile:  "",
		NodeID:   "",
		Event:    false,
		Interval: opcua.DefaultSubscriptionInterval,
		Debug:    false,
	}

	debug.Enable = opts.Debug
	log.SetFlags(0)

	ctx := context.Background()

	endpoints, err := opcua.GetEndpoints(ctx, opts.Endpoint)
	if err != nil {
		log.Fatal(err)
	}
	ep := opcua.SelectEndpoint(endpoints, opts.Policy, ua.MessageSecurityModeFromString(opts.Mode))
	if ep == nil {
		log.Fatal("Failed to find suitable endpoint")
	}
	ep.EndpointURL = opts.Endpoint

	fmt.Println("*", ep.SecurityPolicyURI, ep.SecurityMode)

	clientOpts := []opcua.Option{
		opcua.SecurityPolicy(opts.Policy),
		opcua.SecurityModeString(opts.Mode),
		opcua.CertificateFile(opts.CertFile),
		opcua.PrivateKeyFile(opts.KeyFile),
		opcua.AuthAnonymous(),
		opcua.SecurityFromEndpoint(ep, ua.UserTokenTypeAnonymous),
	}

	c, err := opcua.NewClient(ep.EndpointURL, clientOpts...)
	if err != nil {
		log.Fatal(err)
	}
	if err := c.Connect(ctx); err != nil {
		log.Fatal(err)
	}
	defer c.Close(ctx)

	notifyCh := make(chan *opcua.PublishNotificationData)

	sub, err := c.Subscribe(ctx, &opcua.SubscriptionParameters{
		Interval: opts.Interval,
	}, notifyCh)
	if err != nil {
		log.Fatal(err)
	}
	defer sub.Cancel(ctx)
	log.Printf("Created subscription with id %v", sub.SubscriptionID)

	id, err := ua.ParseNodeID(opts.NodeID)
	if err != nil {
		log.Fatal(err)
	}

	var miCreateRequest *ua.MonitoredItemCreateRequest
	var eventFieldNames []string
	if opts.Event {
		miCreateRequest, eventFieldNames = eventRequest(id)
	} else {
		miCreateRequest = valueRequest(id, opts.HandleID)
	}
	res, err := sub.Monitor(ctx, ua.TimestampsToReturnBoth, miCreateRequest)
	if err != nil || res.Results[0].StatusCode != ua.StatusOK {
		log.Fatal(err)
	}

	// read from subscription's notification channel until ctx is cancelled
	for {
		select {
		case <-ctx.Done():
			return
		case res := <-notifyCh:
			if res.Error != nil {
				log.Print(res.Error)
				continue
			}

			switch x := res.Value.(type) {
			case *ua.DataChangeNotification:
				for _, item := range x.MonitoredItems {
					data := item.Value.Value.Value()
					log.Printf("MonitoredItem with client handle %v = %v", item.ClientHandle, data)
				}

			case *ua.EventNotificationList:
				for _, item := range x.Events {
					log.Printf("Event for client handle: %v\n", item.ClientHandle)
					for i, field := range item.EventFields {
						log.Printf("%v: %v of Type: %T", eventFieldNames[i], field.Value(), field.Value())
					}
					log.Println()
				}

			default:
				log.Printf("what's this publish result? %T", res.Value)
			}
		}
	}
}

func valueRequest(nodeID *ua.NodeID, id int) *ua.MonitoredItemCreateRequest {
	handle := uint32(id)
	return opcua.NewMonitoredItemCreateRequestWithDefaults(nodeID, ua.AttributeIDValue, handle)
}

func eventRequest(nodeID *ua.NodeID) (*ua.MonitoredItemCreateRequest, []string) {
	fieldNames := []string{"EventId", "EventType", "Severity", "Time", "Message"}
	selects := make([]*ua.SimpleAttributeOperand, len(fieldNames))

	for i, name := range fieldNames {
		selects[i] = &ua.SimpleAttributeOperand{
			TypeDefinitionID: ua.NewNumericNodeID(0, id.BaseEventType),
			BrowsePath:       []*ua.QualifiedName{{NamespaceIndex: 0, Name: name}},
			AttributeID:      ua.AttributeIDValue,
		}
	}

	wheres := &ua.ContentFilter{
		Elements: []*ua.ContentFilterElement{
			{
				FilterOperator: ua.FilterOperatorGreaterThanOrEqual,
				FilterOperands: []*ua.ExtensionObject{
					{
						EncodingMask: 1,
						TypeID: &ua.ExpandedNodeID{
							NodeID: ua.NewNumericNodeID(0, id.SimpleAttributeOperand_Encoding_DefaultBinary),
						},
						Value: ua.SimpleAttributeOperand{
							TypeDefinitionID: ua.NewNumericNodeID(0, id.BaseEventType),
							BrowsePath:       []*ua.QualifiedName{{NamespaceIndex: 0, Name: "Severity"}},
							AttributeID:      ua.AttributeIDValue,
						},
					},
					{
						EncodingMask: 1,
						TypeID: &ua.ExpandedNodeID{
							NodeID: ua.NewNumericNodeID(0, id.LiteralOperand_Encoding_DefaultBinary),
						},
						Value: ua.LiteralOperand{
							Value: ua.MustVariant(uint16(0)),
						},
					},
				},
			},
		},
	}

	filter := ua.EventFilter{
		SelectClauses: selects,
		WhereClause:   wheres,
	}

	filterExtObj := ua.ExtensionObject{
		EncodingMask: ua.ExtensionObjectBinary,
		TypeID: &ua.ExpandedNodeID{
			NodeID: ua.NewNumericNodeID(0, id.EventFilter_Encoding_DefaultBinary),
		},
		Value: filter,
	}

	handle := uint32(42)
	req := &ua.MonitoredItemCreateRequest{
		ItemToMonitor: &ua.ReadValueID{
			NodeID:       nodeID,
			AttributeID:  ua.AttributeIDEventNotifier,
			DataEncoding: &ua.QualifiedName{},
		},
		MonitoringMode: ua.MonitoringModeReporting,
		RequestedParameters: &ua.MonitoringParameters{
			ClientHandle:     handle,
			DiscardOldest:    true,
			Filter:           &filterExtObj,
			QueueSize:        10,
			SamplingInterval: 1.0,
		},
	}

	return req, fieldNames
}
