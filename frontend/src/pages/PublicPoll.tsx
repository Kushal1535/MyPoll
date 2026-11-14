
import { useCallback, useEffect, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import './PublicPoll.css'
import {
  Activity,
  CheckCircle2,
  LoaderCircle,
  Vote,
  Users,
  Radio,
} from 'lucide-react'
import {
  api,
  getWebSocketUrl,
  type Poll,
} from '../services/api'

// Keep a stable voter ID for each poll in this browser.
function getVoterId(pollId: string): string {
  const storageKey = `mypoll_voter_id_${pollId}`

  let voterId = localStorage.getItem(storageKey)

  if (!voterId) {
    voterId = crypto.randomUUID()
    localStorage.setItem(storageKey, voterId)
  }

  return voterId
}

export default function PublicPoll() {
  const { id } = useParams<{ id: string }>()

  const [poll, setPoll] = useState<Poll | null>(null)
  const [totalVotes, setTotalVotes] = useState(0)
  const [selectedOption, setSelectedOption] = useState('')
  const [loading, setLoading] = useState(true)
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState('')
  const [success, setSuccess] = useState(false)
  const [liveConnected, setLiveConnected] = useState(false)

  const loadPoll = useCallback(async () => {
    if (!id) return

    try {
      const response = await api.getPoll(id)

      setPoll(response.poll)
      setTotalVotes(response.totalVotes)
      setError('')
    } catch (err) {
      setError(
        err instanceof Error ? err.message : 'Unable to load this poll.',
      )
    } finally {
      setLoading(false)
    }
  }, [id])

  useEffect(() => {
    setLoading(true)
    void loadPoll()
  }, [loadPoll])

  // Receive live vote updates through Redis Pub/Sub and WebSocket.
  useEffect(() => {
    if (!id) return

    const socket = new WebSocket(getWebSocketUrl(id))

    socket.onopen = () => setLiveConnected(true)

    socket.onmessage = () => {
      void loadPoll()
    }

    socket.onclose = () => setLiveConnected(false)
    socket.onerror = () => setLiveConnected(false)

    return () => {
      socket.close()
    }
  }, [id, loadPoll])

  async function handleVote() {
    if (!id || !selectedOption || !poll || submitting) return

    if (!poll.isActive) {
      setError('This poll is currently closed.')
      return
    }

    setSubmitting(true)
    setError('')

    try {
      const voterId = getVoterId(id)

      await api.vote(id, selectedOption, voterId)

      setSuccess(true)
      await loadPoll()
    } catch (err) {
      setError(
        err instanceof Error ? err.message : 'Unable to submit your vote.',
      )
    } finally {
      setSubmitting(false)
    }
  }

  if (loading) {
    return (
      <main className="public-poll-page">
        <div className="public-loading">
          <LoaderCircle className="spin" size={30} />
          <p>Loading poll...</p>
        </div>
      </main>
    )
  }

  if (!poll) {
    return (
      <main className="public-poll-page">
        <div className="public-poll-card public-error-card">
          <Activity size={38} />
          <h2>Poll unavailable</h2>
          <p>{error || 'This poll could not be found.'}</p>
          <Link to="/login" className="public-home-link">
            Go to MyPoll
          </Link>
        </div>
      </main>
    )
  }

  return (
    <main className="public-poll-page">
      <div className="public-poll-shell">
        <header className="public-header">
          <Link to="/" className="public-brand">
            <span className="public-brand-icon">
              <Activity size={23} />
            </span>
            <span>MyPoll</span>
          </Link>

          <div className="public-live-status">
            <span className={liveConnected ? 'live-dot connected' : 'live-dot'} />
            {liveConnected ? 'Live updates on' : 'Connecting...'}
          </div>
        </header>

        <section className="public-poll-card">
          <div className="public-poll-topline">
            <span className="public-eyebrow">
              <Vote size={14} />
              COMMUNITY POLL
            </span>

            <span className={`public-status ${poll.isActive ? 'active' : 'closed'}`}>
              {poll.isActive ? 'Active' : 'Closed'}
            </span>
          </div>

          <h1 className="public-poll-title">{poll.question}</h1>

          {poll.description && (
            <p className="public-poll-description">{poll.description}</p>
          )}

          <div className="public-vote-summary">
            <div className="public-vote-summary-icon">
              <Users size={18} />
            </div>
            <div>
              <strong>{totalVotes}</strong>
              <span>{totalVotes === 1 ? 'vote submitted' : 'votes submitted'}</span>
            </div>
          </div>

          <div className="public-options-heading">
            <h2>{success ? 'Poll results' : 'Choose your answer'}</h2>
            <span>{poll.options.length} options</span>
          </div>

          <div className="public-options-list">
            {poll.options.map((option) => {
              const percentage =
                totalVotes > 0
                  ? Math.round((option.votes / totalVotes) * 100)
                  : 0

              const isSelected = selectedOption === option.id

              return (
                <button
                  type="button"
                  key={option.id}
                  className={`public-option ${isSelected ? 'selected' : ''} ${
                    success ? 'show-result' : ''
                  }`}
                  onClick={() => {
                    if (!success && poll.isActive && !submitting) {
                      setSelectedOption(option.id)
                    }
                  }}
                  disabled={success || !poll.isActive || submitting}
                  aria-pressed={isSelected}
                >
                  <span className="public-option-content">
                    <span className="public-option-choice">
                      <span className="public-radio">
                        {isSelected && <span />}
                      </span>
                      <span className="public-option-name">{option.text}</span>
                    </span>

                    <span className="public-option-percentage">
                      {percentage}%
                    </span>
                  </span>

                  <span className="public-result-track">
                    <span
                      className="public-result-fill"
                      style={{ width: `${percentage}%` }}
                    />
                  </span>

                  <span className="public-option-votes">
                    {option.votes} {option.votes === 1 ? 'vote' : 'votes'}
                  </span>
                </button>
              )
            })}
          </div>

          {error && (
            <div className="public-error-message" role="alert">
              {error}
            </div>
          )}

          {success ? (
            <div className="public-success-message">
              <CheckCircle2 size={20} />
              Your vote has been submitted successfully!
            </div>
          ) : poll.isActive ? (
            <button
              type="button"
              className="public-submit-button"
              onClick={handleVote}
              disabled={!selectedOption || submitting}
            >
              {submitting ? (
                <>
                  <LoaderCircle className="spin" size={19} />
                  Submitting...
                </>
              ) : (
                <>
                  <Vote size={19} />
                  Submit Vote
                </>
              )}
            </button>
          ) : (
            <div className="public-closed-message">
              This poll is closed. Voting is no longer available.
            </div>
          )}

          <div className="public-poll-footer">
            <span>
              <Radio size={14} />
              Results update automatically
            </span>
            <span>Powered by <strong>MyPoll</strong></span>
          </div>
        </section>

        <p className="public-bottom-note">
          Your opinion matters. Thank you for participating!
        </p>
      </div>
    </main>
  )
}
