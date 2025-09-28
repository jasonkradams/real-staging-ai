export default function Page() {
  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-semibold">Welcome to Virtual Staging AI</h1>
      <p className="text-gray-600">Phase 1: Upload an image and track its processing status via SSE.</p>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        <a href="/upload" className="card">
          <div className="card-header">Upload</div>
          <div className="card-body text-gray-600">
            Presign, upload to S3, and create an image job.
          </div>
        </a>
        <a href="/images" className="card">
          <div className="card-header">Images</div>
          <div className="card-body text-gray-600">
            Enter an image ID to watch status updates in real-time.
          </div>
        </a>
      </div>
    </div>
  )
}
