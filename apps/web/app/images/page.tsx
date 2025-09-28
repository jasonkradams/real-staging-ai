"use client";

import SSEViewer from "@/components/SSEViewer";

export default function ImagesPage() {
  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-semibold">Images</h1>
      <p className="text-gray-600">Enter an image ID and connect to stream real-time status updates (SSE).</p>
      <SSEViewer />
    </div>
  );
}
