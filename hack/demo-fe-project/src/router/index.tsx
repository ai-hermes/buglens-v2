import { createBrowserRouter, Navigate } from 'react-router-dom'
import { AppLayout } from '@/layouts/AppLayout'
import { HomePage } from '@/pages/home'
import { NotFoundPage } from '@/pages/not-found'

export const router = createBrowserRouter([
  {
    path: '/',
    element: <AppLayout />,
    children: [
      {
        index: true,
        element: <Navigate to="/home" replace />,
      },
      {
        path: 'home',
        element: <HomePage />,
      },
    ],
  },
  {
    path: '*',
    element: <NotFoundPage />,
  },
])
