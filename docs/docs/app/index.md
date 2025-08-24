# App Overview

The Harlequin App is a modern React-based web application for managing and visualizing your Permaweb applications and data.

## Features

- **Project Management**: Create, configure, and manage multiple projects
- **Real-time Monitoring**: Monitor deployments and transaction status
- **Data Browser**: Explore Arweave data with advanced filtering
- **Wallet Integration**: Connect popular Arweave wallets
- **Dark/Light Theme**: Customizable UI theme
- **Internationalization**: Multi-language support

## Architecture

Built with modern web technologies:

- **React 18**: Latest React features with concurrent rendering
- **TypeScript**: Full type safety and developer experience
- **Vite**: Lightning-fast development and build
- **Tailwind CSS**: Utility-first styling
- **React Router**: Client-side routing
- **React Query**: Data fetching and caching

## Getting Started

### Running Locally

```bash
# Navigate to the app directory
cd app

# Install dependencies
yarn

# Start development server
yarn dev
```

### Building for Production

```bash
# Build the application
yarn build

# Preview the production build
yarn preview
```

## Configuration

The app uses environment variables for configuration:

```env
# .env.local
VITE_ARWEAVE_GATEWAY=https://arweave.net
VITE_APP_NAME=Harlequin Toolkit
VITE_APP_VERSION=1.0.0
```

## Features Overview

### Project Dashboard

- Overview of all your Permaweb projects
- Quick access to build and deployment status
- Recent activity and notifications

### Data Browser

- Explore transactions and data on Arweave
- Advanced search and filtering capabilities
- Download and view transaction data

### Wallet Manager

- Connect multiple Arweave wallets
- View wallet balance and transactions
- Manage permissions and settings

### Deployment Center

- Deploy applications directly from the UI
- Monitor deployment progress
- View deployment logs and status

## Next Steps

- [Component Library](/app/components) - Reusable UI components
- [State Management](/app/state) - Application state patterns
- [Theming Guide](/app/theming) - Customizing the UI theme
