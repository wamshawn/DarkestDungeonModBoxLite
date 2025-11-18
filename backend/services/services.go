package services

import "context"

type Load func() (service any, startup func(ctx context.Context), shutdown func(ctx context.Context))
