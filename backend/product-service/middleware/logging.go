package middleware

import (
    "context"
    "time"

    "go.uber.org/zap"
    "google.golang.org/grpc"
    "google.golang.org/grpc/status"
)

func LoggingInterceptor(logger *zap.Logger) grpc.UnaryServerInterceptor {
    return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
        start := time.Now()
        
        // Execute the handler
        resp, err := handler(ctx, req)
        
        // Log the request
        duration := time.Since(start)
        if err != nil {
            st, _ := status.FromError(err)
            logger.Error("gRPC request failed",
                zap.String("method", info.FullMethod),
                zap.Duration("duration", duration),
                zap.String("code", st.Code().String()),
                zap.Error(err),
            )
        } else {
            logger.Info("gRPC request successful",
                zap.String("method", info.FullMethod),
                zap.Duration("duration", duration),
            )
        }
        
        return resp, err
    }
}