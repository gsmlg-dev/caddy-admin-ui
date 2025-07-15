"use client";

import { useState, useEffect } from "react";

export const Upstreams = ({ dataPromise }) => {
  const [data, setData] = useState([]);
  useEffect(() => {
    fetch("/reverse_proxy/upstreams")
      .then((resp) => resp.json())
      .then((data) => setData(data));
  }, []);

  return (
    <div className="grid-cols-auto-fit-32 gap-4">
      {(data ?? []).map((site) => (
        <div key={site.address} className="stats shadow shadow-info m-2">
          <div className="stat">
            <div className="stat-title">Requests</div>
            <div className="stat-value text-success">
              {site.num_requests}
              <sub className="text-error mx-4 text-xs">Fails: {site.fails}</sub>
            </div>
            <div className="stat-desc text-info">{site.address}</div>
          </div>
        </div>
      ))}
    </div>
  );
};
