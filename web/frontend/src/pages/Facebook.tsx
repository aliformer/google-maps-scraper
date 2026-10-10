import { useState, useEffect } from "react"
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { JobTable } from "@/components/JobTable"
import { CookieAuth } from "@/components/CookieAuth"
import { getJobs, createJob } from "@/api/client"
import type { Job, CreateJobRequest } from "@/types"
import { Loader2, Facebook } from "lucide-react"

export function FacebookPage() {
  const [jobs, setJobs] = useState<Job[]>([])
  const [loading, setLoading] = useState(true)
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const [jobName, setJobName] = useState("")
  const [username, setUsername] = useState("")
  const [query, setQuery] = useState("")
  const [cookie, setCookie] = useState("")
  const [depth, setDepth] = useState(1)

  const fetchJobs = async () => {
    try {
      const data = await getJobs("facebook")
      setJobs(data)
    } catch (err) {
      console.error("Failed to fetch Facebook jobs:", err)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchJobs()
    const interval = setInterval(fetchJobs, 10000)
    return () => clearInterval(interval)
  }, [])

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!jobName.trim()) return

    setError(null)
    setSubmitting(true)

    const payload: CreateJobRequest = {
      name: jobName,
      type: "facebook",
      username: username.trim(),
      query: query.trim(),
      cookie: cookie.trim(),
      depth,
      keywords: [],
      lang: "en",
      zoom: 1,
      lat: "",
      lon: "",
      fast_mode: false,
      radius: 0,
      email: false,
      extra_reviews: false,
      max_time: 0,
      proxies: [],
    }

    try {
      await createJob(payload)
      setJobName("")
      setUsername("")
      setQuery("")
      setCookie("")
      await fetchJobs()
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to create Facebook job")
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <div>
          <h1 className="text-3xl font-bold tracking-tight flex items-center space-x-2">
            <Facebook className="w-8 h-8 text-blue-600" />
            <span>Facebook Scraper</span>
          </h1>
          <p className="text-muted-foreground">Scrape user posts, profiles, and public pages.</p>
        </div>
        <CookieAuth platform="facebook" />
      </div>

      <Card>
        <CardHeader>
          <CardTitle>New Facebook Scrape Job</CardTitle>
          <CardDescription>Target a user handle or a search query.</CardDescription>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSubmit} className="space-y-4">
            {error && <div className="p-3 text-sm bg-destructive/10 text-destructive rounded-md">{error}</div>}

            <div className="grid md:grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label htmlFor="name">Job Name</Label>
                <Input
                  id="name"
                  placeholder="e.g., Brand Mentions Feed"
                  value={jobName}
                  onChange={(e) => setJobName(e.target.value)}
                  required
                />
              </div>

              <div className="space-y-2">
                <Label htmlFor="username">Username/Page ID (Optional)</Label>
                <Input
                  id="username"
                  placeholder="e.g., zuck"
                  value={username}
                  onChange={(e) => setUsername(e.target.value)}
                />
              </div>
            </div>

            <div className="grid md:grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label htmlFor="query">Search Query (Optional)</Label>
                <Input
                  id="query"
                  placeholder="e.g., Tech events 2024"
                  value={query}
                  onChange={(e) => setQuery(e.target.value)}
                />
              </div>

              <div className="space-y-2">
                <Label htmlFor="depth">Max Scroll Depth / Pages</Label>
                <Input
                  id="depth"
                  type="number"
                  min={1}
                  max={50}
                  value={depth}
                  onChange={(e) => setDepth(parseInt(e.target.value) || 1)}
                />
              </div>
            </div>

            <div className="space-y-2">
              <Label htmlFor="cookie">Auth Cookie / Token (Optional - falls back to cookie pool)</Label>
              <Input
                id="cookie"
                placeholder="c_user=...; xs=..."
                value={cookie}
                onChange={(e) => setCookie(e.target.value)}
              />
            </div>

            <Button type="submit" disabled={submitting}>
              {submitting && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
              Start Facebook Job
            </Button>
          </form>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Facebook Scraping Jobs</CardTitle>
        </CardHeader>
        <CardContent>
          {loading ? (
            <div className="flex justify-center items-center py-8">
              <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
            </div>
          ) : (
            <JobTable jobs={jobs} onJobDeleted={fetchJobs} onViewJob={() => {}} />
          )}
        </CardContent>
      </Card>
    </div>
  )
}
