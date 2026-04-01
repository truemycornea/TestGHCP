import Link from 'next/link';

/**
 * Landing page — redirects authenticated users to /gallery.
 * Unauthenticated users see the marketing hero section.
 */
export default function HomePage() {
  return (
    <main className="flex min-h-screen flex-col items-center justify-center gap-8 px-4 text-center">
      <div className="space-y-4">
        <h1 className="text-5xl font-bold tracking-tight text-primary-600">
          Aura
        </h1>
        <p className="max-w-xl text-lg text-muted-foreground">
          A high-performance, self-hostable media library with AI-powered
          semantic search, face recognition, and smart deduplication.
        </p>
      </div>

      <div className="flex gap-4">
        <Link
          href="/gallery"
          className="rounded-lg bg-primary px-6 py-3 text-sm font-semibold text-primary-foreground shadow hover:bg-primary-600 transition-colors"
        >
          Open Gallery
        </Link>
        <Link
          href="/auth/login"
          className="rounded-lg border border-border px-6 py-3 text-sm font-semibold hover:bg-muted transition-colors"
        >
          Sign In
        </Link>
      </div>

      <div className="grid grid-cols-2 gap-4 text-left sm:grid-cols-3 lg:grid-cols-5 mt-8">
        {features.map((f) => (
          <div key={f.title} className="rounded-lg border border-border p-4 space-y-2">
            <span className="text-2xl">{f.icon}</span>
            <h3 className="font-semibold text-sm">{f.title}</h3>
            <p className="text-xs text-muted-foreground">{f.desc}</p>
          </div>
        ))}
      </div>
    </main>
  );
}

const features = [
  { icon: '🔍', title: 'Semantic Search', desc: 'Natural-language queries powered by CLIP embeddings.' },
  { icon: '🤖', title: 'Hybrid AI', desc: 'Local Ollama models + cloud OpenAI/Vertex AI tagging.' },
  { icon: '🗺️', title: 'Geo Intelligence', desc: 'Map view with reverse geocoding from GPS metadata.' },
  { icon: '👤', title: 'Face Grouping', desc: 'Automatic clustering of people across your library.' },
  { icon: '🔗', title: 'Smart Sharing', desc: 'Password-protected links with expiry and permissions.' },
];
