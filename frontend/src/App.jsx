import { BrowserRouter as Router, Route, Routes } from 'react-router-dom'
import Dashboard from './pages/Dashboard'
import PlanDetails from './pages/PlanDetails'

function App() {
  return (
    <Router>
      <Routes>
        <Route path="/" element={<Dashboard />} />
        <Route path="/plans/:planId" element={<PlanDetails />} />
      </Routes>
    </Router>
  )
}

export default App