
import { Activity, LayoutDashboard, PlusCircle, LogOut } from 'lucide-react'
import { Link, useNavigate } from 'react-router-dom'
import { useAuth } from '../context/AuthContext'

export default function Navbar() {
  const { user, logout } = useAuth()
  const navigate = useNavigate()

  function handleLogout() {
    logout()
    navigate('/login')
  }

  return (
    <header className="navbar">
      <Link to="/dashboard" className="brand navbar-brand">
        <span className="brand-icon">
          <Activity size={22} />
        </span>
        MyPoll
      </Link>

      <nav className="nav-links">
        <Link to="/dashboard">
          <LayoutDashboard size={17} />
          Dashboard
        </Link>

        <Link to="/create-poll">
          <PlusCircle size={17} />
          Create Poll
        </Link>
      </nav>

      <div className="nav-user">
        <span className="user-avatar">
          {user?.name?.charAt(0).toUpperCase() || 'U'}
        </span>

        <span className="user-name">{user?.name}</span>

        <button
          className="logout-button"
          onClick={handleLogout}
          title="Logout"
        >
          <LogOut size={18} />
        </button>
      </div>
    </header>
  )
}
