import React, { useState, useEffect } from 'react';
import { DataItem } from '@dha-team/arbundles';
import { Button } from './ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from './ui/card';
import { Badge } from './ui/badge';
import { FileText, CheckCircle, AlertCircle, Loader2 } from 'lucide-react';
import { useWallet } from '../contexts/WalletContext';

interface SigningInterfaceProps {
  uuid: string;
  data?: string;
  dataSize?: number;
  status?: 'pending' | 'signed' | 'error';
  serverUrl?: string;
  isTestMode?: boolean;
}

export const SigningInterface: React.FC<SigningInterfaceProps> = ({
  uuid,
  data,
  dataSize = 0,
  status = 'pending',
  serverUrl,
  isTestMode = false,
}) => {
  // Get server URL from props or URL search params or default to localhost
  const getServerUrl = (): string => {
    if (serverUrl) return serverUrl;
    const urlParams = new URLSearchParams(window.location.search);
    const serverUrlFromParams = urlParams.get('server');
    return serverUrlFromParams || 'http://localhost:8080';
  };

  const finalServerUrl = getServerUrl();
  const { walletConnected } = useWallet();
  const [isLoading, setIsLoading] = useState(false);
  const [signingStatus, setSigningStatus] = useState<{ type: 'idle' | 'pending' | 'success' | 'error'; message: string }>({
    type: 'idle',
    message: 'Ready to sign'
  });
  const [dataPreview, setDataPreview] = useState<string>(data || '');
  const [actualDataSize, setActualDataSize] = useState<number>(dataSize);

  useEffect(() => {
    fetchDataPreview();
  }, []);

  const fetchDataPreview = async () => {
    try {
      const response = await fetch(`${finalServerUrl}/${uuid}`);
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
    if (isTestMode) {
      setSigningStatus({ type: 'error', message: 'Test mode: Signing is disabled' });
      return;
    }

    if (!walletConnected || !window.arweaveWallet) {
      setSigningStatus({ type: 'error', message: 'Please connect your wallet first' });
      return;
    }

    try {
      setIsLoading(true);
      setSigningStatus({ type: 'pending', message: '⏳ Fetching data to sign...' });

      // Fetch the data to sign
      const response = await fetch(`${finalServerUrl}/${uuid}`);
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


      setSigningStatus({ type: 'pending', message: '⏳ Please confirm signing in your wallet...' });

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

      setSigningStatus({ type: 'pending', message: '⏳ Preparing signed data...' });

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

      setSigningStatus({ type: 'pending', message: '⏳ Submitting signed data...' });

      // Submit signed data back to server - use the raw signed data
      const submitResponse = await fetch(`${finalServerUrl}/${uuid}`, {
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
        setSigningStatus({
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
      setSigningStatus({
        type: 'error',
        message: `❌ Signing failed: ${error.message || 'Unknown error'}`
      });
    } finally {
      setIsLoading(false);
    }
  };

    return (
      <div className="min-h-screen bg-beigeLight">
        <div className="container mx-auto px-4 py-8 max-w-4xl">
          <div className="space-y-6">
            {/* Data Information Card */}
            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <FileText className="h-6 w-6" />
                  Data Information
                </CardTitle>
                <CardDescription>
                  Review the data you're about to sign
                </CardDescription>
              </CardHeader>
              <CardContent>
                <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-4">
                  <div className="bg-beigeMedium/20 p-4 rounded-lg">
                    <label className="block text-sm font-medium text-blackWarm mb-1">Request ID:</label>
                    <span className="font-mono text-xs text-blackTrue break-all">{uuid}</span>
                  </div>
                  <div className="bg-beigeMedium/20 p-4 rounded-lg">
                    <label className="block text-sm font-medium text-blackWarm mb-1">Data Size:</label>
                    <span className="text-blackTrue">{actualDataSize} bytes</span>
                  </div>
                  <div className="bg-beigeMedium/20 p-4 rounded-lg">
                    <label className="block text-sm font-medium text-blackWarm mb-1">Status:</label>
                    <Badge
                      variant={status === 'signed' ? 'default' : 'secondary'}
                      className={status === 'signed' ? 'bg-green-600 hover:bg-green-700' : ''}
                    >
                      {status === 'signed' ? (
                        <>
                          <CheckCircle className="h-3 w-3 mr-1" />
                          Signed
                        </>
                      ) : (
                        <>
                          <CheckCircle className="h-3 w-3 mr-1" />
                          Ready to Sign
                        </>
                      )}
                    </Badge>
                  </div>
                </div>

                {dataPreview && (
                  <div className="mt-4">
                    <h3 className="text-lg font-medium text-blackTrue mb-2">Data Preview:</h3>
                    <pre className="bg-beigeMedium/20 p-4 rounded-lg text-sm overflow-x-auto border border-beigeMedium font-mono text-blackTrue">
                      {dataPreview}
                    </pre>
                  </div>
                )}
              </CardContent>
            </Card>



            {/* Signing Section Card */}
            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <FileText className="h-6 w-6" />
                  Sign Data
                </CardTitle>
                <CardDescription>
                  Sign your data item with your connected wallet
                </CardDescription>
              </CardHeader>
              <CardContent>
                <div className={`p-4 rounded-lg mb-4 border ${
                  signingStatus.type === 'idle' ? 'bg-beigeMedium/20 border-beigeMedium' :
                  signingStatus.type === 'pending' ? 'bg-yellow-50 border-yellow-200' :
                  signingStatus.type === 'success' ? 'bg-green-50 border-green-200' :
                  'bg-red-50 border-red-200'
                }`}>
                  <div className="flex items-center gap-2">
                    {signingStatus.type === 'pending' && <Loader2 className="h-4 w-4 animate-spin" />}
                    {signingStatus.type === 'success' && <CheckCircle className="h-4 w-4 text-green-600" />}
                    {signingStatus.type === 'error' && <AlertCircle className="h-4 w-4 text-red-600" />}
                    <p className="font-medium text-blackTrue">
                      {signingStatus.message}
                    </p>
                  </div>
                </div>

                <Button
                  onClick={signDataItem}
                  disabled={!walletConnected || isLoading}
                  size="lg"
                  className="w-full"
                >
                  {isLoading ? (
                    <>
                      <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                      Signing...
                    </>
                  ) : (
                    <>
                      <FileText className="h-4 w-4 mr-2" />
                      Sign Data Item
                    </>
                  )}
                </Button>
              </CardContent>
            </Card>
          </div>
        </div>
      </div>
    );
};
