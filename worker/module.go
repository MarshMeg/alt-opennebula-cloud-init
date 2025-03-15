package worker

import (
	"fmt"
	"log"
	"net"
	"opennebula-init/types"
	"os"

	"github.com/go-resty/resty/v2"
)

func WorkerInit(controllerIP net.IP) {
	hostname, _ := os.Hostname()

	var sshData types.SSHData
	res, err := resty.New().R().Get(fmt.Sprintf("http://%s:8000/ssh-key?hostname=%s", controllerIP.String(), hostname))
	if err != nil {
		log.Fatal(err)
	}
	if err := resty.New().JSONUnmarshal(res.Body(), sshData); err != nil {
		log.Fatal(err)
	}

	_ = os.WriteFile("/root/.ssh/id_ed25519", []byte(sshData.SecretKey), 0777)
	_ = os.WriteFile("/root/.ssh/id_ed25519.pub", []byte(sshData.PublicKey), 0777)
	_ = os.WriteFile("/var/lib/one/.one/one_auth", []byte(sshData.Passwd), 0777)
}
