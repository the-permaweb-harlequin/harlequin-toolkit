import React from 'react'
import { useSigningData } from '../hooks/useSigningData'
import { SigningInterface } from '../components/SigningInterface'
import { Card, CardContent, CardHeader, CardTitle } from '../components/ui/card'
import { Badge } from '../components/ui/badge'
import { FileText, AlertCircle, Loader2 } from 'lucide-react'

export const TestPage: React.FC = () => {
  const { signingData, isLoading, error } = useSigningData()

  if (isLoading) {
    return (
      <div className="min-h-screen bg-beigeLight flex items-center justify-center">
        <div className="text-center">
          <Loader2 className="h-8 w-8 animate-spin text-redDark mx-auto mb-4" />
          <p className="text-blackWarm">Loading test data...</p>
        </div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="min-h-screen bg-beigeLight flex items-center justify-center">
        <Card className="max-w-md">
          <CardHeader>
            <CardTitle className="flex items-center space-x-2 text-redDark">
              <AlertCircle className="h-5 w-5" />
              <span>Error</span>
            </CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-skinLight">{error}</p>
          </CardContent>
        </Card>
      </div>
    )
  }

  if (!signingData) {
    return (
      <div className="min-h-screen bg-beigeLight flex items-center justify-center">
        <Card className="max-w-md">
          <CardHeader>
            <CardTitle>No Data Available</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-skinLight">No signing data found.</p>
          </CardContent>
        </Card>
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-beigeLight">
      <div className="container mx-auto px-4 py-8">
        {/* Test Page Header */}
        <div className="mb-8">
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center space-x-2">
                <FileText className="h-5 w-5" />
                <span>Development Test Page</span>
                <Badge variant="outline" className="ml-2">
                  Test Mode
                </Badge>
              </CardTitle>
            </CardHeader>
            <CardContent>
              <p className="text-blackBrown mb-4">
                This is a development test page for styling the signing interface.
                The data below is mock data and will not actually sign anything.
              </p>
              <div className="grid grid-cols-1 md:grid-cols-3 gap-4 text-sm">
                <div>
                  <span className="font-semibold text-blackBrown">UUID:</span>
                  <p className="text-blackBrown font-mono break-all">{signingData.uuid}</p>
                </div>
                <div>
                  <span className="font-semibold text-blackBrown">Status:</span>
                  <Badge
                    variant={signingData.status === 'signed' ? 'default' : 'secondary'}
                    className="ml-2"
                  >
                    {signingData.status}
                  </Badge>
                </div>
                <div>
                  <span className="font-semibold text-blackBrown">Data Size:</span>
                  <p className="text-blackBrown">{signingData.dataSize} bytes</p>
                </div>
              </div>
            </CardContent>
          </Card>
        </div>

        {/* Signing Interface */}
        <SigningInterface
          uuid={signingData.uuid}
          data={signingData.data}
          dataSize={signingData.dataSize}
          status={signingData.status}
          isTestMode={true}
        />
      </div>
    </div>
  )
}
