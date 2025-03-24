# Key Points

1. Protocol Translations:
   - Frontend ↔ API Gateway: HTTP/REST
   - API Gateway ↔ Microservices: gRPC
   - Microservices ↔ Database: SQL

2. Advantages of this Architecture:
   - Separation of concerns
   - Independent scaling
   - Protocol optimization (gRPC for internal, REST for external)
   - Service isolation

3. Communication Ports:
   - Next.js Frontend: 3000
   - API Gateway: 8080
   - Product Service: 50051
   - User Service: 50052
   - Other Services: 50053+
   - PostgreSQL: 5432

4. Data Flow:
   Request:  Frontend → API Gateway → Microservice → Database
   Response: Database → Microservice → API Gateway → Frontend

5. Data Flow Example: "Get Product Details"
   Next.js → GET http://api-gateway:8080/products/123 (HTTP).

   API Gateway:
      Validates JWT.
      Converts HTTP → gRPC → calls ProductService.GetProduct(id=123).

   Product Service (gRPC):
      Queries DB: SELECT * FROM products WHERE id=123.

   Response:
      DB → Product Service → API Gateway (gRPC → JSON) → Next.js.