'use client'

import { useState, useEffect } from 'react'
import { Zap, GitBranch, Users, Bell, Target, Package, MonitorDown, Terminal, Wrench, ChevronDown, ExternalLink, Copy, Check, AlertCircle, ChevronRight, Menu, X } from 'lucide-react'

const sections = [
  { id: 'start', label: 'Quick Start' },
  { id: 'features', label: 'Features' },
  { id: 'installation', label: 'Installation' },
  { id: 'docs', label: 'Documentation' },
  { id: 'github', label: 'GitHub', external: 'https://github.com/parzi-val/axle-file-sync' },
]

// Network Topology Animation Component
const NetworkTopology = ({ activeSection }: { activeSection: string }) => {
  const [visibleNodes, setVisibleNodes] = useState(0)
  const [animationComplete, setAnimationComplete] = useState(false)
  const [packets, setPackets] = useState<Array<{ id: number; source: number; target: number; progress: number }>>([])

  // Determine leader node based on active section
  // start: 0 (12 o'clock), features: 2 (3 o'clock), installation: 4 (6 o'clock), docs: 6 (9 o'clock)
  const getLeaderNode = () => {
    switch(activeSection) {
      case 'start': return 0
      case 'features': return 2
      case 'installation': return 4
      case 'docs': return 6
      default: return 0
    }
  }

  const leaderNode = getLeaderNode()

  useEffect(() => {
    const timer = setInterval(() => {
      setVisibleNodes(prev => {
        if (prev < 8) {
          return prev + 1
        } else {
          clearInterval(timer)
          setAnimationComplete(true)
          return prev
        }
      })
    }, 300)

    return () => clearInterval(timer)
  }, [])

  // Generate random packets after animation completes
  useEffect(() => {
    if (!animationComplete) return

    const packetInterval = setInterval(() => {
      const sourceNode = Math.floor(Math.random() * 8)
      const newPackets: Array<{ id: number; source: number; target: number; progress: number }> = []

      // Send packets from source to all other nodes
      for (let i = 0; i < 8; i++) {
        if (i !== sourceNode) {
          newPackets.push({
            id: Date.now() + i,
            source: sourceNode,
            target: i,
            progress: 0
          })
        }
      }

      setPackets(prev => [...prev, ...newPackets])
    }, 5000)

    return () => clearInterval(packetInterval)
  }, [animationComplete])

  // Animate packets
  useEffect(() => {
    const animateInterval = setInterval(() => {
      setPackets(prev => {
        return prev
          .map(packet => ({
            ...packet,
            progress: packet.progress + 0.02
          }))
          .filter(packet => packet.progress <= 1)
      })
    }, 16)

    return () => clearInterval(animateInterval)
  }, [])

  const nodes = Array.from({ length: 8 }, (_, i) => {
    const angle = (i * Math.PI * 2) / 8 - Math.PI / 2
    const radius = 96
    const x = 180 + Math.cos(angle) * radius
    const y = 180 + Math.sin(angle) * radius
    return { x, y, id: i }
  })

  return (
    <div className="fixed right-16 top-1/2 -translate-y-1/2 opacity-20 hover:opacity-30 transition-opacity duration-500 hidden xl:block">
      <svg width="360" height="360" viewBox="0 0 360 360">
        {/* Draw full mesh connections */}
        {nodes.map((node, i) => (
          nodes.slice(i + 1).map((targetNode, j) => (
            <line
              key={`line-${i}-${i + j + 1}`}
              x1={node.x}
              y1={node.y}
              x2={targetNode.x}
              y2={targetNode.y}
              stroke="currentColor"
              strokeWidth="0.5"
              strokeOpacity={i < visibleNodes && (i + j + 1) < visibleNodes ? 0.8 : 0}
              className="transition-all duration-500 text-gray-400"
            />
          ))
        ))}

        {/* Draw animated packets */}
        {packets.map(packet => {
          const source = nodes[packet.source]
          const target = nodes[packet.target]
          const x = source.x + (target.x - source.x) * packet.progress
          const y = source.y + (target.y - source.y) * packet.progress

          return (
            <circle
              key={packet.id}
              cx={x}
              cy={y}
              r="3"
              fill="#10b981"
              opacity={0.8 * (1 - packet.progress * 0.5)}
            />
          )
        })}

        {/* Draw nodes */}
        {nodes.map((node, i) => (
          <g key={`node-${i}`}>
            <circle
              cx={node.x}
              cy={node.y}
              r={i < visibleNodes ? 10 : 0}
              fill={i === leaderNode ? '#10b981' : '#3b82f6'}
              className={`transition-all duration-500 ${
                animationComplete && i === leaderNode ? 'animate-pulse' : ''
              }`}
              opacity="0.8"
            />
            {i < visibleNodes && animationComplete && (
              <circle
                cx={node.x}
                cy={node.y}
                r="15"
                fill="none"
                stroke={i === leaderNode ? '#10b981' : '#3b82f6'}
                strokeWidth="0.5"
                strokeOpacity="0.2"
                className="animate-ping"
              />
            )}
          </g>
        ))}
      </svg>
    </div>
  )
}

export default function Home() {
  const [activeSection, setActiveSection] = useState('start')
  const [sidebarOpen, setSidebarOpen] = useState(false)

  const handleSectionClick = (sectionId: string) => {
    setActiveSection(sectionId)
    setSidebarOpen(false)
  }

  return (
    <div className="flex min-h-screen">
      <NetworkTopology activeSection={activeSection} />

      {/* Mobile Menu Button */}
      <button
        onClick={() => setSidebarOpen(!sidebarOpen)}
        className="lg:hidden fixed bottom-4 left-4 z-50 p-3 bg-black/80 border border-white/20 rounded-full backdrop-blur-sm shadow-lg hover:bg-black/90 transition-all duration-200"
      >
        {sidebarOpen ? <X className="w-5 h-5 text-white" /> : <Menu className="w-5 h-5 text-white" />}
      </button>

      {/* Sidebar */}
      <aside className={`fixed left-0 top-0 h-full w-64 border-r border-white/5 bg-black/30 backdrop-blur-xl p-8 z-40 transition-transform duration-300 lg:translate-x-0 ${
        sidebarOpen ? 'translate-x-0' : '-translate-x-full'
      }`}>
        <div className="mb-8">
          <h1 className="text-2xl font-bold text-white tracking-tight">axle</h1>
          <p className="text-sm text-neutral-400 mt-1 font-medium">v0.1.1</p>
        </div>

        <div className="border-t border-white/5 mb-6"></div>

        <nav className="space-y-1">
          {sections.map((section) => (
            section.external ? (
              <a
                key={section.id}
                href={section.external}
                target="_blank"
                rel="noopener noreferrer"
                className="flex items-center gap-2 px-3 py-2 text-sm text-neutral-400 hover:text-white transition-all duration-200"
              >
                {section.label}
                <ExternalLink className="w-3 h-3" />
              </a>
            ) : (
              <button
                key={section.id}
                onClick={() => handleSectionClick(section.id)}
                className={`block w-full text-left px-3 py-2 text-sm rounded-md transition-all duration-200 ${
                  activeSection === section.id
                    ? 'text-white bg-white/10'
                    : 'text-neutral-400 hover:text-white hover:bg-white/5'
                }`}
              >
                {section.label}
              </button>
            )
          ))}
        </nav>

        <div className="border-t border-white/5 mt-8 pt-8">
          <p className="text-xs text-neutral-600 leading-relaxed">
            Real-time file sync<br />
            for hackathon teams.
          </p>
        </div>
      </aside>

      {/* Mobile Overlay */}
      {sidebarOpen && (
        <div
          className="lg:hidden fixed inset-0 bg-black/50 z-30"
          onClick={() => setSidebarOpen(false)}
        />
      )}

      {/* Main Content */}
      <main className="flex-1 lg:ml-64 p-6 sm:p-8 md:p-12 lg:p-16 max-w-4xl w-full">
        {activeSection === 'start' && <QuickStart setActiveSection={setActiveSection} />}
        {activeSection === 'features' && <Features />}
        {activeSection === 'installation' && <Installation />}
        {activeSection === 'docs' && <Documentation />}
      </main>

      {/* Network Topology Animation */}
      <NetworkTopology activeSection={activeSection} />
    </div>
  )
}

function QuickStart({ setActiveSection }: { setActiveSection: (section: string) => void }) {
  return (
    <section className="animate-fadeIn">
      <div className="mb-8 md:mb-12 lg:mb-16 mt-8 md:mt-12 lg:mt-0">
        <h2 className="text-3xl sm:text-4xl md:text-5xl lg:text-6xl font-bold text-white mb-3 md:mb-6 tracking-tight">axle.</h2>
        <p className="text-base sm:text-lg md:text-xl text-neutral-400 leading-relaxed mb-3 md:mb-4">
          Real-time file synchronization.<br />
          No accounts. No cloud. Just sync.
        </p>
        <p className="text-xs sm:text-sm text-neutral-500 leading-relaxed max-w-2xl">
          Built for hackathon teams and rapid prototyping groups who need instant code sharing without the complexity.
          Perfect for 24-48 hour sprints where every second counts. Sync your entire codebase across team members
          with Git-powered patches over Redis pub/sub.
        </p>
      </div>

      <div className="bg-black/50 backdrop-blur-sm border border-white/10 rounded-lg md:rounded-xl p-4 md:p-6 mb-8 md:mb-12 overflow-x-auto">
        <div className="flex items-center gap-2 mb-3 md:mb-4">
          <div className="w-2 h-2 md:w-3 md:h-3 rounded-full bg-red-500/80"></div>
          <div className="w-2 h-2 md:w-3 md:h-3 rounded-full bg-yellow-500/80"></div>
          <div className="w-2 h-2 md:w-3 md:h-3 rounded-full bg-green-500/80"></div>
          <span className="ml-3 md:ml-4 text-xs text-neutral-500">terminal</span>
        </div>
        <pre className="text-xs sm:text-sm leading-relaxed whitespace-pre overflow-x-auto">
          <span className="text-neutral-500">$</span> <span className="text-green-400">axle init</span> <span className="text-blue-400">--team</span> hackathon-2024 <span className="text-blue-400">--username</span> alice{'\n'}
          <span className="text-neutral-500">$</span> <span className="text-green-400">axle start</span>
        </pre>
      </div>

      <div className="flex flex-col sm:flex-row gap-3 sm:gap-4">
        <button
          onClick={() => setActiveSection('installation')}
          className="px-4 sm:px-6 py-2.5 sm:py-3 bg-white text-black font-medium rounded text-sm sm:text-base hover:bg-neutral-200 transition-all duration-200"
        >
          Get Started
        </button>
        <a
          href="https://github.com/parzi-val/axle-file-sync"
          target="_blank"
          rel="noopener noreferrer"
          className="px-4 sm:px-6 py-2.5 sm:py-3 border border-white/20 text-white rounded text-sm sm:text-base hover:bg-white/5 transition-all duration-200 text-center"
        >
          View Source
        </a>
      </div>
    </section>
  )
}

function Features() {
  const features = [
    { title: 'Git-based sync', desc: 'Redis pub/sub coordination', Icon: Zap },
    { title: '5 conflict modes', desc: 'theirs, mine, merge, backup, interactive', Icon: GitBranch },
    { title: 'Team presence', desc: 'See who\'s coding in real-time', Icon: Users },
    { title: 'Priority chat', desc: 'Desktop notifications for urgent messages', Icon: Bell },
    { title: 'Smart detection', desc: 'Auto-configures .gitignore by stack', Icon: Target },
    { title: '10MB limit', desc: 'Optimized for code, not assets', Icon: Package },
  ]

  return (
    <section className="animate-fadeIn">
      <h2 className="text-2xl sm:text-3xl font-bold text-white mb-6 md:mb-8">Features</h2>
      <div className="border-t border-white/10 mb-8 md:mb-12"></div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-4 md:gap-6">
        {features.map((feature, i) => (
          <div key={i} className="group p-6 bg-black/30 border border-white/10 rounded-lg hover:border-white/20 transition-all duration-200">
            <div className="flex items-start gap-4">
              <feature.Icon className="w-5 h-5 text-neutral-400 group-hover:text-white transition-colors mt-0.5" />
              <div>
                <h3 className="text-white font-semibold mb-1">{feature.title}</h3>
                <p className="text-sm text-neutral-400">{feature.desc}</p>
              </div>
            </div>
          </div>
        ))}
      </div>

      <div className="mt-12 p-4 md:p-8 bg-black/30 border border-emerald-500/20 rounded-lg">
        <h3 className="text-base md:text-lg font-semibold text-emerald-300 mb-4 md:mb-6">Conflict Resolution Strategies</h3>
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <tbody className="divide-y divide-white/5">
              <tr>
                <td className="py-2 pr-4">
                  <code className="text-emerald-400 bg-emerald-400/10 px-2 py-0.5 rounded whitespace-nowrap">theirs</code>
                </td>
                <td className="py-2 text-neutral-400">Accept remote changes</td>
              </tr>
              <tr>
                <td className="py-2 pr-4">
                  <code className="text-emerald-400 bg-emerald-400/10 px-2 py-0.5 rounded whitespace-nowrap">mine</code>
                </td>
                <td className="py-2 text-neutral-400">Keep local changes</td>
              </tr>
              <tr>
                <td className="py-2 pr-4">
                  <code className="text-emerald-400 bg-emerald-400/10 px-2 py-0.5 rounded whitespace-nowrap">merge</code>
                </td>
                <td className="py-2 text-neutral-400">Create conflict markers</td>
              </tr>
              <tr>
                <td className="py-2 pr-4">
                  <code className="text-emerald-400 bg-emerald-400/10 px-2 py-0.5 rounded whitespace-nowrap">backup</code>
                </td>
                <td className="py-2 text-neutral-400">Create .backup files</td>
              </tr>
              <tr>
                <td className="py-2 pr-4">
                  <code className="text-emerald-400 bg-emerald-400/10 px-2 py-0.5 rounded whitespace-nowrap">interactive</code>
                </td>
                <td className="py-2 text-neutral-400">Open conflicts in VS Code</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </section>
  )
}

function Installation() {
  const [copiedIndex, setCopiedIndex] = useState<number | null>(null)

  const installations = [
    {
      label: 'Windows',
      command: 'irm https://raw.githubusercontent.com/parzi-val/axle-file-sync/main/scripts/install.ps1 | iex',
      Icon: MonitorDown,
    },
    {
      label: 'macOS / Linux',
      command: 'curl -sSL https://raw.githubusercontent.com/parzi-val/axle-file-sync/main/scripts/install.sh | bash',
      Icon: Terminal,
    },
    {
      label: 'From Source',
      command: 'go install github.com/parzi-val/axle-file-sync@latest',
      Icon: Wrench,
    },
  ]

  const copyToClipboard = (text: string, index: number) => {
    navigator.clipboard.writeText(text)
    setCopiedIndex(index)
    setTimeout(() => setCopiedIndex(null), 2000)
  }

  return (
    <section className="animate-fadeIn" data-section="installation">
      <h2 className="text-2xl sm:text-3xl font-bold text-white mb-6 md:mb-8">Installation</h2>
      <div className="border-t border-white/10 mb-8 md:mb-12"></div>

      <div className="space-y-6">
        {installations.map((item, i) => (
          <div key={i} className="group">
            <div className="flex items-center gap-3 mb-3">
              <item.Icon className="w-4 h-4 text-neutral-500" />
              <h3 className="text-sm font-medium text-neutral-500 uppercase tracking-wider">{item.label}</h3>
            </div>
            <div className="relative bg-black/40 border border-white/10 rounded-lg p-3 md:p-5 hover:border-white/20 transition-all duration-200">
              <pre className="text-xs sm:text-sm overflow-x-auto scrollbar-hide">
                <code className="block sm:inline-block whitespace-pre-wrap sm:whitespace-nowrap break-all sm:break-normal">
                  {item.label === 'Windows' && (
                    <>
                      <span className="text-green-400">irm</span>{' '}
                      <span className="text-neutral-400">https://raw.githubusercontent.com/.../install.ps1</span>
                      <span className="hidden sm:inline">{' '}</span>
                      <span className="block sm:inline mt-1 sm:mt-0">
                        <span className="text-neutral-500">|</span>{' '}
                        <span className="text-green-400">iex</span>
                      </span>
                    </>
                  )}
                  {item.label === 'macOS / Linux' && (
                    <>
                      <span className="text-green-400">curl</span>{' '}
                      <span className="text-blue-400">-sSL</span>{' '}
                      <span className="text-neutral-400">https://raw.githubusercontent.com/.../install.sh</span>
                      <span className="hidden sm:inline">{' '}</span>
                      <span className="block sm:inline mt-1 sm:mt-0">
                        <span className="text-neutral-500">|</span>{' '}
                        <span className="text-green-400">bash</span>
                      </span>
                    </>
                  )}
                  {item.label === 'From Source' && (
                    <>
                      <span className="text-green-400">go install</span>{' '}
                      <span className="text-neutral-400">github.com/parzi-val/axle-file-sync@latest</span>
                    </>
                  )}
                </code>
              </pre>
              <button
                onClick={() => copyToClipboard(item.command, i)}
                className="absolute right-2 sm:right-4 top-2 sm:top-1/2 sm:-translate-y-1/2 p-1.5 sm:p-2 bg-white/5 border border-white/10 rounded hover:bg-white/10 transition-all duration-200"
              >
                {copiedIndex === i ? <Check className="w-3 h-3" /> : <Copy className="w-3 h-3" />}
              </button>
            </div>
          </div>
        ))}
      </div>

      <div className="mt-8 md:mt-12 p-3 md:p-4 bg-black/30 border border-amber-500/20 rounded-lg">
        <div className="flex items-start gap-2 md:gap-3">
          <AlertCircle className="w-4 h-4 text-amber-500 mt-0.5 flex-shrink-0" />
          <div className="flex-1">
            <p className="text-xs text-amber-200/70 leading-relaxed">
              <strong>Requirements:</strong> Git and Redis must be installed. Redis should be running on <code className="bg-black/30 px-1 py-0.5 rounded text-amber-300 text-xs">localhost:6379</code> or specify custom host with <code className="bg-black/30 px-1 py-0.5 rounded text-amber-300 text-xs">--host</code> flag
            </p>
          </div>
        </div>
      </div>
    </section>
  )
}

function Documentation() {
  const [expandedSection, setExpandedSection] = useState<string | null>('getting-started')

  const docSections = [
    {
      id: 'getting-started',
      title: 'Getting Started',
      content: (
        <div className="space-y-6">
          <div className="bg-red-500/10 border border-red-500/20 rounded-lg p-4 mb-6">
            <div className="flex items-start gap-3">
              <AlertCircle className="w-5 h-5 text-red-400 mt-0.5 flex-shrink-0" />
              <div className="text-sm text-red-300">
                <strong>Prerequisites:</strong> Redis must be running before using axle!
                <div className="mt-2 text-xs text-red-300/80">
                  Start Redis with: <code className="bg-black/30 px-1 py-0.5 rounded">redis-server</code> or <code className="bg-black/30 px-1 py-0.5 rounded">docker run -p 6379:6379 redis</code>
                </div>
              </div>
            </div>
          </div>
          <div className="bg-black/50 border border-white/10 rounded-lg p-5">
            <h4 className="text-base font-semibold text-white mb-3">1. Team Lead Setup</h4>
            <p className="text-sm text-neutral-300 mb-4">The team lead creates a new sync session with a unique team name and password.</p>
            <div className="bg-black/50 border border-white/10 rounded p-4 overflow-x-auto">
              <code className="text-sm whitespace-nowrap">
                <span className="text-green-400">axle init</span> <span className="text-blue-400">--team</span> hackathon-2024 <span className="text-blue-400">--username</span> alice <span className="text-blue-400">--password</span> secret123
              </code>
            </div>
          </div>
          <div className="bg-black/50 border border-white/10 rounded-lg p-5">
            <h4 className="text-base font-semibold text-white mb-3">2. Start Syncing</h4>
            <p className="text-sm text-neutral-300 mb-4">Begin file synchronization with your preferred conflict resolution strategy.</p>
            <div className="bg-black/50 border border-white/10 rounded p-4 overflow-x-auto">
              <code className="text-sm whitespace-nowrap">
                <span className="text-green-400">axle start</span> <span className="text-blue-400">--conflict</span> merge
              </code>
            </div>
          </div>
          <div className="bg-black/50 border border-white/10 rounded-lg p-5">
            <h4 className="text-base font-semibold text-neutral-300 mb-3">3. Team Members Join</h4>
            <p className="text-sm text-neutral-300 mb-4">Other team members join using the same team name and password.</p>
            <div className="bg-black/50 border border-white/10 rounded p-4 overflow-x-auto">
              <code className="text-sm whitespace-nowrap">
                <span className="text-green-400">axle join</span> <span className="text-blue-400">--team</span> hackathon-2024 <span className="text-blue-400">--username</span> bob <span className="text-blue-400">--password</span> secret123
              </code>
            </div>
          </div>
        </div>
      )
    },
    {
      id: 'commands',
      title: 'Command Reference',
      content: (
        <div className="space-y-6">
          <div className="bg-black/50 border border-white/10 rounded-lg p-4 md:p-5">
            <div className="flex flex-col sm:flex-row sm:items-center gap-2 sm:gap-3 mb-4">
              <code className="text-green-400 font-semibold text-base">init</code>
              <span className="text-sm text-neutral-400">Creates new team</span>
            </div>
            <div className="space-y-3 sm:pl-2">
              <div className="flex flex-col sm:grid sm:grid-cols-[120px_1fr] md:grid-cols-[140px_1fr] gap-1 sm:gap-4 items-start">
                <code className="text-sm text-blue-400">--team</code>
                <span className="text-sm text-neutral-300 break-words">Team name (required)</span>
              </div>
              <div className="flex flex-col sm:grid sm:grid-cols-[120px_1fr] md:grid-cols-[140px_1fr] gap-1 sm:gap-4 items-start">
                <code className="text-xs sm:text-sm text-blue-400">--username</code>
                <span className="text-xs sm:text-sm text-neutral-300 break-words">Your username (required)</span>
              </div>
              <div className="flex flex-col sm:grid sm:grid-cols-[120px_1fr] md:grid-cols-[140px_1fr] gap-1 sm:gap-4 items-start">
                <code className="text-xs sm:text-sm text-blue-400">--password</code>
                <span className="text-xs sm:text-sm text-neutral-300 break-words">Team password (optional, prompts if not provided)</span>
              </div>
              <div className="flex flex-col sm:grid sm:grid-cols-[120px_1fr] md:grid-cols-[140px_1fr] gap-1 sm:gap-4 items-start">
                <code className="text-xs sm:text-sm text-blue-400">--host</code>
                <span className="text-xs sm:text-sm text-neutral-300 break-words">Redis host (default: localhost:6379)</span>
              </div>
            </div>
          </div>

          <div className="bg-black/50 border border-white/10 rounded-lg p-4 md:p-5">
            <div className="flex flex-col sm:flex-row sm:items-center gap-2 sm:gap-3 mb-4">
              <code className="text-green-400 font-semibold text-base">join</code>
              <span className="text-sm text-neutral-400">Join existing team</span>
            </div>
            <div className="space-y-3 sm:pl-2">
              <div className="flex flex-col sm:grid sm:grid-cols-[120px_1fr] md:grid-cols-[140px_1fr] gap-1 sm:gap-4 items-start">
                <code className="text-xs sm:text-sm text-blue-400">--team</code>
                <span className="text-xs sm:text-sm text-neutral-300 break-words">Team to join (required)</span>
              </div>
              <div className="flex flex-col sm:grid sm:grid-cols-[120px_1fr] md:grid-cols-[140px_1fr] gap-1 sm:gap-4 items-start">
                <code className="text-xs sm:text-sm text-blue-400">--username</code>
                <span className="text-xs sm:text-sm text-neutral-300 break-words">Your username (required)</span>
              </div>
              <div className="flex flex-col sm:grid sm:grid-cols-[120px_1fr] md:grid-cols-[140px_1fr] gap-1 sm:gap-4 items-start">
                <code className="text-xs sm:text-sm text-blue-400">--password</code>
                <span className="text-xs sm:text-sm text-neutral-300 break-words">Team password (optional, prompts if not provided)</span>
              </div>
              <div className="flex flex-col sm:grid sm:grid-cols-[120px_1fr] md:grid-cols-[140px_1fr] gap-1 sm:gap-4 items-start">
                <code className="text-xs sm:text-sm text-blue-400">--host</code>
                <span className="text-xs sm:text-sm text-neutral-300 break-words">Redis host (default: localhost:6379)</span>
              </div>
            </div>
          </div>

          <div className="bg-black/50 border border-white/10 rounded-lg p-4 md:p-5">
            <div className="flex flex-col sm:flex-row sm:items-center gap-2 sm:gap-3 mb-4">
              <code className="text-green-400 font-semibold text-base">start</code>
              <span className="text-sm text-neutral-400">Begin synchronization</span>
            </div>
            <div className="space-y-3 sm:pl-2">
              <div className="flex flex-col sm:grid sm:grid-cols-[120px_1fr] md:grid-cols-[140px_1fr] gap-1 sm:gap-4 items-start">
                <code className="text-xs sm:text-sm text-blue-400">--conflict</code>
                <span className="text-xs sm:text-sm text-neutral-300 break-words">Resolution strategy: theirs, mine, merge, backup, interactive</span>
              </div>
              <div className="flex flex-col sm:grid sm:grid-cols-[120px_1fr] md:grid-cols-[140px_1fr] gap-1 sm:gap-4 items-start">
                <code className="text-sm text-blue-400">--verbose</code>
                <span className="text-sm text-neutral-300 break-words">Enable debug logging</span>
              </div>
            </div>
          </div>

          <div className="bg-black/50 border border-white/10 rounded-lg p-4 md:p-5">
            <div className="flex flex-col sm:flex-row sm:items-center gap-2 sm:gap-3 mb-4">
              <code className="text-green-400 font-semibold text-base">chat</code>
              <span className="text-sm text-neutral-400">Send team message</span>
            </div>
            <div className="space-y-3 sm:pl-2">
              <div className="flex flex-col sm:grid sm:grid-cols-[120px_1fr] md:grid-cols-[140px_1fr] gap-1 sm:gap-4 items-start">
                <code className="text-sm text-blue-400">-p</code>
                <span className="text-sm text-neutral-300 break-words">Priority flag - triggers desktop notification</span>
              </div>
              <div className="flex flex-col sm:grid sm:grid-cols-[120px_1fr] md:grid-cols-[140px_1fr] gap-1 sm:gap-4 items-start">
                <code className="text-sm text-blue-400">"message"</code>
                <span className="text-sm text-neutral-300 break-words">Your message text (quote if spaces)</span>
              </div>
            </div>
          </div>

          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <div className="bg-black/50 border border-white/10 rounded-lg p-4 md:p-5">
              <div className="flex flex-col sm:flex-row sm:items-center gap-2 sm:gap-3">
                <code className="text-green-400 font-semibold text-base">team</code>
                <span className="text-sm text-neutral-400 break-words">Show online members</span>
              </div>
            </div>
            <div className="bg-black/50 border border-white/10 rounded-lg p-4 md:p-5">
              <div className="flex flex-col sm:flex-row sm:items-center gap-2 sm:gap-3">
                <code className="text-green-400 font-semibold text-base">stats</code>
                <span className="text-sm text-neutral-400 break-words">Display sync statistics</span>
              </div>
            </div>
          </div>
        </div>
      )
    },
    {
      id: 'conflict-resolution',
      title: 'Conflict Resolution',
      content: (
        <div className="space-y-6">
          <p className="text-sm text-neutral-300 mb-6">When file conflicts occur, axle provides 5 strategies to handle them:</p>

          <div className="space-y-4">
            <div className="bg-black/50 border border-white/10 rounded-lg p-5">
              <div className="flex items-baseline gap-3 mb-2">
                <code className="text-green-400 font-semibold text-base">theirs</code>
                <span className="text-sm text-neutral-400">Accept remote</span>
              </div>
              <p className="text-sm text-neutral-300">Always accept changes from other team members. Good for read-only files.</p>
            </div>

            <div className="bg-black/50 border border-white/10 rounded-lg p-5">
              <div className="flex items-baseline gap-3 mb-2">
                <code className="text-green-400 font-semibold text-base">mine</code>
                <span className="text-sm text-neutral-400">Keep local</span>
              </div>
              <p className="text-sm text-neutral-300">Always keep your local changes. Useful when you're the primary editor.</p>
            </div>

            <div className="bg-black/50 border border-white/10 rounded-lg p-5">
              <div className="flex items-baseline gap-3 mb-2">
                <code className="text-green-400 font-semibold text-base">merge</code>
                <span className="text-sm text-neutral-400">Git-style markers</span>
              </div>
              <p className="text-sm text-neutral-300">Creates conflict markers ({'<<<<<<<'} {'======='} {'>>>>>>>'}) for manual resolution.</p>
            </div>

            <div className="bg-black/50 border border-white/10 rounded-lg p-5">
              <div className="flex items-baseline gap-3 mb-2">
                <code className="text-green-400 font-semibold text-base">backup</code>
                <span className="text-sm text-neutral-400">Create .backup</span>
              </div>
              <p className="text-sm text-neutral-300">Accept remote but save local version as filename.backup for comparison.</p>
            </div>

            <div className="bg-black/50 border border-white/10 rounded-lg p-5">
              <div className="flex items-baseline gap-3 mb-2">
                <code className="text-green-400 font-semibold text-base">interactive</code>
                <span className="text-sm text-neutral-400">VS Code diff</span>
              </div>
              <p className="text-sm text-neutral-300">Opens conflicts in VS Code for visual resolution (requires code in PATH).</p>
            </div>
          </div>

          <div className="mt-6 p-4 bg-amber-500/10 border border-amber-500/20 rounded-lg">
            <p className="text-sm text-amber-300">
              <strong>Tip:</strong> Start with 'merge' for most projects, use 'interactive' for critical files.
            </p>
          </div>
        </div>
      )
    },
    {
      id: 'architecture',
      title: 'How It Works',
      content: (
        <div className="space-y-6">
          <div className="bg-black/50 border border-white/10 rounded-lg p-5">
            <h4 className="text-base font-semibold text-white mb-3">Synchronization Flow</h4>
            <ol className="text-sm text-neutral-300 space-y-2">
              <li><span className="text-neutral-500">1.</span> File watcher detects local changes</li>
              <li><span className="text-neutral-500">2.</span> Git creates a patch of the changes</li>
              <li><span className="text-neutral-500">3.</span> Patch is published to Redis pub/sub channel</li>
              <li><span className="text-neutral-500">4.</span> All team members receive the patch</li>
              <li><span className="text-neutral-500">5.</span> Patch is applied using chosen conflict strategy</li>
            </ol>
          </div>

          <div className="bg-black/50 border border-white/10 rounded-lg p-5">
            <h4 className="text-base font-semibold text-white mb-3">Smart Filtering</h4>
            <ul className="text-sm text-neutral-300 space-y-2">
              <li>• Respects .gitignore patterns</li>
              <li>• Auto-detects framework patterns (node_modules, venv, etc.)</li>
              <li>• Skips files over 10MB</li>
              <li>• Ignores binary and build artifacts</li>
            </ul>
          </div>

          <div className="bg-black/50 border border-white/10 rounded-lg p-5">
            <h4 className="text-base font-semibold text-white mb-3">Team Presence</h4>
            <ul className="text-sm text-neutral-300 space-y-2">
              <li>• Heartbeat every 30 seconds</li>
              <li>• Shows who's actively coding</li>
              <li>• Auto-cleanup on disconnect</li>
              <li>• Team chat with priority notifications</li>
            </ul>
          </div>
        </div>
      )
    },
    {
      id: 'troubleshooting',
      title: 'Troubleshooting',
      content: (
        <div className="space-y-6">
          <div className="bg-red-500/10 border border-red-500/20 rounded-lg p-5">
            <h4 className="text-base font-semibold text-red-300 mb-3">⚠️ Redis Connection Failed (Most Common Issue)</h4>
            <div className="text-sm text-neutral-300 space-y-3">
              <p className="text-red-300/90 font-medium">Redis is REQUIRED for axle to work. Ensure Redis is running:</p>
              <div className="bg-black/50 border border-white/10 rounded p-3">
                <code className="text-sm text-green-400">redis-cli ping</code>
                <span className="text-xs text-neutral-400 ml-2"># Should return PONG</span>
              </div>
              <p>Start Redis:</p>
              <div className="bg-black/50 border border-white/10 rounded p-3 space-y-1">
                <div><code className="text-sm text-green-400">redis-server</code> <span className="text-xs text-neutral-400"># Local install</span></div>
                <div><code className="text-sm text-green-400">brew services start redis</code> <span className="text-xs text-neutral-400"># macOS</span></div>
                <div><code className="text-sm text-green-400">docker run -p 6379:6379 redis</code> <span className="text-xs text-neutral-400"># Docker</span></div>
              </div>
              <p>Use custom host if needed:</p>
              <div className="bg-black/50 border border-white/10 rounded p-3">
                <code className="text-sm">
                  <span className="text-green-400">axle init</span> <span className="text-blue-400">--host</span> redis.example.com:6379
                </code>
              </div>
            </div>
          </div>

          <div className="bg-black/50 border border-white/10 rounded-lg p-5">
            <h4 className="text-base font-semibold text-white mb-3">Sync Not Working</h4>
            <ul className="text-sm text-neutral-300 space-y-2">
              <li>• Check if Git is initialized: <code className="text-green-400">git status</code></li>
              <li>• Verify team password matches</li>
              <li>• Run with verbose mode: <code className="text-green-400">axle start</code> <code className="text-blue-400">--verbose</code></li>
              <li>• Check firewall/network settings</li>
            </ul>
          </div>

          <div className="bg-black/50 border border-white/10 rounded-lg p-5">
            <h4 className="text-base font-semibold text-white mb-3">File Too Large</h4>
            <p className="text-sm text-neutral-300 mb-3">Files over 10MB are automatically skipped. Add large files to .gitignore:</p>
            <div className="bg-black/50 border border-white/10 rounded p-3">
              <code className="text-sm text-neutral-400">
                # .gitignore{'\n'}
                *.mp4{'\n'}
                *.zip{'\n'}
                node_modules/
              </code>
            </div>
          </div>
        </div>
      )
    }
  ]

  return (
    <section className="animate-fadeIn">
      <h2 className="text-2xl sm:text-3xl font-bold text-white mb-6 md:mb-8">Documentation</h2>
      <div className="border-t border-white/10 mb-8 md:mb-12"></div>

      <div className="space-y-4">
        {docSections.map((section) => (
          <div key={section.id} className="border border-white/10 rounded-lg overflow-hidden">
            <button
              onClick={() => setExpandedSection(expandedSection === section.id ? null : section.id)}
              className="w-full px-4 md:px-6 py-3 md:py-4 bg-black/30 hover:bg-black/40 transition-colors duration-200 flex items-center justify-between"
            >
              <h3 className="text-base md:text-lg font-semibold text-white">{section.title}</h3>
              <ChevronDown className={`w-4 md:w-5 h-4 md:h-5 text-neutral-400 transition-transform duration-200 ${
                expandedSection === section.id ? 'rotate-180' : ''
              }`} />
            </button>
            {expandedSection === section.id && (
              <div className="p-4 md:p-6 border-t border-white/10 bg-black/20">
                {section.content}
              </div>
            )}
          </div>
        ))}
      </div>
    </section>
  )
}