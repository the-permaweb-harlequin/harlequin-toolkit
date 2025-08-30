import { useEffect, useState } from 'react'
import { SigningInterface } from './components/SigningInterface'

function App() {
  const [signingData, setSigningData] = useState<{
    uuid: string;
    serverUrl: string;
    dataSize: number;
  } | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string>('');

  useEffect(() => {
    // Parse UUID from URL path (e.g., /sign/uuid-here)
    const path = window.location.pathname;
    const pathMatch = path.match(/\/sign\/([a-f0-9-]+)/);

    if (pathMatch) {
      const uuid = pathMatch[1];
      // Get server URL from current location
      const serverUrl = `${window.location.protocol}//${window.location.host}`;
      // We'll fetch data size from the server when we load the signing request
      setSigningData({ uuid, serverUrl, dataSize: 0 });
    } else {
      setError('No signing request found. Invalid URL format.');
    }
    setLoading(false);
  }, []);

  if (loading) {
    return (
      <div className="min-h-screen bg-gradient-to-br from-blue-50 via-white to-indigo-50 flex items-center justify-center">
        <div className="text-center">
          <div className="loading-spinner mx-auto mb-4"></div>
          <p className="text-lg text-gray-600">Loading signing interface...</p>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="min-h-screen bg-gradient-to-br from-red-50 via-white to-pink-50 flex items-center justify-center">
        <div className="max-w-md mx-auto text-center bg-white rounded-xl shadow-lg p-8 border border-red-200">
          <h1 className="text-3xl font-bold text-red-600 mb-4">❌ Error</h1>
          <p className="text-gray-700 mb-2">{error}</p>
          <p className="text-gray-500">Please check the URL and try again.</p>
        </div>
      </div>
    );
  }

  if (!signingData) {
    return (
      <div className="min-h-screen bg-gradient-to-br from-blue-50 via-white to-indigo-50 flex items-center justify-center">
        <div className="max-w-lg mx-auto text-center bg-white rounded-xl shadow-lg p-8 border border-gray-200">
          <h1 className="text-4xl font-bold text-gray-800 mb-4">🎭 Harlequin Remote Signing</h1>
          <p className="text-lg text-gray-600 mb-2">Welcome to the remote signing interface.</p>
          <p className="text-gray-500">To sign a data item, you need a valid signing URL with a UUID parameter.</p>
        </div>
      </div>
    );
  }

  return (
    <SigningInterface
      uuid={signingData.uuid}
      serverUrl={signingData.serverUrl}
      dataSize={signingData.dataSize}
    />
  );
}

export default App
