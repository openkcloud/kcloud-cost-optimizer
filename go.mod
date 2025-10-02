module github.com/kcloud-opt/policy

go 1.21

require (
	github.com/gin-gonic/gin v1.9.1
	github.com/gin-contrib/cors v1.5.0
	github.com/spf13/cobra v1.8.0
	github.com/spf13/viper v1.17.0
	go.uber.org/zap v1.26.0
	go.uber.org/zap/zapcore v1.26.0
	
	// Kubernetes
	k8s.io/api v0.28.3
	k8s.io/apimachinery v0.28.3
	k8s.io/client-go v0.28.3
	sigs.k8s.io/controller-runtime v0.16.3
	
	// Database
	github.com/go-redis/redis/v8 v8.11.5
	gorm.io/gorm v1.25.5
	gorm.io/driver/postgres v1.5.3
	
	// Policy Engine
	github.com/open-policy-agent/opa v0.58.0
	github.com/expr-lang/expr v1.15.0
	
	// Monitoring
	github.com/prometheus/client_golang v1.17.0
	
	// YAML/JSON
	gopkg.in/yaml.v3 v3.0.1
	github.com/tidwall/gjson v1.17.0
	github.com/xeipuuv/gojsonschema v1.2.0
	
	// Testing
	github.com/stretchr/testify v1.8.4
)