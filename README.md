# E-Commerce Platform

A modern, scalable e-commerce platform built with a microservices architecture. This project implements a complete e-commerce solution with features like product management, user authentication, order processing, payment integration, and more.

## 🏗️ Architecture

The platform is built using a microservices architecture with the following components:

### Frontend
- **Next.js E-Commerce Frontend**: Modern, responsive UI built with Next.js, React, and Tailwind CSS
- [Frontend Documentation](./frontend/README.md)

### Backend Microservices
- **API Gateway**: Central entry point for all client requests
- **User Service**: Handles user authentication, authorization, and profile management
- **Product Service**: Manages product catalog, categories, and product details
- **Order Service**: Processes orders and manages order history
- **Payment Service**: Handles payment processing and transactions
- **Inventory Service**: Manages product inventory and stock levels
- **Admin Service**: Provides administrative functions and dashboard
- **Recommendation Service**: Generates product recommendations

## 🚀 Features

### Frontend
- Modern, responsive UI with Next.js and Tailwind CSS
- Redux for state management
- NextAuth.js for authentication
- Framer Motion animations
- SWR for data fetching
- Form validation with Zxcvbn

### Backend
- Microservices architecture with Go
- gRPC for service-to-service communication
- PostgreSQL for data storage
- Redis for caching
- JWT for authentication
- Docker containerization
- Database migrations with golang-migrate

## 📦 Prerequisites

### Frontend
- Node.js (v18 or higher)
- npm or yarn
- Git

### Backend
- Go 1.21 or higher
- PostgreSQL 14 or higher
- Redis 6 or higher
- Docker and Docker Compose
- Make (optional, for using Makefile commands)

## 🛠️ Installation

### Clone the Repository
```bash
git clone https://github.com/louai60/e-commerce_project.git
cd e-commerce_project
```

### Frontend Setup
```bash
cd frontend
npm install
# or
yarn install

# Create .env.local file
cp .env.example .env.local
# Edit .env.local with your configuration
```

### Backend Setup
```bash
cd backend

# Initialize development environment
./init-dev.bat  # Windows
# or
./init-dev.sh   # Linux/Mac

# Start all services
./dev.bat       # Windows
# or
./dev.sh        # Linux/Mac
```

## 🚀 Development

### Frontend Development
```bash
cd frontend
npm run dev
# or
yarn dev
```

### Backend Development
```bash
cd backend
./dev.bat  # Windows
# or
./dev.sh   # Linux/Mac
```

## 📁 Project Structure

```
e-commerce_project/
├── frontend/                # Next.js frontend application
│   ├── src/                 # Source code
│   │   ├── app/             # Next.js 13+ app directory
│   │   ├── components/      # Reusable UI components
│   │   ├── contexts/        # React Context providers
│   │   ├── hooks/           # Custom React hooks
│   │   ├── lib/             # Utility functions
│   │   ├── pages/           # Next.js pages (legacy)
│   │   ├── redux/           # Redux store and slices
│   │   ├── services/        # API service functions
│   │   └── types/           # TypeScript type definitions
│   └── public/              # Static assets
│
├── backend/                 # Backend microservices
│   ├── api-gateway/         # API Gateway service
│   ├── user-service/        # User management service
│   ├── product-service/     # Product catalog service
│   ├── order-service/       # Order processing service
│   ├── payment-service/     # Payment processing service
│   ├── inventory-service/   # Inventory management service
│   ├── admin-service/       # Administrative functions
│   ├── recommendation-service/ # Product recommendations
│   ├── shared/              # Shared utilities and types
│   └── common/              # Common code across services
│
└── docs/                    # Project documentation
```

## 🏗️ Build

### Frontend Build
```bash
cd frontend
npm run build
# or
yarn build
```

### Backend Build
```bash
cd backend
go build -o bin/api-gateway ./api-gateway
go build -o bin/user-service ./user-service
# ... build other services
```

## 🧪 Testing

### Frontend Testing
```bash
cd frontend
npm run lint
# or
yarn lint
```

### Backend Testing
```bash
cd backend
go test ./...
```

## 🔧 Configuration

### Frontend
- `next.config.js` - Next.js configuration
- `tailwind.config.ts` - Tailwind CSS configuration
- `tsconfig.json` - TypeScript configuration
- `.eslintrc.json` - ESLint configuration

### Backend
- `.env` files in each service directory
- `docker-compose.yml` for container configuration
- Database migration files in each service

## 📚 Dependencies

### Frontend
See [Frontend README](./frontend/README.md) for detailed frontend dependencies.

### Backend
- Go 1.21+
- gRPC
- PostgreSQL
- Redis
- Docker
- golang-migrate
- JWT
- Zap (logging)
- Viper (configuration)

## 🤝 Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📝 License

This project is licensed under the MIT License - see the LICENSE file for details.

## 👥 Authors

- Louai - Initial work

## 🙏 Acknowledgments

- Next.js team for the amazing framework
- Go team for the excellent language and tools
- All contributors and maintainers