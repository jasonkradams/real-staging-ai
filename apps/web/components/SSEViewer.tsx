"use client";

import { useEffect, useRef, useState } from "react";

export default function SSEViewer() {
  const [imageId, setImageId] = useState("");
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

  const connect = () => {
    if (!imageId) return;
    if (esRef.current) {
      esRef.current.close();
      esRef.current = null;
    }
    const base = process.env.NEXT_PUBLIC_API_BASE || '/api';
    let url = `${base}/v1/events?image_id=${encodeURIComponent(imageId)}`;
    // Append access_token from localStorage for auth (EventSource can't set headers)
    if (typeof window !== 'undefined') {
      const token = localStorage.getItem('token');
      if (token) {
        url += `&access_token=${encodeURIComponent(token)}`;
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
        const data = (ev as MessageEvent).data as string;
        setLog((prev) => [
          `[${new Date().toLocaleTimeString()}] job_update ${data}`,
          ...prev,
        ]);
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
