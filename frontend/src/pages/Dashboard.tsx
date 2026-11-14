
import { useCallback, useEffect, useState } from 'react'
import {
  Plus,
  Copy,
  ExternalLink,
  RefreshCw,
  BarChart3,
  ClipboardList,
  Users,
  CheckCircle2,
} from 'lucide-react'
import { Link } from 'react-router-dom'
import Navbar from '../components/Navbar'
import {
  api,
  getWebSocketUrl,
  type MyPollItem,
} from '../services/api'

interface VoteUpdateMessage {
  pollId: string
  optionId: string
  votes: number
}

export default function Dashboard() {
  const [polls, setPolls] = useState<MyPollItem[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [copiedId, setCopiedId] = useState('')

  const loadPolls = useCallback(async () => {
    setLoading(true)
    setError('')

    try {
      const response = await api.getMyPolls()
      setPolls(response.polls ?? [])
    } catch (err) {
      setError(
        err instanceof Error ? err.message : 'Unable to load polls'
      )
    } finally {
      setLoading(false)
    }
  }, [])

  // Load the creator's polls initially.
  useEffect(() => {
    void loadPolls()
  }, [loadPolls])

  // Keep the poll IDs stable so vote updates do not reconnect
  // every WebSocket whenever the poll counts change.
  const pollIds = polls
    .map((item) => item.poll.id)
    .join(',')

  // Subscribe to live vote updates for every poll owned by the creator.
  useEffect(() => {
    if (!pollIds) return

    const ids = pollIds.split(',')
    const sockets: WebSocket[] = []

    ids.forEach((pollId) => {
      const socket = new WebSocket(getWebSocketUrl(pollId))

      socket.onopen = () => {
        console.log(`Live updates connected for poll: ${pollId}`)
      }

      socket.onmessage = (event) => {
        try {
          const update = JSON.parse(
            event.data
          ) as VoteUpdateMessage

          if (
            update.pollId !== pollId ||
            typeof update.optionId !== 'string' ||
            typeof update.votes !== 'number'
          ) {
            return
          }

          setPolls((currentPolls) =>
            currentPolls.map((item) => {
              if (item.poll.id !== update.pollId) {
                return item
              }

              const updatedOptions = item.poll.options.map(
                (option) =>
                  option.id === update.optionId
                    ? { ...option, votes: update.votes }
                    : option
              )

              const updatedTotalVotes = updatedOptions.reduce(
                (sum, option) => sum + option.votes,
                0
              )

              return {
                ...item,
                poll: {
                  ...item.poll,
                  options: updatedOptions,
                },
                totalVotes: updatedTotalVotes,
              }
            })
          )
        } catch (err) {
          console.error('Unable to process live vote update:', err)
        }
      }

      socket.onerror = () => {
        console.error(`WebSocket error for poll: ${pollId}`)
      }

      socket.onclose = () => {
        console.log(`WebSocket disconnected for poll: ${pollId}`)
      }

      sockets.push(socket)
    })

    // Clean up connections when polls change or Dashboard unmounts.
    return () => {
      sockets.forEach((socket) => {
        socket.close()
      })
    }
  }, [pollIds])

  const totalVotes = polls.reduce(
    (sum, item) => sum + item.totalVotes,
    0
  )

  const activePolls = polls.filter(
    (item) => item.poll.isActive
  ).length

  async function copyPollLink(pollId: string) {
    const link = `${window.location.origin}/poll/${pollId}`

    try {
      await navigator.clipboard.writeText(link)
      setCopiedId(pollId)
      window.setTimeout(() => setCopiedId(''), 2000)
    } catch {
      setError(
        'Unable to copy link. Please copy the URL from your browser.'
      )
    }
  }

  return (
    <div className="app-layout">
      <Navbar />

      <main className="dashboard-container">
        <section className="dashboard-heading">
          <div>
            <span className="eyebrow">OVERVIEW</span>
            <h1>Welcome back!</h1>
            <p>
              Manage your polls and see what people are thinking.
            </p>
          </div>

          <Link
            to="/create-poll"
            className="primary-button create-link"
          >
            <Plus size={19} />
            Create New Poll
          </Link>
        </section>

        <section className="stats-grid">
          <div className="stat-card">
            <div className="stat-icon purple">
              <ClipboardList size={22} />
            </div>
            <div>
              <span>Total Polls</span>
              <strong>{loading ? '—' : polls.length}</strong>
            </div>
          </div>

          <div className="stat-card">
            <div className="stat-icon blue">
              <Users size={22} />
            </div>
            <div>
              <span>Total Votes</span>
              <strong>{loading ? '—' : totalVotes}</strong>
            </div>
          </div>

          <div className="stat-card">
            <div className="stat-icon green">
              <CheckCircle2 size={22} />
            </div>
            <div>
              <span>Active Polls</span>
              <strong>{loading ? '—' : activePolls}</strong>
            </div>
          </div>
        </section>

        <section className="poll-section">
          <div className="section-heading">
            <div>
              <h2>My Polls</h2>
              <p>Your polls and their latest voting activity.</p>
            </div>

            <button
              className="refresh-button"
              onClick={() => void loadPolls()}
              disabled={loading}
            >
              <RefreshCw
                size={16}
                className={loading ? 'spin' : ''}
              />
              Refresh
            </button>
          </div>

          {error && <div className="error-banner">{error}</div>}

          {loading ? (
            <div className="empty-state">
              <RefreshCw className="spin" size={25} />
              <p>Loading your polls...</p>
            </div>
          ) : polls.length === 0 ? (
            <div className="empty-state">
              <div className="empty-icon">
                <BarChart3 size={30} />
              </div>

              <h3>No polls yet</h3>
              <p>
                Create your first poll and start collecting opinions.
              </p>

              <Link
                to="/create-poll"
                className="primary-button create-link"
              >
                <Plus size={18} />
                Create Your First Poll
              </Link>
            </div>
          ) : (
            <div className="poll-list">
              {polls.map(({ poll, totalVotes }) => (
                <article className="poll-card" key={poll.id}>
                  <div className="poll-card-top">
                    <div>
                      <span
                        className={`status ${
                          poll.isActive ? 'active' : 'closed'
                        }`}
                      >
                        {poll.isActive ? 'Active' : 'Closed'}
                      </span>

                      <h3>{poll.question}</h3>

                      {poll.description && (
                        <p>{poll.description}</p>
                      )}
                    </div>

                    <div className="poll-vote-count">
                      <strong>{totalVotes}</strong>
                      <span>
                        {totalVotes === 1 ? 'Vote' : 'Votes'}
                      </span>
                    </div>
                  </div>

                  <div className="poll-options-summary">
                    {poll.options.map((option) => (
                      <div
                        className="option-summary"
                        key={option.id}
                      >
                        <span>{option.text}</span>
                        <strong>{option.votes}</strong>
                      </div>
                    ))}
                  </div>

                  <div className="poll-card-footer">
                    <span>
                      Created{' '}
                      {new Date(
                        poll.createdAt
                      ).toLocaleDateString()}
                    </span>

                    <div className="poll-actions">
                      <button
                        className="secondary-button"
                        onClick={() => void copyPollLink(poll.id)}
                      >
                        <Copy size={15} />
                        {copiedId === poll.id
                          ? 'Copied!'
                          : 'Copy link'}
                      </button>

                      <Link
                        className="secondary-button"
                        to={`/poll/${poll.id}`}
                      >
                        <ExternalLink size={15} />
                        View poll
                      </Link>
                    </div>
                  </div>
                </article>
              ))}
            </div>
          )}
        </section>
      </main>
    </div>
  )
}
