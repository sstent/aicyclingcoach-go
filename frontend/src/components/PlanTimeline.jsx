import React, { useState } from 'react';

const PlanTimeline = ({ plan, versions }) => {
  const [expandedWeeks, setExpandedWeeks] = useState({});

  const toggleWeek = (weekNumber) => {
    setExpandedWeeks(prev => ({
      ...prev,
      [weekNumber]: !prev[weekNumber]
    }));
  };

  return (
    <div className="plan-timeline bg-white rounded-lg shadow-md p-5">
      <div className="header flex justify-between items-center mb-6">
        <div>
          <h2 className="text-xl font-bold text-gray-800">{plan.name || 'Training Plan'}</h2>
          <p className="text-gray-600">Version {plan.version} • Created {new Date(plan.created_at).toLocaleDateString()}</p>
        </div>
        <div className="bg-blue-100 text-blue-800 px-3 py-1 rounded-full text-sm">
          {plan.jsonb_plan.overview.focus.replace(/_/g, ' ')}
        </div>
      </div>

      {versions.length > 1 && (
        <div className="version-history mb-8">
          <h3 className="text-lg font-medium text-gray-800 mb-3">Version History</h3>
          <div className="overflow-x-auto">
            <table className="min-w-full divide-y divide-gray-200">
              <thead className="bg-gray-50">
                <tr>
                  <th className="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Version</th>
                  <th className="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Created</th>
                  <th className="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Trigger</th>
                  <th className="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Changes</th>
                </tr>
              </thead>
              <tbody className="bg-white divide-y divide-gray-200">
                {versions.map(version => (
                  <tr key={version.id} className={version.id === plan.id ? 'bg-blue-50' : ''}>
                    <td className="px-4 py-3 whitespace-nowrap text-sm font-medium text-gray-900">
                      v{version.version}
                    </td>
                    <td className="px-4 py-3 whitespace-nowrap text-sm text-gray-500">
                      {new Date(version.created_at).toLocaleDateString()}
                    </td>
                    <td className="px-4 py-3 whitespace-nowrap text-sm text-gray-500 capitalize">
                      {version.evolution_trigger?.replace(/_/g, ' ') || 'initial'}
                    </td>
                    <td className="px-4 py-3 text-sm text-gray-500">
                      {version.changes_summary || 'Initial version'}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}

      <div className="plan-overview bg-gray-50 p-4 rounded-md mb-6">
        <h3 className="text-lg font-medium text-gray-800 mb-2">Plan Overview</h3>
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <div className="metric-card">
            <span className="text-gray-500">Duration</span>
            <span className="text-xl font-bold text-gray-800">
              {plan.jsonb_plan.overview.duration_weeks} weeks
            </span>
          </div>
          <div className="metric-card">
            <span className="text-gray-500">Weekly Hours</span>
            <span className="text-xl font-bold text-gray-800">
              {plan.jsonb_plan.overview.total_weekly_hours} hours
            </span>
          </div>
          <div className="metric-card">
            <span className="text-gray-500">Focus</span>
            <span className="text-xl font-bold text-gray-800 capitalize">
              {plan.jsonb_plan.overview.focus.replace(/_/g, ' ')}
            </span>
          </div>
        </div>
      </div>

      <div className="weekly-schedule">
        <h3 className="text-lg font-medium text-gray-800 mb-4">Weekly Schedule</h3>
        
        {plan.jsonb_plan.weeks.map((week, weekIndex) => (
          <div key={weekIndex} className="week-card border border-gray-200 rounded-md mb-4 overflow-hidden">
            <div 
              className="week-header bg-gray-100 p-3 flex justify-between items-center cursor-pointer hover:bg-gray-200"
              onClick={() => toggleWeek(weekIndex)}
            >
              <h4 className="font-medium text-gray-800">Week {week.week_number} • {week.focus.replace(/_/g, ' ')}</h4>
              <div className="flex items-center">
                <span className="text-sm text-gray-600 mr-2">
                  {week.total_hours} hours • {week.workouts.length} workouts
                </span>
                <svg 
                  className={`w-5 h-5 text-gray-500 transform transition-transform ${
                    expandedWeeks[weekIndex] ? 'rotate-180' : ''
                  }`} 
                  fill="none" 
                  viewBox="0 0 24 24" 
                  stroke="currentColor"
                >
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
                </svg>
              </div>
            </div>
            
            {expandedWeeks[weekIndex] && (
              <div className="workouts-list p-4 bg-white">
                {week.workouts.map((workout, workoutIndex) => (
                  <div key={workoutIndex} className="workout-item border-b border-gray-100 py-3 last:border-0">
                    <div className="flex justify-between">
                      <div>
                        <span className="font-medium text-gray-800 capitalize">{workout.type.replace(/_/g, ' ')}</span>
                        <span className="text-gray-600 ml-2">• {workout.day}</span>
                      </div>
                      <span className="text-gray-600">{workout.duration_minutes} min</span>
                    </div>
                    
                    <div className="mt-1 flex flex-wrap gap-2">
                      <span className="px-2 py-1 bg-blue-100 text-blue-800 text-xs rounded-full capitalize">
                        {workout.intensity.replace(/_/g, ' ')}
                      </span>
                      {workout.route_id && (
                        <span className="px-2 py-1 bg-green-100 text-green-800 text-xs rounded-full">
                          Route: {workout.route_name || workout.route_id}
                        </span>
                      )}
                      <span className="px-2 py-1 bg-purple-100 text-purple-800 text-xs rounded-full">
                        TSS: {workout.tss_target || 'N/A'}
                      </span>
                    </div>
                    
                    {workout.description && (
                      <p className="mt-2 text-gray-700 text-sm">{workout.description}</p>
                    )}
                  </div>
                ))}
              </div>
            )}
          </div>
        ))}
      </div>
    </div>
  );
};

export default PlanTimeline;