import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'

const Dashboard = () => {
  const [dashboardData, setDashboardData] = useState(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)

  useEffect(() => {
    const fetchDashboardData = async () => {
      try {
        const response = await fetch('/api/dashboard', {
          headers: {
            'X-API-Key': process.env.REACT_APP_API_KEY
          }
        })
        
        if (!response.ok) {
          throw new Error(`Dashboard load failed: ${response.statusText}`)
        }
        
        const data = await response.json()
        setDashboardData(data)
        setError(null)
      } catch (err) {
        console.error('Dashboard error:', err)
        setError(err.message)
      } finally {
        setLoading(false)
      }
    }

    fetchDashboardData()
  }, [])

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-50">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-blue-500 mx-auto"></div>
          <p className="mt-4 text-gray-600">Loading your training dashboard...</p>
        </div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-50">
        <div className="text-center text-red-600">
          <h2 className="text-xl font-semibold">Error loading dashboard</h2>
          <p className="mt-2">{error}</p>
          <button
            onClick={() => window.location.reload()}
            className="mt-4 px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600"
          >
            Try Again
          </button>
        </div>
      </div>
    )
  }

  return (
    <div className="dashboard bg-gray-50 min-h-screen p-4 md:p-6">
      <div className="max-w-7xl mx-auto">
        {/* Metrics Overview */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 mb-6">
          <div className="bg-white p-4 rounded-lg shadow">
            <h3 className="text-gray-500 text-sm font-medium">Weekly Volume</h3>
            <p className="text-2xl font-bold text-gray-900">
              {dashboardData.metrics.weekly_volume?.toFixed(1) || 0} hours
            </p>
          </div>
          <div className="bg-white p-4 rounded-lg shadow">
            <h3 className="text-gray-500 text-sm font-medium">Plan Progress</h3>
            <p className="text-2xl font-bold text-gray-900">
              {dashboardData.metrics.plan_progress || 0}%
            </p>
          </div>
        </div>

        {/* Recent Workouts */}
        <div className="bg-white rounded-lg shadow p-6 mb-6">
          <h2 className="text-xl font-semibold mb-4">Recent Workouts</h2>
          <div className="space-y-4">
            {dashboardData.recent_workouts.map(workout => (
              <div key={workout.id} className="border-b pb-4">
                <div className="flex justify-between items-center">
                  <div>
                    <h3 className="font-medium">{new Date(workout.start_time).toLocaleDateString()}</h3>
                    <p className="text-gray-600">{workout.activity_type}</p>
                  </div>
                  <div className="text-right">
                    <p className="text-gray-900 font-medium">
                      {(workout.distance_m / 1000).toFixed(1)} km
                    </p>
                    <p className="text-sm text-gray-500">
                      {Math.floor(workout.duration_seconds / 60)} minutes
                    </p>
                  </div>
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* Current Plan */}
        {dashboardData.current_plan && (
          <div className="bg-white rounded-lg shadow p-6 mb-6">
            <div className="flex justify-between items-center mb-4">
              <h2 className="text-xl font-semibold">Current Training Plan</h2>
              <Link 
                to={`/plans/${dashboardData.current_plan.id}`}
                className="text-blue-500 hover:text-blue-700"
              >
                View Details →
              </Link>
            </div>
            <div className="flex items-center">
              <div className="flex-1">
                <h3 className="font-medium">{dashboardData.current_plan.name}</h3>
                <p className="text-gray-600">
                  {dashboardData.current_plan.duration_weeks} week plan
                </p>
              </div>
              <div className="w-24 h-2 bg-gray-200 rounded-full overflow-hidden">
                <div 
                  className="h-full bg-blue-500" 
                  style={{ width: `${dashboardData.current_plan.progress}%` }}
                ></div>
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  )
}

export default Dashboard