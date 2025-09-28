"use client";

import { useEffect, useState } from "react";
import { apiFetch } from "@/lib/api";

type Project = {
  id: string
  name: string
}
type ProjectListResponse = {
  projects: Project[]
}

export default function UploadPage() {
  const [file, setFile] = useState<File | null>(null)
  const [projectId, setProjectId] = useState("")
  const [status, setStatus] = useState<string>("")
  const [imageId, setImageId] = useState<string>("")
  const [projects, setProjects] = useState<Project[]>([])
  const [newProjectName, setNewProjectName] = useState("")

  async function loadProjects() {
    try {
      setStatus("Loading projects...")
      const res = await apiFetch<ProjectListResponse>("/v1/projects")
      setProjects(res.projects || [])
      if (!projectId && res.projects && res.projects.length > 0) {
        setProjectId(res.projects[0].id)
      }
      setStatus("")
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : String(err)
      setStatus(message)
    }
  }

  async function createProject() {
    if (!newProjectName.trim()) {
      setStatus("Please provide a project name.")
      return
    }
    try {
      setStatus("Creating project...")
      const created = await apiFetch<Project>("/v1/projects", {
        method: "POST",
        body: JSON.stringify({ name: newProjectName.trim() }),
      })
      setNewProjectName("")
      // refresh list and select newly created project
      await loadProjects()
      setProjectId(created.id)
      setStatus("Project created.")
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : String(err)
      setStatus(message)
    }
  }

  useEffect(() => {
    loadProjects()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  async function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    setStatus("")
    setImageId("")
    if (!file) {
      setStatus("Please select a file.")
      return
    }
    if (!projectId) {
      setStatus("Please select or create a project.")
      return
    }

    try {
      // 1) Presign
      setStatus("Presigning upload URL...")
      const presign = await apiFetch<{ upload_url: string; file_key: string }>(
        "/v1/uploads/presign",
        {
          method: "POST",
          body: JSON.stringify({
            filename: file.name,
            content_type: file.type || "application/octet-stream",
            file_size: file.size,
          }),
        }
      )

      // 2) Upload to S3 via presigned URL
      setStatus("Uploading to S3...")
      const putRes = await fetch(presign.upload_url, {
        method: "PUT",
        headers: {
          "Content-Type": file.type || "application/octet-stream",
        },
        body: file,
      })
      if (!putRes.ok) {
        throw new Error(`Upload failed: ${putRes.status}`)
      }

      // 3) Create Image with original_url pointing to the uploaded object URL
      // Derive the base from the presigned URL to avoid hardcoding bucket/host
      const u = new URL(presign.upload_url)
      const originalUrl = `${u.origin}${u.pathname}`

      setStatus("Creating image job...")
      const created = await apiFetch<{ id: string }>("/v1/images", {
        method: "POST",
        body: JSON.stringify({ project_id: projectId, original_url: originalUrl }),
      })

      setImageId(created.id)
      setStatus("Image created! Use the Images page to watch status via SSE.")
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : String(err)
      setStatus(message)
    }
  }

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-semibold">Upload</h1>

      <div className="card">
        <div className="card-header">Projects</div>
        <div className="card-body space-y-4">
          <div className="flex items-end gap-2">
            <div className="flex-1">
              <label className="block text-sm mb-1">New Project Name</label>
              <input
                className="input w-full"
                value={newProjectName}
                onChange={(e) => setNewProjectName(e.target.value)}
                placeholder="e.g. Living Room Set"
              />
            </div>
            <button type="button" className="btn" onClick={createProject}>
              Create
            </button>
          </div>
          <div className="flex items-center gap-2">
            <button type="button" className="btn" onClick={loadProjects}>
              Refresh Projects
            </button>
            <select
              className="input"
              value={projectId}
              onChange={(e) => setProjectId(e.target.value)}
            >
              <option value="">Select a project...</option>
              {projects.map((p) => (
                <option key={p.id} value={p.id}>
                  {p.name} ({p.id.slice(0, 8)}…)
                </option>
              ))}
            </select>
          </div>
        </div>
      </div>

      <form onSubmit={onSubmit} className="card">
        <div className="card-header">Presign & Upload</div>
        <div className="card-body space-y-4">
          <div>
            <label className="block text-sm mb-1">Selected Project</label>
            <input className="input" value={projectId} readOnly placeholder="Select a project above" />
          </div>
          <div>
            <label className="block text-sm mb-1">File</label>
            <input
              type="file"
              onChange={(e) => setFile(e.target.files?.[0] || null)}
              className="block w-full"
            />
          </div>
          <div className="flex items-center gap-2">
            <button className="btn" type="submit">
              Upload & Create Image
            </button>
            {status && <span className="text-sm text-gray-600">{status}</span>}
          </div>
        </div>
      </form>

      {imageId && (
        <div className="card">
          <div className="card-header">Created</div>
          <div className="card-body">
            <div className="text-sm text-gray-600">
              Image ID: <code className="px-1 py-0.5 bg-gray-100 rounded">{imageId}</code>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
