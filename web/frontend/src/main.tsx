import React from 'react'
import ReactDOM from 'react-dom/client'
import { BrowserRouter, Routes, Route } from 'react-router-dom'
import { Layout } from './components/Layout'
import { LandingPage } from './pages/Landing'
import { GoogleMapsPage } from './pages/GoogleMaps'
import { TwitterPage } from './pages/Twitter'
import { ThreadsPage } from './pages/Threads'
import { FacebookPage } from './pages/Facebook'
import { TikTokPage } from './pages/TikTok'
import { LoginPage } from './pages/Login'
import { ProtectedRoute } from './components/ProtectedRoute'
import { AuthProvider } from './context/AuthContext'
import './index.css'

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <AuthProvider>
      <BrowserRouter>
        <Routes>
          <Route path="/login" element={<LoginPage />} />
          <Route element={<ProtectedRoute />}>
            <Route path="/" element={<Layout />}>
              <Route index element={<LandingPage />} />
              <Route path="gmaps" element={<GoogleMapsPage />} />
              <Route path="twitter" element={<TwitterPage />} />
              <Route path="threads" element={<ThreadsPage />} />
              <Route path="facebook" element={<FacebookPage />} />
              <Route path="tiktok" element={<TikTokPage />} />
            </Route>
          </Route>
        </Routes>
      </BrowserRouter>
    </AuthProvider>
  </React.StrictMode>,
)
