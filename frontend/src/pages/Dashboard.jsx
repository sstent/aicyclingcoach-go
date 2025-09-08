import React, { useState, useEffect } from 'react';
import GarminSync from '../components/GarminSync';
import WorkoutCharts from '../components/WorkoutCharts';
import PlanTimeline from '../components/PlanTimeline';
import WorkoutAnalysis from '../components/WorkoutAnalysis';

const Dashboard = () => {
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [dashboardData, setDashboardData] = useState({
    recentWorkouts: [],
    upcomingWorkouts: [],
    currentPlan: null,
    planVersions: [],
    lastAnalysis: null,
    syncStatus: null,
    metrics: {}
  });

  useEffect(() => {
    const fetchDashboardData = async () => {
      try {
        setLoading(true);
        const response = await fetch('/api/dashboard', {
          headers: {
            'X-API-Key': process.env.REACT_APP_API_KEY
          }
        });
        
        if (!response.ok) {
          throw new Error(`Failed to load dashboard: ${response.statusText}`);
        }
        
        const data = await response.json();
        setDashboardData(data);
        setError(null);
      } catch (err) {
        console.error('Dashboard load error:', err);
        setError(err.message);
      } finally {
        setLoading(false);
      }
    };
    
    fetchDashboardData();
  }, []);

  const handleSyncComplete = (newSyncStatus) => {
    setDashboardData(prev => ({
      ...prev,
      syncStatus: newSyncStatus
    }));
  };

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-50">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-blue-500 mx-auto"></div>
          <p className="mt-4 text-gray-600">Loading your training dashboard...</p>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-50">
        <div className="max-w-md p-6 bg-white rounded-lg shadow-md">
          <div className="text-red-500 text-center mb-4">
            <svg xmlns="http://www.w3.org/2000/svg" className="h-12 w-12 mx-auto" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
            </svg>
          </div>
          <h2 className="text-xl font-bold text-gray-800 mb-2">Dashboard Error</h2>
          <p className="text-gray-600 mb-4">{error}</p>
          <button 
            onClick={() => window.location.reload()}
            className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 transition-colors"
          >
            Try Again
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="dashboard bg-gray-50 min-h-screen p-4 md:p-6">
      <div className="max-w-7xl mx-auto">
        <div className="mb-6">
          <h1 className="text-2xl md:text-3xl font-bold text-gray-900">Training Dashboard</h1>
          <p className="text-gray-600">Your personalized cycling training overview</p>
        </div>

        {/* Stats Overview */}
        <div className="grid grid-cols-1 md:grid-cols-4 gap-4 mb-6">
          <div className="bg-white p-4 rounded-lg shadow">
            <h3 className="text-gray-500 text-sm font-medium">Weekly Hours</h3>
            <p className="text-2xl font-bold text-gray-900">
              {dashboardData.metrics.weekly_hours || '0'}h
            </p>
          </div>
          <div className="bg-white p-4 rounded-lg shadow">
            <h3 className="text-gray-500 text-sm font-medium">Workouts This Week</h3>
            <p className="text-2xl font-bold text-gray-900">
              {dashboardData.metrics.workouts_this_week || '0'}
            </p>
          </div>
          <div className="bg-white p-4 rounded-lg shadow">
            <h3 className="text-gray-500 text-sm font-medium">Plan Progress</h3>
            <p className="text-2xl font-bold text-gray-900">
              {dashboardData.metrics.plan_progress || '0'}%
            </p>
          </div>
          <div className="bg-white p-4 rounded-lg shadow">
            <h3 className="text-gray-500 text-sm font-medium">Fitness Level</h3>
            <p className="text-2xl font-bold text-gray-900 capitalize">
              {dashboardData.metrics.fitness_level || 'N/A'}
            </p>
          </div>
        </div>

        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
          {/* Left Column */}
          <div className="lg:col-span-2 space-y-6">
            {/* Garmin Sync */}
            <div className="bg-white rounded-lg shadow-md p-5">
              <GarminSync onSyncComplete={handleSyncComplete} />
            </div>

            {/* Current Plan */}
            {dashboardData.currentPlan && (
              <div className="bg-white rounded-lg shadow-md p-5">
                <h2 className="text-xl font-bold text-gray-800 mb-4">Current Training Plan</h2>
                <PlanTimeline 
                  plan={dashboardData.currentPlan} 
                  versions={dashboardData.planVersions} 
                />
              </div>
            )}

            {/* Recent Analysis */}
            {dashboardData.lastAnalysis && (
              <div className="bg-white rounded-lg shadow-md p-5">
                <h2 className="text-xl font-bold text-gray-800 mb-4">Latest Workout Analysis</h2>
                <WorkoutAnalysis 
                  workout={dashboardData.lastAnalysis.workout} 
                  analysis={dashboardData.lastAnalysis} 
                />
              </div>
            )}
          </div>

          {/* Right Column */}
          <div className="space-y-6">
            {/* Upcoming Workouts */}
            <div className="bg-white rounded-lg shadow-md p-5">
              <h2 className="text-xl font-bold text-gray-800 mb-4">Upcoming Workouts</h2>
              {dashboardData.upcomingWorkouts.length > 0 ? (
                <div className="space-y-3">
                  {dashboardData.upcomingWorkouts.map(workout => (
                    <div key={workout.id} className="border-b border-gray-100 pb-3 last:border-0">
                      <div className="flex justify-between items-start">
                        <div>
                          <h3 className="font-medium text-gray-800 capitalize">
                            {workout.type.replace(/_/g, ' ')}
                          </h3>
                          <p className="text-sm text-gray-600">
                            {new Date(workout.scheduled_date).toLocaleDateString()} • {workout.duration_minutes} min
                          </p>
                        </div>
                        <span className="px-2 py-1 bg-blue-100 text-blue-800 text-xs rounded-full capitalize">
                          {workout.intensity.replace(/_/g, ' ')}
                        </span>
                      </div>
                      {workout.description && (
                        <p className="mt-1 text-sm text-gray-700">{workout.description}</p>
                      )}
                    </div>
                  ))}
                </div>
              ) : (
                <p className="text-gray-500 italic">No upcoming workouts scheduled</p>
              )}
            </div>

            {/* Recent Workouts */}
            <div className="bg-white rounded-lg shadow-md p-5">
              <h2 className="text-xl font-bold text-gray-800 mb-4">Recent Workouts</h2>
              {dashboardData.recentWorkouts.length > 0 ? (
                <div className="space-y-3">
                  {dashboardData.recentWorkouts.map(workout => (
                    <div key={workout.id} className="border-b border-gray-100 pb-3 last:border-0">
                      <div className="flex justify-between">
                        <div>
                          <h3 className="font-medium text-gray-800 capitalize">
                            {workout.activity_type || 'Cycling'}
                          </h3>
                          <p className="text-sm text-gray-600">
                            {new Date(workout.start_time).toLocaleDateString()} • {Math.round(workout.duration_seconds / 60)} min
                          </p>
                        </div>
                        <div className="text-right">
                          <span className="block text-sm font-medium">
                            {workout.distance_m ? `${(workout.distance_m / 1000).toFixed(1)} km` : ''}
                          </span>
                          {workout.analysis && workout.analysis.performance_score && (
                            <span className="text-xs px-2 py-0.5 bg-green-100 text-green-800 rounded-full">
                              Score: {workout.analysis.performance_score}/10
                            </span>
                          )}
                        </div>
                      </div>
                      {workout.analysis && workout.analysis.performance_summary && (
                        <p className="mt-1 text-sm text-gray-700 line-clamp-2">
                          {workout.analysis.performance_summary}
                        </p>
                      )}
                    </div>
                  ))}
                </div>
              ) : (
                <p className="text-gray-500 italic">No recent workouts recorded</p>
              )}
            </div>

            {/* Quick Actions */}
            <div className="bg-white rounded-lg shadow-md p-5">
              <h2 className="text-xl font-bold text-gray-800 mb-4">Quick Actions</h2>
              <div className="grid grid-cols-2 gap-3">
                <button className="px-3 py-2 bg-blue-600 text-white rounded-md text-sm font-medium hover:bg-blue-700 transition-colors">
                  Generate New Plan
                </button>
                <button className="px-3 py-2 bg-green-600 text-white rounded-md text-sm font-medium hover:bg-green-700 transition-colors">
                  Add Custom Workout
                </button>
                <button className="px-3 py-2 bg-purple-600 text-white rounded-md text-sm font-medium hover:bg-purple-700 transition-colors">
                  View All Routes
                </button>
                <button className="px-3 py-2 bg-yellow-600 text-white rounded-md text-sm font-medium hover:bg-yellow-700 transition-colors">
                  Update Rules
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default Dashboard;