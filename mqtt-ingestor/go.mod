module home_automation_dashboard/mqtt-ingestor

go 1.20

require home_automation_dashboard/shared v0.0.0

replace home_automation_dashboard/shared => ../shared

require (
	github.com/eclipse/paho.mqtt.golang v1.5.0
	go.mongodb.org/mongo-driver v1.17.1
)

require (
	github.com/golang/snappy v0.0.4 // indirect
	github.com/gorilla/websocket v1.5.3 // indirect
	github.com/klauspost/compress v1.13.6 // indirect
	github.com/montanaflynn/stats v0.7.1 // indirect
	github.com/xdg-go/pbkdf2 v1.0.0 // indirect
	github.com/xdg-go/scram v1.1.2 // indirect
	github.com/xdg-go/stringprep v1.0.4 // indirect
	github.com/youmark/pkcs8 v0.0.0-20240726163527-a2c0da244d78 // indirect
	golang.org/x/crypto v0.26.0 // indirect
	golang.org/x/net v0.27.0 // indirect
	golang.org/x/sync v0.8.0 // indirect
	golang.org/x/text v0.17.0 // indirect
)
