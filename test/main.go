package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/gopcua/opcua"
	"github.com/gopcua/opcua/id"
	"github.com/gopcua/opcua/ua"
)

type SubOptions struct {
	Endpoint string
	Policy   string
	Mode     string
	CertFile string
	KeyFile  string
	NodeID   string
	Event    bool
	Interval time.Duration
}

func main() {
	options := SubOptions{
		Endpoint: "opc.tcp://localhost:4840",
		Policy:   "",
		Mode:     "",
		CertFile: "",
		KeyFile:  "",
		NodeID:   "",
		Event:    false,
		Interval: opcua.DefaultSubscriptionInterval,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	ep := getEndpoint(ctx, options.Endpoint, options.Policy, options.Mode)
	c := connectClient(ctx, ep, options.Policy, options.Mode, options.CertFile, options.KeyFile)
	defer c.Close(ctx)

	sub := createSubscription(ctx, c, options.Interval)
	defer sub.Cancel(ctx)

	id := parseNodeID(options.NodeID)
	miCreateRequest, eventFieldNames := createMonitoredItemRequest(id, options.Event)
	monitorItem(ctx, sub, id, miCreateRequest)

	readNotifications(ctx, sub)
}

func getEndpoint(ctx context.Context, endpoint, policy, mode string) *ua.EndpointDescription {
	endpoints, err := opcua.GetEndpoints(ctx, endpoint)
	if err != nil {
		log.Fatal(err)
	}
	ep := opcua.SelectEndpoint(endpoints, policy, ua.MessageSecurityModeFromString(mode))
	if ep == nil {
		log.Fatal("Failed to find suitable endpoint")
	}
	ep.EndpointURL = endpoint
	fmt.Println("*", ep.SecurityPolicyURI, ep.SecurityMode)
	return ep
}

func connectClient(ctx context.Context, ep *ua.EndpointDescription, policy, mode, certFile, keyFile string) *opcua.Client {
	opts := []opcua.Option{
		opcua.SecurityPolicy(policy),
		opcua.SecurityModeString(mode),
		opcua.CertificateFile(certFile),
		opcua.PrivateKeyFile(keyFile),
		opcua.AuthAnonymous(),
		opcua.SecurityFromEndpoint(ep, ua.UserTokenTypeAnonymous),
	}

	c, err := opcua.NewClient(ep.EndpointURL, opts...)
	if err != nil {
		log.Fatal(err)
	}
	if err := c.Connect(ctx); err != nil {
		log.Fatal(err)
	}
	return c
}

func createSubscription(ctx context.Context, c *opcua.Client, interval time.Duration) *opcua.Subscription {
	notifyCh := make(chan *opcua.PublishNotificationData)
	sub, err := c.Subscribe(ctx, &opcua.SubscriptionParameters{
		Interval: interval,
	}, notifyCh)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Created subscription with id %v", sub.SubscriptionID)
	return sub
}

func parseNodeID(nodeID string) *ua.NodeID {
	id, err := ua.ParseNodeID(nodeID)
	if err != nil {
		log.Fatal(err)
	}
	return id
}

func createMonitoredItemRequest(id *ua.NodeID, event bool) (*ua.MonitoredItemCreateRequest, []string) {
	if event {
		return eventRequest(id)
	}
	return valueRequest(id), nil
}

func monitorItem(ctx context.Context, sub *opcua.Subscription, id *ua.NodeID, miCreateRequest *ua.MonitoredItemCreateRequest) {
	res, err := sub.Monitor(ctx, ua.TimestampsToReturnBoth, miCreateRequest)
	if err != nil || res.Results[0].StatusCode != ua.StatusOK {
		log.Fatal(err)
	}
}

func readNotifications(ctx context.Context, sub *opcua.Subscription) {
	for {
		select {
		case <-ctx.Done():
			return
		case res := <-sub.Notifs:
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

func valueRequest(nodeID *ua.NodeID) *ua.MonitoredItemCreateRequest {
	handle := uint32(42)
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
