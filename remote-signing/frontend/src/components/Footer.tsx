export const Footer = () => {
  return (
    <footer className="bg-beigeLight border-t border-beigeMedium mt-8">
      <div className="container mx-auto px-4 py-6">
        <div className="text-center">
          <p className="text-sm text-blackWarm">
            Powered by{' '}
            <span className="font-semibold text-blackTrue">Harlequin Toolkit</span>
            {' '}• Secure • Decentralized
          </p>
          <p className="text-xs text-blackWarm mt-2">
            Sign your data items securely with your Arweave wallet
          </p>
        </div>
      </div>
    </footer>
  )
}
