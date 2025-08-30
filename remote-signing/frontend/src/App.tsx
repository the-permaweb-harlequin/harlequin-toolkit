import { useEffect, useState } from 'react'
import { SigningInterface } from './components/SigningInterface'
import { Navbar } from './components/Navbar'
import { Footer } from './components/Footer'
import { WalletProvider } from './contexts/WalletContext'
import { TestPage } from './pages/TestPage'

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

    // Check if we're on the test page
    if (path === '/test') {
      setSigningData(null); // Test page will handle its own data
      setLoading(false);
      return;
    }

    const pathMatch = path.match(/\/sign\/([a-f0-9-]+)/);

    if (pathMatch) {
      const uuid = pathMatch[1];
      // Get server URL from search params or default to current location
      const urlParams = new URLSearchParams(window.location.search);
      const serverUrl = urlParams.get('server') || `${window.location.protocol}//${window.location.host}`;
      // We'll fetch data size from the server when we load the signing request
      setSigningData({ uuid, serverUrl, dataSize: 0 });
    } else {
      setError('No signing request found. Invalid URL format.');
    }
    setLoading(false);
  }, []);

  if (loading) {
    return (
      <div className="min-h-screen bg-beigeLight flex items-center justify-center">
        <div className="text-center">
          <div className="loading-spinner mx-auto mb-4"></div>
          <p className="text-lg text-black-warm">Loading signing interface...</p>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="min-h-screen bg-beigeLight flex items-center justify-center">
        <div className="max-w-md mx-auto text-center bg-beigeLight rounded-xl shadow-lg p-8 border border-red-200">
          <h1 className="text-3xl font-bold text-redDark mb-4">❌ Error</h1>
          <p className="text-blackTrue mb-2">{error}</p>
          <p className="text-blackWarm">Please check the URL and try again.</p>
        </div>
      </div>
    );
  }

  // Check if we're on the test page
  if (window.location.pathname === '/test') {
    return (
      <WalletProvider>
        <div className="min-h-screen bg-beigeLight">
          <Navbar />
          <TestPage />
          <Footer />
        </div>
      </WalletProvider>
    );
  }

  if (!signingData) {
    return (
      <div className="min-h-screen bg-beigeLight flex items-center justify-center">
        <div className="max-w-lg mx-auto text-center bg-beigeLight rounded-xl shadow-lg p-8 border border-beigeMedium">
          <h1 className="text-4xl font-bold text-blackTrue mb-4">🎭 Harlequin Remote Signing</h1>
          <p className="text-lg text-blackWarm mb-2">Welcome to the remote signing interface.</p>
          <p className="text-blackWarm">To sign a data item, you need a valid signing URL with a UUID parameter.</p>
        </div>
      </div>
    );
  }

  return (
    <WalletProvider>
      <div className="min-h-screen bg-beigeLight">
        <Navbar />
        <SigningInterface
          uuid={signingData.uuid}
          serverUrl={signingData.serverUrl}
          dataSize={signingData.dataSize}
        />
        <Footer />
      </div>
    </WalletProvider>
  );
}

export default App
