package auth

import (
	"context"
	"log"
	"sync"

	"github.com/welibekov/cluster-api-provider-ptc/pkg/ptc/auth/local"
	"github.com/welibekov/cluster-api-provider-ptc/pkg/ptc/auth/types"
)

const authClientDriverDefault = "local"

var (
	// Global instance of the AuthManager
	GlobalClient types.ClientAuthManager
	onceClient   sync.Once
)

func InitClient() {
	onceClient.Do(func() {
		authClientDriver := authClientDriverDefault

		switch authClientDriver {
		case "local":
			GlobalClient = local.NewFromConfig(context.TODO())
		default:
			log.Fatalf("unknown auth driver: %s\n", authClientDriver)
		}
	})
}
