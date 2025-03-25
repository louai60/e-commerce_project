package clients

// import (
//     "github.com/louai60/e-commerce_project/backend/api-gateway/config"
//     userPb "github.com/louai60/e-commerce_project/backend/user-service/proto"
//     "google.golang.org/grpc"
//     "google.golang.org/grpc/credentials/insecure"
// )

// type ServiceClients struct {
//     UserClient    userPb.UserServiceClient
//     userConn     *grpc.ClientConn
//     // Add other service clients here
// }

// func NewServiceClients(cfg *config.Config) (*ServiceClients, error) {
//     // Initialize user service client
//     userConn, err := grpc.Dial(
//         cfg.Services.UserService.Address,
//         grpc.WithTransportCredentials(insecure.NewCredentials()),
//     )
//     if err != nil {
//         return nil, err
//     }

//     return &ServiceClients{
//         UserClient: userPb.NewUserServiceClient(userConn),
//         userConn:  userConn,
//     }, nil
// }

// func (s *ServiceClients) Close() {
//     if s.userConn != nil {
//         s.userConn.Close()
//     }
//     // Close other connections
