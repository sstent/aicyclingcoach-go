import React from 'react';
import { 
  LineChart, Line, XAxis, YAxis, CartesianGrid, 
  Tooltip, Legend, ResponsiveContainer 
} from 'recharts';

const WorkoutCharts = ({ timeSeries }) => {
  // Transform timestamp to minutes from start for X-axis
  const formatTimeSeries = (data) => {
    if (!data || data.length === 0) return [];
    
    const startTime = new Date(data[0].timestamp);
    return data.map(point => ({
      ...point,
      time: (new Date(point.timestamp) - startTime) / 60000, // Convert to minutes
      heart_rate: point.heart_rate || null,
      power: point.power || null,
      cadence: point.cadence || null
    }));
  };

  const formattedData = formatTimeSeries(timeSeries);
  
  return (
    <div className="workout-charts bg-white p-4 rounded-lg shadow-md">
      <h3 className="text-lg font-medium text-gray-800 mb-4">Workout Metrics</h3>
      
      <ResponsiveContainer width="100%" height={300}>
        <LineChart
          data={formattedData}
          margin={{ top: 5, right: 30, left: 20, bottom: 5 }}
        >
          <CartesianGrid strokeDasharray="3 3" stroke="#f0f0f0" />
          <XAxis 
            dataKey="time" 
            label={{ 
              value: 'Time (minutes)', 
              position: 'insideBottomRight', 
              offset: -5 
            }}
            domain={['dataMin', 'dataMax']}
            tickCount={6}
          />
          <YAxis yAxisId="left" orientation="left" stroke="#8884d8">
            <Label value="HR (bpm) / Cadence (rpm)" angle={-90} position="insideLeft" />
          </YAxis>
          <YAxis yAxisId="right" orientation="right" stroke="#82ca9d">
            <Label value="Power (W)" angle={90} position="insideRight" />
          </YAxis>
          <Tooltip 
            formatter={(value, name) => [`${value} ${name === 'power' ? 'W' : name === 'heart_rate' ? 'bpm' : 'rpm'}`, name]}
            labelFormatter={(value) => `Time: ${value.toFixed(1)} min`}
          />
          <Legend />
          <Line 
            yAxisId="left"
            type="monotone"
            dataKey="heart_rate"
            name="Heart Rate"
            stroke="#8884d8"
            strokeWidth={2}
            dot={false}
            activeDot={{ r: 6 }}
            isAnimationActive={false}
          />
          <Line 
            yAxisId="right"
            type="monotone"
            dataKey="power"
            name="Power"
            stroke="#82ca9d"
            strokeWidth={2}
            dot={false}
            activeDot={{ r: 6 }}
            isAnimationActive={false}
          />
          <Line 
            yAxisId="left"
            type="monotone"
            dataKey="cadence"
            name="Cadence"
            stroke="#ffc658"
            strokeWidth={2}
            dot={false}
            activeDot={{ r: 6 }}
            isAnimationActive={false}
          />
        </LineChart>
      </ResponsiveContainer>
      
      <div className="mt-4 text-sm text-gray-500">
        <p>Note: Charts show metrics over time during the workout. Hover over points to see exact values.</p>
      </div>
    </div>
  );
};

export default WorkoutCharts;