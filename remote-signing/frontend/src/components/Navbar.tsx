import { Github, BookOpen, Wallet } from 'lucide-react'
import { Button } from './ui/button'
import { useWallet } from '../contexts/WalletContext'

export const Navbar = () => {
  const { walletConnected, walletAddress, isLoading, connectWallet, disconnectWallet } = useWallet();
  return (
    <nav className="bg-beigeLight border-b border-beigeMedium sticky top-0 z-50">
      <div className="container mx-auto px-4 py-3">
        <div className="flex items-center justify-between">
          {/* Left side - Logo and Title */}
          <div className="flex items-center space-x-3">
                               <div className="flex items-center space-x-2">
                     <img
                       src="/harlequin_mascot.png"
                       alt="Harlequin Mascot"
                       className="h-8 w-auto object-contain"
                     />
                   </div>
            <div>
              <h1 className="text-xl font-bold text-blackTrue">Harlequin</h1>
              <p className="text-xs text-blackWarm">Remote Signer</p>
            </div>
          </div>

          {/* Right side - Navigation */}
          <div className="flex items-center space-x-2">
            {/* Wallet Connection */}
            <div className="flex items-center space-x-2">
              {!walletConnected ? (
                <Button
                  onClick={connectWallet}
                  disabled={isLoading}
                  variant="default"
                  size="sm"
                  className="flex items-center space-x-2"
                >
                  <Wallet className="h-4 w-4" />
                  <span>{isLoading ? 'Connecting...' : 'Connect Wallet'}</span>
                </Button>
              ) : (
                <div className="flex items-center space-x-2">
                  <div className="text-right">
                    <div className="text-xs text-blackWarm">Connected</div>
                    <div className="text-xs font-mono text-blackTrue">
                      {walletAddress?.slice(0, 6)}...{walletAddress?.slice(-4)}
                    </div>
                  </div>
                  <Button
                    onClick={disconnectWallet}
                    variant="outline"
                    size="sm"
                  >
                    Disconnect
                  </Button>
                </div>
              )}
            </div>

            {/* External Links */}
            <div className="flex items-center space-x-1">
              <Button
                variant="ghost"
                size="sm"
                asChild
              >
                <a
                  href="https://github.com/the-permaweb-harlequin/harlequin-toolkit/tree/main/remote-signing"
                  target="_blank"
                  rel="noopener noreferrer"
                  className="flex items-center space-x-1"
                >
                  <Github className="h-4 w-4" />
                  <span className="hidden sm:inline">GitHub</span>
                </a>
              </Button>

              <Button
                variant="ghost"
                size="sm"
                asChild
              >
                <a
                  href="https://docs_harlequin.ar.io"
                  target="_blank"
                  rel="noopener noreferrer"
                  className="flex items-center space-x-1"
                >
                  <BookOpen className="h-4 w-4" />
                  <span className="hidden sm:inline">Docs</span>
                </a>
              </Button>
            </div>
          </div>
        </div>
      </div>
    </nav>
  )
}
