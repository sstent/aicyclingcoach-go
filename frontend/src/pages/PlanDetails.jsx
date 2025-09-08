import { useParams } from 'react-router-dom'
import PlanTimeline from '../components/PlanTimeline'

const PlanDetails = () => {
  const { planId } = useParams()

  return (
    <div className="min-h-screen bg-gray-50 p-4 md:p-6">
      <div className="max-w-4xl mx-auto">
        <PlanTimeline planId={planId} />
      </div>
    </div>
  )
}

export default PlanDetails