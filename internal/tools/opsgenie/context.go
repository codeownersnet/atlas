package opsgenie

import (
	"context"

	"github.com/codeownersnet/atlas/pkg/atlassian/opsgenie"
)

// Context key for storing Opsgenie client
type contextKey string

const opsgenieClientKey contextKey = "opsgenie_client"

// WithOpsgenieClient adds an Opsgenie client to the context
func WithOpsgenieClient(ctx context.Context, client *opsgenie.Client) context.Context {
	return context.WithValue(ctx, opsgenieClientKey, client)
}

// GetOpsgenieClient retrieves the Opsgenie client from the context
func GetOpsgenieClient(ctx context.Context) *opsgenie.Client {
	client, ok := ctx.Value(opsgenieClientKey).(*opsgenie.Client)
	if !ok {
		return nil
	}
	return client
}
