package main

import (
	"bufio"
	"fmt"
	"log"
	"main/api"
	"main/types"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/gopcua/opcua"
	"github.com/gopcua/opcua/id"
	"github.com/gopcua/opcua/ua"
)

func clearConsole() {
	cmd := exec.Command("cmd", "/c", "cls")
	cmd.Stdout = os.Stdout
	cmd.Run()
}

func switchStatusState(number *int) {
	if *number == 1 {
		*number = 0
	} else {
		*number = 1
	}
}

func wait(milliseconds int) {
	time.Sleep(time.Duration(milliseconds) * time.Millisecond)
}

func removeDuplicateStr(strSlice []string) []string {
	allKeys := make(map[string]bool)
	list := []string{}
	for _, item := range strSlice {
		if _, value := allKeys[item]; !value {
			allKeys[item] = true
			list = append(list, item)
		}
	}
	return list
}

func MainLoop(cfgs *[]types.EndpointConfig) {

	reader := bufio.NewReader(os.Stdin)

	for {
		clearConsole()
		consoleLog.Println("~~~OPCUA LOGGER~~~")
		consoleLog.Println("1. Изменить интервал запроса.\n2. Вкл/выкл тэгов.\n3. Добавить новый endpoint\nq для выхода")

		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if input == "q" {
			os.Exit(0)
			break
		}

		option, err := strconv.Atoi(input)

		if err != nil {
			consoleLog.Printf("\n\nЭто не число.\n")
			wait(500)
			continue
		} else {
			switch option {

			// case 1:
			// 	SubscribtionManagerLoop(cfgs)

			case 2:
				TurnOnTagLoop(*cfgs)

			case 3:
				NewEndpointLoop(*cfgs)
				if len(*cfgs) != 0 {
					if (*cfgs)[0].Client == nil {
						*cfgs = (*cfgs)[1:]
					}
					for i := range *cfgs {

						if len((*cfgs)[i].Tags) != 0 {
							(*cfgs)[i].Tags = GetAllNodes((*cfgs)[i])
						}
					}

				}

			default:
				consoleLog.Printf("\n\nНеизвестная команда.\n")
				wait(500)
			}

		}
	}
}

func TurnOnTagLoop(cfgs []types.EndpointConfig) {

	reader := bufio.NewReader(os.Stdin)

	for {
		clearConsole()

		consoleLog.Println("Введите номер сервера")
		for i, cfg := range cfgs {
			consoleLog.Printf("%d. %s", i+1, cfg.Endpoint)
		}
		consoleLog.Println("q для выхода")

		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		if input == "q" {
			break
		}

		serverID, err := strconv.Atoi(input)
		serverID--
		if err != nil {
			consoleLog.Printf("\n\nЭто не число.\n")
			wait(500)
			continue
		}

		if serverID < 0 || serverID > len(cfgs) {
			consoleLog.Printf("\n\nТакого сервера нет в списке.\n")
			wait(500)
			continue
		}

		for {
			clearConsole()

			for _, tag := range cfgs[serverID].Tags {
				consoleLog.Printf("ID %s. Тэг: %s. Статус: %d\n", tag.ID, tag.Name, tag.Enabled)
			}
			consoleLog.Println("Введите номера тэгов через пробел, чтобы вкл/выкл их. Напишите q для выхода")
			input, _ := reader.ReadString('\n')
			input = strings.TrimSpace(input)
			tagIDs_str := removeDuplicateStr(strings.Split(input, " "))
			if input == "q" {
				break
			}
			for i := range tagIDs_str {
				temp, err := strconv.Atoi(tagIDs_str[i])
				temp--
				if err != nil {
					consoleLog.Printf("Это не число.%v\n", temp)
					wait(500)
					break
				}

				if temp < 0 || temp >= len(cfgs[serverID].Tags) {
					consoleLog.Printf("Неверный номер тэга.%v\n", temp-1)
					wait(500)
					break
				}
				switchStatusState(&cfgs[serverID].Tags[temp].Enabled)
			}
		}

	}
}

func NewEndpointLoop(cfgs []types.EndpointConfig) {
	reader := bufio.NewReader(os.Stdin)
	for {
		clearConsole()

		consoleLog.Println("~~~ДОБАВЛЕНИЕ АДРЕСА~~~")
		consoleLog.Println("Введите адрес сервера (прим. localhost:55000)")
		consoleLog.Println("q для выхода")

		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if input == "q" {
			break
		}

		endpoint := fmt.Sprintf("opc.tcp://%s", input)
		c, err := opcua.NewClient(endpoint, opcua.SecurityMode(ua.MessageSecurityModeNone))
		if err != nil {
			consoleLog.Println("Не удалось создать клиент")
			wait(3000)
			continue
		}
		if err := c.Connect(ctx); err != nil {
			consoleLog.Println("Не удалось подключиться к серверу")
			wait(3000)
			continue
		}
		cfg := types.EndpointConfig{
			Client:   c,
			Endpoint: endpoint,
		}
		cfgs = append(cfgs, cfg)
		consoleLog.Printf("Соединение с %s установлено", endpoint)
		wait(1000)
	}
}

// func SubscribtionManagerLoop(cfgs *[]types.EndpointConfig) {

// 	reader := bufio.NewReader(os.Stdin)
// 	for {
// 		clearConsole()

// 		consoleLog.Println("~~~ИЗМЕНЕНИЕ ИНТЕРВАЛА ЗАПРОСА~~~")
// 		consoleLog.Println("Введите число в секундах\nq для выхода")
// 		consoleLog.Printf("Текущий интервал: %d секунд", cfgs.Interval)

// 		input, _ := reader.ReadString('\n')
// 		input = strings.TrimSpace(input)

// 		if input == "q" {
// 			break
// 		}

// 		seconds, err := strconv.Atoi(input)
// 		if err != nil {
// 			consoleLog.Printf("\n\nЭто не число.\n")
// 			wait(500)
// 			continue
// 		} else {
// 			cfg.Interval = seconds
// 		}
// 	}
// }

func GetAllNodes(cfg types.EndpointConfig) []types.Tag {
	log.Println("Getting root node")
	cfg.Client.Connect(ctx)
	root := cfg.Client.Node(ua.NewTwoByteNodeID(id.ObjectsFolder))
	log.Println("Success")

	log.Println("Getting list of nodes")
	nodeList, err := api.Browse(ctx, root, "", 0)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Success")

	log.Println("Filling config entity with tags")
	if err = api.FillEndpointConfig(&cfg, nodeList); err != nil {
		log.Fatal(err)
	}
	log.Println("Success")
	wait(3000)
	return cfg.Tags

}
