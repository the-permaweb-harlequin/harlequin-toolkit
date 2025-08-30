import React, { createContext, useContext, useState, useEffect, type ReactNode } from 'react';

interface WalletContextType {
  walletConnected: boolean;
  walletAddress: string;
  isLoading: boolean;
  connectWallet: () => Promise<void>;
  disconnectWallet: () => Promise<void>;
}

const WalletContext = createContext<WalletContextType | undefined>(undefined);

export const useWallet = () => {
  const context = useContext(WalletContext);
  if (context === undefined) {
    throw new Error('useWallet must be used within a WalletProvider');
  }
  return context;
};

interface WalletProviderProps {
  children: ReactNode;
}

export const WalletProvider: React.FC<WalletProviderProps> = ({ children }) => {
  const [walletConnected, setWalletConnected] = useState(false);
  const [walletAddress, setWalletAddress] = useState('');
  const [isLoading, setIsLoading] = useState(false);

  useEffect(() => {
    checkWalletConnection();
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
      throw new Error('Wander/ArConnect wallet not detected. Please install the extension.');
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
    } catch (error) {
      console.error('Failed to connect wallet:', error);
      throw error;
    } finally {
      setIsLoading(false);
    }
  };

  const disconnectWallet = async () => {
    if (window.arweaveWallet) {
      try {
        // eslint-disable-next-line @typescript-eslint/ban-ts-comment
        // @ts-ignore
        await window.arweaveWallet.disconnect();
        setWalletConnected(false);
        setWalletAddress('');
      } catch (error) {
        console.error('Failed to disconnect wallet:', error);
        throw error;
      }
    }
  };

  const value = {
    walletConnected,
    walletAddress,
    isLoading,
    connectWallet,
    disconnectWallet,
  };

  return (
    <WalletContext.Provider value={value}>
      {children}
    </WalletContext.Provider>
  );
};
