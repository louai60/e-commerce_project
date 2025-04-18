# Next.js E-Commerce Frontend

A modern, responsive e-commerce frontend built with Next.js, TypeScript, and Tailwind CSS. This project implements a robust e-commerce solution with features like product browsing, cart management, user authentication, and more.

## 🚀 Features

- **Modern Tech Stack**: Built with Next.js 15, React 19, and TypeScript
- **Responsive Design**: Fully responsive UI using Tailwind CSS
- **State Management**: Redux Toolkit for global state management
- **Authentication**: Secure authentication using NextAuth.js
- **UI Components**: 
  - Custom components with Framer Motion animations
  - Swiper for carousels and sliders
  - React Hot Toast for notifications
  - Lucide React for icons
- **API Integration**: Axios for API calls with SWR for data fetching
- **Form Validation**: Zxcvbn for password strength validation

## 📦 Prerequisites

- Node.js (v18 or higher)
- npm or yarn
- Git

## 🛠️ Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/louai60/e-commerce_project.git
   cd frontend
   ```

2. Install dependencies:
   ```bash
   npm install
   # or
   yarn install
   ```

3. Create a `.env.local` file in the root directory and add your environment variables:
   ```env
   NEXT_PUBLIC_API_URL=your_api_url
   NEXTAUTH_SECRET=your_nextauth_secret
   NEXTAUTH_URL=http://localhost:3000
   ```

## 🚀 Development

To run the development server:

```bash
npm run dev
# or
yarn dev
```

Open [http://localhost:3000](http://localhost:3000) with your browser to see the result.

## 📁 Project Structure

```
src/
├── app/          # Next.js 13+ app directory
├── components/   # Reusable UI components
├── contexts/     # React Context providers
├── hooks/        # Custom React hooks
├── lib/          # Utility functions and configurations
├── pages/        # Next.js pages (legacy)
├── redux/        # Redux store and slices
├── services/     # API service functions
└── types/        # TypeScript type definitions
```

## 🏗️ Build

To create a production build:

```bash
npm run build
# or
yarn build
```

To start the production server:

```bash
npm run start
# or
yarn start
```

## 🧪 Testing

```bash
npm run lint
# or
yarn lint
```

## 🔧 Configuration

- `next.config.js` - Next.js configuration
- `tailwind.config.ts` - Tailwind CSS configuration
- `tsconfig.json` - TypeScript configuration
- `.eslintrc.json` - ESLint configuration

## 📚 Dependencies

### Core
- Next.js 15.2.3
- React 19.0.0
- TypeScript 5.2.2

### Styling
- Tailwind CSS 3.3.3
- Framer Motion 12.6.3
- Swiper 10.2.0

### State Management & Data Fetching
- Redux Toolkit 2.6.1
- SWR 2.3.3
- Axios 1.8.4

### Authentication & Security
- NextAuth.js 4.24.11
- Zxcvbn 4.4.2

### UI Components
- Lucide React 0.487.0
- React Hot Toast 2.4.1
- React Icons 5.5.0

## 🤝 Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📝 License

This project is licensed under the MIT License - see the LICENSE file for details.

## 👥 Authors

- Your Name - Initial work

## 🙏 Acknowledgments

- Next.js team for the amazing framework
- Vercel for hosting and deployment
- All contributors and maintainers
