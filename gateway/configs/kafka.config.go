package configs

import "os"

type KafkaConfig struct {
	BootstrapServers string
	SecurityProtocol string
	Topics           map[string]string
}

func LoadKafkaConfig() *KafkaConfig {
	return &KafkaConfig{
		BootstrapServers: os.Getenv("KAFKA_BOOTSTRAP_SERVERS"),
		SecurityProtocol: os.Getenv("KAFKA_SECURITY_PROTOCOL"),
		Topics: map[string]string{
			"post":          os.Getenv("POST_TOPIC"),
			"user":          os.Getenv("USER_TOPIC"),
			"moderation":    os.Getenv("MODERATION_TOPIC"),
			"auth":          os.Getenv("AUTH_TOPIC"),
			"friend":        os.Getenv("FRIEND_TOPIC"),
			"message":       os.Getenv("MESSAGE_TOPIC"),
			"notification":  os.Getenv("NOTIFICATION_TOPIC"),
			"onlinehistory": os.Getenv("ONLINEHISTORY_TOPIC"),
			"otp":           os.Getenv("OTP_TOPIC"),
			"privacy":       os.Getenv("PRIVACY_TOPIC"),
		},
	}
}
