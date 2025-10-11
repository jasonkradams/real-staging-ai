"use client";

import { useEffect, useMemo, useState, useCallback } from "react";
import JSZip from "jszip";
import NextImage from "next/image";

import { 
  FolderOpen, 
  RefreshCw, 
  Grid3x3, 
  List, 
  Download, 
  CheckCircle2,
  ExternalLink,
  Loader2,
  Image as ImageIcon,
  AlertCircle,
  Check,
  X
} from "lucide-react";

import SSEViewer from "@/components/SSEViewer";
import { apiFetch } from "@/lib/api";
import { cn, formatRelativeTime } from "@/lib/utils";

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

export default function ImagesPage() {
  const [projects, setProjects] = useState<Project[]>([]);
  const [selectedProjectId, setSelectedProjectId] = useState<string>("");
  const [images, setImages] = useState<ImageRecord[]>([]);
  const [selectedImageIds, setSelectedImageIds] = useState<Set<string>>(new Set());
  const [statusMessage, setStatusMessage] = useState<string>("");
  const [loadingProjects, setLoadingProjects] = useState(false);
  const [loadingImages, setLoadingImages] = useState(false);
  const [viewMode, setViewMode] = useState<'grid' | 'list'>('grid');
  const [imageUrls, setImageUrls] = useState<Record<string, { original?: string; staged?: string }>>({});
  const [hoveredImageId, setHoveredImageId] = useState<string | null>(null);
  const [downloadType, setDownloadType] = useState<'original' | 'staged'>('staged');

  const selectedProject = useMemo(
    () => projects.find((project) => project.id === selectedProjectId) ?? null,
    [projects, selectedProjectId]
  );

  const toggleImageSelection = useCallback((imageId: string) => {
    setSelectedImageIds(prev => {
      const newSet = new Set(prev);
      if (newSet.has(imageId)) {
        newSet.delete(imageId);
      } else {
        newSet.add(imageId);
      }
      return newSet;
    });
  }, []);

  const selectAll = useCallback(() => {
    setSelectedImageIds(new Set(images.map(img => img.id)));
  }, [images]);

  const clearSelection = useCallback(() => {
    setSelectedImageIds(new Set());
  }, []);

  // Fetch presigned URL for viewing
  async function getPresignedUrl(imageId: string, kind: 'original' | 'staged'): Promise<string | null> {
    try {
      const params = new URLSearchParams({ kind });
      const res = await apiFetch<{ url: string }>(`/v1/images/${imageId}/presign?${params.toString()}`);
      return res?.url || null;
    } catch (err: unknown) {
      console.error('Failed to get presigned URL:', err);
      return null;
    }
  }

  // Open presigned URL in new tab (for download/view)
  async function openPresigned(imageId: string, kind: 'original' | 'staged') {
    try {
      const url = await getPresignedUrl(imageId, kind);
      if (url) {
        window.open(url, '_blank', 'noopener,noreferrer');
      }
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : String(err);
      setStatusMessage(message);
    }
  }

  // Prefetch image URLs for display
  const prefetchImageUrls = useCallback(async (imageList: ImageRecord[]) => {
    const urlMap: Record<string, { original?: string; staged?: string }> = {};
    
    // Fetch all URLs in parallel (with throttling to avoid overwhelming the API)
    const chunks = [];
    for (let i = 0; i < imageList.length; i += 5) {
      chunks.push(imageList.slice(i, i + 5));
    }

    for (const chunk of chunks) {
      await Promise.all(
        chunk.map(async (image) => {
          const [originalUrl, stagedUrl] = await Promise.all([
            getPresignedUrl(image.id, 'original'),
            image.staged_url ? getPresignedUrl(image.id, 'staged') : Promise.resolve(null)
          ]);
          
          urlMap[image.id] = {
            original: originalUrl || undefined,
            staged: stagedUrl || undefined
          };
        })
      );
    }

    setImageUrls(urlMap);
  }, []);

  // Prefetch image on hover (for smooth transitions)
  const handleImageHover = useCallback((imageId: string) => {
    setHoveredImageId(imageId);
    
    const urls = imageUrls[imageId];
    if (urls?.original && typeof window !== "undefined") {
      const existing = document.querySelector(`img[src="${urls.original}"]`);
      if (!existing) {
        const preloadImg = new window.Image();
        preloadImg.src = urls.original;
      }
    }
  }, [imageUrls]);

  // Download image to local file system
  async function downloadImage(imageId: string, kind: 'original' | 'staged') {
    try {
      // Get presigned URL with download parameter
      const params = new URLSearchParams({ kind, download: '1' });
      const res = await apiFetch<{ url: string }>(`/v1/images/${imageId}/presign?${params.toString()}`);
      
      if (res?.url) {
        // Create a temporary anchor element to trigger download
        const link = document.createElement('a');
        link.href = res.url;
        link.download = ''; // Let browser determine filename from Content-Disposition header
        document.body.appendChild(link);
        link.click();
        document.body.removeChild(link);
      }
    } catch (err: unknown) {
      console.error('Failed to download image:', err);
      const message = err instanceof Error ? err.message : String(err);
      setStatusMessage(`Download failed: ${message}`);
    }
  }

  async function downloadSelected() {
    const count = selectedImageIds.size;
    
    // Single file - download directly
    if (count === 1) {
      const imageId = Array.from(selectedImageIds)[0];
      const image = images.find(img => img.id === imageId);
      
      if (downloadType === 'staged' && image?.staged_url) {
        await downloadImage(imageId, 'staged');
      } else if (downloadType === 'original') {
        await downloadImage(imageId, 'original');
      }
      return;
    }
    
    // Multiple files - create zip
    try {
      setStatusMessage(`Preparing ${count} ${downloadType} image(s) for download...`);
      
      const zip = new JSZip();
      let completed = 0;
      
      // Fetch all images and add to zip
      for (const imageId of selectedImageIds) {
        const image = images.find(img => img.id === imageId);
        
        // Skip if requested type not available
        if (downloadType === 'staged' && !image?.staged_url) continue;
        if (downloadType === 'original' && !image) continue;
        
        try {
          // Get presigned URL
          const params = new URLSearchParams({ kind: downloadType });
          const res = await apiFetch<{ url: string }>(`/v1/images/${imageId}/presign?${params.toString()}`);
          
          if (res?.url) {
            // Fetch the image as blob
            const response = await fetch(res.url);
            const blob = await response.blob();
            
            // Generate filename
            const extension = blob.type.split('/')[1] || 'jpg';
            const roomType = image?.room_type || 'image';
            const filename = `${roomType.replace(/\s+/g, '-')}_${downloadType}_${imageId.substring(0, 8)}.${extension}`;
            
            // Add to zip
            zip.file(filename, blob);
            
            completed++;
            setStatusMessage(`Preparing ${completed}/${count} images...`);
          }
        } catch (err) {
          console.error(`Failed to fetch image ${imageId}:`, err);
        }
      }
      
      if (completed === 0) {
        setStatusMessage('No images available for download');
        setTimeout(() => setStatusMessage(''), 3000);
        return;
      }
      
      // Generate zip file
      setStatusMessage('Creating zip file...');
      const zipBlob = await zip.generateAsync({ type: 'blob' });
      
      // Trigger download
      const link = document.createElement('a');
      link.href = URL.createObjectURL(zipBlob);
      const projectName = selectedProject?.name.replace(/\s+/g, '-') || 'images';
      link.download = `${projectName}_${downloadType}_${new Date().toISOString().split('T')[0]}.zip`;
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
      
      // Clean up
      URL.revokeObjectURL(link.href);
      
      setStatusMessage(`Downloaded ${completed} image(s) as zip file`);
      setTimeout(() => setStatusMessage(''), 3000);
      
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : String(err);
      setStatusMessage(`Download failed: ${message}`);
      setTimeout(() => setStatusMessage(''), 5000);
    }
  }

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
        setStatusMessage("No projects found. Create one from the Upload page.");
        return;
      }

      if (!selectedProjectId || !list.some((project) => project.id === selectedProjectId)) {
        setSelectedProjectId(list[0].id);
      }
      setStatusMessage("");
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
      setImageUrls({});
      return;
    }
    setLoadingImages(true);
    setStatusMessage("Loading images...");
    try {
      const res = await apiFetch<ImageListResponse>(`/v1/projects/${projectId}/images`);
      const list = res.images ?? [];
      setImages(list);
      setSelectedImageIds(new Set());
      setStatusMessage(list.length === 0 ? "No images found for this project yet." : "");
      
      // Prefetch image URLs for display
      if (list.length > 0) {
        prefetchImageUrls(list);
      }
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : String(err);
      setStatusMessage(message);
      setImages([]);
      setImageUrls({});
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
      return;
    }
    loadImages(selectedProjectId);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [selectedProjectId]);

  return (
    <div className="space-y-8">
      {/* Header */}
      <div>
        <h1 className="text-3xl font-bold mb-2">
          <span className="gradient-text">Image Gallery</span>
        </h1>
        <p className="text-gray-600">
          View and manage your virtually staged images across all projects
        </p>
      </div>

      {/* Project Selection */}
      <div className="card">
        <div className="card-header">
          <div className="flex items-center gap-2">
            <FolderOpen className="h-5 w-5 text-blue-600" />
            <span>Project</span>
          </div>
        </div>
        <div className="card-body">
          <div className="flex flex-col sm:flex-row gap-3">
            <div className="flex-1">
              <select
                className="input"
                value={selectedProjectId}
                onChange={(e) => {
                  setSelectedProjectId(e.target.value);
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
            <button 
              type="button" 
              className="btn btn-ghost" 
              onClick={loadProjects}
              disabled={loadingProjects}
            >
              <RefreshCw className={cn("h-4 w-4", loadingProjects && "animate-spin")} />
              Refresh
            </button>
          </div>
          {selectedProject && (
            <p className="text-sm text-gray-600 mt-3">
              Viewing <span className="font-medium">{selectedProject.name}</span> • {images.length} image{images.length !== 1 ? 's' : ''}
            </p>
          )}
        </div>
      </div>

      {/* Toolbar */}
      {images.length > 0 && (
        <div className="card">
          <div className="card-body">
            <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
              {/* Selection Controls */}
              <div className="flex items-center gap-3">
                <button
                  onClick={selectAll}
                  className="btn btn-ghost text-sm"
                  disabled={images.length === 0}
                >
                  <Check className="h-4 w-4" />
                  Select All ({images.length})
                </button>
                <button
                  onClick={clearSelection}
                  className="btn btn-ghost text-sm"
                  disabled={selectedImageIds.size === 0}
                >
                  <X className="h-4 w-4" />
                  Clear
                </button>
                {selectedImageIds.size > 0 && (
                  <span className="text-sm font-medium text-blue-600">
                    {selectedImageIds.size} selected
                  </span>
                )}
              </div>

              {/* View Mode & Actions */}
              <div className="flex items-center gap-2">
                {selectedImageIds.size > 0 && (
                  <>
                    {/* Download Type Toggle */}
                    <div className="flex rounded-lg border border-gray-200 p-1 mr-2">
                      <button
                        onClick={() => setDownloadType('staged')}
                        className={cn(
                          "px-3 py-1 text-xs rounded transition-colors",
                          downloadType === 'staged'
                            ? "bg-blue-100 text-blue-600 font-medium"
                            : "text-gray-600 hover:bg-gray-100"
                        )}
                        title="Download staged (AI-enhanced) images"
                      >
                        Staged
                      </button>
                      <button
                        onClick={() => setDownloadType('original')}
                        className={cn(
                          "px-3 py-1 text-xs rounded transition-colors",
                          downloadType === 'original'
                            ? "bg-blue-100 text-blue-600 font-medium"
                            : "text-gray-600 hover:bg-gray-100"
                        )}
                        title="Download original (unprocessed) images"
                      >
                        Original
                      </button>
                    </div>
                    <button
                      onClick={downloadSelected}
                      className="btn btn-secondary"
                    >
                      <Download className="h-4 w-4" />
                      Download {downloadType} ({selectedImageIds.size})
                    </button>
                  </>
                )}
                <div className="flex rounded-lg border border-gray-200 p-1">
                  <button
                    onClick={() => setViewMode('grid')}
                    className={cn(
                      "p-2 rounded transition-colors",
                      viewMode === 'grid' 
                        ? "bg-blue-100 text-blue-600" 
                        : "text-gray-600 hover:bg-gray-100"
                    )}
                  >
                    <Grid3x3 className="h-4 w-4" />
                  </button>
                  <button
                    onClick={() => setViewMode('list')}
                    className={cn(
                      "p-2 rounded transition-colors",
                      viewMode === 'list' 
                        ? "bg-blue-100 text-blue-600" 
                        : "text-gray-600 hover:bg-gray-100"
                    )}
                  >
                    <List className="h-4 w-4" />
                  </button>
                </div>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Loading State */}
      {loadingImages && (
        <div className="flex flex-col items-center justify-center py-16">
          <Loader2 className="h-12 w-12 text-blue-600 animate-spin mb-4" />
          <p className="text-gray-600">Loading images...</p>
        </div>
      )}

      {/* Empty State */}
      {!loadingImages && images.length === 0 && selectedProjectId && (
        <div className="card">
          <div className="card-body">
            <div className="flex flex-col items-center justify-center py-16 text-center">
              <div className="rounded-full bg-gray-100 p-4 mb-4">
                <ImageIcon className="h-12 w-12 text-gray-400" />
              </div>
              <h3 className="text-lg font-semibold text-gray-900 mb-2">No images yet</h3>
              <p className="text-gray-600 mb-4 max-w-md">
                Upload your first property image to get started with AI-powered virtual staging
              </p>
              <a href="/upload" className="btn btn-primary">
                Upload Image
              </a>
            </div>
          </div>
        </div>
      )}

      {/* Grid View */}
      {!loadingImages && images.length > 0 && viewMode === 'grid' && (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {images.map((image) => {
            const urls = imageUrls[image.id];
            const stagedSrc = urls?.staged ?? null;
            const originalSrc = urls?.original ?? null;

            return (
              <div
                key={image.id}
                className={cn(
                  "card group cursor-pointer transition-all duration-200",
                  selectedImageIds.has(image.id) && "ring-2 ring-blue-500 shadow-xl"
                )}
                onClick={() => toggleImageSelection(image.id)}
                onMouseEnter={() => handleImageHover(image.id)}
                onMouseLeave={() => setHoveredImageId(null)}
              >
                <div className="relative aspect-video bg-gray-100 overflow-hidden rounded-t-2xl">
                  {/* Image Preview - Show staged by default, original on hover */}
                  {urls ? (
                    <>
                    {/* Staged Image (shown by default) */}
                    {typeof stagedSrc === "string" && (
                      <NextImage
                        src={stagedSrc}
                        alt="Staged"
                        fill
                        className={cn(
                          "absolute inset-0 object-cover transition-all duration-300",
                          hoveredImageId === image.id ? "opacity-0" : "opacity-100 group-hover:scale-105"
                        )}
                        loading="lazy"
                        sizes="(max-width: 768px) 100vw, (max-width: 1200px) 50vw, 33vw"
                      />
                    )}
                    {/* Original Image (shown on hover) */}
                    {typeof originalSrc === "string" && (
                      <NextImage
                        src={originalSrc}
                        alt="Original"
                        fill
                        className={cn(
                          "absolute inset-0 object-cover transition-opacity duration-300",
                          hoveredImageId === image.id ? "opacity-100" : "opacity-0"
                        )}
                        loading="lazy"
                        sizes="(max-width: 768px) 100vw, (max-width: 1200px) 50vw, 33vw"
                      />
                    )}
                    {/* Fallback if no images loaded yet */}
                    {!stagedSrc && !originalSrc && (
                      <div className="flex items-center justify-center h-full">
                        <Loader2 className="h-16 w-16 text-gray-300 animate-spin" />
                      </div>
                    )}
                  </>
                ) : (
                  <div className="flex items-center justify-center h-full">
                    <Loader2 className="h-16 w-16 text-gray-300 animate-spin" />
                  </div>
                  )}

                  {/* Selection Indicator */}
                  <div className={cn(
                  "absolute top-3 left-3 flex items-center justify-center h-6 w-6 rounded-full border-2 transition-all",
                  selectedImageIds.has(image.id)
                    ? "bg-blue-600 border-blue-600"
                    : "bg-white border-white group-hover:border-blue-400"
                )}>
                  {selectedImageIds.has(image.id) && (
                    <Check className="h-4 w-4 text-white" />
                  )}
                </div>

                {/* Status Badge */}
                <div className="absolute top-3 right-3">
                  <span className={cn(
                    "badge",
                    `badge-status-${image.status}`
                  )}>
                    {image.status}
                  </span>
                </div>

                {/* Original/Staged Indicator */}
                {imageUrls[image.id]?.original && imageUrls[image.id]?.staged && (
                  <div className="absolute bottom-3 left-3 z-10">
                    <span className="badge bg-black/70 text-white text-xs">
                      {hoveredImageId === image.id ? "Original" : "Staged"}
                    </span>
                  </div>
                )}

                {/* Quick Actions Bar - Bottom */}
                <div className="absolute bottom-0 inset-x-0 bg-gradient-to-t from-black/70 to-transparent opacity-0 group-hover:opacity-100 transition-opacity p-3 flex items-center justify-center gap-2">
                  <button
                    onClick={(e) => {
                      e.stopPropagation();
                      openPresigned(image.id, 'original');
                    }}
                    className="btn btn-secondary btn-sm"
                    title="Open Original in New Tab"
                  >
                    <ExternalLink className="h-3 w-3" />
                    <span className="hidden sm:inline">Original</span>
                  </button>
                  {image.staged_url && (
                    <button
                      onClick={(e) => {
                        e.stopPropagation();
                        openPresigned(image.id, 'staged');
                      }}
                      className="btn btn-primary btn-sm"
                      title="Open Staged in New Tab"
                    >
                      <ExternalLink className="h-3 w-3" />
                      <span className="hidden sm:inline">Staged</span>
                    </button>
                  )}
                </div>
              </div>

              {/* Card Content */}
              <div className="p-4 space-y-2">
                <div className="flex items-start justify-between gap-2">
                  <div className="flex-1 min-w-0">
                    <p className="text-sm font-medium text-gray-900 truncate">
                      {image.room_type || 'Untitled'}
                    </p>
                    <p className="text-xs text-gray-500">
                      {formatRelativeTime(image.updated_at)}
                    </p>
                  </div>
                  {image.style && (
                    <span className="text-xs px-2 py-1 bg-gray-100 text-gray-700 rounded-full">
                      {image.style}
                    </span>
                  )}
                </div>

                {image.error && (
                  <div className="flex items-start gap-2 text-xs text-red-600 bg-red-50 p-2 rounded">
                    <AlertCircle className="h-3 w-3 mt-0.5 flex-shrink-0" />
                    <span className="line-clamp-2">{image.error}</span>
                  </div>
                )}
              </div>
            </div>
          );
          })}
        </div>
      )}

      {/* List View */}
      {!loadingImages && images.length > 0 && viewMode === 'list' && (
        <div className="card">
          <div className="card-body p-0">
            <div className="overflow-x-auto">
              <table className="min-w-full divide-y divide-gray-200">
                <thead className="bg-gray-50">
                  <tr>
                    <th className="w-12 px-4 py-3"></th>
                    <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Preview</th>
                    <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Details</th>
                    <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Status</th>
                    <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Updated</th>
                    <th className="px-4 py-3 text-right text-xs font-medium text-gray-500 uppercase">Actions</th>
                  </tr>
                </thead>
                <tbody className="bg-white divide-y divide-gray-200">
                  {images.map((image) => {
                    const thumbSrc = imageUrls[image.id]?.staged ?? null;

                    return (
                      <tr
                        key={image.id}
                        className={cn(
                          "hover:bg-gray-50 transition-colors cursor-pointer",
                          selectedImageIds.has(image.id) && "bg-blue-50"
                        )}
                        onClick={() => toggleImageSelection(image.id)}
                      >
                      <td className="px-4 py-4">
                        <div
                          className={cn(
                            "flex items-center justify-center h-5 w-5 rounded border-2 transition-all",
                            selectedImageIds.has(image.id)
                              ? "bg-blue-600 border-blue-600"
                              : "border-gray-300"
                          )}
                        >
                          {selectedImageIds.has(image.id) && (
                            <Check className="h-3 w-3 text-white" />
                          )}
                        </div>
                      </td>
                      <td className="px-4 py-4">
                        <div className="h-16 w-24 rounded-lg overflow-hidden bg-gray-100">
                          {typeof thumbSrc === "string" ? (
                            <NextImage
                              src={thumbSrc}
                              alt="Preview"
                              width={96}
                              height={64}
                              className="h-full w-full object-cover"
                            />
                          ) : imageUrls[image.id] ? (
                            <div className="flex items-center justify-center h-full">
                              <ImageIcon className="h-8 w-8 text-gray-300" />
                            </div>
                          ) : (
                            <div className="flex items-center justify-center h-full">
                              <Loader2 className="h-6 w-6 text-gray-300 animate-spin" />
                            </div>
                          )}
                        </div>
                      </td>
                      <td className="px-4 py-4">
                        <div>
                          <p className="text-sm font-medium text-gray-900">
                            {image.room_type || 'Untitled'}
                          </p>
                          {image.style && (
                            <p className="text-xs text-gray-500">{image.style}</p>
                          )}
                        </div>
                      </td>
                      <td className="px-4 py-4">
                        <span className={cn("badge badge-status-" + image.status)}>
                          {image.status}
                        </span>
                        {image.error && (
                          <p className="text-xs text-red-600 mt-1 max-w-xs truncate">
                            {image.error}
                          </p>
                        )}
                      </td>
                      <td className="px-4 py-4 text-sm text-gray-600">
                        {formatRelativeTime(image.updated_at)}
                      </td>
                      <td className="px-4 py-4 text-right">
                        <div className="flex items-center justify-end gap-2">
                          <button
                            onClick={(e) => {
                              e.stopPropagation();
                              openPresigned(image.id, 'original');
                            }}
                            className="text-gray-600 hover:text-blue-600 transition-colors"
                            title="View Original"
                          >
                            <ExternalLink className="h-4 w-4" />
                          </button>
                          {image.staged_url && (
                            <button
                              onClick={(e) => {
                                e.stopPropagation();
                                openPresigned(image.id, 'staged');
                              }}
                              className="text-blue-600 hover:text-blue-700 transition-colors"
                              title="View Staged"
                            >
                              <CheckCircle2 className="h-4 w-4" />
                            </button>
                          )}
                        </div>
                      </td>
                    </tr>
                  );
                  })}
                </tbody>
              </table>
            </div>
          </div>
        </div>
      )}

      {/* SSE Live Updates */}
      {selectedProjectId && (
        <div className="card">
          <div className="card-header">
            <span>Live Updates</span>
          </div>
          <div className="card-body">
            <SSEViewer
              onStatus={(status) => {
                if (status === "ready" || status === "error") {
                  loadImages(selectedProjectId);
                }
              }}
            />
          </div>
        </div>
      )}

      {/* Status Message */}
      {statusMessage && (
        <div className="fixed bottom-4 right-4 bg-white shadow-lg rounded-lg p-4 border border-gray-200 max-w-sm animate-in">
          <p className="text-sm text-gray-700">{statusMessage}</p>
        </div>
      )}
    </div>
  );
}
