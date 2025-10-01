import type { Metadata } from 'next'
import './globals.css'

export const metadata: Metadata = {
  title: 'Axle - Real-time File Sync',
  description: 'Real-time file synchronization for hackathon teams. No accounts. No cloud. Just sync.',
}

export default function RootLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <html lang="en">
      <head>
        <link rel="icon" type="image/svg+xml" href="/axle-file-sync/favicon.svg" />
        <link rel="preconnect" href="https://fonts.googleapis.com" />
        <link rel="preconnect" href="https://fonts.gstatic.com" crossOrigin="" />
        <link href="https://fonts.googleapis.com/css2?family=JetBrains+Mono:wght@400;500;600&display=swap" rel="stylesheet" />
      </head>
      <body className="font-mono bg-dark text-neutral-300 antialiased">
        {children}
      </body>
    </html>
  )
}