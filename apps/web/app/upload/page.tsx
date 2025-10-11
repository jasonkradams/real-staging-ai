"use client";

import { useEffect, useState, useCallback } from "react";
import { Upload as UploadIcon, FolderOpen, Plus, RefreshCw, CheckCircle2, Loader2, FileImage } from "lucide-react";
import { apiFetch } from "@/lib/api";
import { cn } from "@/lib/utils";

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
  const [roomType, setRoomType] = useState("")
  const [style, setStyle] = useState("")
  const [status, setStatus] = useState<string>("")
  const [imageId, setImageId] = useState<string>("")
  const [projects, setProjects] = useState<Project[]>([])
  const [newProjectName, setNewProjectName] = useState("")
  const [isUploading, setIsUploading] = useState(false)
  const [isDragging, setIsDragging] = useState(false)

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

  const handleDragOver = useCallback((e: React.DragEvent) => {
    e.preventDefault()
    setIsDragging(true)
  }, [])

  const handleDragLeave = useCallback((e: React.DragEvent) => {
    e.preventDefault()
    setIsDragging(false)
  }, [])

  const handleDrop = useCallback((e: React.DragEvent) => {
    e.preventDefault()
    setIsDragging(false)
    const droppedFile = e.dataTransfer.files?.[0]
    if (droppedFile && droppedFile.type.startsWith('image/')) {
      setFile(droppedFile)
    }
  }, [])

  async function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    setStatus("")
    setImageId("")
    setIsUploading(true)
    if (!file) {
      setStatus("Please select a file.")
      setIsUploading(false)
      return
    }
    if (!projectId) {
      setStatus("Please select or create a project.")
      setIsUploading(false)
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
      const body: { project_id: string; original_url: string; room_type?: string; style?: string } = {
        project_id: projectId,
        original_url: originalUrl,
      }
      if (roomType) body.room_type = roomType
      if (style) body.style = style

      const created = await apiFetch<{ id: string }>("/v1/images", {
        method: "POST",
        body: JSON.stringify(body),
      })

      setImageId(created.id)
      setStatus("Success! Image created and queued for staging.")
      setIsUploading(false)
      // Reset form after successful upload
      setTimeout(() => {
        setFile(null)
        setRoomType("")
        setStyle("")
      }, 2000)
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : String(err)
      setStatus(message)
      setIsUploading(false)
    }
  }

  return (
    <div className="space-y-8">
      {/* Header */}
      <div>
        <h1 className="text-3xl font-bold mb-2">
          <span className="gradient-text">Upload & Stage</span>
        </h1>
        <p className="text-gray-600">
          Upload property photos and transform them with AI-powered virtual staging
        </p>
      </div>

      {/* Project Management */}
      <div className="card">
        <div className="card-header">
          <div className="flex items-center gap-2">
            <FolderOpen className="h-5 w-5 text-blue-600" />
            <span>Project Selection</span>
          </div>
        </div>
        <div className="card-body space-y-4">
          {/* Create New Project */}
          <div className="flex flex-col sm:flex-row gap-3">
            <div className="flex-1">
              <label className="block text-sm font-medium text-gray-700 mb-2">Create New Project</label>
              <input
                className="input"
                value={newProjectName}
                onChange={(e) => setNewProjectName(e.target.value)}
                placeholder="e.g., Downtown Condo Staging"
                onKeyDown={(e) => e.key === 'Enter' && createProject()}
              />
            </div>
            <button 
              type="button" 
              className="btn btn-secondary sm:mt-7" 
              onClick={createProject}
              disabled={!newProjectName.trim()}
            >
              <Plus className="h-4 w-4" />
              Create
            </button>
          </div>

          {/* Select Existing Project */}
          <div className="flex flex-col sm:flex-row gap-3">
            <div className="flex-1">
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Select Project
              </label>
              <select
                className="input"
                value={projectId}
                onChange={(e) => setProjectId(e.target.value)}
              >
                <option value="">Choose a project...</option>
                {projects.map((p) => (
                  <option key={p.id} value={p.id}>
                    {p.name}
                  </option>
                ))}
              </select>
            </div>
            <button 
              type="button" 
              className="btn btn-ghost sm:mt-7" 
              onClick={loadProjects}
            >
              <RefreshCw className="h-4 w-4" />
              Refresh
            </button>
          </div>
        </div>
      </div>

      {/* Upload Form */}
      <form onSubmit={onSubmit} className="card">
        <div className="card-header">
          <div className="flex items-center gap-2">
            <UploadIcon className="h-5 w-5 text-blue-600" />
            <span>Upload Image</span>
          </div>
        </div>
        <div className="card-body space-y-6">
          {/* Drag and Drop Zone */}
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-3">Property Image</label>
            <div
              onDragOver={handleDragOver}
              onDragLeave={handleDragLeave}
              onDrop={handleDrop}
              className={cn(
                "relative rounded-xl border-2 border-dashed transition-all duration-200 p-8",
                isDragging 
                  ? "border-blue-500 bg-blue-50" 
                  : "border-gray-300 hover:border-gray-400",
                file && "border-green-500 bg-green-50"
              )}
            >
              <input
                type="file"
                onChange={(e) => setFile(e.target.files?.[0] || null)}
                className="absolute inset-0 w-full h-full opacity-0 cursor-pointer"
                accept="image/jpeg,image/png,image/webp"
              />
              <div className="flex flex-col items-center justify-center text-center space-y-3">
                {file ? (
                  <>
                    <div className="rounded-xl bg-green-100 p-3">
                      <CheckCircle2 className="h-8 w-8 text-green-600" />
                    </div>
                    <div>
                      <p className="font-medium text-gray-900">{file.name}</p>
                      <p className="text-sm text-gray-600 mt-1">
                        {(file.size / 1024 / 1024).toFixed(2)} MB
                      </p>
                    </div>
                    <button
                      type="button"
                      onClick={(e) => {
                        e.preventDefault()
                        setFile(null)
                      }}
                      className="text-sm text-blue-600 hover:text-blue-700 font-medium"
                    >
                      Choose different file
                    </button>
                  </>
                ) : (
                  <>
                    <div className="rounded-xl bg-blue-100 p-3">
                      <FileImage className="h-8 w-8 text-blue-600" />
                    </div>
                    <div>
                      <p className="font-medium text-gray-900">
                        Drag & drop your image here
                      </p>
                      <p className="text-sm text-gray-600 mt-1">
                        or click to browse • Max 10MB
                      </p>
                    </div>
                    <p className="text-xs text-gray-500">
                      Supports: JPG, PNG, WEBP
                    </p>
                  </>
                )}
              </div>
            </div>
          </div>

          {/* Staging Options */}
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Room Type <span className="text-gray-400 font-normal">(optional)</span>
              </label>
              <select
                className="input"
                value={roomType}
                onChange={(e) => setRoomType(e.target.value)}
              >
                <option value="">Auto-detect</option>
                <option value="living_room">Living Room</option>
                <option value="bedroom">Bedroom</option>
                <option value="kitchen">Kitchen</option>
                <option value="bathroom">Bathroom</option>
                <option value="dining_room">Dining Room</option>
                <option value="office">Office</option>
                <option value="entryway">Entryway</option>
                <option value="outdoor">Outdoor/Patio</option>
              </select>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Furniture Style <span className="text-gray-400 font-normal">(optional)</span>
              </label>
              <select
                className="input"
                value={style}
                onChange={(e) => setStyle(e.target.value)}
              >
                <option value="">Default</option>
                <option value="modern">Modern</option>
                <option value="contemporary">Contemporary</option>
                <option value="traditional">Traditional</option>
                <option value="industrial">Industrial</option>
                <option value="scandinavian">Scandinavian</option>
                <option value="rustic">Rustic</option>
                <option value="coastal">Coastal</option>
                <option value="bohemian">Bohemian</option>
                <option value="minimalist">Minimalist</option>
                <option value="mid-century modern">Mid-Century Modern</option>
              </select>
            </div>
          </div>

          {/* Submit Button */}
          <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4 pt-2">
            <button 
              className="btn btn-primary w-full sm:w-auto" 
              type="submit"
              disabled={isUploading || !file || !projectId}
            >
              {isUploading ? (
                <>
                  <Loader2 className="h-4 w-4 animate-spin" />
                  Processing...
                </>
              ) : (
                <>
                  <UploadIcon className="h-4 w-4" />
                  Upload & Stage Image
                </>
              )}
            </button>
            
            {status && (
              <div className={cn(
                "text-sm font-medium px-4 py-2 rounded-lg",
                status.includes("Success") || status.includes("created")
                  ? "bg-green-100 text-green-700"
                  : status.includes("Error") || status.includes("failed")
                  ? "bg-red-100 text-red-700"
                  : "bg-blue-100 text-blue-700"
              )}>
                {status}
              </div>
            )}
          </div>
        </div>
      </form>

      {/* Success Message */}
      {imageId && (
        <div className="card border-green-200 bg-green-50/50 animate-in">
          <div className="card-body">
            <div className="flex items-start gap-4">
              <div className="rounded-xl bg-green-100 p-2">
                <CheckCircle2 className="h-6 w-6 text-green-600" />
              </div>
              <div className="flex-1">
                <h3 className="font-semibold text-green-900 mb-1">Image Successfully Queued!</h3>
                <p className="text-sm text-green-700 mb-3">
                  Your image has been uploaded and is being processed by our AI staging system.
                </p>
                <div className="flex flex-col sm:flex-row gap-3">
                  <a 
                    href="/images" 
                    className="inline-flex items-center gap-2 text-sm font-medium text-green-700 hover:text-green-800"
                  >
                    View in Images Dashboard →
                  </a>
                  <div className="text-xs text-green-600">
                    Image ID: <code className="bg-green-100 px-2 py-1 rounded">{imageId}</code>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
