import { BrowserRouter, Routes, Route } from 'react-router-dom'
import './App.css'
import Auth from './pages/Auth/Auth'
import MainLayout from './layouts/MainLayout'
import Dashboard from './pages/Dashboard/Dashboard'
import Tasks from './pages/Tasks/Tasks'
import Leaderboard from './pages/Leaderboard/Leaderboard'
import Rewards from './pages/Rewards/Rewards'

function App() {

  return (
    <BrowserRouter>
      <Routes>
        <Route path='/auth' element={<Auth/>}/>
        <Route element={<MainLayout/>}>
          <Route path='/' element={<Dashboard/>}/>
          <Route path='/tasks' element={<Tasks/>}/>
          <Route path='/leaderboard' element={<Leaderboard/>}/>
          <Route path='/rewards' element={<Rewards/>}/>
        </Route>
      </Routes>
    </BrowserRouter>
  )
}

export default App
