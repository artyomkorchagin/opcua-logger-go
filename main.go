package main

import (
	"context"
	"log"
	"main/types"
	"os"
	"time"
)

const (
	THREAD_LIMITER_AMOUNT = 8
)

var (
	consoleLog = log.New(os.Stdout, "", log.LstdFlags)
	ctx        = context.Background()
)

func main() {

	log.SetFlags(3)
	consoleLog.SetFlags(0)
	f, err := os.OpenFile("testlogfile", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		consoleLog.Println("Не удалось создать лог файл")
		time.Sleep(time.Duration(3) * time.Second)
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()

	log.SetOutput(f)
	log.Println("Starting app")
	consoleLog.Println("Загружается конфигурация...")

	// var wg sync.WaitGroup

	// msk := time.FixedZone("MSK", 3*60*60)

	log.Println("Creating config")
	endpointCfgs := types.NewEndpointConfig()
	log.Println("Success")

	fi, err := os.Stat("configs/service_cfg.yaml")

	if os.IsNotExist(err) || fi.Size() == 0 {
		// if config doesnt exist
		log.Println("Generating YAML")
		types.GenerateYaml(endpointCfgs)
		log.Println("Success")

	} else {
		log.Println("Loading from YAML")
		endpointCfgs, err = types.GetEndpoints()
		if err != nil {
			log.Fatal(err)
		}
		log.Println("Success")
	}

	MainLoop(endpointCfgs)

	// for {

	// 	log.Println("Connecting to Database")
	// 	db := api.ConnectToDB()
	// 	defer db.Close()
	// 	log.Println("Success")

	// 	log.Println("Refreshing config")
	// 	types.GenerateYaml(endpointCfgs)

	// 	log.Println("Success")

	// 	log.Println("Updating tags")
	// 	for _, cfg := range *endpointCfgs {
	// 		for _, tag := range cfg.Tags {
	// 			wg.Add(1)
	// 			go func(t types.Tag) {
	// 				defer wg.Done()
	// 				api.UpdateTagsTable(db, tag)
	// 			}(tag)
	// 		}
	// 		wg.Wait()
	// 	}
	// 	log.Println("Success")

	// 	log.Println("Getting addresses")
	// 	devices := []types.DeviceLog{}
	// 	for _, cfg := range *endpointCfgs {
	// 		for _, v := range cfg.Tags {
	// 			go func(tag types.Tag) {
	// 				if d := api.RequestNodeAdressesFromTag(tag); d.Address != "" {
	// 					devices = append(devices, d)
	// 				}
	// 			}(v)
	// 		}
	// 	}
	// 	log.Println("Success")

	// devices := api.RequestNodeAdressesFromDB(db)

	// limiter := make(chan int, THREAD_LIMITER_AMOUNT)
	// for _, device := range devices {
	// 	log.Println("Parsing address to node id")
	// 	id, err := ua.ParseNodeID(device.Address)
	// 	if err != nil {
	// 		log.Fatalf("invalid node id: %v", err)
	// 	}
	// 	log.Println("Success")

	// 	req := &ua.ReadRequest{
	// 		MaxAge: 2000,
	// 		NodesToRead: []*ua.ReadValueID{
	// 			{NodeID: id},
	// 		},
	// 		TimestampsToReturn: ua.TimestampsToReturnBoth,
	// 	}

	// 	log.Println("Retrieving info from the nodes")
	// 	var resp *ua.ReadResponse
	// 	for {
	// 		resp, err = c.Read(ctx, req)
	// 		if err == nil {
	// 			break
	// 		}

	// 		// Following switch contains known errors that can be retried by the user.
	// 		// Best practice is to do it on read operations.
	// 		switch {
	// 		case err == io.EOF && c.State() != opcua.Closed:
	// 			// has to be retried unless user closed the connection
	// 			time.After(1 * time.Second)
	// 			continue

	// 		case errors.Is(err, ua.StatusBadSessionIDInvalid):
	// 			// Session is not activated has to be retried. Session will be recreated internally.
	// 			time.After(1 * time.Second)
	// 			continue

	// 		case errors.Is(err, ua.StatusBadSessionNotActivated):
	// 			// Session is invalid has to be retried. Session will be recreated internally.
	// 			time.After(1 * time.Second)
	// 			continue

	// 		case errors.Is(err, ua.StatusBadSecureChannelIDInvalid):
	// 			// secure channel will be recreated internally.
	// 			time.After(1 * time.Second)
	// 			continue

	// 		default:
	// 			log.Fatalf("Read failed: %s", err)
	// 		}
	// 	}
	// 	log.Println("Success")

	// 	if resp.Results[0] == nil || resp.Results[0].Value == nil {
	// 		continue
	// 	}
	// 	device.Timestamp = resp.Results[0].SourceTimestamp.In(msk).Format("02.01.2006 15:04:05")
	// 	device.Value = fmt.Sprintf("%v", resp.Results[0].Value.Value())

	// 	switch resp.Results[0].Status {

	// 	case ua.StatusOK:
	// 		device.Quality = "GOOD"

	// 	case ua.StatusBad:
	// 		device.Quality = "BAD"

	// 	default:
	// 		device.Quality = "UNKNOWN"
	// 	}

	// 	log.Println("Inserting data to Accumulation table")
	// 	limiter <- 1
	// 	wg.Add(1)
	// 	go func(d types.DeviceLog) {
	// 		defer wg.Done()
	// 		api.InsertNewDataEntry(db, d)
	// 		<-limiter
	// 	}(device)

	// }
	// wg.Wait()
	// log.Println("Success. Got info of", len(devices), "elements")
	// }
}
