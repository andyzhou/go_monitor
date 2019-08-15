package base

import (
	"google.golang.org/grpc/stats"
	"context"
)

/*
 * Basic function face
 */

//connect ctx key info
type ConnCtxKey struct{}

//basic face info
type Basic struct {}

//get connect tag from context
func (b *Basic) GetConnTagFromContext(ctx context.Context) (*stats.ConnTagInfo, bool) {
	tag, ok := ctx.Value(ConnCtxKey{}).(*stats.ConnTagInfo)
	return tag, ok
}