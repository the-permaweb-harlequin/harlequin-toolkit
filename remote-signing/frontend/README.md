# Harlequin Remote Signing Frontend

This is the frontend application for the Harlequin Remote Signing service. It can be deployed as a static site and configured to connect to different signing servers.

## 🚀 Usage

### Test Page

For development and testing, visit the test page:

```
http://localhost:8080/test
```

### Signing Pages

For actual signing workflows, use URLs with UUIDs:

```
http://localhost:8080/sign/<uuid>
```

## 🔧 Server Configuration

The frontend can connect to different signing servers using URL parameters:

### Default (Same Host)

If no server parameter is provided, it defaults to the same host:

```
http://localhost:8080/sign/123e4567-e89b-12d3-a456-426614174000
```

### Custom Server

Specify a different server using the `server` parameter:

```
http://localhost:8080/sign/123e4567-e89b-12d3-a456-426614174000?server=https://my-signing-server.com
```

### Test Page with Custom Server

```
http://localhost:8080/test?server=https://my-signing-server.com
```

## 📦 Deployment

### Static Site Deployment

The frontend can be deployed as a static site to any hosting service:

1. Build the application:

   ```bash
   yarn build
   ```

2. Deploy the `dist/` folder to your hosting service (Netlify, Vercel, GitHub Pages, etc.)

3. Configure the server URL when linking to the signing interface:
   ```html
   <a href="https://my-frontend.com/sign/123e4567-e89b-12d3-a456-426614174000?server=https://my-signing-server.com">
     Sign Document
   </a>
   ```

### Environment Configuration

You can also set a default server URL by modifying the `getServerUrl()` function in the components.

## 🎨 Development

### Running Locally

```bash
yarn dev
```

### Building for Production

```bash
yarn build
```

### Test Mode

The test page (`/test`) provides mock data for development and styling without requiring a real signing server.

## 🔗 Integration Examples

### CLI Integration

When the CLI opens the signing interface, it can specify the server:

```go
url := fmt.Sprintf("https://my-frontend.com/sign/%s?server=%s", uuid, serverURL)
```

### Direct Links

Create direct links to signing requests:

```html
<a href="https://my-frontend.com/sign/abc123?server=https://signing.example.com"> Sign Document </a>
```

## 🛠️ Architecture

- **React + TypeScript**: Modern frontend framework
- **Tailwind CSS**: Utility-first styling
- **shadcn/ui**: Component library
- **Vite**: Build tool and dev server
- **URL Parameters**: Server configuration via query strings

The frontend is designed to be completely decoupled from the signing server, making it easy to deploy separately and connect to different server instances.
