package global

import (
	"search-radius/pkg/common/cache"
	"search-radius/pkg/logger"
	"search-radius/pkg/mq/kafka"
	"search-radius/pkg/settings"
)

var (
	Logger        *logger.LoggerZap
	Config        *settings.Config
	KafkaProducer kafka.Producer
	Redis         cache.CacheEngine
)
