import { useState, useEffect } from "react"
import { JobForm } from "@/components/JobForm"
import { JobTable } from "@/components/JobTable"
import { JobMapModal } from "@/components/JobMapModal"
import { getJobs, createJob, getJobPlaces } from "@/api/client"
import type { Job, CreateJobRequest, Place } from "@/types"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Loader2 } from "lucide-react"

export function GoogleMapsPage() {
  const [jobs, setJobs] = useState<Job[]>([])
  const [error, setError] = useState<string | null>(null)
  const [loading, setLoading] = useState(true)
  const [selectedJob, setSelectedJob] = useState<{ id: string; name: string } | null>(null)
  const [places, setPlaces] = useState<Place[]>([])
  const [mapLoading, setMapLoading] = useState(false)
  const [mapError, setMapError] = useState<string | null>(null)

  const fetchJobs = async () => {
    try {
      const data = await getJobs("gmaps")
      setJobs(data)
    } catch (err) {
      console.error("Failed to fetch jobs:", err)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchJobs()
    const interval = setInterval(fetchJobs, 10000)
    return () => clearInterval(interval)
  }, [])

  const handleSubmit = async (jobData: CreateJobRequest) => {
    setError(null)
    try {
      await createJob(jobData)
      await fetchJobs()
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to create job")
      throw err
    }
  }

  const handleViewJob = async (job: Job) => {
    setSelectedJob({ id: job.id, name: job.name })
    setPlaces([])
    setMapError(null)
    setMapLoading(true)

    try {
      const parsedPlaces = await getJobPlaces(job.id)
      setPlaces(parsedPlaces)
    } catch (err) {
      setMapError(err instanceof Error ? err.message : "Failed to load map data")
    } finally {
      setMapLoading(false)
    }
  }

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Google Maps Scraper</h1>
          <p className="text-muted-foreground">Extract places, emails, phone numbers, and reviews.</p>
        </div>
      </div>

      <JobForm onSubmit={handleSubmit} error={error} />

      <Card>
        <CardHeader>
          <CardTitle>Jobs</CardTitle>
        </CardHeader>
        <CardContent>
          {loading ? (
            <div className="flex justify-center items-center py-8">
              <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
            </div>
          ) : (
            <JobTable jobs={jobs} onJobDeleted={fetchJobs} onViewJob={handleViewJob} />
          )}
        </CardContent>
      </Card>

      {selectedJob && (
        <JobMapModal
          job={selectedJob}
          onClose={() => setSelectedJob(null)}
          places={places}
          loading={mapLoading}
          error={mapError}
        />
      )}
    </div>
  )
}
