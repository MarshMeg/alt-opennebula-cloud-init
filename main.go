package main

import (
	"fmt"
	"net"
	"opennebula-init/controller"
)

func main() {
	var controllerIp net.IP

	getMeta(&controllerIp)

	if controllerIp.String() == "127.0.0.1" {
		controller.ControllerInit()
	}
}

func getMeta(ip *net.IP) {
	// Запрашиваем ввод данных
	var ipStr string
	fmt.Print("Введите IPv4 контролирующей ноды (127.0.0.1): ")
	_, err := fmt.Scanf("%s", &ipStr)
	if err != nil {
		ipStr = "127.0.0.1"
	}

	*ip = net.ParseIP(ipStr)
}
