package controller

import (
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/google/uuid"
)

func ControllerInit() {
	var nodesCount uint32
	fmt.Print("Введите кол-во нод в кластере (1): ")
	_, err := fmt.Scanf("%d", &nodesCount)
	if err != nil {
		nodesCount = 1
	}
	var passwd uuid.UUID = uuid.New()

	if err = os.WriteFile("/var/lib/one/.one/one_auth", []byte(fmt.Sprintf("oneadmin:%s", passwd.String())), 0644); err != nil {
		log.Fatal(err)
	}

	files, err := os.ReadDir("/root/.ssh")
	if err != nil {
		log.Fatalf("Ошибка при чтении директории: %v", err)
	}

	var keyReady bool = false
	for _, file := range files {
		if file.Name() == "id_ed25519" {
			keyReady = true
		}
	}

	if !keyReady {
		cmd := exec.Command("/usr/bin/ssh-keygen", "-t", "ed25519", "-f", "/root/.ssh/id_ed25519", "-C", "root@cloud.local", "-N", "")
		output, err := cmd.CombinedOutput()
		fmt.Println("\nKey not found, run generate:")
		fmt.Printf("%s\n", output)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
	}

	_ = exec.Command("systemctl", "start", "opennebula").Run()
	_ = exec.Command("systemctl", "start", "opennebula-sunstone").Run()

	var nodes []NodeApplyConfig = RunApi(nodesCount, passwd)
	var keyscanList []byte = []byte{}

	for _, node := range nodes {
		hosts, err := os.ReadFile("/etc/hosts")
		if err != nil {
			fmt.Println(err)
		}
		os.WriteFile("/etc/hosts", append(hosts, []byte(fmt.Sprintf("%s %s", node.Host, node.Name))...), 0644)

		keyscanList = append(keyscanList, []byte(fmt.Sprintf("%s %s ", node.Host, node.Name))...)
	}

	if err = exec.Command("ssh-keyscan", string(keyscanList), ">>", "/var/lib/one/.ssh/known_hosts").Run(); err != nil {
		log.Fatal(err)
	}

	for _, node := range nodes {
		if err = exec.Command("scp", "-rp", "/var/lib/one/.ssh", fmt.Sprintf("%s:/var/lib/one/", node.Name)).Run(); err != nil {
			log.Fatal(err)
		}

		if err = exec.Command("onehost", "create", node.Name, "--im kvm", "--vm kvm").Run(); err != nil {
			log.Fatal(err)
		}
	}
}
