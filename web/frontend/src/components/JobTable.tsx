import { useState } from "react"
import type { Job } from "@/types"
import { deleteJob, downloadJobFile } from "@/api/client"
import { Button } from "@/components/ui/button"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"
import { StatusBadge } from "@/components/ui/status-badge"
import { Eye, Download, Trash2, Loader2 } from "lucide-react"

interface JobTableProps {
  jobs: Job[]
  onJobDeleted: () => void
  onViewJob: (job: Job) => void
}

export function JobTable({ jobs = [], onJobDeleted, onViewJob }: JobTableProps) {
  const [deleting, setDeleting] = useState<string | null>(null)
  const [downloading, setDownloading] = useState<string | null>(null)
  const safeJobs = jobs || []

  const handleDownload = async (job: Job) => {
    setDownloading(job.id)
    try {
      await downloadJobFile(job.id, `${job.name || job.id}.csv`)
    } catch (err) {
      alert(`Failed to download: ${err}`)
    } finally {
      setDownloading(null)
    }
  }

  const handleDelete = async (id: string) => {
    if (!confirm("Are you sure you want to delete this job?")) return

    setDeleting(id)
    try {
      await deleteJob(id)
      onJobDeleted()
    } catch (err) {
      alert(`Failed to delete job: ${err}`)
    } finally {
      setDeleting(null)
    }
  }

  const formatDate = (dateStr?: string) => {
    if (!dateStr) return "-"
    const date = new Date(dateStr)
    if (isNaN(date.getTime())) return "-"
    return date.toLocaleString("en-US", {
      year: "numeric",
      month: "short",
      day: "2-digit",
      hour: "2-digit",
      minute: "2-digit",
    })
  }

  if (safeJobs.length === 0) {
    return (
      <div className="text-center py-12 text-muted-foreground">
        No jobs yet. Create your first scraping job!
      </div>
    )
  }

  return (
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead>Job ID</TableHead>
          <TableHead>Name</TableHead>
          <TableHead>Date</TableHead>
          <TableHead>Status</TableHead>
          <TableHead className="text-right">Actions</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {safeJobs.map((job) => (
          <TableRow key={job.id}>
            <TableCell className="font-mono text-xs">
              {job.id ? `${job.id.substring(0, 8)}...` : "N/A"}
            </TableCell>
            <TableCell className="font-medium">{job.name}</TableCell>
            <TableCell className="text-sm text-muted-foreground">
              {formatDate(job.date)}
            </TableCell>
            <TableCell>
              <StatusBadge status={job.status} />
            </TableCell>
            <TableCell className="text-right">
              <div className="flex justify-end gap-2">
                <Button
                  size="sm"
                  variant="outline"
                  onClick={() => onViewJob(job)}
                  disabled={job.status !== "ok"}
                >
                  <Eye className="h-4 w-4" />
                </Button>
                {job.status === "ok" && (
                  <Button
                    size="sm"
                    variant="outline"
                    onClick={() => handleDownload(job)}
                    disabled={downloading === job.id}
                  >
                    {downloading === job.id ? (
                      <Loader2 className="h-4 w-4 animate-spin" />
                    ) : (
                      <Download className="h-4 w-4" />
                    )}
                  </Button>
                )}
                <Button
                  size="sm"
                  variant="destructive"
                  onClick={() => handleDelete(job.id)}
                  disabled={deleting === job.id}
                >
                  {deleting === job.id ? (
                    <Loader2 className="h-4 w-4 animate-spin" />
                  ) : (
                    <Trash2 className="h-4 w-4" />
                  )}
                </Button>
              </div>
            </TableCell>
          </TableRow>
        ))}
      </TableBody>
    </Table>
  )
}
