package ray

type Middleware func(next Handler) Handler
