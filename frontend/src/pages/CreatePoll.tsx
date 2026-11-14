
import { useState, type FormEvent } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import {
  ArrowLeft,
  Plus,
  Trash2,
  Activity,
  CircleHelp,
  AlignLeft,
} from 'lucide-react'
import { api } from '../services/api'
import Navbar from '../components/Navbar'

export default function CreatePoll() {
  const navigate = useNavigate()

  const [question, setQuestion] = useState('')
  const [description, setDescription] = useState('')
  const [options, setOptions] = useState(['', ''])
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  function updateOption(index: number, value: string) {
    setOptions((current) =>
      current.map((option, i) => (i === index ? value : option))
    )
  }

  function addOption() {
    if (options.length < 10) {
      setOptions((current) => [...current, ''])
    }
  }

  function removeOption(index: number) {
    if (options.length > 2) {
      setOptions((current) =>
        current.filter((_, i) => i !== index)
      )
    }
  }

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    setError('')

    const cleanQuestion = question.trim()
    const cleanOptions = options.map((option) => option.trim())

    if (!cleanQuestion) {
      setError('Please enter a poll question.')
      return
    }

    if (cleanOptions.some((option) => !option)) {
      setError('Please fill in every answer option.')
      return
    }

    const uniqueOptions = new Set(
      cleanOptions.map((option) => option.toLowerCase())
    )

    if (uniqueOptions.size !== cleanOptions.length) {
      setError('Duplicate options are not allowed.')
      return
    }

    setLoading(true)

    try {
      await api.createPoll(
        cleanQuestion,
        description.trim(),
        cleanOptions
      )

      navigate('/dashboard')
    } catch (err) {
      setError(
        err instanceof Error
          ? err.message
          : 'Unable to create poll. Please try again.'
      )
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="app-layout">
      <Navbar />

      <main className="form-page">
        <div className="form-page-inner">
          <Link to="/dashboard" className="back-link">
            <ArrowLeft size={17} />
            Back to dashboard
          </Link>

          <section className="create-card">
            <div className="create-heading">
              <div className="create-icon">
                <Activity size={25} />
              </div>

              <span className="eyebrow">NEW POLL</span>

              <h1>Create a poll</h1>

              <p>
                Ask a question and let people share their opinions
                in real time.
              </p>
            </div>

            <form onSubmit={handleSubmit} className="create-form">
              <div className="form-field-group">
                <label htmlFor="question">
                  <CircleHelp size={16} />
                  Poll question
                </label>

                <input
                  id="question"
                  className="standard-input"
                  type="text"
                  placeholder="e.g. Which technology should we learn next?"
                  value={question}
                  onChange={(event) =>
                    setQuestion(event.target.value)
                  }
                  maxLength={200}
                  required
                />

                <span className="field-hint">
                  Keep your question clear and easy to understand.
                </span>
              </div>

              <div className="form-field-group">
                <label htmlFor="description">
                  <AlignLeft size={16} />
                  Description
                  <span className="optional-label">Optional</span>
                </label>

                <textarea
                  id="description"
                  className="standard-input description-input"
                  placeholder="Add some context to your poll..."
                  value={description}
                  onChange={(event) =>
                    setDescription(event.target.value)
                  }
                  maxLength={500}
                />
              </div>

              <div className="options-section">
                <div className="options-heading">
                  <div>
                    <h3>Answer options</h3>
                    <p>Add between 2 and 10 choices.</p>
                  </div>

                  <span className="option-count">
                    {options.length}/10
                  </span>
                </div>

                <div className="options-list">
                  {options.map((option, index) => (
                    <div className="option-input-row" key={index}>
                      <span className="option-number">
                        {index + 1}
                      </span>

                      <input
                        className="standard-input"
                        type="text"
                        placeholder={`Enter option ${index + 1}`}
                        value={option}
                        onChange={(event) =>
                          updateOption(index, event.target.value)
                        }
                        maxLength={100}
                        required
                        aria-label={`Answer option ${index + 1}`}
                      />

                      <button
                        type="button"
                        className="icon-button"
                        onClick={() => removeOption(index)}
                        disabled={options.length <= 2}
                        aria-label={`Remove option ${index + 1}`}
                      >
                        <Trash2 size={17} />
                      </button>
                    </div>
                  ))}
                </div>

                <button
                  type="button"
                  className="add-option-button"
                  onClick={addOption}
                  disabled={options.length >= 10}
                >
                  <Plus size={17} />
                  Add another option
                </button>
              </div>

              {error && (
                <div className="error-banner">
                  {error}
                </div>
              )}

              <div className="create-form-footer">
                <p>
                  You can share your poll with anyone after creating it.
                </p>

                <button
                  type="submit"
                  className="primary-button submit-poll-button"
                  disabled={loading}
                >
                  {loading ? 'Creating poll...' : 'Create Poll'}
                  {!loading && <Plus size={18} />}
                </button>
              </div>
            </form>
          </section>
        </div>
      </main>
    </div>
  )
}
