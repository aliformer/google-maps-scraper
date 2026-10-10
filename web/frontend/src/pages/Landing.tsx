import { Link } from "react-router-dom"
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { MapPin, Twitter, AtSign, Facebook, Video, ArrowRight, Zap, Shield, Database } from "lucide-react"

export function LandingPage() {
  const scrapers = [
    {
      title: "Google Maps",
      description: "Extract businesses, reviews, emails, phones, ratings, coordinates and images at scale.",
      icon: MapPin,
      path: "/gmaps",
      color: "text-blue-500",
    },
    {
      title: "X (Twitter)",
      description: "Scrape user profiles, tweets, hashtags, metrics, replies, and conversation threads.",
      icon: Twitter,
      path: "/twitter",
      color: "text-sky-500",
    },
    {
      title: "Threads",
      description: "Harvest public threads, creator posts, engagement metrics, and community replies.",
      icon: AtSign,
      path: "/threads",
      color: "text-emerald-500",
    },
    {
      title: "Facebook",
      description: "Scrape user posts, profile details, pages, and engagement metrics.",
      icon: Facebook,
      path: "/facebook",
      color: "text-blue-600",
    },
    {
      title: "TikTok",
      description: "Scrape user video feeds, video captions, engagement stats, and hashtags.",
      icon: Video,
      path: "/tiktok",
      color: "text-pink-500",
    },
  ]

  const features = [
    {
      title: "High Performance",
      description: "Built on Go concurrency, Playwright automation, and ScrapeMate orchestration.",
      icon: Zap,
    },
    {
      title: "Proxy & Rate Control",
      description: "Built-in proxy rotation, user-agent randomization, and request throttling.",
      icon: Shield,
    },
    {
      title: "Multi-Storage Ready",
      description: "Export directly to CSV, JSON, PostgreSQL, or Amazon S3 buckets.",
      icon: Database,
    },
  ]

  return (
    <div className="space-y-12 max-w-5xl mx-auto">
      <div className="text-center space-y-4 py-8">
        <h1 className="text-4xl sm:text-5xl font-extrabold tracking-tight">
          Universal Data Extraction Platform
        </h1>
        <p className="text-lg text-muted-foreground max-w-2xl mx-auto">
          High-performance distributed scrapers for Google Maps, X, Threads, Facebook, and TikTok. Run jobs locally or scale across clusters.
        </p>
      </div>

      <div className="grid md:grid-cols-3 gap-6">
        {scrapers.map((s) => {
          const Icon = s.icon
          return (
            <Card key={s.title} className="flex flex-col justify-between hover:border-primary/50 transition-colors">
              <CardHeader>
                <div className="flex items-center space-x-3 mb-2">
                  <Icon className={`w-8 h-8 ${s.color}`} />
                  <CardTitle>{s.title}</CardTitle>
                </div>
                <CardDescription className="text-sm">{s.description}</CardDescription>
              </CardHeader>
              <CardContent>
                <Button asChild className="w-full">
                  <Link to={s.path} className="flex items-center justify-center space-x-2">
                    <span>Open Scraper</span>
                    <ArrowRight className="w-4 h-4" />
                  </Link>
                </Button>
              </CardContent>
            </Card>
          )
        })}
      </div>

      <div className="pt-8 border-t">
        <h2 className="text-2xl font-bold text-center mb-8">Platform Features</h2>
        <div className="grid md:grid-cols-3 gap-8">
          {features.map((f) => {
            const Icon = f.icon
            return (
              <div key={f.title} className="flex flex-col items-center text-center space-y-2">
                <div className="p-3 bg-muted rounded-full mb-2">
                  <Icon className="w-6 h-6 text-primary" />
                </div>
                <h3 className="font-semibold">{f.title}</h3>
                <p className="text-sm text-muted-foreground">{f.description}</p>
              </div>
            )
          })}
        </div>
      </div>
    </div>
  )
}
