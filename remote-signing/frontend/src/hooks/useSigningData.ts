import { useState, useEffect } from 'react'

export interface SigningData {
  uuid: string
  data: string
  dataSize: number
  status: 'pending' | 'signed' | 'error'
  createdAt: string
  signedAt?: string
  dataItemId?: string
}

export interface UseSigningDataReturn {
  signingData: SigningData | null
  isLoading: boolean
  error: string | null
  serverUrl: string
  refreshData: () => Promise<void>
}

// Test data for development
const TEST_SIGNING_DATA: SigningData = {
  uuid: 'test-uuid-12345-67890-abcdef',
  data: `Hello, Arweave!
This is a sample file for testing the Harlequin Remote Signing Library.
The library provides a simple way to upload data for signing and automatically handles:
- Server startup and management
- Web interface for signing
- Wallet integration with Wander/ArConnect
- Automatic bundler upload
- Real-time status updates

This is a longer piece of text to test how the interface handles larger content blocks.
It should wrap properly and maintain good readability.`,
  dataSize: 454,
  status: 'pending',
  createdAt: new Date().toISOString(),
}

export const useSigningData = (): UseSigningDataReturn => {
  const [signingData, setSigningData] = useState<SigningData | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  // Get server URL from URL search params or default to localhost
  const getServerUrl = (): string => {
    const urlParams = new URLSearchParams(window.location.search)
    const serverUrl = urlParams.get('server')
    return serverUrl || 'http://localhost:8080'
  }

  const [serverUrl] = useState(getServerUrl())

  const loadSigningData = async () => {
    setIsLoading(true)
    setError(null)

    try {
      // Check if we're on the test page
      const isTestPage = window.location.pathname === '/test'

      if (isTestPage) {
        // Use test data for development
        setSigningData(TEST_SIGNING_DATA)
        setIsLoading(false)
        return
      }

      // Extract UUID from URL path
      const pathParts = window.location.pathname.split('/')
      const uuid = pathParts[pathParts.length - 1]

      if (!uuid || uuid === 'test') {
        throw new Error('No UUID provided')
      }

      // Load real data from server
      const response = await fetch(`${serverUrl}/${uuid}`)

      if (!response.ok) {
        throw new Error(`Failed to load signing data: ${response.statusText}`)
      }

      const data = await response.json()

      setSigningData({
        uuid: data.uuid || uuid,
        data: data.data || '',
        dataSize: data.data_size || 0,
        status: data.is_signed ? 'signed' : 'pending',
        createdAt: data.created_at || new Date().toISOString(),
        signedAt: data.signed_at,
        dataItemId: data.data_item_id,
      })
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unknown error occurred')
    } finally {
      setIsLoading(false)
    }
  }

  const refreshData = async () => {
    await loadSigningData()
  }

  useEffect(() => {
    loadSigningData()
  }, [serverUrl])

  return {
    signingData,
    isLoading,
    error,
    serverUrl,
    refreshData,
  }
}
