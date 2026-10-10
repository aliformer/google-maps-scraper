import { Link, Outlet, useLocation } from "react-router-dom"
import { MapPin, Twitter, AtSign, Facebook, Video, Home, Github, LogOut, User } from "lucide-react"
import { useAuth } from "@/context/AuthContext"
import { Button } from "@/components/ui/button"

export function Layout() {
  const location = useLocation()
  const { user, signOut } = useAuth()
  const title = import.meta.env.VITE_PROJECT_NAME
  const navItems = [
    { path: "/", label: "Home", icon: Home },
    { path: "/gmaps", label: "Google Maps", icon: MapPin },
    { path: "/twitter", label: "X (Twitter)", icon: Twitter },
    { path: "/threads", label: "Threads", icon: AtSign },
    { path: "/facebook", label: "Facebook", icon: Facebook },
    { path: "/tiktok", label: "TikTok", icon: Video },
  ]

  return (
    <div className="min-h-screen bg-background text-foreground flex flex-col">
      <header className="border-b bg-card">
        <div className="container mx-auto px-4 h-16 flex items-center justify-between">
          <div className="flex items-center space-x-6">
            <Link to="/" className="font-bold text-xl flex items-center space-x-2">
              <span className="bg-primary text-primary-foreground px-2 py-1 rounded">SP</span>
              <span>{title}</span>
            </Link>
            <nav className="flex items-center space-x-1">
              {navItems.map((item) => {
                const Icon = item.icon
                const isActive = location.pathname === item.path
                return (
                  <Link
                    key={item.path}
                    to={item.path}
                    className={`flex items-center space-x-2 px-3 py-2 rounded-md text-sm font-medium transition-colors ${isActive
                      ? "bg-accent text-accent-foreground"
                      : "text-muted-foreground hover:bg-accent/50 hover:text-foreground"
                      }`}
                  >
                    <Icon className="w-4 h-4" />
                    <span className="hidden sm:inline">{item.label}</span>
                  </Link>
                )
              })}
            </nav>
          </div>
          <div className="flex items-center space-x-4">
            <a
              href="https://github.com/gosom/google-maps-scraper"
              target="_blank"
              rel="noreferrer"
              className="text-muted-foreground hover:text-foreground transition-colors hidden sm:block"
            >
              <Github className="w-5 h-5" />
            </a>
            {user && (
              <div className="flex items-center space-x-3 border-l pl-4 ml-2">
                <div className="flex items-center space-x-2 text-sm text-muted-foreground hidden md:flex">
                  <User className="w-4 h-4" />
                  <span className="truncate max-w-[150px]">{user.email}</span>
                </div>
                <Button variant="ghost" size="sm" onClick={signOut}>
                  <LogOut className="w-4 h-4 mr-2" />
                  <span className="hidden sm:inline">Sign Out</span>
                </Button>
              </div>
            )}
          </div>
        </div>
      </header>
      <main className="flex-1 container mx-auto px-4 py-8">
        <Outlet />
      </main>
    </div>
  )
}
