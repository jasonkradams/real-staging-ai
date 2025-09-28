"use client";

import { useEffect, useMemo, useState } from "react";

import SSEViewer from "@/components/SSEViewer";
import { apiFetch } from "@/lib/api";

type Project = {
  id: string;
  name: string;
};

type ProjectListResponse = {
  projects: Project[];
};

type ImageRecord = {
  id: string;
  project_id: string;
  original_url: string;
  staged_url?: string | null;
  status: string;
  error?: string | null;
  room_type?: string | null;
  style?: string | null;
  seed?: number | null;
  created_at: string;
  updated_at: string;
};

type ImageListResponse = {
  images: ImageRecord[];
};

const dateFormatter =
  typeof Intl !== "undefined"
    ? new Intl.DateTimeFormat(undefined, { dateStyle: "short", timeStyle: "medium" })
    : null;

function formatDate(value?: string | null) {
  if (!value) return "-";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return value;
  }
  return dateFormatter ? dateFormatter.format(date) : date.toISOString();
}

export default function ImagesPage() {
  const [projects, setProjects] = useState<Project[]>([]);
  const [selectedProjectId, setSelectedProjectId] = useState<string>("");
  const [images, setImages] = useState<ImageRecord[]>([]);
  const [selectedImageId, setSelectedImageId] = useState<string>("");
  const [statusMessage, setStatusMessage] = useState<string>("");
  const [loadingProjects, setLoadingProjects] = useState(false);
  const [loadingImages, setLoadingImages] = useState(false);

  const selectedProject = useMemo(
    () => projects.find((project) => project.id === selectedProjectId) ?? null,
    [projects, selectedProjectId]
  );

  const selectedImage = useMemo(
    () => images.find((image) => image.id === selectedImageId) ?? null,
    [images, selectedImageId]
  );

  async function loadProjects() {
    setLoadingProjects(true);
    setStatusMessage("Loading projects...");
    try {
      const res = await apiFetch<ProjectListResponse>("/v1/projects");
      const list = res.projects ?? [];
      setProjects(list);
      if (list.length === 0) {
        setSelectedProjectId("");
        setImages([]);
        setSelectedImageId("");
        setStatusMessage("No projects found. Create one from the Upload page.");
        return;
      }

      if (!selectedProjectId || !list.some((project) => project.id === selectedProjectId)) {
        setSelectedProjectId(list[0].id);
      }
      setStatusMessage(`Loaded ${list.length} project${list.length === 1 ? "" : "s"}.`);
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : String(err);
      setStatusMessage(message);
    } finally {
      setLoadingProjects(false);
    }
  }

  async function loadImages(projectId: string) {
    if (!projectId) {
      setImages([]);
      setSelectedImageId("");
      return;
    }
    setLoadingImages(true);
    setStatusMessage("Loading images...");
    try {
      const res = await apiFetch<ImageListResponse>(`/v1/projects/${projectId}/images`);
      const list = res.images ?? [];
      setImages(list);
      if (list.length === 0) {
        setSelectedImageId("");
        setStatusMessage("No images found for this project yet.");
        return;
      }
      const currentSelection = list.some((image) => image.id === selectedImageId)
        ? selectedImageId
        : list[0].id;
      setSelectedImageId(currentSelection);
      setStatusMessage(`Loaded ${list.length} image${list.length === 1 ? "" : "s"}.`);
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : String(err);
      setStatusMessage(message);
      setImages([]);
      setSelectedImageId("");
    } finally {
      setLoadingImages(false);
    }
  }

  useEffect(() => {
    loadProjects();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  useEffect(() => {
    if (!selectedProjectId) {
      setImages([]);
      setSelectedImageId("");
      return;
    }
    loadImages(selectedProjectId);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [selectedProjectId]);

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-semibold">Images</h1>
          <p className="text-sm text-gray-600">Browse staged images by project and monitor job updates.</p>
        </div>
        <div className="text-sm text-gray-500">
          {statusMessage && <span>{statusMessage}</span>}
        </div>
      </div>

      <div className="card">
        <div className="card-header">Project selection</div>
        <div className="card-body space-y-4">
          <div className="flex flex-wrap items-end gap-3">
            <div className="min-w-[220px] flex-1">
              <label className="block text-sm mb-1">Project</label>
              <select
                className="input w-full"
                value={selectedProjectId}
                onChange={(event) => {
                  setSelectedProjectId(event.target.value);
                  setStatusMessage("");
                }}
                disabled={loadingProjects}
              >
                <option value="">Select a project...</option>
                {projects.map((project) => (
                  <option key={project.id} value={project.id}>
                    {project.name}
                  </option>
                ))}
              </select>
            </div>
            <button type="button" className="btn" onClick={loadProjects} disabled={loadingProjects}>
              {loadingProjects ? "Loading..." : "Refresh Projects"}
            </button>
            <button
              type="button"
              className="btn"
              onClick={() => selectedProjectId && loadImages(selectedProjectId)}
              disabled={!selectedProjectId || loadingImages}
            >
              {loadingImages ? "Loading..." : "Refresh Images"}
            </button>
          </div>
          {selectedProject && (
            <p className="text-sm text-gray-600">
              Viewing images for <span className="font-medium">{selectedProject.name}</span>.
            </p>
          )}
        </div>
      </div>

      <div className="card">
        <div className="card-header">Images</div>
        <div className="card-body space-y-4">
          {images.length === 0 ? (
            <div className="text-sm text-gray-600">
              No images yet. Upload one from the Upload page to populate this list.
            </div>
          ) : (
            <div className="overflow-x-auto">
              <table className="min-w-full text-sm">
                <thead className="text-left text-xs uppercase text-gray-500">
                  <tr>
                    <th className="py-2 pr-4">Image</th>
                    <th className="py-2 pr-4">Status</th>
                    <th className="py-2 pr-4">Original</th>
                    <th className="py-2 pr-4">Staged</th>
                    <th className="py-2 pr-4">Updated</th>
                  </tr>
                </thead>
                <tbody>
                  {images.map((image) => {
                    const isSelected = image.id === selectedImageId;
                    return (
                      <tr
                        key={image.id}
                        className={`align-top transition-colors ${
                          isSelected ? "bg-indigo-50" : "hover:bg-gray-50"
                        }`}
                      >
                        <td className="py-2 pr-4">
                          <button
                            type="button"
                            className="font-medium text-indigo-600 hover:underline"
                            onClick={() => setSelectedImageId(image.id)}
                          >
                            {image.id}
                          </button>
                        </td>
                        <td className="py-2 pr-4">
                          <span className="inline-flex items-center rounded-full bg-gray-100 px-2 py-0.5 text-xs font-medium capitalize text-gray-700">
                            {image.status}
                          </span>
                          {image.error && (
                            <div className="mt-1 text-xs text-red-600">{image.error}</div>
                          )}
                        </td>
                        <td className="py-2 pr-4">
                          <a
                            href={image.original_url}
                            target="_blank"
                            rel="noreferrer"
                            className="text-indigo-600 hover:underline"
                          >
                            Original
                          </a>
                        </td>
                        <td className="py-2 pr-4">
                          {image.staged_url ? (
                            <a
                              href={image.staged_url}
                              target="_blank"
                              rel="noreferrer"
                              className="text-indigo-600 hover:underline"
                            >
                              View
                            </a>
                          ) : (
                            <span className="text-xs text-gray-500">Pending</span>
                          )}
                        </td>
                        <td className="py-2 pr-4 text-xs text-gray-600">
                          <div>Created: {formatDate(image.created_at)}</div>
                          <div>Updated: {formatDate(image.updated_at)}</div>
                        </td>
                      </tr>
                    );
                  })}
                </tbody>
              </table>
            </div>
          )}

          {selectedImage && (
            <div className="rounded border border-gray-200 bg-gray-50 px-4 py-3 text-sm text-gray-700">
              <div className="font-semibold text-gray-800">Selected image</div>
              <div className="mt-1 grid gap-1 text-xs sm:grid-cols-2">
                <div>
                  <span className="font-medium">ID:</span> {selectedImage.id}
                </div>
                <div>
                  <span className="font-medium">Status:</span> {selectedImage.status}
                </div>
                {selectedImage.room_type && (
                  <div>
                    <span className="font-medium">Room:</span> {selectedImage.room_type}
                  </div>
                )}
                {selectedImage.style && (
                  <div>
                    <span className="font-medium">Style:</span> {selectedImage.style}
                  </div>
                )}
                {selectedImage.seed != null && (
                  <div>
                    <span className="font-medium">Seed:</span> {selectedImage.seed}
                  </div>
                )}
                {selectedImage.error && (
                  <div className="text-red-600">
                    <span className="font-medium">Error:</span> {selectedImage.error}
                  </div>
                )}
              </div>
            </div>
          )}
        </div>
      </div>

      <div className="card">
        <div className="card-header">Live updates</div>
        <div className="card-body">
          <SSEViewer initialImageId={selectedImageId || undefined} />
        </div>
      </div>
    </div>
  );
}
