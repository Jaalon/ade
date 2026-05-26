import { NavLink } from 'react-router-dom'
import type { ReactNode } from 'react'

const navItems = [
  { to: '/', label: 'Dashboard' },
  { to: '/projects', label: 'Projets' },
  { to: '/plugins', label: 'Plugins' },
  { to: '/workflows', label: 'Workflows' },
  { to: '/reports', label: 'Rapports' },
]

export default function Layout({ children }: { children: ReactNode }) {
  return (
    <div className="layout">
      <header className="header">
        <div className="header-brand">ADE Orchestrator</div>
        <nav className="header-nav">
          {navItems.map(item => (
            <NavLink
              key={item.to}
              to={item.to}
              end={item.to === '/'}
              className={({ isActive }) => `nav-link${isActive ? ' active' : ''}`}
            >
              {item.label}
            </NavLink>
          ))}
        </nav>
      </header>
      <main className="main">{children}</main>
    </div>
  )
}
