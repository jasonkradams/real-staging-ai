"use client";

import { useEffect, useState } from "react";

export default function TokenBar() {
  const [token, setToken] = useState("");
  const [visible, setVisible] = useState(false);

  useEffect(() => {
    const t = localStorage.getItem("token") || "";
    setToken(t);
  }, []);

  const save = () => {
    localStorage.setItem("token", token);
    setVisible(false);
  };

  return (
    <div className="text-sm text-gray-500">
      <button className="underline" onClick={() => setVisible((v) => !v)}>
        {token ? "Edit Token" : "Set Token"}
      </button>
      {visible && (
        <div className="mt-2 flex items-center gap-2">
          <input
            className="input"
            placeholder="Paste bearer token"
            value={token}
            onChange={(e) => setToken(e.target.value)}
          />
          <button className="btn" onClick={save}>
            Save
          </button>
        </div>
      )}
    </div>
  );
}
