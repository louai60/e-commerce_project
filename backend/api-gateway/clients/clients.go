package clients

import (
    "fmt"

    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"

    "github.com/louai60/e-commerce_project/backend/api-gateway/config"
    // Import service protos
    productpb "github.com/louai60/e-commerce_project/backend/product-service/proto"
    // userpb "github.com/louai60/e-commerce_project/backend/user-service/proto"
    // orderpb "github.com/louai60/e-commerce_project/backend/order-service/proto"
    // cartpb "github.com/louai60/e-commerce_project/backend/cart-service/proto"
    // Import other service protos
)

type ServiceClients struct {
    ProductClient       productpb.ProductServiceClient
    // UserClient         userpb.UserServiceClient
    // OrderClient        orderpb.OrderServiceClient
    // CartClient         cartpb.CartServiceClient
    // Add other service clients
    
    connections []*grpc.ClientConn
}

func NewServiceClients(cfg *config.Config) (*ServiceClients, error) {
    sc := &ServiceClients{
        connections: make([]*grpc.ClientConn, 0),
    }

    // Product Service
    productConn, err := sc.createConnection(cfg.Services.Product)
    if err != nil {
        return nil, fmt.Errorf("failed to connect to product service: %v", err)
    }
    sc.ProductClient = productpb.NewProductServiceClient(productConn)
    sc.connections = append(sc.connections, productConn)

    // User Service
    userConn, err := sc.createConnection(cfg.Services.User)
    if err != nil {
        return nil, fmt.Errorf("failed to connect to user service: %v", err)
    }
    // sc.UserClient = userpb.NewUserServiceClient(userConn)
    sc.connections = append(sc.connections, userConn)

    // Order Service
    orderConn, err := sc.createConnection(cfg.Services.Order)
    if err != nil {
        return nil, fmt.Errorf("failed to connect to order service: %v", err)
    }
    // sc.OrderClient = orderpb.NewOrderServiceClient(orderConn)
    sc.connections = append(sc.connections, orderConn)

    // Add other service connections similarly

    return sc, nil
}

func (sc *ServiceClients) createConnection(cfg config.ServiceConfig) (*grpc.ClientConn, error) {
    return grpc.Dial(
        fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
        grpc.WithTransportCredentials(insecure.NewCredentials()),
    )
}

func (sc *ServiceClients) Close() {
    for _, conn := range sc.connections {
        conn.Close()
    }
}