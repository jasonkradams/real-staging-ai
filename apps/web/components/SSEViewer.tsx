"use client";

import { useEffect, useRef, useState } from "react";

type SSEViewerProps = {
  initialImageId?: string;
  onStatus?: (status: string) => void;
};

export default function SSEViewer({ initialImageId, onStatus }: SSEViewerProps = {}) {
  const [imageId, setImageId] = useState(initialImageId ?? "");
  const [connected, setConnected] = useState(false);
  const [log, setLog] = useState<string[]>([]);
  const esRef = useRef<EventSource | null>(null);

  useEffect(() => {
    return () => {
      if (esRef.current) {
        esRef.current.close();
        esRef.current = null;
      }
    };
  }, []);

  useEffect(() => {
    if (initialImageId !== undefined && initialImageId !== imageId) {
      setImageId(initialImageId);
    }
  }, [initialImageId, imageId]);

  const connect = async () => {
    if (!imageId) return;
    if (esRef.current) {
      esRef.current.close();
      esRef.current = null;
    }
    const base = process.env.NEXT_PUBLIC_API_BASE || "/api";
    let url = `${base}/v1/events?image_id=${encodeURIComponent(imageId)}`;
    
    // Fetch access token from Auth0 session (EventSource can't set headers)
    if (typeof window !== "undefined") {
      try {
        const response = await fetch('/auth/access-token');
        if (response.ok) {
          const data = await response.json();
          if (data.accessToken) {
            url += `&access_token=${encodeURIComponent(data.accessToken)}`;
          }
        }
      } catch (error) {
        console.error('Failed to get access token for SSE:', error);
      }
    }
    
    const es = new EventSource(url);
    esRef.current = es;

    es.onopen = () => {
      setConnected(true);
      setLog((prev) => [
        `[${new Date().toLocaleTimeString()}] connected`,
        ...prev,
      ]);
    };

    es.onerror = () => {
      setConnected(false);
      setLog((prev) => [
        `[${new Date().toLocaleTimeString()}] error (see network tab)`,
        ...prev,
      ]);
    };

    es.addEventListener("connected", () => {
      setLog((prev) => [
        `[${new Date().toLocaleTimeString()}] event: connected`,
        ...prev,
      ]);
    });

    es.addEventListener("job_update", (ev) => {
      try {
        const dataStr = (ev as MessageEvent).data as string;
        setLog((prev) => [
          `[${new Date().toLocaleTimeString()}] job_update ${dataStr}`,
          ...prev,
        ]);
        try {
          const parsed = JSON.parse(dataStr);
          const status = parsed?.status as string | undefined;
          if (status && onStatus) onStatus(status);
        } catch {
          // ignore parse errors for callback
        }
      } catch {
        setLog((prev) => [
          `[${new Date().toLocaleTimeString()}] job_update (unparseable)`,
          ...prev,
        ]);
      }
    });
  };

  const disconnect = () => {
    if (esRef.current) {
      esRef.current.close();
      esRef.current = null;
    }
    setConnected(false);
  };

  return (
    <div className="space-y-4">
      <div className="flex items-end gap-2">
        <div className="flex-1">
          <label className="block text-sm mb-1">Image ID</label>
          <input
            className="input"
            placeholder="image UUID"
            value={imageId}
            onChange={(e) => setImageId(e.target.value)}
          />
        </div>
        {!connected ? (
          <button className="btn" onClick={connect} disabled={!imageId}>
            Connect
          </button>
        ) : (
          <button className="btn" onClick={disconnect}>
            Disconnect
          </button>
        )}
      </div>

      <div className="card">
        <div className="card-header">Events</div>
        <div className="card-body">
          {log.length === 0 ? (
            <div className="text-sm text-gray-500">No events yet.</div>
          ) : (
            <ul className="text-sm space-y-1">
              {log.map((l, i) => (
                <li key={i} className="font-mono text-gray-700">
                  {l}
                </li>
              ))}
            </ul>
          )}
        </div>
      </div>
    </div>
  );
}
