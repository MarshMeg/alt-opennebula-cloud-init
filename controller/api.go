package controller

import (
	"fmt"
	"net"
	"net/http"
	"opennebula-init/types"
	"os"
	"strings"
	"sync/atomic"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

var (
	requestCounter uint32                   // Счётчик запросов
	shutdownSignal = make(chan interface{}) // Канал для сигнала завершения
)

func RunApi(nodesCount uint32, passwd uuid.UUID) []NodeApplyConfig {
	var secretKey string
	var publicKey string

	secretKeyBytes, _ := os.ReadFile("/root/.ssh/id_ed25519")
	publicKeyBytes, _ := os.ReadFile("/root/.ssh/id_ed25519.pub")
	secretKey = strings.ReplaceAll(string(secretKeyBytes), "\n", "")
	secretKey = strings.ReplaceAll(secretKey, "-----BEGIN OPENSSH PRIVATE KEY-----", "")
	secretKey = strings.ReplaceAll(secretKey, "-----END OPENSSH PRIVATE KEY-----", "")
	publicKey = strings.ReplaceAll(string(publicKeyBytes), "\n", "")

	router := gin.Default()
	var nodes []NodeApplyConfig

	router.GET("/ssh-key", func(c *gin.Context) {
		count := atomic.AddUint32(&requestCounter, 1)
		fmt.Printf("Обработан запрос #%d\n", count)

		c.JSON(200, types.SSHData{
			SecretKey: secretKey,
			PublicKey: publicKey,
			Passwd:    passwd.String(),
		})

		nodes = append(nodes, NodeApplyConfig{
			Host: net.ParseIP(c.ClientIP()),
			Name: c.Query("hostname"),
		})

		if count >= nodesCount {
			fmt.Println("Данные для подключения отправлены. Остановка сервера...")
			close(shutdownSignal) // Закрываем канал для завершения
		}
	})

	// Запускаем сервер в отдельной горутине
	server := &http.Server{
		Addr:    "0.0.0.0:8000",
		Handler: router,
	}
	go func() {
		fmt.Println("Сервер запущен на http://0.0.0.0:8000")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("Ошибка сервера: %v\n", err)
		}
	}()

	// Ожидаем сигнала завершения
	<-shutdownSignal

	// Остановка сервера
	fmt.Println("Остановка сервера...")
	if err := server.Close(); err != nil {
		fmt.Printf("Ошибка при остановке сервера: %v\n", err)
	}

	fmt.Println("Сервер остановлен.")

	return nodes
}
