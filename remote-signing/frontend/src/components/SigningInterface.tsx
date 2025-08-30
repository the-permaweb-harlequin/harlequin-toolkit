import React, { useState, useEffect } from 'react';
import { DataItem } from '@dha-team/arbundles';

interface SigningInterfaceProps {
  uuid: string;
  serverUrl: string;
  dataSize: number;
}

// Use the existing arweaveWallet type from arconnect but extend it
interface ExtendedArweaveWallet {
  connect: (permissions: string[]) => Promise<void>;
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  signDataItem?: (data: { data: Uint8Array; tags: Array<{ name: string; value: string }> }) => Promise<any>;
  disconnect: () => Promise<void>;
  getActiveAddress: () => Promise<string>;
}

export const SigningInterface: React.FC<SigningInterfaceProps> = ({
  uuid,
  serverUrl,
  dataSize,
}) => {
  const [walletConnected, setWalletConnected] = useState(false);
  const [walletAddress, setWalletAddress] = useState<string>('');
  const [isLoading, setIsLoading] = useState(false);
  const [status, setStatus] = useState<{ type: 'idle' | 'pending' | 'success' | 'error'; message: string }>({
    type: 'idle',
    message: 'Ready to sign'
  });
  const [dataPreview, setDataPreview] = useState<string>('');
  const [actualDataSize, setActualDataSize] = useState<number>(dataSize);

  useEffect(() => {
    checkWalletConnection();
    fetchDataPreview();
  }, []);

  const checkWalletConnection = async () => {
    if (window.arweaveWallet) {
      try {
        // eslint-disable-next-line @typescript-eslint/ban-ts-comment
        // @ts-ignore
        const address = await window.arweaveWallet.getActiveAddress();
        if (address) {
          setWalletConnected(true);
          setWalletAddress(address);
        }
      } catch {
        console.log('Wallet not connected yet');
      }
    }
  };

  const connectWallet = async () => {
    if (!window.arweaveWallet) {
      setStatus({ type: 'error', message: 'Wander/ArConnect wallet not detected. Please install the extension.' });
      return;
    }

    try {
      setIsLoading(true);
      // eslint-disable-next-line @typescript-eslint/ban-ts-comment
      // @ts-ignore
      await window.arweaveWallet.connect(['ACCESS_ADDRESS', 'SIGN_TRANSACTION']);
      // eslint-disable-next-line @typescript-eslint/ban-ts-comment
      // @ts-ignore
      const address = await window.arweaveWallet.getActiveAddress();
      setWalletConnected(true);
      setWalletAddress(address);
      setStatus({ type: 'success', message: `Connected to wallet: ${address.slice(0, 8)}...${address.slice(-8)}` });
    } catch (error) {
      console.error('Failed to connect wallet:', error);
      setStatus({ type: 'error', message: 'Failed to connect to wallet. Please try again.' });
    } finally {
      setIsLoading(false);
    }
  };

  const disconnectWallet = async () => {
    if (window.arweaveWallet) {
      try {
        const wallet = window.arweaveWallet as ExtendedArweaveWallet;
        await wallet.disconnect();
        setWalletConnected(false);
        setWalletAddress('');
        setStatus({ type: 'idle', message: 'Wallet disconnected' });
      } catch (error) {
        console.error('Failed to disconnect wallet:', error);
      }
    }
  };

  const fetchDataPreview = async () => {
    try {
      const response = await fetch(`${serverUrl}/${uuid}`);
      if (response.ok) {
        const dataItemJson = await response.json();

        // Show a preview of the actual data content
        const dataContent = dataItemJson.data || '';
        const preview = dataContent.length > 200 ? dataContent.slice(0, 200) + '...' : dataContent;
        setDataPreview(preview);

        // Update actual data size (size of the original data)
        setActualDataSize(new TextEncoder().encode(dataContent).length);
      }
    } catch (error) {
      console.error('Failed to fetch data preview:', error);
    }
  };



  const signDataItem = async () => {
    if (!walletConnected || !window.arweaveWallet) {
      setStatus({ type: 'error', message: 'Please connect your wallet first' });
      return;
    }

    try {
      setIsLoading(true);
      setStatus({ type: 'pending', message: '⏳ Fetching data to sign...' });

      // Fetch the data to sign
      const response = await fetch(`${serverUrl}/${uuid}`);
      if (!response.ok) {
        throw new Error(`Failed to fetch data: ${response.status}`);
      }

      // Parse the JSON DataItem structure sent by the CLI
      const dataItemJson = await response.json();

      console.log('Received DataItem JSON:', dataItemJson);

      // Extract the components
      const data = new TextEncoder().encode(dataItemJson.data);
      const tags = dataItemJson.tags || [];
      const anchor = dataItemJson.anchor || '';

      console.log({
        data: data,
        tags: tags,
        anchor: anchor,
        dataSize: data.length
      })


      setStatus({ type: 'pending', message: '⏳ Please confirm signing in your wallet...' });

      // Sign the data item using wallet's signDataItem method
      // eslint-disable-next-line @typescript-eslint/ban-ts-comment
      // @ts-ignore
      const signed = await window.arweaveWallet.signDataItem({
        data: data,
        tags:tags,
        anchor:anchor
      });

      console.log({
        signed
      })

      setStatus({ type: 'pending', message: '⏳ Preparing signed data...' });

      // Handle different types of signed data from wallet
      let signedData;

      console.log('Signed data type:', typeof signed, signed.constructor?.name);

      if (signed instanceof ArrayBuffer) {
        signedData = Buffer.from(signed);
      } else if (signed instanceof Uint8Array) {
        signedData = Buffer.from(signed);
      } else if (typeof signed === 'object' && signed.constructor === Object) {
        // If it's a plain object, it might need special handling
        console.log('Signed data is object:', signed);
        // Try to extract the actual data
        if (signed.data) {
          signedData = Buffer.from(signed.data);
        } else {
          signedData = Buffer.from(Object.values(signed));
        }
      } else {
        // Default conversion - this might be where DataView is needed
        console.log('Converting signed data with default method');
        signedData = Buffer.from(signed);
      }

      // Create DataItem from signed data
      const dataItem = new DataItem(signedData);

      console.log({
        dataItem,
        dataItemId: dataItem.id
      });

      setStatus({ type: 'pending', message: '⏳ Submitting signed data...' });

      // Submit signed data back to server - use the raw signed data
      const submitResponse = await fetch(`${serverUrl}/${uuid}`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/octet-stream'
        },
        body: signed
      });

      console.log({
        submitResponse
      })

      if (submitResponse.ok) {
        setStatus({
          type: 'success',
          message: `✅ Data item signed successfully! ID: ${dataItem.id.slice(0, 12)}...`
        });
      } else {
        const errorData = await submitResponse.json();
        throw new Error(errorData.error || 'Failed to submit signed data');
      }

      // eslint-disable-next-line @typescript-eslint/no-explicit-any
    } catch (error: any) {
      console.error('Signing failed:', error);
      setStatus({
        type: 'error',
        message: `❌ Signing failed: ${error.message || 'Unknown error'}`
      });
    } finally {
      setIsLoading(false);
    }
  };

    return (
    <div className="min-h-screen bg-gradient-to-br from-blue-50 via-white to-indigo-50">
      <div className="container mx-auto px-4 py-8 max-w-4xl">
        <header className="text-center mb-8">
          <h1 className="text-4xl font-bold text-gray-800 mb-2">🎭 Harlequin Remote Signing</h1>
          <p className="text-lg text-gray-600">Sign your data item securely with your Arweave wallet</p>
        </header>

        <div className="space-y-6">
          {/* Data Information Card */}
          <div className="bg-white rounded-xl shadow-lg p-6 border border-gray-200">
            <h2 className="text-2xl font-semibold text-gray-800 mb-4 flex items-center">
              📋 Data Information
            </h2>
            <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-4">
              <div className="bg-gray-50 p-4 rounded-lg">
                <label className="block text-sm font-medium text-gray-600 mb-1">Request ID:</label>
                <span className="mono text-xs text-gray-800 break-all">{uuid}</span>
              </div>
              <div className="bg-gray-50 p-4 rounded-lg">
                <label className="block text-sm font-medium text-gray-600 mb-1">Data Size:</label>
                <span className="text-gray-800">{actualDataSize} bytes</span>
              </div>
              <div className="bg-gray-50 p-4 rounded-lg">
                <label className="block text-sm font-medium text-gray-600 mb-1">Status:</label>
                <span className="font-medium text-green-600">
                  🟢 Ready to Sign
                </span>
              </div>
            </div>

            {dataPreview && (
              <div className="mt-4">
                <h3 className="text-lg font-medium text-gray-700 mb-2">Data Preview:</h3>
                <pre className="bg-gray-100 p-4 rounded-lg text-sm overflow-x-auto border">
                  {dataPreview}
                </pre>
              </div>
            )}
          </div>

          {/* Wallet Connection Card */}
          <div className="bg-white rounded-xl shadow-lg p-6 border border-gray-200">
            <h2 className="text-2xl font-semibold text-gray-800 mb-4 flex items-center">
              🔐 Wallet Connection
            </h2>
            {!walletConnected ? (
              <button
                onClick={connectWallet}
                disabled={isLoading}
                className={`btn btn-lg w-full ${isLoading ? 'btn-disabled' : 'btn-primary'}`}
              >
                {isLoading ? (
                  <span className="flex items-center justify-center">
                    <div className="loading-spinner mr-2"></div>
                    Connecting...
                  </span>
                ) : (
                  'Connect Wander/ArConnect Wallet'
                )}
              </button>
            ) : (
              <div className="bg-green-50 border border-green-200 rounded-lg p-4">
                <div className="flex items-center justify-between">
                  <div className="flex items-center space-x-3">
                    <span className="text-2xl">✅</span>
                    <div>
                      <div className="text-sm font-medium text-gray-600">Connected to:</div>
                      <div className="mono text-sm text-gray-800 break-all">{walletAddress}</div>
                    </div>
                  </div>
                  <button onClick={disconnectWallet} className="btn btn-secondary btn-sm">
                    Disconnect
                  </button>
                </div>
              </div>
            )}
          </div>

          {/* Signing Section Card */}
          <div className="bg-white rounded-xl shadow-lg p-6 border border-gray-200">
            <h2 className="text-2xl font-semibold text-gray-800 mb-4 flex items-center">
              ✍️ Sign Data
            </h2>

            <div className={`p-4 rounded-lg mb-4 border ${
              status.type === 'idle' ? 'bg-gray-50 border-gray-200' :
              status.type === 'pending' ? 'bg-yellow-50 border-yellow-200' :
              status.type === 'success' ? 'bg-green-50 border-green-200' :
              'bg-red-50 border-red-200'
            }`}>
              <p className={`status-${status.type} font-medium`}>
                {status.message}
              </p>
            </div>

            <button
              onClick={signDataItem}
              disabled={!walletConnected || isLoading}
              className={`btn btn-lg w-full ${
                !walletConnected || isLoading ? 'btn-disabled' : 'btn-success'
              }`}
            >
              {isLoading ? (
                <span className="flex items-center justify-center">
                  <div className="loading-spinner mr-2"></div>
                  Signing...
                </span>
              ) : (
                'Sign Data Item'
              )}
            </button>
          </div>
        </div>

        <footer className="text-center mt-8 text-gray-500">
          <p>Powered by Harlequin Toolkit • Secure • Decentralized</p>
        </footer>
      </div>
    </div>
  );
};
