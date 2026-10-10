import { useState } from "react"
import { Button } from "@/components/ui/button"
import { Label } from "@/components/ui/label"
import { Textarea } from "@/components/ui/textarea"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog"
import {
  Accordion,
  AccordionContent,
  AccordionItem,
  AccordionTrigger,
} from "@/components/ui/accordion"
import { KeyRound, CheckCircle2, XCircle, Loader2 } from "lucide-react"

interface CookieAuthProps {
  platform: "twitter" | "tiktok" | "threads" | "facebook"
  onSave?: (cookies: string) => void
}

const platformInfo: Record<string, { name: string; domain: string; requiredCookies: string[] }> = {
  twitter: {
    name: "X (Twitter)",
    domain: "x.com",
    requiredCookies: ["auth_token", "ct0"],
  },
  tiktok: {
    name: "TikTok",
    domain: "tiktok.com",
    requiredCookies: ["sessionid", "tt_chain_token"],
  },
  threads: {
    name: "Threads",
    domain: "threads.net",
    requiredCookies: ["sessionid", "csrftoken"],
  },
  facebook: {
    name: "Facebook",
    domain: "facebook.com",
    requiredCookies: ["c_user", "xs"],
  },
}

export function CookieAuth({ platform, onSave }: CookieAuthProps) {
  const [open, setOpen] = useState(false)
  const [cookies, setCookies] = useState("")
  const [saving, setSaving] = useState(false)
  const [status, setStatus] = useState<"idle" | "success" | "error">("idle")
  const [errorMsg, setErrorMsg] = useState("")

  const info = platformInfo[platform]

  const handleSave = async () => {
    if (!cookies.trim()) return

    setSaving(true)
    setStatus("idle")
    setErrorMsg("")

    try {
      const response = await fetch(`/api/v1/auth/cookies/${platform}`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ cookies: cookies.trim() }),
      })

      if (!response.ok) {
        const err = await response.json()
        throw new Error(err.error || "Failed to save cookies")
      }

      setStatus("success")
      onSave?.(cookies)
      setTimeout(() => setOpen(false), 1500)
    } catch (err) {
      setStatus("error")
      setErrorMsg(err instanceof Error ? err.message : "Failed to save cookies")
    } finally {
      setSaving(false)
    }
  }

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <Button variant="outline" size="sm">
          <KeyRound className="w-4 h-4 mr-2" />
          Set {info.name} Cookies
        </Button>
      </DialogTrigger>
      <DialogContent className="max-w-2xl">
        <DialogHeader>
          <DialogTitle>Authenticate with {info.name}</DialogTitle>
          <DialogDescription>
            Paste your session cookies to enable scraping. Cookies are stored locally on the server.
          </DialogDescription>
        </DialogHeader>

        <Accordion type="single" collapsible className="w-full">
          <AccordionItem value="instructions">
            <AccordionTrigger>How to get cookies from {info.name}</AccordionTrigger>
            <AccordionContent>
              <ol className="list-decimal list-inside space-y-2 text-sm text-muted-foreground">
                <li>Open <strong>{info.domain}</strong> in your browser and log in</li>
                <li>Open DevTools (F12 or Cmd+Option+I)</li>
                <li>Go to <strong>Application</strong> → <strong>Cookies</strong> → <strong>https://{info.domain}</strong></li>
                <li>
                  Copy these cookies: <code className="bg-muted px-1 rounded">{info.requiredCookies.join(", ")}</code>
                </li>
                <li>
                  Format as: <code className="bg-muted px-1 rounded">name1=value1; name2=value2</code>
                </li>
              </ol>
              <div className="mt-3 p-2 bg-muted rounded text-xs">
                <strong>Quick method:</strong> In DevTools Console, run:{" "}
                <code className="select-all">document.cookie</code> and copy the output.
              </div>
            </AccordionContent>
          </AccordionItem>
        </Accordion>

        <div className="space-y-2">
          <Label htmlFor="cookies">Paste Cookies</Label>
          <Textarea
            id="cookies"
            placeholder={`${info.requiredCookies[0]}=xxx; ${info.requiredCookies[1]}=yyy; ...`}
            value={cookies}
            onChange={(e) => setCookies(e.target.value)}
            rows={4}
            className="font-mono text-sm"
          />
        </div>

        {status === "success" && (
          <div className="flex items-center gap-2 text-green-600">
            <CheckCircle2 className="w-4 h-4" />
            <span>Cookies saved successfully!</span>
          </div>
        )}

        {status === "error" && (
          <div className="flex items-center gap-2 text-destructive">
            <XCircle className="w-4 h-4" />
            <span>{errorMsg}</span>
          </div>
        )}

        <div className="flex justify-end gap-2">
          <Button variant="ghost" onClick={() => setOpen(false)}>
            Cancel
          </Button>
          <Button onClick={handleSave} disabled={saving || !cookies.trim()}>
            {saving && <Loader2 className="w-4 h-4 mr-2 animate-spin" />}
            Save Cookies
          </Button>
        </div>
      </DialogContent>
    </Dialog>
  )
}
